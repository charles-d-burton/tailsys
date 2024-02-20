package main

import (
	"context"
  "log"
)

func main() {
  ctx := context.Background()
  err := startCLI(ctx)
  if err != nil {
    log.Fatal(err)
  }
}
