package assert

import "log"

func True(v bool) {
	if !v {
		log.Fatal("assertion failure")
	}
}
