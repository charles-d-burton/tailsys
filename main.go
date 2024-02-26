package main

import (
	"log"
  "github.com/charles-d-burton/tailsys/cmd"
)

func main() {
	err := cmd.StartCLI()
	if err != nil {
		log.Fatal(err)
	}
}
