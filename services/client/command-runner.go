package client

import (
	pb "github.com/charles-d-burton/tailsys/commands"
	"golang.org/x/net/context"
)

//CommandServer struct to contain command runner rpc
type CommandServer struct {
  pb.UnimplementedCommandRunnerServer
}

//RegisterCommandRunner registers the RPC call and implements behavior for command runner
func (c *CommandServer) RegisterCommandRunner(ctx context.Context, in *pb.CommandRequest) (*pb.CommandResonse, error) {
  return &pb.CommandResonse{}, nil
}


