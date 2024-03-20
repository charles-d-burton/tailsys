package coordination

import (
	"context"
	"database/sql"
	"fmt"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/data/queries"
  "google.golang.org/protobuf/proto"
)

// RegistrationServer struct to contain proto for gRPC
type RegistrationServer struct {
	pb.UnimplementedRegistrationServer
	DevMode  bool
	ID       string
	Hostname string
	DB       *sql.DB
}

func (r *RegistrationServer) createRegistration(nrr *pb.NodeRegistrationRequest) error {

	data, err := proto.Marshal(nrr)
	if err != nil {
		return err
	}

	clientName := nrr.GetInfo().Hostname

	nhr := &queries.RegisteredHostsData{
		Hostname: nrr.GetInfo().Hostname,
		Key:      nrr.GetKey().Key,
		Data:     data,
	}
	// clientKey := nrr.Key.GetKey()
	fmt.Printf("registering %s\n", clientName)
	err = queries.InsertHostRegistration(r.DB, nhr)

	if err != nil {
		fmt.Println("could not create bucket")
		return err
	}
	return nil
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
		Hostname: r.Hostname,
	}, nil
}
