package registry

import (
	"context"
	"testing"
	"time"

	"service-center/pkg/api/pb"

	"google.golang.org/grpc/metadata"
)

func TestRegisterDiscoverAndDeregister(t *testing.T) {
	svr := NewRegistryServer()
	ctx := context.Background()

	// Register an instance
	inst := &pb.ServiceInstance{ServiceName: "user", InstanceId: "id-1", Address: "127.0.0.1", Port: 8080}
	resp, err := svr.Register(ctx, &pb.RegisterRequest{Instance: inst})
	if err != nil || !resp.Success {
		t.Fatalf("register failed: resp=%v err=%v", resp, err)
	}

	// Discover should return the instance
	svc, err := svr.Discover(ctx, &pb.DiscoverRequest{ServiceName: "user"})
	if err != nil {
		t.Fatalf("discover err: %v", err)
	}
	if got := len(svc.Instances); got != 1 {
		t.Fatalf("discover instances = %d, want 1", got)
	}

	// Heartbeat should succeed
	hb, err := svr.Heartbeat(ctx, &pb.HeartbeatRequest{ServiceName: "user", InstanceId: "id-1"})
	if err != nil || !hb.Success {
		t.Fatalf("heartbeat failed: resp=%v err=%v", hb, err)
	}

	// Deregister
	der, err := svr.Deregister(ctx, &pb.DeregisterRequest{ServiceName: "user", InstanceId: "id-1"})
	if err != nil || !der.Success {
		t.Fatalf("deregister failed: resp=%v err=%v", der, err)
	}

	// Discover again should be empty
	svc2, err := svr.Discover(ctx, &pb.DiscoverRequest{ServiceName: "user"})
	if err != nil {
		t.Fatalf("discover err: %v", err)
	}
	if got := len(svc2.Instances); got != 0 {
		t.Fatalf("discover instances after deregister = %d, want 0", got)
	}
}

// fakeWatchServer implements grpc.ServerStreamingServer[pb.WatchResponse] with a minimal surface
// so we can capture messages in tests.
type fakeWatchServer struct {
	msgs []*pb.WatchResponse
	fail bool
	ctx  context.Context
}

func (f *fakeWatchServer) Send(resp *pb.WatchResponse) error {
	if f.fail {
		return fmtError("send fail")
	}
	f.msgs = append(f.msgs, resp)
	return nil
}

func (f *fakeWatchServer) Context() context.Context {
	if f.ctx != nil {
		return f.ctx
	}
	return context.Background()
}

// Implement remaining methods of embedded grpc.ServerStream to satisfy interface
func (f *fakeWatchServer) SetHeader(md metadata.MD) error { return nil }
func (f *fakeWatchServer) SendHeader(md metadata.MD) error { return nil }
func (f *fakeWatchServer) SetTrailer(md metadata.MD)      {}
func (f *fakeWatchServer) RecvMsg(m any) error            { return nil }
func (f *fakeWatchServer) SendMsg(m any) error            { return nil }

// fmtError is defined locally to avoid importing fmt in tests
func fmtError(msg string) error { return &simpleError{s: msg} }

type simpleError struct{ s string }

func (e *simpleError) Error() string { return e.s }

func TestWatchInitialSnapshotAndAddDeleteEvents(t *testing.T) {
	svr := NewRegistryServer()
	ctx := context.Background()

	// Seed with two instances
	_, _ = svr.Register(ctx, &pb.RegisterRequest{Instance: &pb.ServiceInstance{ServiceName: "order", InstanceId: "id-a", Address: "10.0.0.1", Port: 9000}})
	_, _ = svr.Register(ctx, &pb.RegisterRequest{Instance: &pb.ServiceInstance{ServiceName: "order", InstanceId: "id-b", Address: "10.0.0.2", Port: 9001}})

	// Watch ADD events
	fw := &fakeWatchServer{}
	if err := svr.Watch(&pb.WatchRequest{ServiceName: "order", WatchEvent: int32(pb.EventType_EVENT_ADD)}, fw); err != nil {
		t.Fatalf("watch err: %v", err)
	}

	// Initial snapshot should contain existing instances as EVENT_ADD
	if len(fw.msgs) != 2 {
		t.Fatalf("initial snapshot messages = %d, want 2", len(fw.msgs))
	}
	for _, m := range fw.msgs {
		if m.EventType != pb.EventType_EVENT_ADD {
			t.Fatalf("initial message event type = %v, want EVENT_ADD", m.EventType)
		}
	}

	// Register a new instance -> should broadcast an EVENT_ADD asynchronously
	_, _ = svr.Register(ctx, &pb.RegisterRequest{Instance: &pb.ServiceInstance{ServiceName: "order", InstanceId: "id-c", Address: "10.0.0.3", Port: 9002}})
	// Wait briefly for async goroutine
	time.Sleep(50 * time.Millisecond)
	if got := len(fw.msgs); got < 3 {
		t.Fatalf("messages after new register = %d, want >=3", got)
	}
	last := fw.msgs[len(fw.msgs)-1]
	if last.EventType != pb.EventType_EVENT_ADD || last.Instance == nil || last.Instance.InstanceId != "id-c" {
		t.Fatalf("last message = %+v, want EVENT_ADD for id-c", last)
	}

	// Watch DELETE events and then remove an instance
	fwDel := &fakeWatchServer{}
	if err := svr.Watch(&pb.WatchRequest{ServiceName: "order", WatchEvent: int32(pb.EventType_EVENT_DELETE)}, fwDel); err != nil {
		t.Fatalf("watch delete err: %v", err)
	}
	// DELETE watcher usually doesn't need initial snapshot; our AddWatcher sends ADD snapshot only.
	// Remove one instance
	_, _ = svr.Deregister(ctx, &pb.DeregisterRequest{ServiceName: "order", InstanceId: "id-b"})
	// Wait for async deletion broadcast
	time.Sleep(50 * time.Millisecond)
	if len(fwDel.msgs) == 0 {
		t.Fatalf("no delete messages received, want at least 1")
	}
	del := fwDel.msgs[len(fwDel.msgs)-1]
	if del.EventType != pb.EventType_EVENT_DELETE || del.Instance == nil || del.Instance.InstanceId != "id-b" {
		t.Fatalf("delete message = %+v, want EVENT_DELETE for id-b", del)
	}
}
