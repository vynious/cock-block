package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.con/vynious/cock-block/crypto"
	"github.con/vynious/cock-block/util"
)




func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hashedBlock := HashBlock(block)
	assert.Equal(t, 32, len(hashedBlock))
}

func TestSignBlock(t *testing.T) {
	var (
		block = util.RandomBlock()
		privKey = crypto.GeneratePrivateKey()
		pubKey = privKey.PublicKey()
	)

	sig := SignBlock(privKey, block)
	assert.Equal(t, 64, len(sig.Bytes()))
	assert.True(t, sig.Verify(pubKey, HashBlock(block)))
}