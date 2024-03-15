package coordination

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/google/uuid"
	"github.com/nutsdb/nutsdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	registrationBucket = "node-registration"
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

	if co.DB == nil {
		return errors.New("datastore not initialized")
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

func (co *Coordinator) WithDataDir(dir string) Option {
	return func(co *Coordinator) error {
		return co.StartDB(dir)
	}
} // StartRPCCoordinationServer Register the gRPC server endpoints and start the server
func (co *Coordinator) StartRPCCoordinationServer(ctx context.Context) error {
	pb.RegisterPingerServer(co.GRPCServer, &services.Pinger{})
	pb.RegisterRegistrationServer(co.GRPCServer, &RegistrationServer{
		DevMode: co.devMode,
		DB:      co.DB,
		//TODO: This is randomized on startup, we should persist and load
		ID: uuid.NewString(),
	})
	fmt.Println("rpc server starting to serve traffic")
  co.StartPingService(ctx)
	return co.GRPCServer.Serve(co.Listener)
}

func (co *Coordinator)pingNodes(ctx context.Context, limit int) {
  if limit < 1 {
    limit = 10
  }

  //Fill the semaphore pool
  sem := make(chan struct{}, limit)

  fmt.Println("starting ping ticker")
  ticker := time.NewTicker(time.Second * 15) 
  for range ticker.C {
    if err := co.DB.View(func(tx *nutsdb.Tx) error {
      fmt.Println("getting keys")
      kvs, err := tx.GetKeys(registrationBucket)
      if err != nil {
        return err
      }
      for _, key := range kvs{
        sem <- struct{}{} //Block until sem has space
        go co.ping(ctx, sem, key) 

      }
      return nil
    }); err != nil {
      fmt.Println(err)
    }
  }
}

func (co *Coordinator)ping(ctx context.Context, sem chan struct{}, key []byte) {
  defer func() {<-sem}() //make space in sem

  err := co.DB.View(func(tx *nutsdb.Tx) error {
    data, err := tx.Get(registrationBucket, key)
    if err != nil {
      return err
    }
    var node pb.NodeRegistrationRequest
    err = proto.Unmarshal(data, &node)
    if err != nil {
      return err
    }
    fmt.Println("pinging node")
    fmt.Println(node.GetInfo().Ip)
    host := node.GetInfo().Ip
    ctxTo, cancel := context.WithTimeout(ctx, time.Second * 2)
    defer cancel()

    //TODO: Probably need to set the tailnet fqdn at some point
    conn, err := grpc.DialContext(ctxTo, host, 
      grpc.WithTransportCredentials(insecure.NewCredentials()),
      grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error){
        return co.TSServer.Dial(ctx, "-", host + ":6655") 
      }),
    )
    if err != nil {
      return err
    }
    defer conn.Close()

    p := pb.NewPingerClient(conn)
    r, err := p.Ping(ctx, &pb.PingRequest{
      Ping: timestamppb.Now(),
    })
    fmt.Println("latency: ", r.InboundLatency)
    
    return nil
  })
  if err != nil {
    fmt.Println(err)
  }
}

func (co *Coordinator) StartPingService(ctx context.Context) {
  fmt.Println("starting node pings in the background")
  go co.pingNodes(ctx, 10)
}
