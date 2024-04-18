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


func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	return pk.Sign(HashBlock(b))
}

func HashHeader(header *proto.Header) []byte{ 
	h, err := pb.Marshal(header)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(h)
	return hash[:]
}