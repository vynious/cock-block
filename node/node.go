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
		logger:   lg.Sugar(),
	}
}

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
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

	// bootstrap the network with list of known remote networks
	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	return grpcServer.Serve(ln)
}

/*
Handshake only works when the node (grpc client) is doing an outbound connection to the other node (grpc server).
The caller of the Handshake is the grpc client and the owner of version, v is the grpc server.
This function returns the version of the version of the node (grpc client)
This method establishes a connection and exchanges version information.
*/
func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
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

// bootstrapNetwork initializes the network by connecting to the provided addresses and adding peers.
//
// Parameters:
// - addrs: a slice of strings representing the addresses to connect to.
// Returns an error if the connection or handshake fails.
func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		
		if !n.canConnectionWith(addr) {
			continue
		}
		
		n.logger.Debugw("dialing remote peer", "we", n.listenAddr, "remote", addr)
		c, v, err := n.dialRemoteNode(addr)
		if err != nil {
			return err
		}
		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) canConnectionWith(addr string) bool {
	if n.listenAddr == addr {
		return false
	}

	connectedPeers := n.getPeerList()
	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}
	return true
}

// addPeer adds the node A (grpc node client) to the peer list of node (grpc node server)
// as well as its peer list from the version
func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	// handle the logic to decide whether to accept/drop the incoming node connection
	n.peers[c] = v
	

	// connect to all peers into the received lists of peers (remote nodes)
	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}

	n.logger.Debugw("new peer connected", "we", n.listenAddr, "addr", v.ListenAddr, "height", v.Height)

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
		PeerList:   n.getPeerList(),
	}
}

// dialRemoteNode initiates a connection to a remote node using the provided address.
//
// Parameters:
// - addr: the address of the remote node to connect to.
// Returns:
// - proto.NodeClient: the client connection to the remote node.
// - *proto.Version: the version information received after a successful connection.
// - error: an error if the connection or handshake fails.
func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}
	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}
	return c, v, nil
}

// getPeerList interates over the current peers of the node
// and get their listen address (grpc client) through their version (version.ListenAddr)
func (n *Node) getPeerList() []string {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	peers := []string{}
	for _, version := range n.peers {
		peers = append(peers, version.ListenAddr)
	}
	return peers
}
