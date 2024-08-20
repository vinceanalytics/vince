package ro2

import (
	"fmt"
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
	})
	err = db.Flush()
	require.NoError(t, err)
	require.Equal(t, uint64(1), db.seq.Load())
	db.Close()

	db, err = Open(dir)
	require.NoError(t, err)
	defer db.Close()
	db.View(func(tx *Tx) error {
		tx.ExtractBSI(0, timestampField, nil, func(row uint64, c int64) {
			fmt.Println(row, c)
		})
		return nil
	})
	require.Equal(t, uint64(1), db.seq.Load())
}
