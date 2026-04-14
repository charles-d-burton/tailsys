package commander 

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/models"
	"github.com/google/uuid"
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
			return errors.New("no coordination server configured")
		}
		params := strings.Split(server, ":")
		if len(params) < 2 {
			return fmt.Errorf("server hostname %s not formatted correctly, should be <server>:<port>", server)
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

func (cl *Client) SendCommand(ctx context.Context, cmd string, pattern string) error {
	for i := 0; i < 5; i++ {

		if cl.TLSConfig == nil {
			tls, err := cl.getTlSConfig()
			if err != nil {
				return err
			}
			cl.TLSConfig = tls
		}

		client, err := cl.HTTPClient(cl.TLSConfig)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to create http client: %w", err))
			time.Sleep(5 * time.Second)
			continue
		}

		commandReq := &models.CommanderRequest{
			Command: cmd,
			Pattern: pattern,
		}

		data, err := yaml.Marshal(commandReq)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("https://%s/nodes/command", cl.CoordinationServer)
		hreq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
		if err != nil {
			return err
		}
		hreq.Header.Set("Content-Type", "application/x-yaml")

		resp, err := client.Do(hreq)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to send command: %w", err))
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("received non-200 response: %d\n", resp.StatusCode)
			time.Sleep(5 * time.Second)
			continue
		}

		var r models.AggregateResponses
		if err := yaml.NewDecoder(resp.Body).Decode(&r); err != nil {
			return err
		}

		for _, res := range r.Response {
			fmt.Printf("command: %s ran with exit code %d\n", cmd, res.ExitCode)
			fmt.Printf("  result: %s\n", string(res.Output))
		}
		return nil
	}
	return nil
}

func (cl *Client) GetNodes(ctx context.Context, pattern string) error {
	for i := 0; i < 5; i++ {

		if cl.TLSConfig == nil {
			tls, err := cl.getTlSConfig()
			if err != nil {
				return err
			}
			cl.TLSConfig = tls
		}

		client, err := cl.HTTPClient(cl.TLSConfig)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to create http client: %w", err))
			time.Sleep(5 * time.Second)
			continue
		}

		url := fmt.Sprintf("https://%s/nodes?pattern=%s", cl.CoordinationServer, pattern)
		hreq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(hreq)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to get nodes: %w", err))
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("received non-200 response: %d\n", resp.StatusCode)
			time.Sleep(5 * time.Second)
			continue
		}

		var r models.NodeQueryResponse
		if err := yaml.NewDecoder(resp.Body).Decode(&r); err != nil {
			return err
		}

		for _, res := range r.Nodes {
			fmt.Printf("found node %s\n", res)
		}
		return nil
	}
	return nil
}
