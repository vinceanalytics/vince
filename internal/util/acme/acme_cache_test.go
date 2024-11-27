package acme

import (
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/acme/autocert"
)

func TestCache(t *testing.T) {
	db, err := pebble.Open(t.TempDir(), nil)
	require.NoError(t, err)
	defer db.Close()

	ca := New(db)

	key := "test"

	_, err = ca.Get(nil, key)
	require.Equal(t, autocert.ErrCacheMiss, err)

	require.NoError(t, ca.Put(nil, key, []byte(key)))

	value, err := ca.Get(nil, key)
	require.NoError(t, err)
	require.Equal(t, key, string(value))
}
