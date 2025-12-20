package registry

import (
	"service-center/pkg/api/pb"
	"sync"

	grpc "google.golang.org/grpc"
)

type Service struct {
	ServiceName   string
	Instances     serviceInstances
	svcMutex      sync.RWMutex
	eventWatchers map[pb.EventType][]grpc.ServerStreamingServer[pb.WatchResponse]
}

func NewService(name string) *Service {
	return &Service{
		ServiceName:   name,
		Instances:     newServiceInstances(),
		eventWatchers: make(map[pb.EventType][]grpc.ServerStreamingServer[pb.WatchResponse]),
	}
}

func (s *Service) AddInstance(inst *pb.ServiceInstance) {
	s.svcMutex.Lock()
	s.Instances.AddInstance(inst)
	watcherCopy := make([]grpc.ServerStreamingServer[pb.WatchResponse], len(s.eventWatchers[pb.EventType_EVENT_ADD]))
	copy(watcherCopy, s.eventWatchers[pb.EventType_EVENT_ADD])
	s.svcMutex.Unlock()
	go func() {
		watcherToBeRemoved := []grpc.ServerStreamingServer[pb.WatchResponse]{}
		for _, watcher := range watcherCopy {
			if err := informWatchers(watcher, pb.EventType_EVENT_ADD, inst); err != nil {
				watcherToBeRemoved = append(watcherToBeRemoved, watcher)
				continue
			}
		}
		if len(watcherToBeRemoved) > 0 {
			s.removeWatchers(pb.EventType_EVENT_ADD, watcherToBeRemoved)
			s.removeWatchers(pb.EventType_EVENT_DELETE, watcherToBeRemoved)
		}
	}()
}

func (s *Service) removeWatchers(eventType pb.EventType, watcherToBeRemoved []grpc.ServerStreamingServer[pb.WatchResponse]) {
	s.svcMutex.Lock()
	defer s.svcMutex.Unlock()
	remainingWatchers := []grpc.ServerStreamingServer[pb.WatchResponse]{}
	for _, existingWatcher := range s.eventWatchers[eventType] {
		shouldRemove := false
		for _, toBeRemoved := range watcherToBeRemoved {
			if existingWatcher == toBeRemoved {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			remainingWatchers = append(remainingWatchers, existingWatcher)
		}
	}
	s.eventWatchers[eventType] = remainingWatchers
}

func (s *Service) RemoveInstance(instanceID string) {
	s.svcMutex.Lock()
	inst := s.Instances.DropInstance(instanceID)
	watcherCopy := make([]grpc.ServerStreamingServer[pb.WatchResponse], len(s.eventWatchers[pb.EventType_EVENT_DELETE]))
	copy(watcherCopy, s.eventWatchers[pb.EventType_EVENT_DELETE])
	s.svcMutex.Unlock()
	go func() {
		watcherToBeRemoved := []grpc.ServerStreamingServer[pb.WatchResponse]{}
		for _, watcher := range watcherCopy {
			if err := informWatchers(watcher, pb.EventType_EVENT_DELETE, inst); err != nil {
				watcherToBeRemoved = append(watcherToBeRemoved, watcher)
				continue
			}
		}
		if len(watcherToBeRemoved) > 0 {
			s.removeWatchers(pb.EventType_EVENT_DELETE, watcherToBeRemoved)
			s.removeWatchers(pb.EventType_EVENT_ADD, watcherToBeRemoved)
		}
	}()
}

func (s *Service) GetInstances() []*pb.ServiceInstance {
	s.svcMutex.RLock()
	defer s.svcMutex.RUnlock()
	return s.Instances.GetAllInstances()
}

func (s *Service) AddWatcher(eventType pb.EventType, watcher grpc.ServerStreamingServer[pb.WatchResponse]) {
	s.svcMutex.Lock()
	defer s.svcMutex.Unlock()
	for _, inst := range s.Instances.GetAllInstances() {
		if err := informWatchers(watcher, eventType, inst); err != nil {
			continue
		}
	}
	s.eventWatchers[eventType] = append(s.eventWatchers[eventType], watcher)
}

func informWatchers(watcher grpc.ServerStreamingServer[pb.WatchResponse], eventType pb.EventType, inst *pb.ServiceInstance) error {
	return watcher.Send(&pb.WatchResponse{
		EventType: eventType,
		Instance:  inst,
	})
}
