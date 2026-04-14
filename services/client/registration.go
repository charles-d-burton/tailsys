package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/data/queries"
	"github.com/charles-d-burton/tailsys/models"
	"gopkg.in/yaml.v3"
)

// RegisterWithCoordinationServer generate the registration request and send it to the coordination server
func (cl *Client) RegisterWithCoordinationServer(ctx context.Context, addr string) error {
	for i := 0; i < 5; i++ {
		fmt.Println("coordination server address: ", addr)
		ctxTo, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		sconfig, err := cl.getTlSConfig()
		if err != nil {
			return err
		}

		client, err := cl.HTTPClient(sconfig)
		if err != nil {
			return err
		}

		fmt.Println("attempting to send registration request")

		req := models.NodeRegistrationRequest{
			Info: models.SysInfo{
				Hostname: cl.Hostname,
				Port:     cl.Port,
				Type:     models.LINUX,
				IP:       cl.Hostname,
				LastSeen: time.Now(),
			},
			Key:        models.Key{Key: cl.ID},
			SystemType: models.CLIENT,
			TLSKey:     cl.TLSConfig.TLSKey,
			TLSCert:    cl.TLSConfig.TLSCert,
		}
		fmt.Println(req)

		data, err := yaml.Marshal(req)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("https://%s/register", addr)
		hreq, err := http.NewRequestWithContext(ctxTo, "POST", url, bytes.NewReader(data))
		if err != nil {
			return err
		}
		hreq.Header.Set("Content-Type", "application/x-yaml")

		resp, err := client.Do(hreq)
		if err != nil {
			fmt.Println(err)
			time.Sleep(3 * time.Second)
			if i == 4 {
				return fmt.Errorf("unable to connect to coordination server: %s", addr)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("received non-200 response: %d\n", resp.StatusCode)
			time.Sleep(3 * time.Second)
			continue
		}

		var r models.NodeRegistrationResponse
		if err := yaml.NewDecoder(resp.Body).Decode(&r); err != nil {
			return err
		}

		fmt.Println("registering response")
		err = cl.addRegistration(&r)
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

func (cl *Client) addRegistration(r *models.NodeRegistrationResponse) error {
	fmt.Println("registering server")
	err := queries.SetRegisteredCoordinationServer(cl.DB, r.Hostname, r.Key.Key)
	if err != nil {
		fmt.Println("could not register server")
		return err
	}
	return nil
}
