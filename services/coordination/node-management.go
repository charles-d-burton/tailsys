package coordination

import (
	"context"
	"fmt"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/golang/protobuf/proto"
	"github.com/nutsdb/nutsdb"
)

// RegistrationServer struct to contain proto for gRPC
type RegistrationServer struct {
	pb.UnimplementedRegistrationServer
	DevMode bool
	ID      string
	DB      *nutsdb.DB
}

func (r *RegistrationServer) createRegistration(nrr *pb.NodeRegistrationRequest) error {
  clientName := nrr.GetInfo().Hostname
	// clientKey := nrr.Key.GetKey()
	fmt.Printf("registering %s\n", clientName)
	err := r.DB.Update(func(tx *nutsdb.Tx) error {
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

	return r.DB.Update(func(tx *nutsdb.Tx) error {
		data, err := proto.Marshal(nrr)
		if err != nil {
			return err
		}

		fmt.Println("recording request")
		err = tx.Put(registrationBucket, []byte(clientName), data, 0)
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
	in.Accepted = r.DevMode
	if r.DevMode {
		fmt.Println("running in dev mode, accepting all incoming connections")
	}
	if err := r.createRegistration(in); err != nil {
		return nil, err
	}

	return &pb.NodeRegistrationResponse{
		Accepted: r.DevMode,
		Key:      &pb.Key{Key: r.ID},
	}, nil
}
