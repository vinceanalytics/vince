package ro2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func TestRange(t *testing.T) {
	db, err := newDB(t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	ts, _ := time.Parse(time.RFC822, time.RFC822)
	b := roaring64.New()
	ts = ts.UTC()
	for i := range 5 {
		ro.BSI(b, uint64(i), ts.Add(time.Duration(i)*time.Hour).UnixMilli())
	}
	err = db.Update(func(tx *Tx) error {
		return tx.Add(0, 0, b)
	})
	require.NoError(t, err)
	var all *roaring64.Bitmap
	db.View(func(tx *Tx) error {
		all = tx.Cmp(0, 0, roaring64.RANGE,
			ts.UnixMilli(),
			ts.Add(5*time.Hour).UnixMilli(),
		)
		return nil
	})
	require.Equal(t, []uint64{0, 1, 2, 3, 4}, all.ToArray())

	var subset *roaring64.Bitmap
	db.View(func(tx *Tx) error {
		subset = tx.Cmp(0, 0, roaring64.RANGE,
			ts.Add(1*time.Hour).UnixMilli(),
			ts.Add(2*time.Hour).UnixMilli(),
		)
		return nil
	})
	require.Equal(t, []uint64{1, 2}, subset.ToArray())
}
