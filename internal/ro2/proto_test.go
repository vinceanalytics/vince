package ro2

import (
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
	err = db.Add(&v1.Model{
		Timestamp: 1,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), db.seq.Load())
	db.Close()

	db, err = Open(dir)
	require.NoError(t, err)
	defer db.Close()
	require.Equal(t, uint64(1), db.seq.Load())
}
