package service

import blockcypher "github.com/blockcypher/gobcy"

// Syncer is the interface for syncing account balance.
type Syncer interface {
	Balance(address string) (int, error)
	Coin() string
	Chain() string
}

// NewSyncer return new blockCypher instance.
func NewSyncer(coin, chain, token string) (*blockCypher, error) {
	b := &blockCypher{chain: chain, coin: coin}
	b.api = blockcypher.API{Token: token, Coin: b.coin, Chain: b.chain}

	return b, nil
}

type blockCypher struct {
	chain string
	coin  string
	token string

	api blockcypher.API
}

var _ Syncer = (*blockCypher)(nil)

func (b *blockCypher) Balance(address string) (int, error) {
	addr, err := b.api.GetAddrFull(address, nil)
	if err != nil {
		return 0, err
	}
	return addr.Balance, nil
}

func (b *blockCypher) Coin() string {
	return b.coin
}

func (b *blockCypher) Chain() string {
	return b.chain
}
