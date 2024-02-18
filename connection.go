package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tailscale/tailscale-client-go/tailscale"
	"tailscale.com/tsnet"
)

type Tailnet struct {
	ClientID     string
	ClientSecret string
	AuthKey      string
	Hostname     string
	Scopes       []string
	Tags         []string
}

type Option func(tn *Tailnet) error

func (tn *Tailnet) NewConnection(ctx context.Context, opts ...Option) (*tsnet.Server, error) {
	for _, opt := range opts {
		err := opt(tn)
		if err != nil {
			return nil, err
		}
	}

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
			return nil, err
		}
		key, err := client.CreateKey(ctx, capabilities, topts...)
		if err != nil {
			return nil, err
		}
		tn.AuthKey = key.Key
	}

	if tn.AuthKey == "" {
		return nil, errors.New("no auth key set")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	tn.Hostname = hostname + "-tailsys"

	srv := &tsnet.Server{
		Hostname:  tn.Hostname,
		AuthKey:   tn.AuthKey,
		Ephemeral: true,
	}

	return srv, nil
}

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

func (tn *Tailnet) WithAPIKey(key string) Option {
	return func(tn *Tailnet) error {
		tn.AuthKey = key
		return nil
	}
}

func (tn *Tailnet) WithScopes(scopes ...string) Option {
	return func(tn *Tailnet) error {
		if scopes != nil {
			tn.Scopes = scopes
		}
		return nil
	}
}

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
