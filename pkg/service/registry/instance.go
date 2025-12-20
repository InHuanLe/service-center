package registry

import "service-center/pkg/api/pb"

type serviceInstances map[string]*pb.ServiceInstance

func (si serviceInstances) AddInstance(inst *pb.ServiceInstance) {
	si[inst.InstanceId] = inst
}
func (si serviceInstances) DropInstance(instanceID string) *pb.ServiceInstance {
	inst, ok := si[instanceID]
	if ok {
		delete(si, instanceID)
		return inst
	}
	return nil
}
func (si serviceInstances) GetAllInstances() []*pb.ServiceInstance {
	instances := make([]*pb.ServiceInstance, 0, len(si))
	for _, inst := range si {
		instances = append(instances, inst)
	}
	return instances
}
func newServiceInstances() serviceInstances {
	return make(serviceInstances)
}
