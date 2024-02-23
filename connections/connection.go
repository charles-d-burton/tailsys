package connections

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tailscale/tailscale-client-go/tailscale"
	"google.golang.org/grpc"
	"tailscale.com/tsnet"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/charles-d-burton/tailsys/services/client"
	"github.com/charles-d-burton/tailsys/services/coordination"
)

// Tailnet main struct to hold connection to the tailnet information
type Tailnet struct {
	ClientID     string
	ClientSecret string
	AuthKey      string
	Hostname     string
  Addr string
	Scopes       []string
	Tags         []string
	Client       *tailscale.Client
  GRPCServer *grpc.Server
  listener net.Listener
}

// Option function to set different options on the tailnet config
type Option func(tn *Tailnet) error

// NewConnection setup a connection to the tailnet
func (tn *Tailnet) NewConnection(ctx context.Context, opts ...Option) (*tsnet.Server, error) {
	for _, opt := range opts {
		err := opt(tn)
		if err != nil {
			return nil, err
		}
	}

	err := tn.initClient(ctx)
	if err != nil {
		return nil, err
	}

	srv := &tsnet.Server{
		Hostname:  tn.Hostname,
		AuthKey:   tn.AuthKey,
		Ephemeral: true,
	}

	return srv, nil
}

func (tn *Tailnet) initClient(ctx context.Context) error {
	var capabilities tailscale.KeyCapabilities
	capabilities.Devices.Create.Reusable = true
	capabilities.Devices.Create.Ephemeral = true
	capabilities.Devices.Create.Tags = tn.Tags
	capabilities.Devices.Create.Preauthorized = true

	var topts []tailscale.CreateKeyOption
	topts = append(topts, tailscale.WithKeyExpiry(10*time.Second))

	if useOauth(tn.ClientID, tn.ClientSecret) {
		client, err := tailscale.NewClient(
			"",
			"-",
			tailscale.WithOAuthClientCredentials(tn.ClientID, tn.ClientSecret, tn.Scopes),
		)
		if err != nil {
			return err
		}
		key, err := client.CreateKey(ctx, capabilities, topts...)
		if err != nil {
			return err
		}
		tn.AuthKey = key.Key
		tn.Client = client
		return nil
	}

	if tn.AuthKey == "" {
		return errors.New("must set one of oauth keys or api key")
	}

	client, err := tailscale.NewClient(tn.AuthKey, "-")
	if err != nil {
		return err
	}
	tn.Client = client

	return nil
}

// GetDevices returns a list of devices that are conected to the configured tailnet
func (tn *Tailnet) GetDevices(ctx context.Context) ([]tailscale.Device, error) {
	return tn.Client.Devices(ctx)
}

// WithOauth sets up the tailnet connection using an oauth credential
func (tn *Tailnet) WithOauth(clientId, clientSecret string) Option {
	return func(tn *Tailnet) error {
		if clientId == "" {
			return errors.New("client id not set")
		}

		if clientSecret == "" {
			return errors.New("client secret not set")
		}
		tn.ClientID = clientId
		tn.ClientSecret = clientSecret
		return nil
	}
}

// WithAPIKey sets the Option to connect to the tailnet with a preconfigured Auth key
func (tn *Tailnet) WithAuthKey(key string) Option {
	return func(tn *Tailnet) error {
		tn.AuthKey = key
		return nil
	}
}

// WithScopes sets the Oauth scopes to configure for the connection
func (tn *Tailnet) WithScopes(scopes ...string) Option {
	return func(tn *Tailnet) error {
		if scopes != nil {
			tn.Scopes = scopes
		}
		return nil
	}
}

// WithTags sets the tags that were configured with the oauth connection
func (tn *Tailnet) WithTags(tags ...string) Option {
	return func(tn *Tailnet) error {
		if tags != nil {
			for _, tag := range tags {
				stag := strings.Split(tag, ":")
				if len(stag) < 2 {
					return errors.New(fmt.Sprintf("tag %s mush be in format tag:<tag>", tag))
				}
			}
			tn.Tags = tags
		}
		return nil
	}
}

func (tn *Tailnet) WithHostname(hostname string) Option {
  return func(tn *Tailnet) error {
    if hostname == "" {
      hostname, err := os.Hostname()
      if err != nil {
        return err
      }
      tn.Hostname = hostname + "-tailsys"
      return nil
    }
    tn.Hostname = hostname
    return nil
  }
}

func (tn *Tailnet) createRPCServer(ctx context.Context, srv *tsnet.Server) error {

	// devices, err := tn.GetDevices(ctx)
	// if err != nil {
	//   return(err)
	// }
	// for _, device := range devices {
	//   fmt.Println(device)
	// }

	if err := srv.Start(); err != nil {
		return err
	}

	ln, err := srv.Listen("tcp", ":6655")
	if err != nil {
		return err
	}
  tn.Addr = ln.Addr().String()

	s := grpc.NewServer()

  tn.listener = ln
  tn.GRPCServer = s

	return nil
}

func (tn *Tailnet) StartRPCClientMode(ctx context.Context) error {
  pb.RegisterPingerServer(tn.GRPCServer, &services.Pinger{})
  pb.RegisterCommandRunnerServer(tn.GRPCServer, &client.CommandServer{})
  return tn.GRPCServer.Serve(tn.listener)

}

func (tn *Tailnet) StartRPCCoordinationServer(ctx context.Context) error {
  pb.RegisterPingerServer(tn.GRPCServer, &services.Pinger{})
  pb.RegisterRegistrationServer(tn.GRPCServer, &coordination.RegistrationServer{})
  return tn.GRPCServer.Serve(tn.listener)
}


func useOauth(clientId, clientSecret string) bool {
	if clientId == "" || clientSecret == "" {
		return false
	}
	return true
}

func useAPIKey(key string) bool {
	if key == "" {
		return false
	}
	return true
}
