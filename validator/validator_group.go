package validator

import (
	"bytes"
	"math"
	"sort"

	"github.com/herdius/herdius-core/crypto/merkle"
)

// Group ...
type Group struct {
	// NOTE: persisted via reflect, must be exported.
	Validators []*Validator `json:"validators"`

	// cached (unexported)
	totalStakingPower int64
}

//-------------------------------------
// Implements sort for sorting validators by address.

// ValidatorsByAddress : Sort validators by address
type ValidatorsByAddress []*Validator

func (valz ValidatorsByAddress) Len() int {
	return len(valz)
}

func (valz ValidatorsByAddress) Less(i, j int) bool {
	return bytes.Compare(valz[i].Address, valz[j].Address) == -1
}

func (valz ValidatorsByAddress) Swap(i, j int) {
	it := valz[i]
	valz[i] = valz[j]
	valz[j] = it
}

// NewValidatorGroup initializes a Group by copying over the
// values from `valz`, a list of Validators. If valz is nil or empty,
// the new Group will have an empty list of Validators.
func NewValidatorGroup(valz []*Validator) *Group {
	validators := make([]*Validator, len(valz))
	for i, val := range valz {
		validators[i] = val.Copy()
	}
	sort.Sort(ValidatorsByAddress(validators))
	vals := &Group{
		Validators: validators,
	}

	return vals
}

// IsNilOrEmpty : Nil or empty validator sets are invalid.
func (vals *Group) IsNilOrEmpty() bool {
	return vals == nil || len(vals.Validators) == 0
}

// Copy each validator into a new Group
func (vals *Group) Copy() *Group {
	validators := make([]*Validator, len(vals.Validators))
	for i, val := range vals.Validators {
		validators[i] = val.Copy()
	}
	return &Group{
		Validators:        validators,
		totalStakingPower: vals.totalStakingPower,
	}
}

// HasAddress returns true if address given is in the validator set, false -
// otherwise.
func (vals *Group) HasAddress(address []byte) bool {
	idx := sort.Search(len(vals.Validators), func(i int) bool {
		return bytes.Compare(address, vals.Validators[i].Address) <= 0
	})
	return idx < len(vals.Validators) && bytes.Equal(vals.Validators[idx].Address, address)
}

// GetByAddress returns an index of the validator with address and validator
// itself if found. Otherwise, -1 and nil are returned.
func (vals *Group) GetByAddress(address []byte) (index int, val *Validator) {
	idx := sort.Search(len(vals.Validators), func(i int) bool {
		return bytes.Compare(address, vals.Validators[i].Address) <= 0
	})
	if idx < len(vals.Validators) && bytes.Equal(vals.Validators[idx].Address, address) {
		return idx, vals.Validators[idx].Copy()
	}
	return -1, nil
}

// GetByIndex returns the validator's address and validator itself by index.
// It returns nil values if index is less than 0 or greater or equal to
// len(ValidatorSet.Validators).
func (vals *Group) GetByIndex(index int) (address []byte, val *Validator) {
	if index < 0 || index >= len(vals.Validators) {
		return nil, nil
	}
	val = vals.Validators[index]
	return val.Address, val.Copy()
}

// Size returns the length of the validator set.
func (vals *Group) Size() int {
	return len(vals.Validators)
}

// TotalVotingPower returns the sum of the voting powers of all validators.
func (vals *Group) TotalVotingPower() int64 {
	if vals.totalStakingPower == 0 {
		sum := int64(0)
		for _, val := range vals.Validators {
			// mind overflow
			sum = safeAddClip(sum, val.StakingPower)
		}
		// if sum > MaxTotalStakingPower {
		// 	panic(fmt.Sprintf(
		// 		"Total staking power should be guarded to not exceed %v; got: %v",
		// 		MaxTotalStakingPower,
		// 		sum))
		// }
		vals.totalStakingPower = sum
	}
	return vals.totalStakingPower
}

// Hash returns the Merkle root hash build using validators (as leaves) in the
// set.
func (vals *Group) Hash() []byte {
	if len(vals.Validators) == 0 {
		return nil
	}
	bzs := make([][]byte, len(vals.Validators))
	for i, val := range vals.Validators {
		bzs[i] = val.Bytes()
	}
	return merkle.SimpleHashFromByteSlices(bzs)
}

// Add adds val to the validator set and returns true. It returns false if val
// is already in the set.
func (vals *Group) Add(val *Validator) (added bool) {
	val = val.Copy()
	idx := sort.Search(len(vals.Validators), func(i int) bool {
		return bytes.Compare(val.Address, vals.Validators[i].Address) <= 0
	})
	if idx >= len(vals.Validators) {
		vals.Validators = append(vals.Validators, val)
		// Invalidate cache
		vals.totalStakingPower = 0
		return true
	} else if bytes.Equal(vals.Validators[idx].Address, val.Address) {
		return false
	} else {
		newValidators := make([]*Validator, len(vals.Validators)+1)
		copy(newValidators[:idx], vals.Validators[:idx])
		newValidators[idx] = val
		copy(newValidators[idx+1:], vals.Validators[idx:])
		vals.Validators = newValidators
		// Invalidate cache
		vals.totalStakingPower = 0
		return true
	}
}

// Update updates the ValidatorSet by copying in the val.
// If the val is not found, it returns false; otherwise,
// it returns true.
func (vals *Group) Update(val *Validator) (updated bool) {
	index, sameVal := vals.GetByAddress(val.Address)
	if sameVal == nil {
		return false
	}

	vals.Validators[index] = val.Copy()
	// Invalidate cache
	vals.totalStakingPower = 0
	return true
}

// Remove deletes the validator with address. It returns the validator removed
// and true. If returns nil and false if validator is not present in the set.
func (vals *Group) Remove(address []byte) (val *Validator, removed bool) {
	idx := sort.Search(len(vals.Validators), func(i int) bool {
		return bytes.Compare(address, vals.Validators[i].Address) <= 0
	})
	if idx >= len(vals.Validators) || !bytes.Equal(vals.Validators[idx].Address, address) {
		return nil, false
	}
	removedVal := vals.Validators[idx]
	newValidators := vals.Validators[:idx]
	if idx+1 < len(vals.Validators) {
		newValidators = append(newValidators, vals.Validators[idx+1:]...)
	}
	vals.Validators = newValidators
	// Invalidate cache
	vals.totalStakingPower = 0
	return removedVal, true
}

///////////////////////////////////////////////////////////////////////////////
// Safe addition/subtraction

func safeAdd(a, b int64) (int64, bool) {
	if b > 0 && a > math.MaxInt64-b {
		return -1, true
	} else if b < 0 && a < math.MinInt64-b {
		return -1, true
	}
	return a + b, false
}

func safeSub(a, b int64) (int64, bool) {
	if b > 0 && a < math.MinInt64+b {
		return -1, true
	} else if b < 0 && a > math.MaxInt64+b {
		return -1, true
	}
	return a - b, false
}

func safeAddClip(a, b int64) int64 {
	c, overflow := safeAdd(a, b)
	if overflow {
		if b < 0 {
			return math.MinInt64
		}
		return math.MaxInt64
	}
	return c
}

func safeSubClip(a, b int64) int64 {
	c, overflow := safeSub(a, b)
	if overflow {
		if b > 0 {
			return math.MinInt64
		}
		return math.MaxInt64
	}
	return c
}
