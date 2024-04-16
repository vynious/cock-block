package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.con/vynious/cock-block/node"
	"github.con/vynious/cock-block/proto"
	"google.golang.org/grpc"
)

func main() {

	node := node.NewNode()
	opts := []grpc.ServerOption{
	}
	grpcServer := grpc.NewServer(opts...)
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	proto.RegisterNodeServer(grpcServer, node)
	log.Println("node running on port: ", ":3000")

	go func() {
		for {
			time.Sleep(2 * time.Second)
			makeTransaction()
		}
	}()

	grpcServer.Serve(ln)
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	c := proto.NewNodeClient(client)
	if _, err = c.HandleTransaction(context.TODO(), &proto.Transaction{}); err != nil {
		log.Fatal(err)
	}
}
