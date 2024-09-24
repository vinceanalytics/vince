package mutex

import (
	"testing"

	"github.com/gernest/rows"
	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/rbf"
)

func TestMutex_extract(t *testing.T) {
	db := rbf.NewDB(t.TempDir(), nil)
	require.NoError(t, db.Open())
	defer db.Close()

	tx, err := db.Begin(true)
	require.NoError(t, err)

	var ls []uint64
	for i := range uint64(5) {
		ls = append(ls, Add(i+1, i))
	}
	tx.Add("i", ls...)
	require.NoError(t, tx.Commit())
	tx, _ = db.Begin(false)
	defer tx.Rollback()

	c, _ := tx.Cursor("i")
	defer c.Close()

	m0 := map[uint64]uint64{}
	require.NoError(t, Extract(c, 0, rows.NewRow(1), func(column, value uint64) error {
		m0[column] = value
		return nil
	}))
	require.Equal(t, map[uint64]uint64{1: 0}, m0)
}
