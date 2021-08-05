package v2ray

import (
	"context"
	"fmt"
	pb "github.com/v2fly/v2ray-core/v4/app/observatory/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	Products = []string{
		"observatory",
	}
	ApiFeed *Feed
)

const ApiFeedBoxSize = 10

func init() {
	ApiFeed = NewSubscriptions(ApiFeedBoxSize)
	for _, product := range Products {
		ApiFeed.RegisterProduct(product)
	}
	go observatoryProducer()
}

func getObservatoryResponse(conn *grpc.ClientConn) (r *pb.GetOutboundStatusResponse, err error) {
	c := pb.NewObservatoryServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err = c.GetOutboundStatus(ctx, &pb.GetOutboundStatusRequest{})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func observatoryProducer() {
	const product = "observatory"
	var conn *grpc.ClientConn
	for {
		if ApiPort() == 0 {
			time.Sleep(5 * time.Second)
			continue
		}
		// Set up a connection to the server.
		if conn == nil {
			c, err := grpc.Dial(net.JoinHostPort("127.0.0.1", strconv.Itoa(ApiPort())), grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Printf("[Warning] observatoryProducer:%v", fmt.Errorf("did not connect: %w", err))
			}
			defer c.Close()
			conn = c
		}
		r, err := getObservatoryResponse(conn)
		if err != nil {
			if status.Code(err) == codes.Unavailable {
				// the connection is reliable, and reconnect
				conn = nil
				continue
			}
			log.Printf("[Warning] observatoryProducer: %v", err)
		} else {
			outboundStatus := r.GetStatus().GetStatus()
			ApiFeed.ProductMessage(product, outboundStatus)
		}
		time.Sleep(5 * time.Second)
	}
}
