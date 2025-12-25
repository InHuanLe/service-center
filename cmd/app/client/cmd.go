package client

import (
	"service-center/pkg/registry"

	"google.golang.org/grpc"
)

type Options struct {
	target string
}

func (o *Options) Target() string {
	return o.target
}

func NewClient(o *Options) (registry.RegistryClient, error) {
	conn, err := grpc.NewClient(o.Target(), grpc.EmptyDialOption{})
	if err != nil {
		return nil, err
	}
	c := registry.NewRegistryClient(conn)
	return c, nil
}
