package main

import (
	"log"
  "github.com/charles-d-burton/tailsys/config"
)

func main() {
	err := config.StartCLI()
	if err != nil {
		log.Fatal(err)
	}
}
