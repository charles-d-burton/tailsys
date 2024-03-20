package coordination

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/charles-d-burton/tailsys/commands"
	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/data/queries"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommanderServer struct {
  pb.UnimplementedCommandManagerServer  
  DB *sql.DB
  CO *Coordinator
  ID string
}

func (c *CommanderServer) GetNodes(ctx context.Context, in *pb.NodeQuery) (*pb.NodeQueryResponse, error) {
  res := &pb.NodeQueryResponse{}
  nodes, err := queries.GetMatchRegisteredHosts(c.DB, in.Pattern)
  if err != nil {
    return nil, err
  }

  names := make([]string, 0)
  for node := range nodes {
    names = append(names, node.Hostname)
  }
  res.Nodes = names
  return res,nil  
}

func (c *CommanderServer) SendCommandToNodes(ctx context.Context, in *pb.CommanderRequest) (*pb.AggregateResponses, error) {
  return &commands.AggregateResponses{}, nil
}

func (c *CommanderServer) SendCommandToNodesStream(cmd *commands.CommanderRequest, stream pb.CommandManager_SendCommandToNodesStreamServer) error {
  return nil
}

func (c *CommanderServer) streamCommand(cmd string, hosts chan *queries.RegisteredHostsData, limit int) (chan *commands.CommandResponse, error ) {
  if limit < 1 {
    limit = 10
  }
  responses := make(chan *commands.CommandResponse, 100)
  defer close(responses) 
  //create the semaphore pool
  sem := make(chan struct{}, limit)
  fmt.Println("starting command processor stream")
  for host := range hosts {
    sem <- struct{}{} 
    go c.sendCommand(cmd, host, sem, responses)
  }
  close(sem)
  for range sem {
    //Block until all of the workers are finished
  }
  return responses, nil
}

func (c *CommanderServer) sendCommand(cmd string, node *queries.RegisteredHostsData , sem chan struct{}, results chan *commands.CommandResponse) {
  defer func() {<-sem}() //make space in the semaphore channel
  req := &pb.NodeRegistrationRequest{}
  if err := proto.Unmarshal(node.Data, req); err != nil {
    fmt.Println(fmt.Errorf("unable to unmarshal req: %w", err))
    return
  }
  ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
  defer cancel()
  conn, err := c.CO.DialContext(ctx, req.Info.Hostname + ":" +req.Info.Port, &connections.TLSConfig{TLSKey: req.Tlskey, TLSCert: req.Tlscert})
  if err != nil {
    fmt.Println(fmt.Errorf("unable to connect to client: %s with err %w", req.Info.Hostname, err))
  }
  cc := pb.NewCommandRunnerClient(conn)
  r, err := cc.Command(ctx, &pb.CommandRequest{
    Requested: timestamppb.Now(),  
    Command: cmd,
    Key: &commands.Key{Key: c.ID},
  })
  if err != nil {
    fmt.Println(fmt.Errorf("unable to send command: %s to host %s with err: %w", cmd, req.Info.Hostname, err))
  }
  results <- r
  fmt.Printf("successfully ran command %s on host %s with output %s\n", cmd, req.Info.Hostname, string(r.Output))
}
