package indexer

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/buffers"
	"github.com/vinceanalytics/vince/internal/closter/events"
	"github.com/vinceanalytics/vince/internal/index"
)

func TestIndexer(t *testing.T) {
	now, _ := time.Parse(time.RFC822, time.RFC822)
	now = now.UTC()
	ls := events.Samples(events.WithNow(func() time.Time { return now }))
	r := events.New(memory.DefaultAllocator).Write(ls)
	defer r.Release()
	idx := New()
	full, err := idx.Index(r)
	require.NoError(t, err)

	t.Run("Sets min,max timestamp", func(t *testing.T) {
		lo := time.UnixMilli(int64(full.Min())).UTC()
		require.Truef(t, now.Equal(lo), "now=%v lo=%v", now, lo)

		up := time.UnixMilli(ls.Items[len(ls.Items)-1].Timestamp)
		hi := time.UnixMilli(int64(full.Max())).UTC()
		require.Truef(t, up.Equal(hi), "now=%v hi=%v", up, hi)
	})

	t.Run("Serialize", func(t *testing.T) {
		id := "full_index"
		b := buffers.Bytes()
		defer b.Release()
		err := index.WriteFull(b, full, id)
		require.NoError(t, err)

		fidx, err := index.NewFileIndex(bytes.NewReader(b.Bytes()))
		require.NoError(t, err)
		err = full.Columns(func(column index.Column) error {
			f, err := fidx.Get(column.Name())
			if err != nil {
				return err
			}
			if !f.Equal(column) {
				return fmt.Errorf("mismatch %q column", column.Name())
			}
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Serialize format", func(t *testing.T) {
		id := "full_index"
		b := buffers.Bytes()
		defer b.Release()
		err := index.WriteFull(b, full, id)
		require.NoError(t, err)
		// os.WriteFile("testdata/"+id, b.Bytes(), 0600)
		data, err := os.ReadFile("testdata/" + id)
		require.NoError(t, err)
		require.True(t, bytes.Equal(b.Bytes(), data))
	})
}
