// +build debug

package bitgo

import "log"

func debug(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}
