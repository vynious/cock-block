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

func TestSignVerifyBlock(t *testing.T) {
	var (
		block = util.RandomBlock()
		privKey = crypto.GeneratePrivateKey()
		pubKey = privKey.PublicKey()
	)

	sig := SignBlock(privKey, block)
	assert.Equal(t, 64, len(sig.Bytes()))
	assert.True(t, sig.Verify(pubKey, HashBlock(block)))
	assert.Equal(t, block.PublicKey, pubKey.Bytes())
	assert.Equal(t, block.Signature, sig.Bytes())
	assert.True(t, VerifyBlock(block))

	invalidPrivKey := crypto.GeneratePrivateKey()
	block.PublicKey = invalidPrivKey.Bytes()
	assert.False(t, VerifyBlock(block))
}
