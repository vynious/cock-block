package crypto

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.Equal(t, len(privKey.Bytes()), privKeyLen)
	pubKey := privKey.PublicKey()
	assert.Equal(t, len(pubKey.Bytes()), pubKeyLen)
}

func TestPrivateKeySign(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	msg := []byte("hello")
	sig := privKey.Sign(msg)
	assert.Equal(t, len(sig.value), ed25519.SignatureSize)
	assert.True(t, sig.Verify(pubKey, msg))

	// invalid message
	assert.False(t, sig.Verify(pubKey, []byte("scammed")))

	// invalid public key, because msg was signed with a different private key
	invalidPrivKey := GeneratePrivateKey()
	invalidPubKey := invalidPrivKey.PublicKey()
	assert.False(t, sig.Verify(invalidPubKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	// fmt.Println(privKey)
	pubKey := privKey.PublicKey()
	// fmt.Println(pubKey)
	addr := pubKey.Address()
	// fmt.Println(addr)
	assert.Equal(t, len(addr.Bytes()), addressLen)
}