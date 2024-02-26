package coordination

import (
	"context"
	"fmt"

	pb "github.com/charles-d-burton/tailsys/commands"
)

//RegistrationServer struct to contain proto for gRPC
type RegistrationServer struct {
	pb.UnimplementedRegistrationServer
}

//RegisterNode send the key to the server to register a node
func (p *RegistrationServer) Register(ctx context.Context, in *pb.NodeRegistrationRequest) (*pb.NodeRegistrationResponse, error) {
  fmt.Println("received coordination request")
  fmt.Println(in)
	return &pb.NodeRegistrationResponse{
    Accepted: true,  
    Key: &pb.Key{Key: "coordination-server-key"},
  }, nil
}

