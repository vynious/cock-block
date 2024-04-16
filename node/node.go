package node

import (
	"context"
	"log"
	"net"
	"sync"

	"github.con/vynious/cock-block/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Node struct {

	mu *sync.Mutex
	peers map[net.Addr]*grpc.ClientConn
	*proto.UnimplementedNodeServer

}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.None, error) {
	p, _ := peer.FromContext(ctx)
	log.Println("retrieve transaction from peer: ", p)
	return nil, nil
}
