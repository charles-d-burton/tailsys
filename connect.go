package main

import (
	"context"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"tailscale.com/tsnet"
)

func connectOauth(ctx context.Context, id, secret string) (*Tailnet, error) {
	var tn  Tailnet
	srv, err := tn.NewConnection(ctx,
		tn.WithOauth(id, secret),
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
	)

	if err != nil {
    return nil, err
	}
  return &tn, connectTailnet(ctx, srv)
}

func connectAuthKey(ctx context.Context, authKey string) (*Tailnet, error) {
	var tn  Tailnet
	srv, err := tn.NewConnection(ctx,
    tn.WithAuthKey(authKey),
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
	)

	if err != nil {
    return nil, err
	}
  return &tn, connectTailnet(ctx, srv)
}

type pingerGRPCServer struct {
  pb.UnimplementedPingerServer
}

func (p *pingerGRPCServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PongResponse, error) {
  time := time.Now()
  latency := time.Sub(in.Ping.AsTime())

  return &pb.PongResponse{
    Ping: timestamppb.New(time), 
    InboundLatency: float32(latency.Milliseconds()),
  }, nil
}

func connectTailnet(ctx context.Context, srv *tsnet.Server) error {

  // devices, err := tn.GetDevices(ctx)
  // if err != nil {
  //   return(err)
  // }
  // for _, device := range devices {
  //   fmt.Println(device)
  // }

	if err := srv.Start(); err != nil {
    return(err)
		// log.Fatalf("can't start tsnet server: %v", err)
	}

	ln, err := srv.Listen("tcp", ":80")
	if err != nil {
		return(err)
	}

  s := grpc.NewServer()
  pb.RegisterPingerServer(s, &pingerGRPCServer{})
  if err := s.Serve(ln); err != nil {
    return err
  }
	// log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintln(w, "Hi there! Welcome to the tailnet!")
	// })))
  return nil
}

