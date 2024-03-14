package coordination

import (
	"context"
	"errors"
	"fmt"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/services"
)

const (
	registrationBucket = "node-registration"
)

// Coordinator holds the runtime variables for the coordination server
type Coordinator struct {
	connections.Tailnet
	services.DataManagement
	devMode bool
}

// Options defines the configuration options function for configuration injection
type Option func(co *Coordinator) error

// NewCoordinator Create a new coordinator instance and set the provided options
func (co *Coordinator) NewCoordinator(ctx context.Context, opts ...Option) error {

	for _, opt := range opts {
		err := opt(co)
		if err != nil {
			return err
		}
	}

	if co.DB == nil {
		return errors.New("datastore not initialized")
	}
	return nil
}

// WithDevMode enable the server to run in dev mode
func (co *Coordinator) WithDevMode(mode bool) Option {
	return func(co *Coordinator) error {
		fmt.Println("setting dev mode to: ", mode)
		co.devMode = mode
		return nil
	}
}

func (co *Coordinator) WithDataDir(dir string) Option {
	return func(co *Coordinator) error {
		return co.StartDB(dir)
	}
}
