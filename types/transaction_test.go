package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.con/vynious/cock-block/crypto"
	"github.con/vynious/cock-block/proto"
	"github.con/vynious/cock-block/util"
)

/*
current balance : 100

-- want to send 5 coins to "AAA" ---

input: TxInput {}, output: TxOutput {}

2 outputs
5 to the destination address
95 to our address

*/

func TestNewTransaction(t *testing.T) {
	fromPrivKey := crypto.GeneratePrivateKey()
	fromPubKey := fromPrivKey.PublicKey()
	fromAddress := fromPrivKey.PublicKey().Address()


	toPrivKey := crypto.GeneratePrivateKey()
	toAddress := toPrivKey.PublicKey().Address()

	input := &proto.TxInput{
		PrevTxHash: util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey: fromPubKey.Bytes(),
	}

	// sending all the balance assuming 100 tokens.

	// amount to send to destination
	output1 := &proto.TxOutput{
		Amount: 5,
		Address: toAddress.Bytes(),
	}

	// amount to send back to myself
	output2 := &proto.TxOutput{
		Amount: 95,
		Address: fromAddress.Bytes(),
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}
	
	// sign the transaction with the private key
	sig := SignTransaction(fromPrivKey, tx)

	// set the signature of the transaction on the input
	input.Signature = sig.Bytes()

	assert.True(t, VerifyTransaction(tx))
}