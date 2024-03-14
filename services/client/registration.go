package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/golang/protobuf/proto"
	"github.com/nutsdb/nutsdb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (cl *Client) RegisterWithCoordinationServer(ctx context.Context, addr string) error {
	for i := 0; i < 5; i++ {
		fmt.Println("coordination server address: ", addr)
		// conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithBackoffMaxDelay(10*time.Second))
		if err != nil {
			return err
		}
		defer conn.Close()

		c := pb.NewRegistrationClient(conn)

		fmt.Println("attempting to send registration request")
    fmt.Println(conn.GetState())
    req := &pb.NodeRegistrationRequest{
			Info: &pb.SysInfo{
				Hostname: cl.Hostname,
				Type:     pb.OSType_LINUX,
				Ip:       cl.Hostname,
				LastSeen: timestamppb.Now(),
			},
			Key:        &pb.Key{Key: cl.ID},
			SystemType: pb.SystemType_CLIENT,
    }
    fmt.Println(req)
		r, err := c.Register(ctx, req)

		if err != nil {
      fmt.Println("unable to send request")
			fmt.Println(err)
			time.Sleep(3 * time.Second)
			if i == 4 {
				return errors.New(fmt.Sprintf("unable to connect to coordation server: %s", addr))
			}
			continue
		}

		fmt.Println("registering response")
		err = cl.addRegistration(r)
		if err != nil {
			return err
		}
		fmt.Println(r)
		break
	}
	return nil
}

func (cl *Client) addRegistration(r *pb.NodeRegistrationResponse) error {
	serverKey := r.Key.GetKey()
	fmt.Println("registering server")

  err := cl.DB.Update(func (tx *nutsdb.Tx) error {
		if !tx.ExistBucket(nutsdb.DataStructureBTree, coordinationBucket) {
			fmt.Println("recording registration status")
			return tx.NewBucket(nutsdb.DataStructureBTree, coordinationBucket)
    }
    return nil
  })

  if err != nil {
    fmt.Println("could not create bucket")
    return err
  }
  
  return cl.DB.Update(func (tx *nutsdb.Tx) error {
		data, err := proto.Marshal(r)
		if err != nil {
			return err
		}

		fmt.Println("recording request")
		err = tx.Put(coordinationBucket, []byte(serverKey), data, 0)
		if err != nil {
      fmt.Println("could not put record")
			return err
		}
    return nil
  })
}
