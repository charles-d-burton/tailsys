package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	connections.Tailnet
}

func NewClient(ctx context.Context) (*Client, error) {
	return &Client{}, nil
}

func (cli *Client) StartRPCClientMode(ctx context.Context) error {
	pb.RegisterPingerServer(cli.GRPCServer, &services.Pinger{})
	pb.RegisterCommandRunnerServer(cli.GRPCServer, &CommandServer{})
	return cli.GRPCServer.Serve(cli.Listener)

}

func (cl *Client) RegisterWithCoordinationServer(ctx context.Context, addr string) error {
	for i := 0; i < 5; i++ {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		defer conn.Close()
		c := pb.NewRegistrationClient(conn)

		r, err := c.Register(ctx, &pb.NodeRegistrationRequest{
			Info: &pb.SysInfo{
				Hostname: cl.Hostname,
				Type:     pb.OSType_LINUX,
				Ip:       cl.Addr,
				LastSeen: timestamppb.Now(),
			},
			Key:        &pb.Key{Key: "randomstring"},
			SystemType: pb.SystemType_CLIENT,
		})

		if err != nil {
			fmt.Println(err)
			time.Sleep(3 * time.Second)
			if i == 4 {
				return errors.New(fmt.Sprintf("unable to connect to coordation server: %s", addr))
			}
			continue
		}
		fmt.Println(r)
		break
	}

	return nil
}

// CommandServer struct to contain command runner rpc
type CommandServer struct {
	pb.UnimplementedCommandRunnerServer
}

// RegisterCommandRunner registers the RPC call and implements behavior for command runner
func (c *CommandServer) RegisterCommandRunner(ctx context.Context, in *pb.CommandRequest) (*pb.CommandResonse, error) {
	return &pb.CommandResonse{}, nil
}
