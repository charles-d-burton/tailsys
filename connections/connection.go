package connections

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tailscale/tailscale-client-go/tailscale"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v3"
	"tailscale.com/tsnet"
)

// Enum type for Auth Type
type AuthType int

// Enum definition for Auth Type
const (
	OAUTH AuthType = iota
	AUTHKEY
	NONE
)

// Tailnet main struct to hold connection to the tailnet information
type Tailnet struct {
  ConfigDir string
	ClientID       string
	ClientSecret   string
	AuthKey        string
	Hostname       string
	Addr           string
	Port           string
  TLSConfig TLSConfig
	Scopes         []string
	Tags           []string
	Client         *tailscale.Client
	TSServer       *tsnet.Server
	GRPCServer     *grpc.Server
	Listener       net.Listener
	TailnetLogging bool
	authType       AuthType
}

type TLSConfig struct {
  TLSKey         string `yaml:"key"`
  TLSCert        string `yaml:"cert"`
}

// Option function to set different options on the tailnet config
type Option func(tn *Tailnet) error

// connect to the tailnet using oauth credentials
func (tn *Tailnet) Connect(ctx context.Context, opts ...Option) error {
	for _, opt := range opts {
		err := opt(tn)
		if err != nil {
			return err
		}
	}

  if err := tn.generateKeyPair(); err != nil {
    return err
  }

	if tn.Hostname == "" {
		h, err := os.Hostname()
		if err != nil {
			return err
		}
		tn.Hostname = h + "-tailsys"
	}

	err := tn.initClient(ctx)
	if err != nil {
		return err
	}

	srv := &tsnet.Server{
		Hostname:  tn.Hostname,
		AuthKey:   tn.AuthKey,
		Ephemeral: true,
		//TODO: this is erroring, I think it's a bug on the tailscale side
		//    Logger: func(string, ...any) {},
	}
	tn.TSServer = srv

	err = tn.createRPCServer()
	if err != nil {
		return err
	}

	return nil
}

func (tn *Tailnet) DialContext(ctx context.Context, addr string, certs *TLSConfig) (*grpc.ClientConn, error) {
  //Plain dialer if not on tsnet
  //Pass in a cancelable context
  pool := x509.NewCertPool()
  pool.AppendCertsFromPEM([]byte(certs.TLSCert))
  pair, err := tls.X509KeyPair([]byte(certs.TLSCert), []byte(certs.TLSKey))
  if err != nil {
    return nil, err
  }

  tc := credentials.NewTLS(&tls.Config{
    Certificates: []tls.Certificate{pair},
    ClientAuth: tls.RequireAndVerifyClientCert,
    ClientCAs: pool,
    RootCAs: pool,
  })

  if tn.authType == NONE {
    conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(tc))
    return conn, err
  }

	return grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(tc),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return tn.TSServer.Dial(ctx, "-", addr)
		}),
	)
}

func (tn *Tailnet) initClient(ctx context.Context) error {
	var capabilities tailscale.KeyCapabilities
	capabilities.Devices.Create.Reusable = true
	capabilities.Devices.Create.Ephemeral = true
	capabilities.Devices.Create.Tags = tn.Tags
	capabilities.Devices.Create.Preauthorized = true

	var topts []tailscale.CreateKeyOption
	topts = append(topts, tailscale.WithKeyExpiry(10*time.Second))

	if tn.authType == OAUTH {
		fmt.Println("connecting with oauth")
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
		tn.reapDeviceID(ctx)
		return nil
	} else if tn.authType == AUTHKEY {
		fmt.Println("connecting with authkey")
		client, err := tailscale.NewClient(tn.AuthKey, "-")
		if err != nil {
			return err
		}
		tn.Client = client
		tn.reapDeviceID(ctx)
	}
	return nil
}

