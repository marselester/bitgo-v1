# Go client for [BitGo.com API v1](https://bitgo.github.io/bitgo-docs/)

[![Documentation](https://godoc.org/github.com/marselester/bitgo-v1?status.svg)](https://godoc.org/github.com/marselester/bitgo-v1)
[![Go Report Card](https://goreportcard.com/badge/github.com/marselester/bitgo-v1)](https://goreportcard.com/report/github.com/marselester/bitgo-v1)
[![Travis CI](https://travis-ci.org/marselester/bitgo-v1.png)](https://travis-ci.org/marselester/bitgo-v1)

This is unofficial API client. There are no plans to implement all resources.

## [List Wallet Unspents](https://bitgo.github.io/bitgo-docs/#list-wallet-unspents)

Gets a list of unspent input transactions for a wallet. For example, we want to request
`2N91XzUxLrSkfDMaRcwQhe9DauhZMhUoxGr` wallet's unspents `250` per API request and
print amounts in BTC. You can stop pagination by cancelling a `ctx` context.

```go
c := bitgo.NewClient(
    bitgo.WithAccesToken("swordfish"),
)
params := url.Values{}
params.Set("limit", "250")
err := c.Wallet.Unspents(ctx, "2N91XzUxLrSkfDMaRcwQhe9DauhZMhUoxGr", params, func(list *bitgo.UnspentList) {
    for _, utxo := range list.Unspents {
        fmt.Printf("%0.8f\n", bitgo.ToBitcoins(utxo.Value))
    }
})
```

There is a CLI program to list all unspensts of a wallet.

```sh
$ go build ./cmd/utxo/
$ ./utxo -token=swordfish -wallet=2N91XzUxLrSkfDMaRcwQhe9DauhZMhUoxGr -limit=250
0.00000117
0.00000001
0.00000001
0.00000562
0.00000001
0.00000562
```

You can use it to get a rough idea about unspents available in the wallet.

```sh
$ ./utxo -token=swordfish -wallet=2N91XzUxLrSkfDMaRcwQhe9DauhZMhUoxGr > unspents.txt
$ cat unspents.txt | sort | uniq -c | sort -n -r
   3 0.00000001
   2 0.00000562
   1 0.00000117
```

## Error Handling

Dave Cheney recommends
[asserting errors for behaviour](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully), not type.

```go
package main

import (
	"fmt"

	"github.com/marselester/bitgo-v1"
	"github.com/pkg/errors"
)

// IsUnauthorized returns true if err caused by authentication problem.
func IsUnauthorized(err error) bool {
	e, ok := errors.Cause(err).(interface {
		IsUnauthorized() bool
	})
	return ok && e.IsUnauthorized()
}

func main() {
	err := bitgo.Error{Type: bitgo.ErrorTypeAuthentication}
	fmt.Println(IsUnauthorized(err))
	fmt.Println(IsUnauthorized(fmt.Errorf("")))
	fmt.Println(IsUnauthorized(nil))
	// Output: true
	// false
	// false
}
```
