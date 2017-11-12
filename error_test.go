package bitgo_test

import (
	"errors"
	"testing"

	"github.com/marselester/bitgo-v1"
)

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{bitgo.Error{Type: bitgo.ErrorTypeAuthentication}, true},
		{bitgo.Error{Type: bitgo.ErrorTypeInvalidRequest}, false},
		{bitgo.Error{Type: bitgo.ErrorTypeRateLimit}, false},
		{bitgo.Error{Type: bitgo.ErrorTypeAPI}, false},
		{bitgo.Error{}, false},
		{errors.New(""), false},
		{nil, false},
	}
	for _, test := range tests {
		got := bitgo.IsUnauthorized(test.err)
		if got != test.want {
			t.Errorf("IsUnauthorized(%#v) = %v, want %v", test.err, got, test.want)
		}
	}
}
