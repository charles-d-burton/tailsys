package client

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	pb "github.com/charles-d-burton/tailsys/commands"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CommandServer struct to contain command runner rpc
type CommandServer struct {
	pb.UnimplementedCommandRunnerServer
}

// RegisterCommandRunner registers the RPC call and implements behavior for command runner
func (c *CommandServer) Command(ctx context.Context, in *pb.CommandRequest) (*pb.CommandResponse, error) {
	fmt.Printf("running command: %s\n", in.Command)
  cnds := strings.Fields(in.Command)
  cmdo := exec.Command(cnds[0], cnds[1:]...)
  out, err := cmdo.Output()
  fmt.Println(string(out))
  if err != nil {
    return &pb.CommandResponse{
      Timestamp: timestamppb.Now(),
      Successful: false,
      Output: []byte(err.Error()),
      ExitCode: 1,
    }, nil
  }
	return &pb.CommandResponse{
		Timestamp:  timestamppb.Now(),
		Successful: true,
		Output:     out,
		ExitCode:   0,
	}, nil
}
