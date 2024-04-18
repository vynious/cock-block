package types

import (
	"crypto/sha256"
	"log"

	"github.con/vynious/cock-block/crypto"
	"github.con/vynious/cock-block/proto"
	pb "google.golang.org/protobuf/proto"
)

// HashBlock returns a SHA256 of the header.
func HashBlock(block *proto.Block) []byte {

	return HashHeader(block.Header)
}

func VerifyBlock(b *proto.Block) bool {
	if len(b.PublicKey) != crypto.PubKeyLen {
		return false
	}
	if len(b.Signature) != crypto.SigLen {
		return false
	}
	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey)
	hash := HashBlock(b)
	return sig.Verify(pubKey, hash)
}


func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	hash := HashBlock(b)
	sig := pk.Sign(hash)
	b.PublicKey = pk.PublicKey().Bytes()
	b.Signature = sig.Bytes()
	return sig
}

func HashHeader(header *proto.Header) []byte{ 
	h, err := pb.Marshal(header)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(h)
	return hash[:]
}