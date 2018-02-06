package bitgo_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/marselester/bitgo-v1"
)

func TestUnspents(t *testing.T) {
	filename := filepath.Join("testdata", "unspents.json")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	want := bitgo.Unspent{
		Address:       "2N26EdwtVNQe6P9QkVgLHGhoWtU5W98ohNB",
		TxHash:        "3246b59fcec99c81e5f59522327b632f5c54e4da42ccb512550ed91a3f9b5ce6",
		TxOutputN:     0,
		Value:         78273186932,
		Script:        "a9146105ee32b12a94436f19592e18b135d206e5f46987",
		RedeemScript:  "522102f90f2bb90f6572af7bf5c7317ebd48311b417b005352ae71c3c79990fea1f60f2102f817f403092d09abbbb955410d1e50fca4d1ee56e145a29dde01e505558dec43210307527a3928d2711212730ef6585d1a82af80d1fe2979e167b7cc1a397c654ba253ae",
		ChainPath:     "/1/117",
		Confirmations: 3474,
		IsChange:      true,
		Instant:       false,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	client := bitgo.NewClient(
		bitgo.WithBaseURL(srv.URL),
	)
	err = client.Wallet.Unspents(context.Background(), "", nil, func(list *bitgo.UnspentList) {
		got := list.Unspents[0]
		if got != want {
			t.Fatalf("should be %#v, not %#v", want, got)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkUnspents(b *testing.B) {
	filename := filepath.Join("testdata", "unspents.json")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		b.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	client := bitgo.NewClient(
		bitgo.WithBaseURL(srv.URL),
	)
	f := func(list *bitgo.UnspentList) {}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Wallet.Unspents(context.Background(), "", nil, f)
	}
}
