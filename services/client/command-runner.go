package client

import (
	"context"
	"fmt"

	pb "github.com/charles-d-burton/tailsys/commands"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CommandServer struct to contain command runner rpc
type CommandServer struct {
	pb.UnimplementedCommandRunnerServer
}

// RegisterCommandRunner registers the RPC call and implements behavior for command runner
func (c *CommandServer) RegisterCommandRunner(ctx context.Context, in *pb.CommandRequest) (*pb.CommandResponse, error) {
  fmt.Printf("running command: %s\n", in.Command) 
	return &pb.CommandResponse{
    Timestamp: timestamppb.Now(),
    Successful: true,
    Output: []byte(fmt.Sprintf("ran command %s\n", in.Command)),
    ExitCode: 0,
  }, nil
}
