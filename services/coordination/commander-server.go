package coordination

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/charles-d-burton/tailsys/commands"
	pb "github.com/charles-d-burton/tailsys/commands"
	"github.com/charles-d-burton/tailsys/connections"
	"github.com/charles-d-burton/tailsys/data/queries"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommanderServer struct {
	sync.WaitGroup
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
    fmt.Println("looked up node: ", node.Hostname)
		names = append(names, node.Hostname)
	}
	res.Nodes = names
	return res, nil
}

func (c *CommanderServer) SendCommandToNodes(ctx context.Context, cmd *pb.CommanderRequest) (*pb.AggregateResponses, error) {
	agg := &commands.AggregateResponses{}
	hosts, err := queries.GetMatchRegisteredHosts(c.DB, cmd.Pattern)
	if err != nil {
		return nil, err
	}

	cmds := c.streamCommand(cmd.Command, hosts, 50)
	for rcmd := range cmds {
		agg.Response = append(agg.Response, rcmd)
	}

	return agg, nil
}

func (c *CommanderServer) SendCommandToNodesStream(cmd *commands.CommanderRequest, stream pb.CommandManager_SendCommandToNodesStreamServer) error {
	hosts, err := queries.GetMatchRegisteredHosts(c.DB, cmd.Pattern)
	if err != nil {
		return err
	}
	cmds := c.streamCommand(cmd.Command, hosts, 50)
	for rcmd := range cmds {
		if err := stream.Send(rcmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *CommanderServer) streamCommand(cmd string, hosts chan *queries.RegisteredHostsData, limit uint16) chan *commands.CommandResponse {
	if limit < 1 {
		limit = 50
	}
	responses := make(chan *commands.CommandResponse, 100)
	go func(hosts chan *queries.RegisteredHostsData, responses chan *commands.CommandResponse) {
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

func (c *CommanderServer) sendCommand(cmd string, node *queries.RegisteredHostsData, sem chan struct{}, results chan *commands.CommandResponse) {
	defer c.Done()           //decrement the wait group
	defer func() { <-sem }() //make space in the semaphore channel
	req := &pb.NodeRegistrationRequest{}
	if err := proto.Unmarshal(node.Data, req); err != nil {
		fmt.Println(fmt.Errorf("unable to unmarshal req: %w", err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
  fmt.Println("connecting to rpc client: ", req.Info.Hostname)
	conn, err := c.CO.DialContext(ctx, req.Info.Hostname+":"+req.Info.Port, &connections.TLSConfig{TLSKey: req.Tlskey, TLSCert: req.Tlscert})
	if err != nil {
		fmt.Println(fmt.Errorf("unable to connect to client: %s with err %w", req.Info.Hostname, err))
		return
	}
	cc := pb.NewCommandRunnerClient(conn)
	r, err := cc.Command(ctx, &pb.CommandRequest{
		Requested: timestamppb.Now(),
		Command:   cmd,
		Key:       &commands.Key{Key: c.ID},
	})

	if err != nil {
		fmt.Println(fmt.Errorf("unable to send command: %s to host %s with err: %w", cmd, req.Info.Hostname, err))
		return
	}
	results <- r
	fmt.Printf("successfully ran command %s on host %s with output %s\n", cmd, req.Info.Hostname, string(r.Output))
}
