package indexer

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/buffers"
	"github.com/vinceanalytics/vince/internal/cluster/events"
	"github.com/vinceanalytics/vince/internal/columns"
	"github.com/vinceanalytics/vince/internal/filters"
	"github.com/vinceanalytics/vince/internal/index"
)

func TestIndexer(t *testing.T) {
	now, _ := time.Parse(time.RFC822, time.RFC822)
	now = now.UTC()
	ls := events.Samples(events.WithNow(func() time.Time { return now }))
	r := events.New(memory.DefaultAllocator).Write(ls)
	defer r.Release()
	idx := New()
	fidx, err := idx.Index(r)
	require.NoError(t, err)
	full := fidx.(*FullIndex)
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
	t.Run("Match", func(t *testing.T) {
		type Case struct {
			d string
			f []*filters.CompiledFilter
			w []uint32
		}

		cases := []Case{
			{d: "Single column exact match", f: []*filters.CompiledFilter{
				{Column: columns.Domain, Value: []byte("vinceanalytics.com")},
			}, w: []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
			{d: "Multi column exact match", f: []*filters.CompiledFilter{
				{Column: columns.Domain, Value: []byte("vinceanalytics.com")},
				{Column: "browser", Value: []byte("Chrome Webview")},
			}, w: []uint32{5, 8}},
		}
		for _, v := range cases {
			t.Run(v.d, func(t *testing.T) {
				o := new(roaring.Bitmap)
				for i := 0; i < int(full.rows); i++ {
					o.Add(uint32(i))
				}
				full.Match(o, v.f)
				require.Equal(t, v.w, o.ToArray())
			})
		}
	})
}
