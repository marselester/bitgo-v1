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
	"time"

	"github.com/marselester/bitgo-v1"
)

func main() {
	baseURL := flag.String("host", "http://0.0.0.0:3080", "BitGo API server base URL.")
	accessToken := flag.String("token", "", "BitGo access token.")
	walletID := flag.String("wallet", "", "BitGo wallet ID (BTC address).")
	target := flag.Float64("target", 0, "The API will attempt to return enough unspents to accumulate to at least this amount of bitcoins.")
	minConfirms := flag.String("min-confirms", "", "Only include unspents with at least this many confirmations.")
	minSize := flag.Float64("min-size", 0, "Only include unspents that are at least this many bitcoins.")
	limit := flag.String("limit", "", "Max number of results to return in a single call (default=100, max=250).")
	skip := flag.String("skip", "", "The starting index number to list from. Default is 0.")
	segwit := flag.Bool("segwit", true, "Include SegWit unspents.")
	waitSeconds := flag.Int("wait", 15, "How many seconds to wait after failed download attempt.")
	flag.Parse()

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

	client := bitgo.NewClient(
		bitgo.WithBaseURL(*baseURL),
		bitgo.WithAccesToken(*accessToken),
	)

	params := url.Values{}
	if *target > 0 {
		params.Set("target", fmt.Sprintf("%d", bitgo.ToSatoshis(*target)))
	}
	if *minConfirms != "" {
		params.Set("minConfirms", *minConfirms)
	}
	if *minSize > 0 {
		params.Set("minSize", fmt.Sprintf("%d", bitgo.ToSatoshis(*minSize)))
	}
	if *limit != "" {
		params.Set("limit", *limit)
	}
	if *skip != "" {
		params.Set("skip", *skip)
	}
	if *segwit {
		params.Set("segwit", "true")
	} else {
		params.Set("segwit", "false")
	}

	downloaded := 0
	for {
		err := client.Wallet.Unspents(ctx, *walletID, params, func(list *bitgo.UnspentList) {
			downloaded = list.Start + list.Count
			log.Printf("utxo: fetched %d/%d unspents", downloaded, list.Total)

			for _, utxo := range list.Unspents {
				fmt.Printf("%0.8f\n", bitgo.ToBitcoins(utxo.Value))
			}
		})
		// Stop when we downloaded everything without errors or
		// when a context was cancelled (user hit Ctrl+C).
		if err == nil || ctx.Err() != nil {
			break
		}

		if apiErr, ok := err.(bitgo.Error); ok {
			log.Printf("utxo: failed to list unspents, %d: %v", apiErr.HTTPStatusCode, apiErr)
		} else {
			log.Printf("utxo: failed to list unspents: %v", err)
		}

		// We shall wait a bit and then try again.
		log.Printf("utxo: retrying in %d seconds...", *waitSeconds)
		time.Sleep(time.Duration(*waitSeconds) * time.Second)
		params.Set("skip", fmt.Sprintf("%d", downloaded))
	}
}
