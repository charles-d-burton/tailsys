package coordination

import (
	"context"

	pb "github.com/charles-d-burton/tailsys/commands"
)

//RegistrationServer struct to contain proto for gRPC
type RegistrationServer struct {
	pb.UnimplementedRegistrationServer
}

//RegisterNode send the key to the server to register a node
func (p *RegistrationServer) RegisterNode(ctx context.Context, in *pb.NodeRegistrationRequest) (*pb.NodeRegistrationResponse, error) {
	return &pb.NodeRegistrationResponse{
  
  }, nil
}
