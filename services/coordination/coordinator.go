package coordination

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/data/queries"
	"github.com/charles-d-burton/tailsys/models"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Coordinator holds the runtime variables for the coordination server
type Coordinator struct {
	connections.Tailnet
	services.DataManagement
	devMode bool
	ID      string
}

// Options defines the configuration options function for configuration injection
type Option func(co *Coordinator) error

// NewCoordinator Create a new coordinator instance and set the provided options
func (co *Coordinator) NewCoordinator(ctx context.Context, opts ...Option) error {

	for _, opt := range opts {
		err := opt(co)
		if err != nil {
			return err
		}
	}
	co.ID = uuid.NewString()
	return nil
}

// WithDevMode enable the server to run in dev mode
func (co *Coordinator) WithDevMode(mode bool) Option {
	return func(co *Coordinator) error {
		fmt.Println("setting dev mode to: ", mode)
		co.devMode = mode
		return nil
	}
}

func (co *Coordinator) StartDatabase(ctx context.Context) error {
	return co.StartDB(co.ConfigDir)
}

// StartHTTPCoordinationServer Register the HTTP server endpoints and start the server
func (co *Coordinator) StartHTTPCoordinationServer(ctx context.Context) error {
	if co.DB == nil {
		return errors.New("datastore not initialized")
	}

	mux := http.NewServeMux()

	pinger := &services.Pinger{DB: co.DB, ID: co.ID}
	mux.HandleFunc("POST /ping", pinger.Ping)

	regServer := &RegistrationServer{
		DevMode:  co.devMode,
		DB:       co.DB,
		Hostname: co.Hostname,
		ID:       co.ID,
	}
	mux.HandleFunc("POST /register", regServer.Register)

	commanderServer := &CommanderServer{
		DB: co.DB,
		CO: co,
		ID: co.ID,
	}
	mux.HandleFunc("GET /nodes", commanderServer.GetNodes)
	mux.HandleFunc("POST /nodes/command", commanderServer.SendCommandToNodes)

	co.HTTPServer.Handler = mux

	fmt.Println("http server starting to serve traffic")
	co.StartPingService(ctx)
	return co.HTTPServer.Serve(co.Listener)
}

func (co *Coordinator) StartPingService(ctx context.Context) {
	fmt.Println("starting node pings in the background")
	go co.pingNodes(ctx, 10)
}

func (co *Coordinator) pingNodes(ctx context.Context, limit uint16) {
	if limit < 1 {
		limit = 50
	}

	//Fill the semaphore pool
	sem := make(chan struct{}, limit)
	defer close(sem)

	fmt.Println("starting ping ticker")
	ticker := time.NewTicker(time.Second * 15)
	pingRunning := false
	for range ticker.C {
		if !pingRunning {
			pingRunning = true
			hosts := queries.GetRegisteredHosts(co.DB)
			for host := range hosts {
				sem <- struct{}{} //Block until sem has space
				go co.ping(ctx, sem, host)
			}
			pingRunning = false
			continue
		}
		fmt.Println("ping already running, waiting on previous attempt")
	}
}

func (co *Coordinator) ping(ctx context.Context, sem chan struct{}, hostRow *queries.RegisteredHostsData) {
	defer func() { <-sem }() //make space in sem

	var node models.NodeRegistrationRequest
	err := yaml.Unmarshal(hostRow.Data, &node)
	if err != nil {
		fmt.Println(fmt.Errorf("error parsing host yaml: %w\n", err))
		return
	}

	lastSeen := time.Since(node.Info.LastSeen)
	if lastSeen.Minutes() > 10 {
		fmt.Printf("node %s has not been seen in more than ten minutes, not pinging\n", hostRow.Hostname)
		return
	}
	fmt.Println("pinging node")
	fmt.Println(node.Info.IP)
	host := hostRow.Hostname
	ctxTo, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	client, err := co.HTTPClient(&connections.TLSConfig{TLSKey: node.TLSKey, TLSCert: node.TLSCert})
	if err != nil {
		fmt.Println(fmt.Errorf("unable to create http client: %w\n", err))
		return
	}

	pingReq := models.PingRequest{
		Ping: time.Now(),
	}
	data, err := yaml.Marshal(pingReq)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to marshal ping request: %w\n", err))
		return
	}

	url := fmt.Sprintf("https://%s:%s/ping", host, node.Info.Port)
	hreq, err := http.NewRequestWithContext(ctxTo, "POST", url, bytes.NewReader(data))
	if err != nil {
		fmt.Println(fmt.Errorf("unable to create http request: %w\n", err))
		return
	}
	hreq.Header.Set("Content-Type", "application/x-yaml")

	resp, err := client.Do(hreq)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to ping host %s with error: %w", host, err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Errorf("non-200 response from host %s: %d", host, resp.StatusCode))
		return
	}

	var r models.PongResponse
	if err := yaml.NewDecoder(resp.Body).Decode(&r); err != nil {
		fmt.Println(fmt.Errorf("unable to decode ping response: %w", err))
		return
	}

	fmt.Println("latency: ", r.InboundLatency)
	node.Info.LastSeen = time.Now()
	data, err = yaml.Marshal(&node)
	if err != nil {
		fmt.Println("unable to marshal registration yaml")
	}

	err = queries.UpdateRegisteredHost(co.DB, &queries.RegisteredHostsData{
		Hostname: host,
		Key:      node.Key.Key,
		Data:     data,
	})

	if err != nil {
		fmt.Println(fmt.Errorf("unable to update registration record: %w", err))
	}
}