func (tn *Tailnet) reapDeviceID(ctx context.Context) error {
	devices, err := tn.Client.Devices(ctx)
	fmt.Printf("FOUND %d DEVICES\n", len(devices))
	if err != nil {
		return err
	}
	for _, device := range devices {
		fmt.Println(device.Hostname)
		if device.Hostname == tn.Hostname {
			fmt.Printf("found device %s with same name %s\n\n", device.Hostname, tn.Hostname)
			err := tn.Client.DeleteDevice(ctx, device.ID)
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (tn *Tailnet) generateKeyPair() error {
  if !tn.checkForKeys() {
    fmt.Println("no keys found, generating new keys")
    priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
      return err
    }
    privDer, err := x509.MarshalPKCS8PrivateKey(priv)
    if err != nil {
      return err
    }
    privPem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDer})

    template := &x509.Certificate{
      SerialNumber: new(big.Int),
      NotAfter: time.Now().Add(time.Hour * 87660), //Ten years
      DNSNames: []string{tn.Hostname},
    }

    certDer, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
    if err != nil {
      return err
    }
    certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})
    tc := TLSConfig{}
    tc.TLSCert = string(certPem)
    tc.TLSKey = string(privPem)
    
    d, err := yaml.Marshal(&tc)
    if err != nil {
      return err
    }
    err = os.MkdirAll(tn.ConfigDir + "/certs", 0750)
    err = os.WriteFile(tn.ConfigDir + "/certs/certs.yaml", d, 0640)
    if err != nil {
      return err
    }
    tn.TLSConfig = tc 

  }
  return nil
}

func (tn *Tailnet) checkForKeys() bool {
  tc := TLSConfig{}
  c, err := os.ReadFile(tn.ConfigDir + "/certs/certs.yaml")
  if err != nil {
    return false
  }
  err = yaml.Unmarshal(c, &tc)
  if err != nil {
    return false
  }
  tn.TLSConfig = tc
  return true
}

// GetDevices returns a list of devices that are connected to the configured tailnet
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

//WithHostname Override the hostname on the tailnet
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

//WithTailnetLogging Enable/Disable logging on the tailnet
func (tn *Tailnet) WithTailnetLogging(enabled bool) Option {
	return func(tn *Tailnet) error {
		tn.TailnetLogging = enabled
		return nil
	}
}

//WithPort Port to bind the grpc server to
func (tn *Tailnet) WithPort(port string) Option {
	return func(tn *Tailnet) error {
		tn.Port = port
		return nil
	}
}

func (tn *Tailnet) WithConfigDir(dir string) Option {
  return func(tn *Tailnet) error {
    tn.ConfigDir = dir
    return nil
  }
}

//createRPCServer create and start the gRPC server
func (tn *Tailnet) createRPCServer() error {

	if tn.authType != NONE {
		if err := tn.TSServer.Start(); err != nil {
			return err
		}
		if tn.Port == "" {
			tn.Port = "6655"
		}
		//TODO: I need to pass a listener port to this
		ln, err := tn.TSServer.Listen("tcp", ":"+tn.Port)
		if err != nil {
			return err
		}
		tn.Addr = ln.Addr().String()
		tn.Listener = ln
	} else {
		ln, err := net.Listen("tcp", ":"+tn.Port)
		if err != nil {
			return err
		}
		tn.Addr = ln.Addr().String()
		tn.Listener = ln
	}

  pool := x509.NewCertPool()
  pool.AppendCertsFromPEM([]byte(tn.TLSConfig.TLSCert))
  pair, err := tls.X509KeyPair([]byte(tn.TLSConfig.TLSCert), []byte(tn.TLSConfig.TLSKey))
  if err != nil {
    return err
  }
  tc := credentials.NewTLS(&tls.Config{
    Certificates: []tls.Certificate{pair},
    ClientAuth: tls.RequireAndVerifyClientCert,
    ClientCAs: pool,
    RootCAs: pool,
  })

	s := grpc.NewServer(grpc.Creds(tc))
	tn.GRPCServer = s

	return nil
}

//getAuthType Determine the type of auth to connect to the tailnet
func (tn *Tailnet) getAuthType() AuthType {
	if tn.ClientID != "" && tn.ClientSecret != "" {
		return OAUTH
	}

	if tn.AuthKey != "" {
		return AUTHKEY

	}
	return NONE
}

