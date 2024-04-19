package main

import (
	"context"
	"log"
	"time"

	"github.con/vynious/cock-block/crypto"
	"github.con/vynious/cock-block/node"
	"github.con/vynious/cock-block/proto"
	"github.con/vynious/cock-block/util"
	"google.golang.org/grpc"
)

func main() {

	

	makeNode(":3000", []string{}, true)
	time.Sleep(time.Second)
	makeNode(":4000", []string{":3000"}, false)
	time.Sleep(2 * time.Second)
	makeNode(":5001", []string{":4000"}, false)

	for {
		time.Sleep(100 * time.Millisecond)
		makeTransaction()
	}
}

func makeNode(listenAddr string, bootstrapNodes []string, isValidator bool) (*node.Node, error) {
	cfg := node.ServerConfig{
		Version:    "cock-blocker-1",
		ListenAddr: listenAddr,
	}
	if isValidator {
		cfg.PrivateKey = crypto.GeneratePrivateKey()
	}
	n := node.NewNode(cfg)
	go func() {
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
	privKey := crypto.GeneratePrivateKey()
	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 0,
				PublicKey:    privKey.PublicKey().Bytes(),
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  99,
				Address: privKey.PublicKey().Address().Bytes(),
			},
		},
	}

	if _, err = c.HandleTransaction(context.TODO(), tx); err != nil {
		log.Fatal(err)
	}
}
