package services

import (
	"context"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Pinger struct {
	pb.UnimplementedPingerServer
}

// Ping GRPC service for the service to ping clients and provide response time
func (p *Pinger) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PongResponse, error) {
	time := time.Now()
	latency := time.Sub(in.Ping.AsTime())

	return &pb.PongResponse{
		Ping:           timestamppb.New(time),
		InboundLatency: float32(latency.Milliseconds()),
	}, nil
}
