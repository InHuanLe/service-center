package server

import (
	"context"
	"service-center/api/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServiceCenterServer struct {
	api.UnimplementedServiceCenterServer
	store Store
}

func NewServiceCenterServer() *ServiceCenterServer {
	return &ServiceCenterServer{
		store: NewInMemoryStore(),
	}
}

func (svr *ServiceCenterServer) DeregisterService(ctx context.Context, deregisterReq *api.DeregisterServiceRequest) (*api.CommonResponse, error) {
	svr.store.Delete(&api.RegisterServiceRequest{
		InstanceId:  deregisterReq.InstanceId,
		ServiceName: deregisterReq.ServiceName,
	})
	return &api.CommonResponse{
		Code:    0,
		Message: "Deregistered successfully",
	}, nil
}

func (svr *ServiceCenterServer) RegisterService(ctx context.Context, registerReq *api.RegisterServiceRequest) (*api.CommonResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RegisterService not implemented")
}

func (svr *ServiceCenterServer) Watch(watchReq *api.WatchRequest, sss grpc.ServerStreamingServer[api.ServicesDiscoverResponse]) error {
	return status.Error(codes.Unimplemented, "method Watch not implemented")
}

func (svr *ServiceCenterServer) Heartbeat(ctx context.Context, heartbeatReq *api.HeartbeatRequest) (*api.CommonResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Heartbeat not implemented")
}

func (svr *ServiceCenterServer) Discover(ctx context.Context, discoverReq *api.DiscoverRequest) (*api.ServicesDiscoverResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Discover not implemented")
}

func (svr *ServiceCenterServer) mustEmbedUnimplementedServiceCenterServer() {}
