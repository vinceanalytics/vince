package ro2

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func TestStore_sequence(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	require.NoError(t, err)
	// zero sequence at the beginning
	require.Equal(t, uint64(0), db.seq.Load())
	db.Buffer(&v1.Model{
		Timestamp: 1,
		Country:   "TZ",
	})
	err = db.Flush()
	require.NoError(t, err)
	require.Equal(t, uint64(1), db.seq.Load())
	db.Close()

	db, err = Open(dir)
	require.NoError(t, err)
	defer db.Close()
	var country string
	var tr string
	var id uint64
	db.View(func(tx *Tx) error {
		tx.searchTranslation(0, CountryField, func(key, val []byte) {
			country = string(key)
			id = binary.BigEndian.Uint64(val)
		})
		tr = tx.Find(CountryField, 1)
		return nil
	})
	require.Equal(t, uint64(1), db.seq.Load())
	require.Equal(t, "TZ", country)
	require.Equal(t, "TZ", tr)
	require.Equal(t, uint64(1), id)
}
