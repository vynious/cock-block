package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
)

/*
generate private key
get public key from private key using
sign message with private key => signature
verify signature with public key generated from private key
*/

const (
	PrivKeyLen = 64
	PubKeyLen  = 32
	SeedLen    = 32
	AddressLen = 20
	SigLen     = 64
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, SeedLen)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		log.Panic("failed to read random seed", err)
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func GenerateNewPrivateKeyFromSeedStr(seed string) *PrivateKey {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		log.Panic(err)
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seedBytes),
	}
}

func GenerateNewPrivateKeyFromSeed(seed []byte) *PrivateKey {
	if len(seed) != SeedLen {
		log.Panic("invalid seed length must be 32")
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func (p *PrivateKey) Bytes() []byte {
	return p.key
}

func (p *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		value: ed25519.Sign(p.key, msg),
	}
}

func (p *PrivateKey) PublicKey() *PublicKey {
	b := make([]byte, PubKeyLen)
	copy(b, p.key[32:])
	return &PublicKey{key: b}
}

func (p *PrivateKey) String() string {
	return hex.EncodeToString(p.Bytes())
}

type PublicKey struct {
	key ed25519.PublicKey
}

func PublicKeyFromBytes(b []byte) *PublicKey {
	if len(b) != PubKeyLen {
		log.Panic("length of the bytes not equals to 32")
	}
	return &PublicKey{
		key: b,
	}
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

func (p *PublicKey) Address() *Address {
	return &Address{
		value: p.key[len(p.key)-AddressLen:],
	}
}

func (p *PublicKey) String() string {
	return hex.EncodeToString(p.Bytes())
}

type Signature struct {
	value []byte
}

func SignatureFromBytes(b []byte) *Signature {
	if len(b) != SigLen {
		log.Panic("length of the bytes not equals to 64")
	}
	return &Signature{
		value: b,
	}
}

func (s *Signature) Bytes() []byte {
	return s.value
}

func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

type Address struct {
	value []byte
}

func AddressFromBytes(b []byte) Address {
	if len(b) != AddressLen {
		panic("length of the address not equals 20")
	}
	return Address{
		value: b,
	}
}

func (a *Address) Bytes() []byte {
	return a.value
}
func (a *Address) String() string {
	return hex.EncodeToString(a.Bytes())
}
