package main

import "github.com/gernest/vince/cmd/run"

//go:generate  capnp compile -Itmp/go-capnp/std -ogo store/store.capnp
func main() {
	run.Main()
}
