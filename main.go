package main

import (
	"context"
	"log"
	"time"

	"github.con/vynious/cock-block/node"
	"github.con/vynious/cock-block/proto"
	"google.golang.org/grpc"
)

func main() {

	makeNode(":3000", []string{})
	time.Sleep(time.Second)
	makeNode(":4000", []string{":3000"})
	time.Sleep(2 * time.Second)
	makeNode(":5001", []string{":4000"})
	select {}
}

func makeNode(listenAddr string, bootstrapNodes []string) (*node.Node, error) {
	n := node.NewNode()
	go func ()  {
		if err := n.Start(listenAddr, bootstrapNodes); err != nil {
			log.Println("err: ", err)
		}
	}()

	return n, nil
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	c := proto.NewNodeClient(client)

	version := &proto.Version{
		Version:    "cock-blocker-0.1",
		Height:     1,
		ListenAddr: ":",
	}
	if _, err = c.Handshake(context.TODO(), version); err != nil {
		log.Fatal(err)
	}
}
