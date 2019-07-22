package service

import (
	"testing"

	"github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/crypto/secp256k1"
	"github.com/herdius/herdius-core/p2p/crypto"
	"github.com/herdius/herdius-core/p2p/network"
	"github.com/stretchr/testify/assert"
)

func TestCreateAndVerifyVote(t *testing.T) {
	privKey := secp256k1.GenPrivKey()

	pubKey := privKey.PubKey()

	keys := &crypto.KeyPair{
		PublicKey:  pubKey.Bytes(),
		PrivateKey: privKey.Bytes(),
		PrivKey:    privKey,
		PubKey:     pubKey,
	}

	builder := network.NewBuilder("dev")
	builder.SetKeys(keys)

	net, err := builder.Build()
	assert.NoError(t, err, "Could not build the network object")

	// Create Dummy Child block
	cbm := &protobuf.ChildBlockMessage{}

	validator := Validator{}
	validatorAddress := "dummy-for-test"
	err = validator.Vote(net, validatorAddress, cbm)
	assert.NoError(t, err, "Failed to create validator vote")

	//Verify Vote
	validatorSign := cbm.Vote.Signature

	isVerified := pubKey.VerifyBytes(cbm.Vote.GetValidator().GetPubKey(), validatorSign)
	assert.True(t, true, isVerified)
	assert.Equal(t, validatorAddress, cbm.Vote.Validator.Address)
}
