package main

import (
	"context"
  "log"
)

func main() {
  ctx := context.Background()
  err := connectTailnet(ctx)
  if err != nil {
    log.Fatal(err)
  }
}
