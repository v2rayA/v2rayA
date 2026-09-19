package v2ray

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2fly/v2ray-core/v5/app/observatory"
	pb "github.com/v2fly/v2ray-core/v5/app/observatory/command"
	"github.com/v2rayA/v2rayA/db/configure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type observatoryTestServer struct {
	pb.UnimplementedObservatoryServiceServer
	calls       atomic.Int32
	unavailable atomic.Bool
}

func (s *observatoryTestServer) GetOutboundStatus(context.Context, *pb.GetOutboundStatusRequest) (*pb.GetOutboundStatusResponse, error) {
	s.calls.Add(1)
	if s.unavailable.Load() {
		return nil, status.Error(codes.Unavailable, "restarting")
	}
	return &pb.GetOutboundStatusResponse{Status: &observatory.ObservationResult{Status: []*observatory.OutboundStatus{{OutboundTag: "unknown", Alive: true}}}}, nil
}

type connectionEvents struct{ ended chan struct{} }

func (s *connectionEvents) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}
func (s *connectionEvents) HandleRPC(context.Context, stats.RPCStats) {}
func (s *connectionEvents) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}
func (s *connectionEvents) HandleConn(_ context.Context, event stats.ConnStats) {
	if _, ok := event.(*stats.ConnEnd); ok {
		select {
		case s.ended <- struct{}{}:
		default:
		}
	}
}

func TestObservatoryUnknownTag(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	events := &connectionEvents{ended: make(chan struct{}, 1)}
	server := grpc.NewServer(grpc.StatsHandler(events))
	observer := &observatoryTestServer{}
	pb.RegisterObservatoryServiceServer(server, observer)
	go server.Serve(listener)
	defer server.Stop()
	if err := configure.AddConnect(configure.NodeRef{TYPE: configure.ServerType, ID: 1}); err != nil {
		t.Fatal(err)
	}
	defer configure.ClearConnects("proxy")
	ProcessManager.mu.Lock()
	previous := ProcessManager.p
	ProcessManager.p = &Process{tag2WhichIndex: map[string]int{}}
	ProcessManager.mu.Unlock()
	defer func() { ProcessManager.mu.Lock(); ProcessManager.p = previous; ProcessManager.mu.Unlock() }()
	box := ApiFeed.SubscribeMessage("observatory")
	defer box.Cancel()
	stop := ObservatoryProducer(listener.Addr().(*net.TCPAddr).Port, []string{"proxy"})
	defer stop()
	select {
	case msg := <-box.Messages:
		statuses := msg.Body.(gin.H)["outboundStatus"].([]OutboundStatus)
		for _, s := range statuses {
			if s.Which != nil {
				t.Errorf("unknown tag attributed to %+v", s.Which)
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no observatory response")
	}
	// Removing the connected list makes index zero out of bounds as well.
	if err := configure.ClearConnects("proxy"); err != nil {
		t.Fatal(err)
	}
	before := observer.calls.Load()
	time.Sleep(1500 * time.Millisecond)
	if calls := observer.calls.Load() - before; calls > 2 {
		t.Errorf("polled %d times in 1.5 seconds", calls)
	}
	observer.unavailable.Store(true)
	select {
	case <-events.ended:
	case <-time.After(3 * time.Second):
		t.Error("unavailable connection was not closed before reconnecting")
	}
}
