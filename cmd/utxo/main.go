// List bitcoin unspent transaction outputs (UTXOs).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/marselester/bitgo-v1"
)

func main() {
	baseURL := flag.String("host", "http://0.0.0.0:3080", "BitGo API server base URL.")
	accessToken := flag.String("token", "", "BitGo access token.")
	walletID := flag.String("wallet", "", "BitGo wallet ID (BTC address).")
	target := flag.String("target", "", "The API will attempt to return enough unspents to accumulate to at least this amount (in satoshis).")
	minConfirms := flag.String("min-confirms", "", "Only include unspents with at least this many confirmations.")
	minSize := flag.String("min-size", "", "Only include unspents that are at least this many satoshis.")
	limit := flag.String("limit", "", "Max number of results to return in a single call (default=100, max=250).")
	skip := flag.String("skip", "", "The starting index number to list from. Default is 0.")
	flag.Parse()

	client := bitgo.New(
		bitgo.WithBaseURL(*baseURL),
		bitgo.WithAccesToken(*accessToken),
	)

	params := url.Values{}
	if *target != "" {
		params.Set("target", *target)
	}
	if *minConfirms != "" {
		params.Set("minConfirms", *minConfirms)
	}
	if *minSize != "" {
		params.Set("minSize", *minSize)
	}
	if *limit != "" {
		params.Set("limit", *limit)
	}
	if *skip != "" {
		params.Set("skip", *skip)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen to Ctrl+C and kill/killall to gracefully stop listing unspents.
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan

		log.Print("utxo: stopping...")
		cancel()
	}()

	err := client.Wallet.Unspents(ctx, *walletID, params, func(list bitgo.UnspentList) {
		for _, utxo := range list.Unspents {
			fmt.Printf("%0.8f\n", bitgo.ToBitcoins(utxo.Value))
		}
	})
	if err != nil {
		log.Fatalf("utxo: failed to list unspents: %v", err)
	}
}
