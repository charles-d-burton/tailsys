package coordination

import (
	"context"
	"errors"
	"fmt" 
  "time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/data/queries"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// StartRPCCoordinationServer Register the gRPC server endpoints and start the server
func (co *Coordinator) StartRPCCoordinationServer(ctx context.Context) error {
	if co.DB == nil {
		return errors.New("datastore not initialized")
	}
	pb.RegisterPingerServer(co.GRPCServer, &services.Pinger{})
	pb.RegisterRegistrationServer(co.GRPCServer, &RegistrationServer{
		DevMode:  co.devMode,
		DB:       co.DB,
		Hostname: co.Hostname,
		//TODO: This is randomized on startup, we should persist and load
		ID: uuid.NewString(),
	})
	fmt.Println("rpc server starting to serve traffic")
	co.StartPingService(ctx)
	return co.GRPCServer.Serve(co.Listener)
}

func (co *Coordinator) StartPingService(ctx context.Context) {
	fmt.Println("starting node pings in the background")
	go co.pingNodes(ctx, 10)
}

func (co *Coordinator) pingNodes(ctx context.Context, limit int) {
	if limit < 1 {
		limit = 10
	}

	//Fill the semaphore pool
	sem := make(chan struct{}, limit)

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

func (co *Coordinator) ping(ctx context.Context, sem chan struct{}, hostRow *queries.RegisteredHostsRow) {
	defer func() { <-sem }() //make space in sem

	var node pb.NodeRegistrationRequest
	err := proto.Unmarshal(hostRow.Data, &node)
	if err != nil {
		fmt.Println(fmt.Errorf("error parsing host proto: %w\n", err))
		return
	}

	lastSeen := time.Now().Sub(node.Info.LastSeen.AsTime())
	if lastSeen.Minutes() > 10 {
		fmt.Printf("node %s has not been seen in more than ten minutes, not pinging\n", hostRow.Hostname)
		return
	}
	fmt.Println("pinging node")
	fmt.Println(node.GetInfo().Ip)
	host := hostRow.Hostname
	ctxTo, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
  conn, err := co.DialContext(ctxTo, host, &connections.TLSConfig{TLSKey: node.Tlskey, TLSCert: node.Tlscert})
	//TODO: Probably need to set the tailnet fqdn at some point
	defer conn.Close()

	if err != nil {
		fmt.Println(fmt.Errorf("error parsing host proto: %w\n", err))
		return
	}

	p := pb.NewPingerClient(conn)
	r, err := p.Ping(ctx, &pb.PingRequest{
		Ping: timestamppb.Now(),
	})
	if err != nil {
		fmt.Println(fmt.Errorf("unable to ping host %s with error: %w", host, err))
		return
	}

	fmt.Println("latency: ", r.InboundLatency)
	node.Info.LastSeen = timestamppb.Now()
	data, err := proto.Marshal(&node)
	if err != nil {
		fmt.Println("unable to marshal registration proto")
	}

	err = queries.UpdateRegisteredHost(co.DB, &queries.RegisteredHostsRow{
		Hostname: host,
		Key:      node.GetKey().Key,
		Data:     data,
	})

	if err != nil {
		fmt.Println(fmt.Errorf("unable to update registration record: %w", err))
	}
}
