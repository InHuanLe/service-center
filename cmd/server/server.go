package main

import (
	"context"
	"net"
	"service-center/pkg/api/pb"
	"service-center/pkg/service/registry"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type RegistryOptions struct {
	Address string
	TimeOut time.Duration
}

func NewRegistryOptions() *RegistryOptions {
	return &RegistryOptions{
		Address: ":50051",
		TimeOut: 10 * time.Second,
	}
}

func (o *RegistryOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Address, "address", "a", ":50051", "The address for the registry server to listen on")
	cmd.Flags().DurationVarP(&o.TimeOut, "timeout", "t", 10*time.Second, "The timeout for server connections")
}

func (o *RegistryOptions) Run(ctx context.Context) error {
	gsvr := grpc.NewServer(grpc.ConnectionTimeout(o.TimeOut))
	nSvr := registry.NewRegistryServer()
	pb.RegisterServiceRegistryServer(gsvr, nSvr)
	listener, err := net.Listen("tcp", o.Address)
	if err != nil {
		return err
	}
	// 等到所有的请求都响应完之后才stop这个listener
	go func() {
		<-ctx.Done()
		gsvr.GracefulStop()
	}()
	return gsvr.Serve(listener)
}
