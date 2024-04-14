package types

import (
	"crypto/sha256"
	"log"

	pb "google.golang.org/protobuf/proto"
	"github.con/vynious/cock-block/proto"
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

