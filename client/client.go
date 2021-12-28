package main

import (
	"bufio"
	logic "chitty-chat/internalLogic"
	proto "chitty-chat/service"
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":5000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	
	c := proto.NewServiceClient(conn)
	defer conn.Close()

	userInput := os.Args[1:]
	readPort := userInput[0]
	readId := userInput[1]

	client := logic.Client{Id: readId, Port: readPort}

	joinClient(c, &client)

	go BroadcastListener(&client)

	for {
		publish(c, &client) // Infinite waiting for a message to be written
	}
}

func joinClient(c proto.ServiceClient, client *logic.Client) {
	client.LamportClock.IncrementLamport()
	log.Printf("Client timestamp :: %v", client.LamportClock.GetTimestamp())

	msg := proto.JoinMessage {
		Id: client.Id,
		Port: client.Port,
		LamportTimestamp: client.LamportClock.GetTimestamp(), // Not actually used
	}

	response, err := c.Join(context.Background(), &msg)
	if (err != nil) {
		log.Fatalf("The client %v could not join", client.Id)
	}

	client.LamportClock.IncrementMaxLamport(response.LamportTimestamp)
	log.Printf("The client %s was succesfully joined at Lamport %v", client.Id, client.LamportClock.GetTimestamp())
}

func publish(c proto.ServiceClient, client *logic.Client) {
	var readContent string
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan() // Scanning user input
	readContent = scanner.Text()
	if (len(readContent) > 4) {
		log.Printf("The message must be less than 128 characters")
	} else {
		client.LamportClock.IncrementLamport()
		log.Printf("Client timestamp :: %v", client.LamportClock.GetTimestamp())

		pub := proto.PublishMessage {
			Id: client.Id,
			Content: readContent,
			LamportTimestamp: client.LamportClock.GetTimestamp(),
		}

		_, err := c.Publish(context.Background(), &pub)
		if (err != nil) {
			log.Fatalf("The message '%v' could not be published", readContent)
		}
	}
}

func BroadcastListener(c *logic.Client) {
	lis, err := net.Listen("tcp", ":"+c.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	proto.RegisterBroadcastServiceServer(grpcServer, c)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}