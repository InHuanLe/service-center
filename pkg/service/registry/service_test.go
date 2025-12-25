package registry

import (
	"context"
	"service-center/pkg/registry"
	"testing"
)

type fakeOptions struct {
	endpoints []string
}

func (o fakeOptions) Endpoints() []string {
	return o.endpoints
}

var opts = fakeOptions{
	endpoints: []string{"http://192.168.104.111:2379"},
}

func TestNewRegistryService(t *testing.T) {
	service, err := NewRegistryService(opts)
	if err != nil {
		t.Fatalf("failed to create registry service: %v", err)
	}
	if service == nil {
		t.Fatal("expected non-nil registry service")
	}
}
func TestRegister(t *testing.T) {
	service, _ := NewRegistryService(opts)
	resp, err := service.Register(context.Background(), &registry.RegisterRequest{
		Name:       "o.Name()",
		InstanceId: "1111",
		Address:    "o.Address()",
		Metadata:   map[string]string{},
		Port:       10079,
		Ttl:        455,
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	t.Logf("Register response: %+v", resp)
}
