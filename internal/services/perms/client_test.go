package perms

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/xligenda/reports/internal/services/perms/pb"
)

// MockPermsServer implements pb.PermsServer for testing
type MockPermsServer struct {
	pb.UnimplementedPermsServer
	ShouldFail     bool
	AlwaysGrant    bool
	CheckCallCount int
}

func (m *MockPermsServer) Check(ctx context.Context, req *pb.AccessRequest) (*pb.AccessResponse, error) {
	m.CheckCallCount++

	if m.ShouldFail {
		return nil, status.Error(codes.Internal, "mock server error")
	}

	if m.AlwaysGrant {
		return &pb.AccessResponse{
			Granted: true,
			Reason:  "Access granted",
		}, nil
	}

	return &pb.AccessResponse{
		Granted: len(req.Id) > 0,
		Reason:  "Test response",
	}, nil
}

// startMockServer starts a mock gRPC server for testing
func startMockServer(t *testing.T, mockServer *MockPermsServer) (string, *grpc.Server) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterPermsServer(grpcServer, mockServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	return listener.Addr().String(), grpcServer
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wantError bool
	}{
		{
			name:      "valid target",
			target:    "127.0.0.1:9999",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.target)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if client != nil {
					client.Close()
				}
			}
		})
	}
}

func TestClientCheck_Success(t *testing.T) {
	mockServer := &MockPermsServer{AlwaysGrant: true}
	address, grpcServer := startMockServer(t, mockServer)
	defer grpcServer.Stop()

	client, err := NewClient(address)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	granted, err := client.Check(ctx, "user123", "create_report", "test.stack")

	assert.NoError(t, err)
	assert.True(t, granted)
	assert.Equal(t, 1, mockServer.CheckCallCount)
}

func TestClientCheck_Denied(t *testing.T) {
	mockServer := &MockPermsServer{AlwaysGrant: false}
	address, grpcServer := startMockServer(t, mockServer)
	defer grpcServer.Stop()

	client, err := NewClient(address)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	granted, err := client.Check(ctx, "", "delete_report", "")

	assert.NoError(t, err)
	assert.False(t, granted)
	assert.Equal(t, 1, mockServer.CheckCallCount)
}

func TestClientCheck_Error(t *testing.T) {
	mockServer := &MockPermsServer{ShouldFail: true}
	address, grpcServer := startMockServer(t, mockServer)
	defer grpcServer.Stop()

	client, err := NewClient(address)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	granted, err := client.Check(ctx, "user456", "some_action", "stack")

	assert.Error(t, err)
	assert.False(t, granted)
	assert.Equal(t, 1, mockServer.CheckCallCount)
}

func TestClientCheck_MultipleRequests(t *testing.T) {
	mockServer := &MockPermsServer{AlwaysGrant: true}
	address, grpcServer := startMockServer(t, mockServer)
	defer grpcServer.Stop()

	client, err := NewClient(address)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		granted, err := client.Check(ctx, "user", "action", "stack")
		assert.NoError(t, err)
		assert.True(t, granted)
	}

	assert.Equal(t, 5, mockServer.CheckCallCount)
}

func TestClientCheck_Timeout(t *testing.T) {
	mockServer := &MockPermsServer{}
	address, grpcServer := startMockServer(t, mockServer)
	defer grpcServer.Stop()

	client, err := NewClient(address)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	granted, err := client.Check(ctx, "user789", "action", "stack")

	assert.Error(t, err)
	assert.False(t, granted)
}

func TestClientCheck_WithContext(t *testing.T) {
	mockServer := &MockPermsServer{AlwaysGrant: true}
	address, grpcServer := startMockServer(t, mockServer)
	defer grpcServer.Stop()

	client, err := NewClient(address)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	granted, err := client.Check(ctx, "user999", "test_action", "test.stack")

	assert.NoError(t, err)
	assert.True(t, granted)
}

func TestClientClose(t *testing.T) {
	mockServer := &MockPermsServer{}
	address, grpcServer := startMockServer(t, mockServer)
	defer grpcServer.Stop()

	client, err := NewClient(address)
	require.NoError(t, err)

	// Single close call - should not error
	err = client.Close()
	assert.NoError(t, err)
}

func TestClientClose_NilConnection(t *testing.T) {
	client := &Client{conn: nil}
	err := client.Close()
	assert.NoError(t, err)
}
