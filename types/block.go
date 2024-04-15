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
	b, err := pb.Marshal(block)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}


func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	return pk.Sign(HashBlock(b))
}