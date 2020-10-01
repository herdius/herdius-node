package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSyncer(t *testing.T) {
	token := os.Getenv("BLOCKCYPHER_API_TOKEN")
	tests := []struct {
		name    string
		coin    string
		chain   string
		address string
	}{
		{"Bitcoin", "btc", "test3", "n4VQ5YdHf7hLQ2gWQYYrcxoE5B7nWuDFNF"},
		{"Litecoin", "ltc", "main", "ltc1q2z5cunycvxsx8szyz9380mvwskhxu5jk6zm0lj"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := NewSyncer(tc.coin, tc.chain, token)
			require.NoError(t, err)
			assert.Equal(t, tc.coin, b.Coin())
			assert.Equal(t, tc.chain, b.Chain())

			balance, err := b.Balance(tc.address)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, balance, 0)
			if balance > 0 {
				t.Log(balance)
			}
		})
	}
}
