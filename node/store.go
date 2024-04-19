package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.con/vynious/cock-block/proto"
	"github.con/vynious/cock-block/types"
)

type TxStorer interface {
	Put(*proto.Transaction) error
	Get(string) (*proto.Transaction, error)
}


type MemoryTxStore struct {
	lock sync.RWMutex
	txx map[string]*proto.Transaction
}

func NewMemoryTxStore() *MemoryTxStore {
	return &MemoryTxStore {
		txx: make(map[string]*proto.Transaction),
	}
}

func (s *MemoryTxStore) Put(tx *proto.Transaction) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	hash := hex.EncodeToString((types.HashTransaction(tx)))
	s.txx[hash] = tx
	return nil
}

func (s *MemoryTxStore) Get(hash string) (*proto.Transaction, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	tx, ok := s.txx[hash]
	if !ok {
		return nil, fmt.Errorf("cannot find tx with hash %s",hash)
	}
	return tx, nil
}



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
