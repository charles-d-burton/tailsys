package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	id := os.Getenv("TS_CLIENT_ID")
	secret := os.Getenv("TS_CLIENT_SECRET")

	ctx := context.Background()
	var li Listener
	srv, err := li.NewConnection(ctx,
		li.WithOauth(id, secret),
		li.WithScopes("devices", "logs:read", "routes:read"),
		li.WithTags("tailsys"),
	)

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
