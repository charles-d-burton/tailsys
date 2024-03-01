package main

import (
	"github.com/charles-d-burton/tailsys/cmd"
	"log"
)

func main() {
	err := cmd.StartCLI()
	if err != nil {
		log.Fatal(err)
	}
}
