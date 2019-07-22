package validator

import (
	"fmt"

	"github.com/herdius/herdius-core/crypto"
	"github.com/herdius/herdius-core/crypto/herhash"
)

type Validator struct {
	Address      Address       `json:"address"`
	PubKey       crypto.PubKey `json:"pub_key"`
	StakingPower int64         `json:"staking_power"`
}

func NewValidator(pubKey crypto.PubKey, stakingPower int64) *Validator {
	return &Validator{
		Address:      pubKey.Address(),
		PubKey:       pubKey,
		StakingPower: stakingPower,
	}
}

// Copy : Creates a new copy of the validator so we can mutate validator properties.
// Panics if the validator is nil.
func (v *Validator) Copy() *Validator {
	vCopy := *v
	return &vCopy
}

func (v *Validator) String() string {
	if v == nil {
		return "nil-Validator"
	}
	return fmt.Sprintf("Validator{%v %v SP:%v}",
		v.Address,
		v.PubKey,
		v.StakingPower)
}

// Hash computes the unique ID of a validator with a given voting power.
func (v *Validator) Hash() []byte {
	return herhash.Sum(v.Bytes())
}

// Bytes computes the unique encoding of a validator with a given staking power.
// These are the bytes that gets hashed in consensus. It excludes address
// as its redundant with the pubkey.
func (v *Validator) Bytes() []byte {
	return cdcEncode(struct {
		PubKey       crypto.PubKey
		StakingPower int64
	}{
		v.PubKey,
		v.StakingPower,
	})
}
