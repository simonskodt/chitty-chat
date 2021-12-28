package main

import (
	logic "chitty-chat/internalLogic"
	proto "chitty-chat/service"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	log.Printf("SERVER STARTED")

	s := logic.Server{}
	go ServerListening(&s)

	for {} // Prevent terminating
}

func ServerListening(s *logic.Server) {
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	proto.RegisterServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
