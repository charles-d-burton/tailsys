package main

import (
	"context"

  pb "github.com/charles-d-burton/tailsys/commands"
  "google.golang.org/grpc"
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

	// log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintln(w, "Hi there! Welcome to the tailnet!")
	// })))
  return nil
}

