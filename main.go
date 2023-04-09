package main

import "github.com/gernest/vince/cmd/run"

// You need to clone capnproto.org/go/capnp/v3 package into tmp
// to access the library files before generating.
//
//go:generate  capnp compile -Itmp/go-capnp/std -ogo store/store.capnp
func main() {
	run.Main()
}
