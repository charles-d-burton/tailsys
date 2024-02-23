package main

import (
	"context"

	pb "github.com/charles-d-burton/tailsys/commands"
	"google.golang.org/grpc"
	"tailscale.com/tsnet")

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
  err = tn.connectGRPCTailnet(ctx, srv)
  if err != nil {
    return nil, err
  }

	return &tn, nil
}

//connect to the tailnet using a pre-generated auth-key
func connectAuthKey(ctx context.Context, authKey string) (*Tailnet, error) {
	var tn Tailnet
	srv, err := tn.NewConnection(ctx,
		tn.WithAuthKey(authKey),

    //TODO: THise needs to be parameterized in the config
		tn.WithScopes("devices", "logs:read", "routes:read"),
		tn.WithTags("tag:tailsys"),
	)

	if err != nil {
		return nil, err
	}

  err = tn.connectGRPCTailnet(ctx, srv)
  if err != nil {
    return nil, err
  }
	return &tn, nil
}


func (tn *Tailnet)connectGRPCTailnet(ctx context.Context, srv *tsnet.Server) error {

	// devices, err := tn.GetDevices(ctx)
	// if err != nil {
	//   return(err)
	// }
	// for _, device := range devices {
	//   fmt.Println(device)
	// }

	if err := srv.Start(); err != nil {
		return err
	}

	ln, err := srv.Listen("tcp", ":6655")
	if err != nil {
		return err
	}
  tn.Addr = ln.Addr().String()

	s := grpc.NewServer()

	pb.RegisterPingerServer(s, &pingerGRPCServer{})
	pb.RegisterRegistrationServer(s, &registrationServer{})

	if err := s.Serve(ln); err != nil {
		return err
	}
  tn.GRPCServer = s

	return nil
}
