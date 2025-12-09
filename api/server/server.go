package server

import (
	"context"
	"service-center/api/api"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	INSTANCE_ADDED   = 0x01
	INSTANCE_REMOVED = 0x02
	INSTANCE_UPDATED = 0x04
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
	err := svr.store.Store(registerReq.InstanceId, registerReq)
	if err != nil {
		return &api.CommonResponse{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	return &api.CommonResponse{
		Code:    0,
		Message: "Registered successfully",
	}, nil
}

func (svr *ServiceCenterServer) Watch(watchReq *api.WatchRequest, sss grpc.ServerStreamingServer[api.ServicesDiscoverResponse]) error {
	// 先实现最简单的，不管什么情况都每10秒发一次全量数据
	go func() {
		for {
			time.Sleep(10 * time.Second)
			serviceInstances, err := svr.store.LoadService(watchReq.ServiceName)
			if err != nil {
				continue
			}
			err = sss.Send(&api.ServicesDiscoverResponse{
				Code:     0,
				Services: []*api.ServiceInstances{serviceInstances},
			})
			if err != nil {
				return
			}
		}
	}()
	return nil
}

func (svr *ServiceCenterServer) Heartbeat(ctx context.Context, heartbeatReq *api.HeartbeatRequest) (*api.CommonResponse, error) {
	err := svr.store.Update(heartbeatReq.InstanceId, nil)
	if err != nil {
		return &api.CommonResponse{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	return &api.CommonResponse{
		Code:    0,
		Message: "Heartbeat received",
	}, nil
}

func (svr *ServiceCenterServer) Discover(ctx context.Context, discoverReq *api.DiscoverRequest) (*api.ServicesDiscoverResponse, error) {
	if discoverReq == nil || discoverReq.ServiceName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid discover request")
	}
	serviceInstances, err := svr.store.LoadService(discoverReq.ServiceName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "service not found")
	}
	if discoverReq.HealthyOnly {
		healthyInstances := make([]*api.RegisterServiceRequest, 0)
		for _, instance := range serviceInstances.Instances {
			if svr.store.Available(instance.InstanceId) {
				healthyInstances = append(healthyInstances, instance)
			}
		}
		serviceInstances.Instances = healthyInstances
	}
	return &api.ServicesDiscoverResponse{
		Code:     0,
		Services: []*api.ServiceInstances{serviceInstances},
	}, nil
}

func (svr *ServiceCenterServer) mustEmbedUnimplementedServiceCenterServer() {}
