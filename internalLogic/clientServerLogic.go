package logic

import (
	proto "chitty-chat/service"
	"context"
	"log"

	"google.golang.org/grpc"
)

type Server struct {
	Clients []Client
	LamportClock Lamport
	proto.UnimplementedServiceServer
}

type Client struct {
	Id string
	Port string
	LamportClock Lamport
	proto.UnimplementedBroadcastServiceServer
}

func (s *Server) Join(context context.Context, joinMessage *proto.JoinMessage) (*proto.JoinResponse, error) {
	client := Client {
		Id: joinMessage.Id,
		Port: joinMessage.Port,
		LamportClock: Lamport {lamportTimestamp: joinMessage.LamportTimestamp},
	}

	s.Clients = append(s.Clients, client)
	s.LamportClock.IncrementMaxLamport(joinMessage.LamportTimestamp) // Clients lamport will always be one
	log.Printf("Server timestamp :: %v", s.LamportClock.GetTimestamp())

	return &proto.JoinResponse{LamportTimestamp: s.LamportClock.GetTimestamp()}, nil
}

func (s *Server) Publish(context context.Context, publishMessage *proto.PublishMessage) (*proto.Empty, error) {
	s.LamportClock.IncrementMaxLamport(publishMessage.GetLamportTimestamp())
	log.Printf("Server timestamp :: %v", s.LamportClock.GetTimestamp())

	log.Printf("Recieved message '%s' from %v", publishMessage.Content, publishMessage.Id)

	broadcastLoop(s, publishMessage)

	return &proto.Empty{}, nil
}

func (c *Client) Broadcast(context context.Context, broadcastMessage *proto.BroadcastMessage) (*proto.Empty, error) {
	c.LamportClock.IncrementMaxLamport(broadcastMessage.GetLamportTimestamp())

	log.Printf("The client %s have published a message with the following content '%s'", broadcastMessage.Id, broadcastMessage.Content)
	log.Printf("Client timestamp :: %v", c.LamportClock.GetTimestamp())

	return &proto.Empty{}, nil
}

func broadcastLoop(s *Server, publishMessage *proto.PublishMessage) {
	s.LamportClock.IncrementLamport()
	log.Printf("Server timestamp :: %v", s.LamportClock.GetTimestamp())

	for _, client := range s.Clients {
		var conn *grpc.ClientConn
		conn, err := grpc.Dial(":"+client.Port, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		
		c := proto.NewBroadcastServiceClient(conn)
		defer conn.Close()

		bMsg := proto.BroadcastMessage {
			Id: publishMessage.Id,
			Content: publishMessage.Content,
			LamportTimestamp: s.LamportClock.GetTimestamp(),
		}

		_, errBr := c.Broadcast(context.Background(), &bMsg)
		if (err != nil) {
			log.Fatalf("The broadcast did not succed :: %s", errBr)
		}
	}
}