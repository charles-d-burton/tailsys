package commander 

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

type Client struct {
	connections.Tailnet
	CoordinationServer string
	ID                 string
	// TLS                *connections.TLSConfig
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

func (cl *Client) WithCoordinationServer(server string) Option {
	return func(cl *Client) error {
		if server == "" {
			return errors.New(fmt.Sprintf("no coodination server configured"))
		}
		params := strings.Split(server, ":")
		if len(params) < 2 {
			return errors.New(fmt.Sprintf("server hostname %s not formatted correctly, should be <server>:<port>", server))
		}
		cl.CoordinationServer = server
		return nil
	}
}

func (cl *Client) getTlSConfig() (*connections.TLSConfig, error) {
	config := connections.TLSConfig{}
	cfile, err := os.ReadFile(cl.ConfigDir + "/certs/server-config.yaml")
	if err != nil {
		return nil, fmt.Errorf("could not find server certs at: %s with err: %w", cl.ConfigDir+"/certs/server-config.yaml", err)
	}

	err = yaml.Unmarshal(cfile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (cl *Client) getConn(ctx context.Context) (*grpc.ClientConn, error) {
	ctxTo, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	return cl.DialContext(ctxTo, cl.CoordinationServer, cl.TLSConfig)
}

func (cl *Client) SendCommand(ctx context.Context, cmd string, pattern string) error {
  for i := range 5 {

    if cl.TLSConfig == nil {
      tls, err := cl.getTlSConfig()
      if err != nil {
        return err
      }
      cl.TLSConfig = tls
    }

    command := &pb.CommanderRequest{
      Command: cmd,
      Pattern: pattern,
    }
    conn, err := cl.getConn(ctx)

    if err != nil {
      if i < 5 {
        fmt.Println(fmt.Errorf("unable to connect: %w", err))
        time.Sleep(5 * time.Second)
        fmt.Println("retrying connection")
        continue
      }
      return err
    }
    cc := pb.NewCommandManagerClient(conn)
    r, err := cc.SendCommandToNodes(ctx, command)
    if err != nil {
      if i < 5 {
        fmt.Println(fmt.Errorf("unable to send command: %w", err))
        time.Sleep(5 * time.Second)
        fmt.Println("retrying sending command")
        continue
      }
      return err
    }
    for _, res := range r.Response {
      fmt.Printf("command: %s ran with exit code %d\n", cmd, res.ExitCode)
      fmt.Printf("  result: %s\n", string(res.GetOutput()))
    }
    return nil
  }
	return nil
}

func (cl *Client) GetNodes(ctx context.Context, pattern string) error {
  for i := range 5 {

    if cl.TLSConfig == nil {
      tls, err := cl.getTlSConfig()
      if err != nil {
        return err
      }
      cl.TLSConfig = tls
    }
    gn := &pb.NodeQuery{
      Pattern: pattern,
    }

    conn, err := cl.getConn(ctx)
    if err != nil {
      if i < 5 {
        fmt.Println(fmt.Errorf("unable to connect: %w", err))
        time.Sleep(5 * time.Second)
        fmt.Println("retrying connection")
        continue
      }
      return err
    }

    cc := pb.NewCommandManagerClient(conn)
    r, err := cc.GetNodes(ctx, gn)
    if err != nil {
      if i < 5 {
        fmt.Println(fmt.Errorf("unable to get nodes: %w", err))
        time.Sleep(5 * time.Second)
        fmt.Println("retrying getting nodes")
        continue
      }
      return err
    }
    for _, res := range r.Nodes {
      fmt.Printf("found node %s\n", res)
    }
    return nil
  }
	return nil
}
