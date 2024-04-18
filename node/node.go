package node

import (
	"context"

	"encoding/hex"
	"net"
	"sync"
	"time"

	"github.con/vynious/cock-block/crypto"
	"github.con/vynious/cock-block/proto"
	"github.con/vynious/cock-block/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const (
	blockTime = time.Second * 5
)

type Mempool struct {
	txx map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txx: make(map[string]*proto.Transaction),
	}
}

// Has checks if the mempool has the transaction inside
func (pool *Mempool) Has(tx *proto.Transaction) bool {
	hash := hex.EncodeToString(types.HashTransaction(tx))
	_, ok := pool.txx[hash]
	return ok
}

func (pool *Mempool) Add(tx *proto.Transaction) bool {

	if pool.Has(tx) {
		return false
	}

	hash := hex.EncodeToString(types.HashTransaction(tx))
	pool.txx[hash] = tx
	return true
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

type Node struct {
	ServerConfig
	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	mempool  *Mempool
	logger   *zap.SugaredLogger
	proto.UnimplementedNodeServer
}

func NewNode(cfg ServerConfig) *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	lg, _ := loggerConfig.Build()

	return &Node{
		peerLock:     sync.RWMutex{},
		peers:        make(map[proto.NodeClient]*proto.Version),
		mempool:      NewMempool(),
		logger:       lg.Sugar(),
		ServerConfig: cfg,
	}
}

// Start starts the Node server on the specified listen address and bootstraps the network with the provided list of bootstrap nodes.
//
// Parameters:
// - listenAddr: the address to listen on for incoming connections.
// - bootstrapNodes: a list of addresses of remote nodes to bootstrap the network connection.
//
// Returns:
// - error: an error if the server fails to start or bootstrap the network.
func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.ListenAddr = listenAddr
	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.ListenAddr)

	// bootstrap the network with list of known remote networks
	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	if n.PrivateKey != nil {
		go n.validatorLoop()
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
	hash := hex.EncodeToString(types.HashTransaction(tx))

	if n.mempool.Add(tx) {
		n.logger.Debugw("received tx", "from", p.Addr, "hash", hash)

		go func() {
			if err := n.broadcast(tx); err != nil {
				n.logger.Errorw("broadcast error", "err", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

func (n *Node) validatorLoop() {
	n.logger.Infow("starting validator loop", "pubKey", n.PrivateKey.PublicKey(), "blockTime", blockTime)
	ticker := time.NewTicker(blockTime)
	for {
		<-ticker.C
		n.logger.Debugw("time to create a new block", "lenTx", len(n.mempool.txx))

		for hash, _ := range n.mempool.txx {
			delete(n.mempool.txx, hash)
		}

	}
}

func (n *Node) broadcast(msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			if _, err := peer.HandleTransaction(context.Background(), v); err != nil {
				return err
			}
		}
	}
	return nil
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

		if !n.canConnectWith(addr) {
			continue
		}

		n.logger.Debugw("dialing remote peer", "we", n.ListenAddr, "remote", addr)
		c, v, err := n.dialRemoteNode(addr)
		if err != nil {
			return err
		}
		n.addPeer(c, v)
	}
	return nil
}

// canConnectWith checks if the address of the remote node is already connected with
// or if it is the same as the current address
func (n *Node) canConnectWith(addr string) bool {
	if n.ListenAddr == addr {
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

	n.logger.Debugw("new peer connected", "we", n.ListenAddr, "addr", v.ListenAddr, "height", v.Height)

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
		ListenAddr: n.ListenAddr,
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
