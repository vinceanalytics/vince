package be

import (
	"testing"

	"github.com/dgraph-io/badger/v4"
)

func TestNew(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("").
		WithInMemory(true).
		WithLoggingLevel(10),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	se, err := NewSession(&Store{db: db})
	if err != nil {
		t.Fatal(err)
	}
	se.Close()
}
