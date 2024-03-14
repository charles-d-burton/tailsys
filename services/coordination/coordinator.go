package coordination

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/google/uuid"
)

const (
	registrationBucket = "node-registration"
)

// Coordinator holds the runtime variables for the coordination server
type Coordinator struct {
	connections.Tailnet
	services.DataManagement
	devMode bool
	ID      string
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

// StartRPCCoordinationServer Register the gRPC server endpoints and start the server
func (co *Coordinator) StartRPCCoordinationServer(ctx context.Context) error {
	pb.RegisterPingerServer(co.GRPCServer, &services.Pinger{})
	pb.RegisterRegistrationServer(co.GRPCServer, &RegistrationServer{
		DevMode: co.devMode,
		DB:      co.DB,
		//TODO: This is randomized on startup, we should persist and load
		ID: uuid.NewString(),
	})

	fmt.Println("rpc server starting to serve traffic")
	return co.GRPCServer.Serve(co.Listener)
}
