// Package oracle stores global truths of the whole runnning system.
package oracle

import "sync/atomic"

var (
	//Records tracks total records stored in the database. This only accounts for
	// records already in the database and excludes records in the batch ingester.
	Records atomic.Uint64

	Listen string

	DataPath string

	Acme struct {
		Email   string
		Domain  string
		Enabled bool
	}

	Endpoint string
	Demo     string

	Profile bool
)
