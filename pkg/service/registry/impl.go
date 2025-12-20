package registry

import (
	"context"
	"fmt"
	"sync"

	"service-center/pkg/api/pb"

	grpc "google.golang.org/grpc"
)

// RegistryServer is a simple in-memory implementation of the generated ServiceRegistryServer.
type RegistryServer struct {
	pb.UnimplementedServiceRegistryServer
	mu       sync.RWMutex
	services map[string]*Service
}

// NewRegistryServer creates a new in-memory RegistryServer.
func NewRegistryServer() *RegistryServer {
	return &RegistryServer{
		services: make(map[string]*Service),
	}
}

func (s *RegistryServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req == nil || req.Instance == nil {
		return &pb.RegisterResponse{Success: false}, nil
	}
	inst := req.Instance

	s.mu.Lock()
	m, ok := s.services[inst.ServiceName]
	if !ok {
		m = NewService(inst.ServiceName)
		s.services[inst.ServiceName] = m
	}
	s.mu.Unlock()
	m.AddInstance(inst)
	return &pb.RegisterResponse{Success: true}, nil
}

func (s *RegistryServer) Deregister(ctx context.Context, req *pb.DeregisterRequest) (*pb.DeregisterResponse, error) {
	if req == nil {
		return &pb.DeregisterResponse{Success: false}, nil
	}
	if m, ok := s.getServices(req.ServiceName); ok {
		m.RemoveInstance(req.InstanceId)
	}
	return &pb.DeregisterResponse{Success: true}, nil
}

func (s *RegistryServer) Discover(ctx context.Context, req *pb.DiscoverRequest) (*pb.Service, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid discover request")
	}
	name := req.ServiceName

	if svcs, ok := s.getServices(name); ok {
		return &pb.Service{
			ServiceName: name,
			Instances:   svcs.GetInstances(),
		}, nil
	}
	return &pb.Service{ServiceName: name}, nil
}

func (s *RegistryServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req == nil {
		return &pb.HeartbeatResponse{Success: false}, nil
	}
	name := req.ServiceName
	id := req.InstanceId

	if m, ok := s.getServices(name); ok {
		// TODO: 太垃了，心跳不应该这样实现
		for _, inst := range m.GetInstances() {
			if inst.InstanceId == id {
				return &pb.HeartbeatResponse{Success: true}, nil
			}
		}
	}
	return &pb.HeartbeatResponse{Success: false}, nil
}

func (s *RegistryServer) Watch(req *pb.WatchRequest, stream grpc.ServerStreamingServer[pb.WatchResponse]) error {
	// Send current state as initial snapshot
	svcName := req.ServiceName
	svc, ok := s.getServices(svcName)
	if !ok {
		return fmt.Errorf("service not found")
	}
	if req.WatchEvent == int32(pb.EventType_EVENT_ADD) {
		svc.AddWatcher(pb.EventType_EVENT_ADD, stream)
		return nil
	}
	if req.WatchEvent == int32(pb.EventType_EVENT_DELETE) {
		svc.AddWatcher(pb.EventType_EVENT_DELETE, stream)
		return nil
	}
	return fmt.Errorf("invalid event type")
}


func (s *RegistryServer) getServices(name string) (*Service, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, ok := s.services[name]
	return svc, ok
}
