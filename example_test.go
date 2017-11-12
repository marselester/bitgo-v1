package bitgo_test

import (
	"fmt"

	"github.com/marselester/bitgo-v1"
)

func ExampleIsUnauthorized() {
	// An error should have IsUnauthorized() bool method.
	fmt.Println(bitgo.IsUnauthorized(
		bitgo.Error{Type: bitgo.ErrorTypeAuthentication},
	))
	// It is ok to pass any error.
	fmt.Println(bitgo.IsUnauthorized(fmt.Errorf("")))
	fmt.Println(bitgo.IsUnauthorized(nil))
	// Output: true
	// false
	// false
}
