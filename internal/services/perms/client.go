package perms

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/xligenda/reports/internal/services/perms/pb"
)

// Client is a gRPC client for the Perms service.
type Client struct {
	conn   *grpc.ClientConn
	client pb.PermsClient
}

// NewClient creates a new Perms gRPC client.
func NewClient(target string) (*Client, error) {
	// Create connection with insecure credentials (no TLS for now)
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", target, err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewPermsClient(conn),
	}, nil
}

// Check verifies if access should be granted for a given request.
// Parameters:
//   - ctx: context for the RPC call
//   - id: user or entity ID requesting access
//   - action: the action being requested (e.g., "create_report", "delete_report")
//   - stack: optional call stack for debugging/audit trail
//
// Returns true if access is granted, false otherwise, or an error if the RPC fails.
func (c *Client) Check(ctx context.Context, id string, action Permission, stack string) (bool, error) {
	// Create a deadline context if none exists
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	request := &pb.AccessRequest{
		Id:              id,
		RequestedAction: action.String(),
		Stack:           stack,
	}

	response, err := c.client.Check(ctx, request)
	if err != nil {
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}

	return response.Granted, nil
}

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
