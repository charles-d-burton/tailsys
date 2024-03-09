package coordination

import (
	"context"
	"fmt"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/google/uuid"
	"github.com/nutsdb/nutsdb"
)

//Coordinator holds the runtime variables for the coordination server
type Coordinator struct {
	connections.Tailnet
	devMode bool
	DB      *nutsdb.DB
}

//Options defines the configuration options function for configuration injection
type Option func(co *Coordinator) error

//NewCoordinator Create a new coordinator instance and set the provided options
func (co *Coordinator) NewCoordinator(ctx context.Context, opts ...Option) error {

	for _, opt := range opts {
		err := opt(co)
		if err != nil {
			return err
		}
	}
	return nil
}

//WithDevMode enable the server to run in dev mode
func (co *Coordinator) WithDevMode(mode bool) Option {
	return func(co *Coordinator) error {
    fmt.Println("setting dev mode to: ", mode)
		co.devMode = mode
		return nil
	}
}

//WithDataDir set the location for the database to be stored
func (co *Coordinator) WithDataDir(dir string) Option {
	return func(co *Coordinator) error {
    fmt.Println("creating database at: ", dir)
    // if err := os.MkdirAll(dir, os.ModePerm); err != nil {
    //   return err
    // }
		db, err := nutsdb.Open(
			nutsdb.DefaultOptions,
			nutsdb.WithDir(dir),
		)
		if err != nil {
			return err
		}
		co.DB = db
    fmt.Println("database is created an online")
		return nil
	}
}

//StartRPCCoordinationServer Register the gRPC server endpoints and start the server
func (co *Coordinator) StartRPCCoordinationServer(ctx context.Context) error {
	pb.RegisterPingerServer(co.GRPCServer, &services.Pinger{})
	pb.RegisterRegistrationServer(co.GRPCServer, &RegistrationServer{
		DevMode: co.devMode,
		DB:      co.DB,
	})
	return co.GRPCServer.Serve(co.Listener)
}

// RegistrationServer struct to contain proto for gRPC
type RegistrationServer struct {
	pb.UnimplementedRegistrationServer
	DevMode bool
	ID      string
	DB      *nutsdb.DB
}

func (r *RegistrationServer) createRegistration(ctx context.Context, nrr *pb.NodeRegistrationRequest) error {
  bucket := nrr.Key.GetKey()
  fmt.Printf("registering %s", bucket)
  return r.DB.Update(func (tx *nutsdb.Tx)error {
    if !tx.ExistBucket(nutsdb.DataStructureBTree, bucket) {
      fmt.Println("bucket does not exist, creating bucket for node")
      return tx.NewBucket(nutsdb.DataStructureBTree, bucket)
    }
    fmt.Println("bucket exists, returning to caller")
    return nil
  })
} 

// Register registers a node with the database when a node sends a request.  Returns the server id so the node can verify further requests
func (r *RegistrationServer) Register(ctx context.Context, in *pb.NodeRegistrationRequest) (*pb.NodeRegistrationResponse, error) {
	fmt.Println("received coordination request")
	fmt.Println(in)
	if r.DevMode {

		id := uuid.New()
		fmt.Println("running in dev mode, accepting all incoming connections")
    if err := r.createRegistration(ctx, in); err != nil {
      return nil, err
    }

		return &pb.NodeRegistrationResponse{
			Accepted: true,
			Key:      &pb.Key{Key: id.String()},
		}, nil
	}

  if err := r.createRegistration(ctx, in); err != nil {
    return nil, err
  }
  
	return &pb.NodeRegistrationResponse{
		Accepted: false,
		Key:      &pb.Key{Key: "coordination-server-key"},
	}, nil
}
