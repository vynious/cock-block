package node

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.con/vynious/cock-block/crypto"
	"github.con/vynious/cock-block/proto"
	"github.con/vynious/cock-block/types"
	"github.con/vynious/cock-block/util"
)

func randomBlock(t *testing.T, chain *Chain) *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	block := util.RandomBlock()
	prevBlock, err := chain.GetBlockByHeight(chain.Height())
	require.Nil(t, err)
	block.Header.PrevHash = types.HashBlock(prevBlock)
	types.SignBlock(privKey, block)
	return block

}
func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
	require.Equal(t, 0, chain.Height())
	_, err := chain.GetBlockByHeight(0)
	require.Nil(t, err)

}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
	for i := 0; i < 100; i++ {
		block := randomBlock(t, chain)
		blockHash := types.HashBlock(block)
		require.Nil(t, chain.AddBlock(block))
		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i + 1)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlockByHeight)
	}

}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
	for i := 0; i < 1; i++ {
		block := randomBlock(t, chain)
		require.Nil(t, chain.AddBlock(block))
		require.Equal(t, chain.Height(), i+1)
	}
}

func TestAddBlockWithTx(t *testing.T) {

	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
		block     = randomBlock(t, chain)
		privKey   = crypto.GenerateNewPrivateKeyFromSeedStr(godSeed)
		recipient = crypto.GeneratePrivateKey().PublicKey().Address().Bytes()
	)

	ftt, err := chain.txStore.Get("1b6adb1c46cecb0a6f4ab1945a98d3dc3df539ec5a25500f9615c9a2f935002e")
	assert.Nil(t, err)
	// fmt.Println(ftt)

	inputs := []*proto.TxInput{
		{
			PrevTxHash:   types.HashTransaction(ftt),
			PrevOutIndex: 0,
			PublicKey:    privKey.PublicKey().Bytes(),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Amount:  100,
			Address: recipient,
		},
		{
			Amount:  900,
			Address: privKey.PublicKey().Address().Bytes(),
		},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}
	block.Transactions = append(block.Transactions, tx)
	require.Nil(t, chain.AddBlock(block))

	txHash := hex.EncodeToString(types.HashTransaction(tx))

	fetchedTx, err := chain.txStore.Get(txHash)
	assert.Nil(t, err)
	assert.Equal(t, tx, fetchedTx)
}
