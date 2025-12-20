package registry

import (
	"testing"

	"service-center/pkg/api/pb"
)

func TestServiceInstanceStore(t *testing.T) {
	si := newServiceInstances()
	inst1 := &pb.ServiceInstance{ServiceName: "svc", InstanceId: "i1"}
	inst2 := &pb.ServiceInstance{ServiceName: "svc", InstanceId: "i2"}

	si.AddInstance(inst1)
	si.AddInstance(inst2)

	all := si.GetAllInstances()
	if len(all) != 2 {
		t.Fatalf("GetAllInstances = %d, want 2", len(all))
	}

	// Drop one and check
	ret := si.DropInstance("i1")
	if ret == nil || ret.InstanceId != "i1" {
		t.Fatalf("DropInstance returned %+v, want i1", ret)
	}
	all2 := si.GetAllInstances()
	if len(all2) != 1 || all2[0].InstanceId != "i2" {
		t.Fatalf("GetAllInstances after drop = %+v, want only i2", all2)
	}
}
