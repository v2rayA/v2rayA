package v2ray

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/devfeel/mapper"
	"github.com/gin-gonic/gin"
	"github.com/v2fly/v2ray-core/v5/app/observatory"
	pb "github.com/v2fly/v2ray-core/v5/app/observatory/command"
	statscommand "github.com/v2fly/v2ray-core/v5/app/stats/command"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ApiProducts = []string{
		"observatory",
		"running_state",
		"traffic",
	}
	ApiFeed *Feed
)

const (
	ApiFeedBoxSize  = 10
	ApiFeedInterval = 1 * time.Second
)

type OutboundStatus struct {
	Alive bool  `json:"alive"`
	Delay int64 `json:"delay"`
	//LastErrorReason string           `json:"last_error_reason"`
	OutboundTag  string           `json:"outbound_tag"`
	Which        *configure.Which `json:"which"`
	LastSeenTime int64            `json:"last_seen_time"`
	LastTryTime  int64            `json:"last_try_time"`
}

func init() {
	mapper.Register(&observatory.OutboundStatus{})

	ApiFeed = NewSubscriptions(ApiFeedBoxSize)
	for _, product := range ApiProducts {
		ApiFeed.RegisterProduct(product)
	}
}

type ObservatoryResp struct {
	OutboundName string
	Resp         *pb.GetOutboundStatusResponse
}

func getObservatoryResponses(conn *grpc.ClientConn, observatoryTags []string) (r []ObservatoryResp, err error) {
	c := pb.NewObservatoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(observatoryTags) == 0 {
		observatoryTags = append(observatoryTags, "")
	}
	for _, tag := range observatoryTags {
		resp, err := c.GetOutboundStatus(ctx, &pb.GetOutboundStatusRequest{
			Tag: tag,
		})
		if err != nil {
			return nil, err
		}
		r = append(r, ObservatoryResp{OutboundName: tag, Resp: resp})
	}
	return r, nil
}

// ObservatoryProducer monitors outbound status through the core's v2ray-compatible
// observatory gRPC service and publishes to ApiFeed.
func ObservatoryProducer(apiPort int, observatoryTags []string) (closeFunc func()) {
	closed := make(chan struct{})
	go func() {
		const product = "observatory"
		var conn *grpc.ClientConn
		defer func() {
			if conn != nil {
				_ = conn.Close()
			}
		}()
		ticker := time.NewTicker(ApiFeedInterval)
		defer ticker.Stop()
		for {
			select {
			case <-closed:
				return
			case <-ticker.C:
			}
			p := ProcessManager.Process()
			if p == nil {
				continue
			}
			// Set up a connection to the server.
			if conn == nil {
				ctx, cancel := context.WithTimeout(context.Background(), ApiFeedInterval)
				c, err := grpc.DialContext(
					ctx,
					net.JoinHostPort("127.0.0.1", strconv.Itoa(apiPort)),
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				cancel()
				if err != nil {
					log.Warn("ObservatoryProducer: did not connect: %v", err)
					continue
				}
				conn = c
			}
			resps, err := getObservatoryResponses(conn, observatoryTags)
			if err != nil {
				if status.Code(err) == codes.Unavailable {
					_ = conn.Close()
					conn = nil
					continue
				}
				log.Warn("ObservatoryProducer: %v", err)
			} else {
				css := configure.GetConnectedServers()
				for _, r := range resps {
					outboundStatus := r.Resp.GetStatus().GetStatus()
					os := make([]OutboundStatus, 0, len(outboundStatus))
					for _, observed := range outboundStatus {
						index, ok := p.tag2WhichIndex[observed.OutboundTag]
						if !ok || index < 0 || index >= css.Len() {
							continue
						}
						var s OutboundStatus
						_ = mapper.AutoMapper(observed, &s)
						s.Which = &configure.Which{NodeRef: *css.Get()[index]}
						os = append(os, s)
					}
					msg := gin.H{
						"outboundName":   r.OutboundName,
						"outboundStatus": os,
					}
					ApiFeed.ProductMessage(product, msg)
				}
			}
		}
	}()
	return func() {
		close(closed)
	}
}

type trafficSample struct {
	Up        float64 `json:"up"`
	Down      float64 `json:"down"`
	UpTotal   int64   `json:"upTotal"`
	DownTotal int64   `json:"downTotal"`
}

type trafficCounter struct {
	up, down int64
	at       time.Time
}

func (c *trafficCounter) sample(stats []*statscommand.Stat, now time.Time) trafficSample {
	var sample trafficSample
	for _, stat := range stats {
		name := stat.GetName()
		if !strings.HasPrefix(name, "outbound>>>") {
			continue
		}
		tag, direction, ok := strings.Cut(strings.TrimPrefix(name, "outbound>>>"), ">>>traffic>>>")
		if !ok {
			continue
		}
		switch tag {
		case "direct", "block", "dns-out", "api-out":
			continue
		}
		switch direction {
		case "uplink":
			sample.UpTotal += stat.GetValue()
		case "downlink":
			sample.DownTotal += stat.GetValue()
		}
	}
	if elapsed := now.Sub(c.at).Seconds(); !c.at.IsZero() && elapsed > 0 {
		// A restarted core can reset counters between successful samples.
		if sample.UpTotal >= c.up {
			sample.Up = float64(sample.UpTotal-c.up) / elapsed
		}
		if sample.DownTotal >= c.down {
			sample.Down = float64(sample.DownTotal-c.down) / elapsed
		}
	}
	c.up, c.down, c.at = sample.UpTotal, sample.DownTotal, now
	return sample
}

// TrafficProducer publishes outbound byte totals and rates until its close function is called.
func TrafficProducer(apiPort int) (closeFunc func()) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(ApiFeedInterval)
		defer ticker.Stop()
		var conn *grpc.ClientConn
		defer func() {
			if conn != nil {
				_ = conn.Close()
			}
		}()
		var counter trafficCounter
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			if conn == nil {
				dialCtx, dialCancel := context.WithTimeout(ctx, ApiFeedInterval)
				c, err := grpc.DialContext(
					dialCtx,
					net.JoinHostPort("127.0.0.1", strconv.Itoa(apiPort)),
					grpc.WithInsecure(),
					grpc.WithBlock(),
				)
				dialCancel()
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Warn("TrafficProducer: did not connect: %v", err)
					continue
				}
				conn = c
			}
			queryCtx, queryCancel := context.WithTimeout(ctx, ApiFeedInterval)
			var response statscommand.QueryStatsResponse
			// The merged core registers xray's service; the v2fly messages share its wire format.
			err := conn.Invoke(queryCtx, "/xray.app.stats.command.StatsService/QueryStats",
				&statscommand.QueryStatsRequest{Pattern: "outbound>>>", Reset_: false}, &response)
			queryCancel()
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				log.Warn("TrafficProducer: %v", err)
				continue
			}
			ApiFeed.ProductMessage("traffic", counter.sample(response.GetStat(), time.Now()))
		}
	}()
	return cancel
}
