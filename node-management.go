package main

import (
	"context"
  "time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
	return &pb.NodeRegistrationResponse{
  
  }, nil
}
