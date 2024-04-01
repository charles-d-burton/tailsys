package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/data/queries"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// RegisterWithCoordinationServer generate the registration request and send it to the coordination server
func (cl *Client) RegisterWithCoordinationServer(ctx context.Context, addr string) error {
	for i := 0; i < 5; i++ {
		fmt.Println("coordination server address: ", addr)
		ctxTo, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()
		sconfig, err := cl.getTlSConfig()
		if err != nil {
			return err
		}

		conn, err := cl.DialContext(ctxTo, addr, sconfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		c := pb.NewRegistrationClient(conn)

		fmt.Println("attempting to send registration request")
		fmt.Println(conn.GetState())

		req := &pb.NodeRegistrationRequest{
			Info: &pb.SysInfo{
				Hostname: cl.Hostname,
				Port:     cl.Port,
				Type:     pb.OSType_LINUX,
				Ip:       cl.Hostname,
				LastSeen: timestamppb.Now(),
			},
			Key:        &pb.Key{Key: cl.ID},
			SystemType: pb.SystemType_CLIENT,
			Tlskey:     cl.TLSConfig.TLSKey,
			Tlscert:    cl.TLSConfig.TLSCert,
		}
		fmt.Println(req)
		r, err := c.Register(ctx, req)

		if err != nil {
			fmt.Println(err)
			time.Sleep(3 * time.Second)
			if i == 4 {
				return errors.New(fmt.Sprintf("unable to connect to coordation server: %s", addr))
			}
			continue
		}

		fmt.Println("registering response")
		err = cl.addRegistration(r)
		if err != nil {
			return err
		}
		fmt.Println(r)
		break
	}
	return nil
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

func (cl *Client) addRegistration(r *pb.NodeRegistrationResponse) error {
	fmt.Println("registering server")
	err := queries.SetRegisteredCoordinationServer(cl.DB, r.GetHostname(), r.GetKey().Key)
	if err != nil {
		fmt.Println("could not register server")
		return err
	}
	return nil
}
