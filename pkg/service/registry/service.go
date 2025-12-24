package registry

import (
	"context"
	"encoding/json"
	pb "service-center/pkg/registry"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type RegistryService struct {
	pb.UnimplementedRegistryServer
	etcdClient *clientv3.Client
}

func NewRegistryService(opts Options) (*RegistryService, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: opts.Endpoints(),
	})
	if err != nil {
		return nil, err
	}
	return &RegistryService{
		etcdClient: etcdClient,
	}, nil
}

func (s *RegistryService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	ttl := req.GetTtl()
	svc := &pb.ServiceInstance{
		InstanceId: req.GetInstanceId(),
		Address:    req.GetAddress(),
		Metadata:   req.GetMetadata(),
		Port:       req.GetPort(),
		Ttl:        ttl,
	}
	// 创建租约
	var leaseID clientv3.LeaseID
	if resp, err := s.etcdClient.Grant(ctx, ttl); err != nil || resp.Error != "" {
		return nil, err
	} else {
		leaseID = resp.ID
	}
	// 序列化服务实例信息
	buffer, err := json.Marshal(svc)
	if err != nil {
		return nil, err
	}
	// 存储服务实例信息到 etcd，附加租约
	if _, err = s.etcdClient.Put(ctx, svc.InstanceId, string(buffer), clientv3.WithLease(leaseID)); err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{
		Success: true,
		LeaseId: int64(leaseID),
	}, nil
}

func (s *RegistryService) Deregister(ctx context.Context, req *pb.DeregisterRequest) (*pb.DeregisterResponse, error) {
	_, err := s.etcdClient.Delete(ctx, req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return &pb.DeregisterResponse{
		Success: true,
	}, nil
}

func (s *RegistryService) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	leaseID := clientv3.LeaseID(req.GetLeaseId())
	// 延长租约
	if _, err := s.etcdClient.KeepAliveOnce(ctx, leaseID); err != nil {
		return nil, err
	}
	return &pb.HeartbeatResponse{
		Success: true,
	}, nil
}

func (s *RegistryService) Watch(req *pb.WatchRequest, stream pb.Registry_WatchServer) error {
	watchCh := s.etcdClient.Watch(s.etcdClient.Ctx(), req.GetServiceName(), clientv3.WithPrefix())
	for watchResp := range watchCh {
		for _, event := range watchResp.Events {
			svc := &pb.ServiceInstance{}
			if err := json.Unmarshal(event.Kv.Value, svc); err != nil {
				return err
			}
			if err := stream.Send(svc); err != nil {
				return err
			}
		}
	}
	return nil
}
