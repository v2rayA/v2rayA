// Package command implements the v2ray-compatible ObservatoryService gRPC endpoint.
// This is structurally derived from v2ray's app/observatory/command/command.go,
// adapted to use xray's interfaces. The gRPC service is registered under the
// v2ray service name "v2ray.core.app.observatory.command.ObservatoryService" so
// that v2rayA's ObservatoryProducer can query it natively.
//
// The request type (GetOutboundStatusRequest) has a Tag field (proto field 1)
// that identifies the balancer group. The response uses xray's GetOutboundStatusResponse
// which is wire-compatible with v2ray's (identical field numbers).
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package command

import (
	"context"

	"github.com/v2rayA/v2raya-core/hint/app/observatory/multiobservatory"
	xray_obs "github.com/xtls/xray-core/app/observatory"
	xray_obs_cmd "github.com/xtls/xray-core/app/observatory/command"
	"github.com/xtls/xray-core/common"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/extension"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// service implements the v2ray ObservatoryService using xray's observatory.
// It is structurally identical to v2ray's command.go service but uses xray's
// extension.Observatory and xray's GetOutboundStatusResponse internally.
type service struct {
	v           *core.Instance
	observatory extension.Observatory
}

// GetOutboundStatus returns the observation result for the given balancer group tag.
// If tag is empty, all groups are aggregated. This matches v2ray's behavior.
func (s *service) GetOutboundStatus(ctx context.Context, req *GetOutboundStatusRequest) (*xray_obs_cmd.GetOutboundStatusResponse, error) {
	if s.observatory == nil {
		return &xray_obs_cmd.GetOutboundStatusResponse{}, nil
	}
	return s.getByTag(ctx, req.GetTag())
}

// getByTag routes the query to the correct observatory group.
// If the observatory is a MultiObservatory (v2rayA multi-group), it uses per-tag lookup.
// Otherwise, it calls GetObservation() to return all results.
func (s *service) getByTag(ctx context.Context, tag string) (*xray_obs_cmd.GetOutboundStatusResponse, error) {
	if mo, ok := s.observatory.(*multiobservatory.MultiObservatory); ok {
		result, err := mo.GetObservationByTag(tag, ctx)
		if err != nil {
			return nil, err
		}
		if obs, ok := result.(*xray_obs.ObservationResult); ok {
			return &xray_obs_cmd.GetOutboundStatusResponse{Status: obs}, nil
		}
	}
	result, err := s.observatory.GetObservation(ctx)
	if err != nil {
		return nil, err
	}
	if obs, ok := result.(*xray_obs.ObservationResult); ok {
		return &xray_obs_cmd.GetOutboundStatusResponse{Status: obs}, nil
	}
	return &xray_obs_cmd.GetOutboundStatusResponse{}, nil
}

// Register implements commander.Service.
// It registers the ObservatoryService under the v2ray gRPC path so v2rayA can
// query it. The response type (xray's GetOutboundStatusResponse) is wire-
// compatible with v2ray's because both have identical proto field numbers.
func (s *service) Register(server *grpc.Server) {
	server.RegisterService(&v2rayObservatoryServiceDesc, s)
}

// v2rayObservatoryServiceDesc is the gRPC ServiceDesc for
// v2ray.core.app.observatory.command.ObservatoryService.
// This mirrors v2ray's RegisterObservatoryServiceServer but without importing
// the v2fly module, to avoid proto type registry conflicts with xray.
var v2rayObservatoryServiceDesc = grpc.ServiceDesc{
	ServiceName: "v2ray.core.app.observatory.command.ObservatoryService",
	HandlerType: (*interface{})(nil), // using raw handler; type is checked via closure
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetOutboundStatus",
			Handler:    getOutboundStatusHandler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "app/observatory/command/command.proto",
}

// getOutboundStatusHandler is the gRPC unary handler for GetOutboundStatus.
// It decodes the request as GetOutboundStatusRequest (which has Tag=field1) so
// the Tag is populated directly — no protowire extraction needed.
func getOutboundStatusHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOutboundStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*service).GetOutboundStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v2ray.core.app.observatory.command.ObservatoryService/GetOutboundStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*service).GetOutboundStatus(ctx, req.(*GetOutboundStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// init registers this service's Config in xray's common registry so xray's commander
// can instantiate it from the TypedMessage in the config file.
func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		s := core.MustFromContext(ctx)
		sv := &service{v: s}
		// RequireFeatures with optional=true: if no observatory is configured, the
		// service still starts but returns empty results.
		_ = s.RequireFeatures(func(obs extension.Observatory) {
			sv.observatory = obs
		}, true)
		return sv, nil
	}))
}

// Ensure proto.Message is imported (used for type assertion in getByTag).
var _ proto.Message = (*GetOutboundStatusRequest)(nil)
