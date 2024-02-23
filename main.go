package main

import (
	"context"
	"log"
  "github.com/charles-d-burton/tailsys/config"
)

func main() {
	ctx := context.Background()
	err := config.StartCLI(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
