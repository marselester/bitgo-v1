package bitgo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// walletService communicates with the wallet API endpoints.
type walletService struct {
	client *Client
}

// Satoshi is the smallest unit of bitcoin.
const Satoshi = 0.00000001

// ToBitcoins converts satoshis to bitcoins.
func ToBitcoins(amount int64) float64 {
	return float64(amount) * Satoshi
}

// ToSatoshis converts bitcoins to satoshis.
func ToSatoshis(amount float64) int64 {
	return int64(amount / Satoshi)
}

// Unspent is an unspent transaction output (UTXO).
type Unspent struct {
	// The address of the unspent input.
	Address string `json:"address"`
	// The hash of the unspent input.
	TxHash string `json:"tx_hash"`
	// The index of the unspent input from tx_hash.
	TxOutputN int `json:"tx_output_n"`
	// The value, in satoshis of the unspent input.
	Value int64 `json:"value"`
	// Output script hash (in hex format).
	Script string `json:"script"`
	// The redeem script.
	RedeemScript string `json:"redeemScript"`
	// The BIP32 path of the unspent output relative to the wallet.
	ChainPath string `json:"chainPath"`
	// Number of blocks seen on and after the unspent transaction was included in a block.
	Confirmations int `json:"confirmations"`
	// Boolean indicating this is an output from a previous spend originating on this wallet,
	// and may be safe to spend even with 0 confirmations.
	IsChange bool `json:"isChange"`
	// Boolean indicating if this unspent can be used to create a BitGo Instant transaction guaranteed against double spends.
	Instant bool `json:"instant"`
}

// ListMeta is a pagination metadata.
type ListMeta struct {
	// Count is a number of records returned in API response, e.g., 2.
	Count int
	// Total is a total number of records, e.g., 5000.
	Total int
	// Start is a starting index number to list from, e.g., 0.
	Start int
}

// UnspentList is a list of unspents as retrieved from a list endpoint.
type UnspentList struct {
	ListMeta
	Unspents []Unspent `json:"unspents"`
}

// Unspents gets a list of unspent transaction outputs (UTXOs) for a wallet.
// It invokes f for each page of results.
// You can filter unspents using query parameters as described in the docs
// https://bitgo.github.io/bitgo-docs/#list-wallet-unspents.
func (s *walletService) Unspents(ctx context.Context, walletID string, queryParams url.Values, f func(*UnspentList)) error {
	path := fmt.Sprintf("wallet/%s/unspents", walletID)
	skip, err := strconv.Atoi(queryParams.Get("skip"))
	if err != nil {
		skip = 0
	}

	for {
		req, err := s.client.NewRequest(ctx, http.MethodGet, path, queryParams, nil)
		if err != nil {
			return err
		}

		v := UnspentList{}
		_, err = s.client.Do(req, &v)
		if err != nil {
			return err
		}
		f(&v)

		skip = skip + v.Count
		stopPagination := skip >= v.Total
		if stopPagination {
			break
		}
		queryParams.Set("skip", strconv.Itoa(skip))
	}

	return nil
}
