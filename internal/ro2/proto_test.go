package ro2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func TestStore_sequence(t *testing.T) {
	dir := t.TempDir()
	db, err := Open(dir)
	require.NoError(t, err)
	// zero sequence at the beginning
	require.Equal(t, uint64(0), db.seq.Load())
	err = db.One(&v1.Model{
		Timestamp: 1,
		Country:   "TZ",
	})
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
		tx.Search(CountryField, nil, func(key []byte, val uint64) {
			country = string(key)
			id = val
		})
		tr = tx.Find(CountryField, 1)
		return nil
	})
	require.Equal(t, uint64(1), db.seq.Load())
	require.Equal(t, "TZ", country)
	require.Equal(t, "TZ", tr)
	require.Equal(t, uint64(1), id)
}

func BenchmarkAddOne(t *testing.B) {
	dir := t.TempDir()
	db, err := Open(dir)
	require.NoError(t, err)
	// zero sequence at the beginning
	require.Equal(t, uint64(0), db.seq.Load())
	m := &v1.Model{
		Timestamp: time.Now().UnixMilli(),
		Country:   "TZ",
	}
	t.ResetTimer()
	t.ReportAllocs()
	for range t.N {
		db.One(m)
	}
}
