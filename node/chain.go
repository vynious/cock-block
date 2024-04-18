package node

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.con/vynious/cock-block/crypto"
	"github.con/vynious/cock-block/proto"
	"github.con/vynious/cock-block/types"
)

// HeaderList stores a list of the hashed headers inside the Block Store
type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: []*proto.Header{},
	}
}

func (list *HeaderList) Add(header *proto.Header) error {
	list.headers = append(list.headers, header)
	return nil
}

func (list *HeaderList) Get(index int) *proto.Header {
	if index > list.Height() {
		log.Panic("index too high")
	}
	return list.headers[index]
}

func (list *HeaderList) Height() int {
	return list.Len() - 1
}

func (list *HeaderList) Len() int {
	return len(list.headers)
}

type Chain struct {
	blockStore BlockStorer
	headers    *HeaderList
}

// NewChain creates a new Chain object with the given BlockStorer. 
// First block will be the genesis block
//
// Parameters:
// - bs: a BlockStorer implementation used to store and retrieve blocks.
//
// Returns:
// - *Chain: a pointer to the newly created Chain object.
func NewChain(bs BlockStorer) *Chain {
	chain := &Chain{
		blockStore: bs,
		headers:    NewHeaderList(),
	}
	chain.addBlock(createGenesisBlock())
	return chain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	// validate new block
	if err := c.ValidateBlock(b); err != nil {
		return err
	}
	return c.addBlock(b)
}

// GetBlockByHeight retrieves a block from the chain by its height.
//
// Parameters:
// - height: an integer representing the height of the block to retrieve.
//
// Returns:
// - *proto.Block: a pointer to the block at the given height.
// - error: an error if the height is too high or if the block retrieval fails.
func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if c.Height() < height {
		return nil, fmt.Errorf("given height (%d) too high - height (%d)", height, c.Height())
	}
	header := c.headers.Get(height)
	hash := types.HashHeader(header)
	return c.GetBlockByHash(hash)
}

// GetBlockByHash retrieves a block from the block store using its hash.
//
// Parameters:
// - hash: a byte slice representing the hash of the block.
//
// Returns:
// - *proto.Block: a pointer to the block with the given hash, or nil if the block does not exist.
// - error: an error if the block store fails to retrieve the block.
func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	// validate the signature of the block
	// validate is the prevHash is the actual hash of the current block.
	if !types.VerifyBlock(b) {
		return fmt.Errorf("invalid block signature")
	}
	
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	hash := types.HashBlock(currentBlock)

	if !bytes.Equal(hash, b.Header.PrevHash) {
		return fmt.Errorf("invalid previous block hash")
	}
	return nil
}

// createGenesisBlock generates the genesis block for the blockchain.
//
// It generates a private key, creates a new block with version 1, signs the block
// with the private key, and returns the generated block.
//
// Returns:
// - *proto.Block: The generated genesis block.
func createGenesisBlock() *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}
	types.SignBlock(privKey, block)
	return block
}

func (c *Chain) addBlock(b *proto.Block) error {
	// add headers of block to list of headers
	c.headers.Add(b.Header)

	// validation for block
	return c.blockStore.Put(b)
}
