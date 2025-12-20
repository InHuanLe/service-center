package registry

import (
	"context"
	"io"

	"google.golang.org/grpc"

	"service-center/pkg/api/pb"
)

// Client wraps the generated gRPC client to provide a small convenience API.
type Client struct {
	c pb.ServiceRegistryClient
}

// NewClientFromConn creates a registry client from an existing grpc connection.
func NewClientFromConn(cc grpc.ClientConnInterface) *Client {
	return &Client{c: pb.NewServiceRegistryClient(cc)}
}

func (c *Client) Register(ctx context.Context, inst *pb.ServiceInstance) (bool, error) {
	res, err := c.c.Register(ctx, &pb.RegisterRequest{Instance: inst})
	if err != nil {
		return false, err
	}
	return res.Success, nil
}

func (c *Client) Deregister(ctx context.Context, serviceName, instanceID string) (bool, error) {
	res, err := c.c.Deregister(ctx, &pb.DeregisterRequest{ServiceName: serviceName, InstanceId: instanceID})
	if err != nil {
		return false, err
	}
	return res.Success, nil
}

func (c *Client) Discover(ctx context.Context, serviceName string) (*pb.Service, error) {
	return c.c.Discover(ctx, &pb.DiscoverRequest{ServiceName: serviceName})
}

func (c *Client) Heartbeat(ctx context.Context, serviceName, instanceID string) (bool, error) {
	res, err := c.c.Heartbeat(ctx, &pb.HeartbeatRequest{ServiceName: serviceName, InstanceId: instanceID})
	if err != nil {
		return false, err
	}
	return res.Success, nil
}

// Watch returns a channel that receives WatchResponse values; the caller should cancel the context to stop.
func (c *Client) Watch(ctx context.Context, serviceName string) (<-chan *pb.WatchResponse, error) {
	stream, err := c.c.Watch(ctx, &pb.WatchRequest{ServiceName: serviceName})
	if err != nil {
		return nil, err
	}
	ch := make(chan *pb.WatchResponse)
	go func() {
		defer close(ch)
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			ch <- msg
		}
	}()
	return ch, nil
}
