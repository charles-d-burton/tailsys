package client

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/google/uuid"
)

type Client struct {
	services.DataManagement
	connections.Tailnet
	ID string
}

type Option func(cl *Client) error

func (cl *Client) NewClient(ctx context.Context, opts ...Option) error {
	for _, opt := range opts {
		err := opt(cl)
		if err != nil {
			return err
		}
	}

	cl.ID = uuid.NewString()
	return nil
}

func (cl *Client) StartDatabase(ctx context.Context) error {
	return cl.StartDB(cl.ConfigDir)
}

func (cl *Client) StartRPCClientMode(ctx context.Context) error {
	fmt.Println("starting grpc client server")
	if cl.DB == nil {
		return errors.New("datastore not initialized")
	}

	pb.RegisterPingerServer(cl.GRPCServer, &services.Pinger{
		DB: cl.DB,
		ID: cl.ID,
	})

	pb.RegisterCommandRunnerServer(cl.GRPCServer, &CommandServer{})

	return cl.GRPCServer.Serve(cl.Listener)
}
