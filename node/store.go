package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.con/vynious/cock-block/proto"
	"github.con/vynious/cock-block/types"
)

type BlockStorer interface {
	Put(*proto.Block) error
	Get(string) (*proto.Block, error)
}

type MemoryBlockStore struct {
	blocksLock sync.RWMutex
	blocks     map[string]*proto.Block
}

func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{
		blocks: make(map[string]*proto.Block),
	}
}
func (s *MemoryBlockStore) Put(b *proto.Block) error {
	s.blocksLock.Lock()
	defer s.blocksLock.Unlock()

	hashHex := hex.EncodeToString(types.HashBlock(b))
	s.blocks[hashHex] = b
	return nil

}

func (s *MemoryBlockStore) Get(hash string) (*proto.Block, error) {
	s.blocksLock.RLock()
	defer s.blocksLock.RUnlock()

	block, ok := s.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash [%s] does not exist", hash)
	}
	return block, nil
}
