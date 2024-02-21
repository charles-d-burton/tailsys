package main

import (
	"context"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"tailscale.com/tsnet"
)

//connect to the tailnet using oauth credentials
func connectOauth(ctx context.Context, id, secret string) (*Tailnet, error) {
	var tn Tailnet
	srv, err := tn.NewConnection(ctx,
		tn.WithOauth(id, secret),
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
	)

	if err != nil {
		return nil, err
	}
  grpcServer, err := connectGRPCTailnet(ctx, srv)
  if err != nil {
    return nil, err
  }
  tn.GRPCServer = grpcServer

	return &tn, nil
}

//connect to the tailnet using a pre-generated auth-key
func connectAuthKey(ctx context.Context, authKey string) (*Tailnet, error) {
	var tn Tailnet
	srv, err := tn.NewConnection(ctx,
		tn.WithAuthKey(authKey),
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
	)

	if err != nil {
		return nil, err
	}

  grpcServer, err := connectGRPCTailnet(ctx, srv)
  if err != nil {
    return nil, err
  }
  tn.GRPCServer = grpcServer
	return &tn, nil
}

type pingerGRPCServer struct {
	pb.UnimplementedPingerServer
}

//Ping GRPC service for the service to ping clients and provide response time
func (p *pingerGRPCServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PongResponse, error) {
	time := time.Now()
	latency := time.Sub(in.Ping.AsTime())

	return &pb.PongResponse{
		Ping:           timestamppb.New(time),
		InboundLatency: float32(latency.Milliseconds()),
	}, nil
}

type registrationServer struct {
	pb.UnimplementedRegistrationServer
}

//RegisterNode send the key to the server to register a node
func (p *registrationServer) RegisterNode(ctx context.Context, in *pb.NodeRegistrationRequest) (*pb.NodeRegistrationResponse, error) {
	return &pb.NodeRegistrationResponse{}, nil
}

func connectGRPCTailnet(ctx context.Context, srv *tsnet.Server) (*grpc.Server, error) {

	// devices, err := tn.GetDevices(ctx)
	// if err != nil {
	//   return(err)
	// }
	// for _, device := range devices {
	//   fmt.Println(device)
	// }

	if err := srv.Start(); err != nil {
		return nil, err
	}

	ln, err := srv.Listen("tcp", ":6655")
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	pb.RegisterPingerServer(s, &pingerGRPCServer{})
	pb.RegisterRegistrationServer(s, &registrationServer{})
	if err := s.Serve(ln); err != nil {
		return nil, err
	}
	return s, nil
}
