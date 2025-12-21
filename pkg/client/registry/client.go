package registry

import (
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"service-center/pkg/api/pb"
)

// Client wraps the generated gRPC client to provide a small convenience API.
type Client struct {
	conn *grpc.ClientConn
	c    pb.ServiceRegistryClient
}

// NewClient creates a new registry client connected to the given server address.
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn: conn,
		c:    pb.NewServiceRegistryClient(conn),
	}, nil
}

// NewClientFromConn creates a registry client from an existing grpc connection.
func NewClientFromConn(cc grpc.ClientConnInterface) *Client {
	return &Client{c: pb.NewServiceRegistryClient(cc)}
}

// Close closes the underlying connection if it was created by NewClient.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
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
