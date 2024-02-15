package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tailscale/tailscale-client-go/tailscale"
	"tailscale.com/tsnet"
)

func main() {
	oauthClientId := os.Getenv("TS_CLIENT_ID")
	oauthClientSecret := os.Getenv("TS_CLIENT_SECRET")

	client, err := tailscale.NewClient(
		"",
		"-",
		tailscale.WithOAuthClientCredentials(oauthClientId, oauthClientSecret, []string{"devices", "logs:read", "routes:read"}),
	)
	if err != nil {
		panic(err)
	}

	var capabilities tailscale.KeyCapabilities
	capabilities.Devices.Create.Reusable = false
	capabilities.Devices.Create.Ephemeral = true
	capabilities.Devices.Create.Tags = []string{"tag:tailsys"}
	capabilities.Devices.Create.Preauthorized = true

	var opts []tailscale.CreateKeyOption
	opts = append(opts, tailscale.WithKeyExpiry(10*time.Second))

	key, err := client.CreateKey(context.Background(), capabilities, opts...)
	if err != nil {
		panic(err)
	}
	log.Printf("authKey: %s", key.Key)

	srv := &tsnet.Server{
		Hostname:  "tailsys-test",
		AuthKey:   key.Key,
		Ephemeral: true,
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("can't start tsnet server: %v", err)
	}

	ln, err := srv.Listen("tcp", ":80")
	if err != nil {
		panic(err)
	}

	log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hi there! Welcome to the tailnet!")
	})))
}
