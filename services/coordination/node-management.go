package coordination

import (
	"context"
	"fmt"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/services"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/nutsdb/nutsdb"
)

// StartRPCCoordinationServer Register the gRPC server endpoints and start the server
func (co *Coordinator) StartRPCCoordinationServer(ctx context.Context) error {
	pb.RegisterPingerServer(co.GRPCServer, &services.Pinger{})
	pb.RegisterRegistrationServer(co.GRPCServer, &RegistrationServer{
		DevMode: co.devMode,
		DB:      co.DB,
	})
	fmt.Println("rpc server starting to serve traffic")
	return co.GRPCServer.Serve(co.Listener)
}

// RegistrationServer struct to contain proto for gRPC
type RegistrationServer struct {
	pb.UnimplementedRegistrationServer
	DevMode bool
	ID      string
	DB      *nutsdb.DB
}

func (r *RegistrationServer) createRegistration(nrr *pb.NodeRegistrationRequest) error {
	clientKey := nrr.Key.GetKey()
	fmt.Printf("registering %s\n", clientKey)
  err := r.DB.Update(func (tx *nutsdb.Tx) error {
		if !tx.ExistBucket(nutsdb.DataStructureBTree, registrationBucket) {
			fmt.Println("recording registration status")
			return tx.NewBucket(nutsdb.DataStructureBTree, registrationBucket)
    }
    return nil
  })

  if err != nil {
    fmt.Println("could not create bucket")
    return err
  }

  return r.DB.Update(func (tx *nutsdb.Tx) error {
		data, err := proto.Marshal(nrr)
		if err != nil {
			return err
		}

		fmt.Println("recording request")
		err = tx.Put(registrationBucket, []byte(clientKey), data, 0)
		if err != nil {
      fmt.Println("could not put record")
			return err
		}
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
		if err := r.createRegistration(in); err != nil {
			return nil, err
		}

		return &pb.NodeRegistrationResponse{
			Accepted: true,
			Key:      &pb.Key{Key: id.String()},
		}, nil
	}

	if err := r.createRegistration(in); err != nil {
		return nil, err
	}

	return &pb.NodeRegistrationResponse{
		Accepted: false,
		Key:      &pb.Key{Key: "coordination-server-key"},
	}, nil
}
