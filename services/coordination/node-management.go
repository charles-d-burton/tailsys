package coordination

import (
	"context"
	"fmt"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/google/uuid"
)

type Coordinator struct {
	connections.Tailnet
  devMode bool
}

type Option func(co *Coordinator) error


func (co *Coordinator) NewCoordinator(ctx context.Context, opts ...Option) error {

  for _, opt := range opts {
    err := opt(co)
    if err != nil {
      return err
    }
  }
	return nil
}

func (co *Coordinator) WithDevMode(mode bool) Option {
  return func(co *Coordinator) error {
    co.devMode = mode
    return nil
  }
}

func (co *Coordinator) StartRPCCoordinationServer(ctx context.Context) error {
	pb.RegisterPingerServer(co.GRPCServer, &services.Pinger{})
	pb.RegisterRegistrationServer(co.GRPCServer, &RegistrationServer{DevMode: co.devMode})
	return co.GRPCServer.Serve(co.Listener)
}

// RegistrationServer struct to contain proto for gRPC
type RegistrationServer struct {
	pb.UnimplementedRegistrationServer
  DevMode bool
  ID string
}

// RegisterNode send the key to the server to register a node
func (p *RegistrationServer) Register(ctx context.Context, in *pb.NodeRegistrationRequest) (*pb.NodeRegistrationResponse, error) {
	fmt.Println("received coordination request")
	fmt.Println(in)
  if p.DevMode {
    id := uuid.New()
    fmt.Println("running in dev mode, accepting all incoming connections")
    return &pb.NodeRegistrationResponse{
      Accepted: true,
      Key:      &pb.Key{Key: id.String()},
    }, nil
  }
	return &pb.NodeRegistrationResponse{
		Accepted: true,
		Key:      &pb.Key{Key: "coordination-server-key"},
	}, nil
}
