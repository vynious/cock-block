package node

import (
	"encoding/hex"
	"fmt"
	"log"

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

func NewChain(bs BlockStorer) *Chain {
	return &Chain{
		blockStore: bs,
		headers:    NewHeaderList(),
	}
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {

	// add headers of block to list of headers
	c.headers.Add(b.Header)

	// validation for block
	return c.blockStore.Put(b)
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
