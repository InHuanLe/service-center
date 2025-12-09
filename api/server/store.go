package server

import (
	"fmt"
	"service-center/api/api"
	"sync"
	"time"
)

type Store interface {
	Store(string, *api.RegisterServiceRequest) error
	Available(string) bool
	Update(string, *api.RegisterServiceRequest) error
	LoadService(string) (*api.ServiceInstances, error)
	LoadInstance(string) (*api.RegisterServiceRequest, error)
	Delete(*api.RegisterServiceRequest)
}

type InMemoryStore struct {
	services      sync.Map // map[string][]instanceId
	instances     sync.Map // map[instanceId]*api.RegisterServiceRequest
	instancesStat sync.Map // map[instanceId]time.Time
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		services:      sync.Map{},
		instances:     sync.Map{},
		instancesStat: sync.Map{},
	}
}

func (s *InMemoryStore) instanceInService(instanceId, serviceName string) bool {
	value, ok := s.services.Load(serviceName)
	if !ok {
		return false
	}
	instanceMap := value.(map[string]struct{})
	_, ok = instanceMap[instanceId]
	return ok
}

func (s *InMemoryStore) registered(instanceId string) bool {
	_, ok := s.instances.Load(instanceId)
	return ok
}

func (s *InMemoryStore) Available(instanceId string) bool {
	value, ok := s.instancesStat.Load(instanceId)
	if !ok {
		return false
	}
	cfg, ok := s.instances.Load(instanceId)
	if !ok {
		return false
	}
	req := cfg.(*api.RegisterServiceRequest)
	heartbeatInterval := time.Duration(req.Ttl) * time.Second
	if heartbeatInterval == 0 {
		heartbeatInterval = 30 * time.Second
	}
	if time.Since(value.(time.Time)) > heartbeatInterval*3 {
		return false
	}
	return s.registered(instanceId) && s.instanceInService(instanceId, req.ServiceName)
}

func (s *InMemoryStore) Store(instanceId string, req *api.RegisterServiceRequest) error {
	instanceState := s.Available(instanceId)
	if s.Available(instanceId) {
		return fmt.Errorf("already stored")
	}
	registered := s.registered(instanceId)
	instanceInService := s.instanceInService(instanceId, req.ServiceName)
	if !instanceInService {
		value, _ := s.services.LoadOrStore(req.ServiceName, make(map[string]struct{}))
		instanceMap := value.(map[string]struct{})
		instanceMap[instanceId] = struct{}{}
		s.services.Store(req.ServiceName, instanceMap)
	}
	if !registered {
		s.instances.Store(instanceId, req)
	}
	if !instanceState {
		s.instancesStat.Store(instanceId, time.Now())
	}
	return nil
}

func (s *InMemoryStore) LoadService(serviceName string) (*api.ServiceInstances, error) {
	value, ok := s.services.Load(serviceName)
	if !ok {
		return &api.ServiceInstances{
			ServiceName: serviceName,
			Instances:   []*api.RegisterServiceRequest{},
		}, fmt.Errorf("service not found")
	}
	instanceMap := value.(map[string]struct{})
	instances := make([]*api.RegisterServiceRequest, 0, len(instanceMap))
	for instanceId := range instanceMap {
		instanceValue, ok := s.instances.Load(instanceId)
		if ok && s.Available(instanceId) {
			instances = append(instances, instanceValue.(*api.RegisterServiceRequest))
		}
	}
	return &api.ServiceInstances{
		ServiceName: serviceName,
		Instances:   instances,
	}, nil
}

func (s *InMemoryStore) LoadInstance(instanceId string) (*api.RegisterServiceRequest, error) {
	value, ok := s.instances.Load(instanceId)
	if !ok {
		return nil, fmt.Errorf("instance not found")
	}
	return value.(*api.RegisterServiceRequest), nil
}

func (s *InMemoryStore) Update(instanceId string, req *api.RegisterServiceRequest) error {
	if !s.Available(instanceId) {
		return fmt.Errorf("instance not available")
	}
	if req == nil {
		s.instancesStat.Store(instanceId, time.Now())
		return nil
	}
	s.instances.Store(instanceId, req)
	return nil
}

func (s *InMemoryStore) Delete(req *api.RegisterServiceRequest) {
	instanceId := req.InstanceId
	serviceName := req.ServiceName

	if value, ok := s.services.Load(serviceName); ok {
		instanceMap := value.(map[string]struct{})
		delete(instanceMap, instanceId)
		s.services.Store(serviceName, instanceMap)
	}
	s.instances.Delete(instanceId)
	s.instancesStat.Delete(instanceId)
}
