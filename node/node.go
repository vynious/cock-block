package node

import (
	"context"
	"log"
	"net"
	"sync"

	"github.con/vynious/cock-block/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Node struct {
	listenAddr string
	version    string
	peerLock   sync.RWMutex
	peers      map[proto.NodeClient]*proto.Version
	logger     *zap.SugaredLogger
	proto.UnimplementedNodeServer
}

func NewNode() *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	lg, _ := loggerConfig.Build()

	return &Node{
		peerLock: sync.RWMutex{},
		peers:    make(map[proto.NodeClient]*proto.Version),
		version:  "cock-blocker-0.1",
		logger: lg.Sugar(),
	}
}

func (n *Node) BootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		c, err := makeNodeClient(addr)
		if err != nil {
			return err
		}
		v, err := c.Handshake(context.TODO(), n.getVersion())
		if err != nil {
			n.logger.Error("handshake error: ", err)
			continue
		}
		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	n.logger.Debugw("new peer connected", "addr", v.ListenAddr, "height", v.Height)
	n.peers[c] = v
}

func (n *Node) removePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "cock-blocker-0.1",
		Height:     0,
		ListenAddr: n.listenAddr,
	}
}

func (n *Node) Start(listenAddr string) error {
	n.listenAddr = listenAddr
	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.listenAddr)
	return grpcServer.Serve(ln)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	// peer, _ := peer.FromContext(ctx)

	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, nil
	}
	n.addPeer(c, v)
	return n.getVersion(), nil
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	p, _ := peer.FromContext(ctx)
	log.Println("retrieve transaction from peer: ", p)
	return &proto.Ack{}, nil
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(c), err
}
