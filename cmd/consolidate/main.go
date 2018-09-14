// Consolidate the unspents currently held in a wallet to a smaller number.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marselester/bitgo-v1"
)

func main() {
	baseURL := flag.String("host", "http://0.0.0.0:3080", "BitGo Express API server base URL.")
	accessToken := flag.String("token", "", "BitGo access token.")
	walletID := flag.String("wallet", "", "BitGo wallet ID.")
	walletPassphrase := flag.String("passphrase", "", "Passphrase of the wallet.")
	numUnspentsToMake := flag.Int("target", 1, "Number of outputs created by the consolidation transaction.")
	limit := flag.Int("limit", 85, "Number of unspents to select.")
	minValue := flag.Float64("min-value", 0, "Ignore unspents smaller than this amount of bitcoins.")
	maxValue := flag.Float64("max-value", 0, "Ignore unspents larger than this amount of bitcoins.")
	feeRate := flag.Int("fee-rate", 0, "The desired fee rate for the transaction in satoshis/kilobyte.")
	minConfirms := flag.Int("min-confirms", 0, "The required number of confirmations for each transaction input.")
	maxIter := flag.Int("max-iter", 1, "Maximum number of consolidation iterations to perform.")
	waitIter := flag.Duration("wait-iter", time.Second, "Wait between consolidation iterations.")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Listen to INT/TERM to gracefully stop consolidation.
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan

		log.Print("consolidate: stopping...")
		cancel()
	}()

	client := bitgo.NewClient(
		bitgo.WithBaseURL(*baseURL),
		bitgo.WithAccesToken(*accessToken),
	)
	params := &bitgo.WalletConsolidateParams{
		NumUnspentsToMake: *numUnspentsToMake,
		Limit:             *limit,
		MinConfirms:       *minConfirms,
		WalletPassphrase:  *walletPassphrase,
		MinValue:          bitgo.ToSatoshis(*minValue),
		MaxValue:          bitgo.ToSatoshis(*maxValue),
		FeeRate:           *feeRate,
	}
	for i := 0; i < *maxIter; i++ {
		tt, err := client.Wallet.Consolidate(ctx, *walletID, params)
		// Print consolidated transaction ID.
		if err == nil {
			for _, tx := range tt {
				fmt.Printf("%s\n", tx.TxID)
			}
			time.Sleep(*waitIter)
			continue
		}

		// Stop when a context was cancelled (user hit Ctrl+C).
		if ctx.Err() != nil {
			break
		}

		if apiErr, ok := err.(bitgo.Error); ok {
			log.Fatalf("consolidate: failed to coalesce unspents, %d: %v", apiErr.HTTPStatusCode, apiErr)
		}
		log.Fatalf("consolidate: failed to coalesce unspents: %v", err)
	}
}
