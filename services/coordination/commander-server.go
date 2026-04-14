package coordination

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/data/queries"
	"github.com/charles-d-burton/tailsys/models"
	"gopkg.in/yaml.v3"
)

type CommanderServer struct {
	sync.WaitGroup
	DB *sql.DB
	CO *Coordinator
	ID string
}

func (c *CommanderServer) GetNodes(w http.ResponseWriter, r *http.Request) {
	pattern := r.URL.Query().Get("pattern")
	nodes, err := queries.GetMatchRegisteredHosts(c.DB, pattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	names := make([]string, 0)
	for node := range nodes {
		fmt.Println("looked up node: ", node.Hostname)
		names = append(names, node.Hostname)
	}
	out := models.NodeQueryResponse{Nodes: names}
	models.WriteYAML(w, http.StatusOK, out)
}

func (c *CommanderServer) SendCommandToNodes(w http.ResponseWriter, r *http.Request) {
	var cmd models.CommanderRequest
	if err := models.ReadYAML(r, &cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hosts, err := queries.GetMatchRegisteredHosts(c.DB, cmd.Pattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	agg := models.AggregateResponses{}
	cmds := c.streamCommand(cmd.Command, hosts, 50)
	for rcmd := range cmds {
		agg.Response = append(agg.Response, *rcmd)
	}

	models.WriteYAML(w, http.StatusOK, agg)
}

func (c *CommanderServer) streamCommand(cmd string, hosts chan *queries.RegisteredHostsData, limit uint16) chan *models.CommandResponse {
	if limit < 1 {
		limit = 50
	}
	responses := make(chan *models.CommandResponse, 100)
	go func(hosts chan *queries.RegisteredHostsData, responses chan *models.CommandResponse) {
		//create the semaphore pool
		sem := make(chan struct{}, limit)
		fmt.Println("starting command processor stream")
		for host := range hosts {
			c.Add(1) //increment the waitgroup
			sem <- struct{}{}
			go c.sendCommand(cmd, host, sem, responses)
		}
		close(sem)
		c.Wait()         //wait for all worker processes to finish
		close(responses) //producer closes
	}(hosts, responses)

	return responses
}

func (c *CommanderServer) sendCommand(cmd string, node *queries.RegisteredHostsData, sem chan struct{}, results chan *models.CommandResponse) {
	defer c.Done()           //decrement the wait group
	defer func() { <-sem }() //make space in the semaphore channel
	req := &models.NodeRegistrationRequest{}
	if err := yaml.Unmarshal(node.Data, req); err != nil {
		fmt.Println(fmt.Errorf("unable to unmarshal req: %w", err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	fmt.Println("connecting to http client: ", req.Info.Hostname)

	client, err := c.CO.HTTPClient(&connections.TLSConfig{TLSKey: req.TLSKey, TLSCert: req.TLSCert})
	if err != nil {
		fmt.Println(fmt.Errorf("unable to create http client for: %s with err %w", req.Info.Hostname, err))
		return
	}

	commandReq := models.CommandRequest{
		Requested: time.Now(),
		Command:   cmd,
		Key:       models.Key{Key: c.ID},
	}

	data, err := yaml.Marshal(commandReq)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to marshal command request: %w", err))
		return
	}

	//TODO: Probably need to set the tailnet fqdn at some point
	url := fmt.Sprintf("https://%s:%s/command", req.Info.Hostname, req.Info.Port)
	hreq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		fmt.Println(fmt.Errorf("unable to create http request: %w", err))
		return
	}
	hreq.Header.Set("Content-Type", "application/x-yaml")

	resp, err := client.Do(hreq)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to send command: %s to host %s with err: %w", cmd, req.Info.Hostname, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Errorf("non-200 response from host %s: %d", req.Info.Hostname, resp.StatusCode))
		return
	}

	var r models.CommandResponse
	if err := yaml.NewDecoder(resp.Body).Decode(&r); err != nil {
		fmt.Println(fmt.Errorf("unable to decode command response from host %s: %w", req.Info.Hostname, err))
		return
	}

	results <- &r
	fmt.Printf("successfully ran command %s on host %s with output %s\n", cmd, req.Info.Hostname, string(r.Output))
}
