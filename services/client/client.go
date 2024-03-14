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

const coordinationBucket = "coordination-server"

type Client struct {
	connections.Tailnet
	services.DataManagement
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

	if cl.DB == nil {
		return errors.New("datastore not initialized")
	}
	cl.ID = uuid.NewString()
	return nil
}

func (cl *Client) WithDataDir(dir string) Option {
	return func(cl *Client) error {
		return cl.StartDB(dir)
	}
}

func (cli *Client) StartRPCClientMode(ctx context.Context) error {
	fmt.Println("starting grpc client server")
	pb.RegisterPingerServer(cli.GRPCServer, &services.Pinger{
		DB: cli.DB,
		ID: cli.ID,
	})
	pb.RegisterCommandRunnerServer(cli.GRPCServer, &CommandServer{})
	return cli.GRPCServer.Serve(cli.Listener)
}
