package validator

import (
	"time"

	"github.com/herdius/herdius-core/crypto"
)

// Address is hex bytes.
type Address = crypto.Address

// Vote represents a prevote, precommit, or commit vote from validators for
// consensus.
type Vote struct {
	Height    int64     `json:"height"`
	Timestamp time.Time `json:"timestamp"`
	//BlockID          BlockID       `json:"block_id"` // zero if vote is nil.
	ValidatorAddress Address `json:"validator_address"`
	ValidatorIndex   int     `json:"validator_index"`
	Signature        []byte  `json:"signature"`
}
