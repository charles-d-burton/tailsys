package services

import (
	"context"
  "fmt"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/nutsdb/nutsdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Pinger struct {
	pb.UnimplementedPingerServer
	DB *nutsdb.DB
	ID string
}

// Ping GRPC service for the service to ping clients and provide response time
func (p *Pinger) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PongResponse, error) {
  fmt.Println("received ping request")
	now := time.Now()
	latency := now.Sub(in.Ping.AsTime())

	return &pb.PongResponse{
		Ping:           timestamppb.New(now),
		InboundLatency: float32(latency.Milliseconds()),
	}, nil
}

