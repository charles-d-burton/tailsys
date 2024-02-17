package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/tailscale/tailscale-client-go/tailscale"
	"tailscale.com/tsnet"
)

type Listener struct {
  ClientID string
  ClientSecret string
  AuthKey string
  Hostname string
  Scopes []string
}

type Option func(l *Listener) error 

func (l *Listener) NewConnection(ctx context.Context, opts ...Option) (*tsnet.Server, error) {
  for _, opt := range opts {
    err := opt(l)
    if err != nil {
      return nil, err
    }
  }

	var capabilities tailscale.KeyCapabilities
	capabilities.Devices.Create.Reusable = true
	capabilities.Devices.Create.Ephemeral = true
	capabilities.Devices.Create.Tags = []string{"tag:tailsys"}
	capabilities.Devices.Create.Preauthorized = true

	var topts []tailscale.CreateKeyOption
	topts = append(topts, tailscale.WithKeyExpiry(10*time.Second))

  if useOauth(l.ClientID, l.ClientSecret) {
    client, err := tailscale.NewClient(
      "",
      "-",
      tailscale.WithOAuthClientCredentials(l.ClientID, l.ClientSecret, l.Scopes),
      )
    if err != nil {
      return  nil, err
    }
    key ,err := client.CreateKey(ctx, capabilities, topts...)
    l.AuthKey = key.Key
  }

  if l.AuthKey == "" {
    return nil, errors.New("no auth key set")
  }


  hostname, err := os.Hostname()
  if err != nil {
    return nil, err
  }
  l.Hostname = hostname + "-tailsys"

  srv := &tsnet.Server{
    Hostname: l.Hostname,
    AuthKey: l.AuthKey,
    Ephemeral: true,
  }

  return srv, nil
}

func(l *Listener) WithOauth(clientId, clientSecret string) Option {
  return func(l *Listener) error {
    if clientId == "" {
      return errors.New("client id not set")
    }

    if clientSecret == "" {
      return errors.New("client secret not set")
    }
    l.ClientID = clientId
    l.ClientSecret = clientSecret
    return nil
  } 
}

func(l *Listener) WithAPIKey(key string) Option {
  return func(l *Listener) error {
    l.AuthKey = key
    return nil
  }
}

func(l *Listener) WithScopes(scopes ...string) Option {
  return func(l *Listener) error {
    if scopes != nil {
      l.Scopes = scopes
    }
    return nil
  }
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
