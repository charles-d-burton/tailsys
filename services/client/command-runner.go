package client

import (
	"context"

	pb "github.com/charles-d-burton/tailsys/commands"
)

// CommandServer struct to contain command runner rpc
type CommandServer struct {
	pb.UnimplementedCommandRunnerServer
}

// RegisterCommandRunner registers the RPC call and implements behavior for command runner
func (c *CommandServer) RegisterCommandRunner(ctx context.Context, in *pb.CommandRequest) (*pb.CommandResonse, error) {
	return &pb.CommandResonse{}, nil
}
