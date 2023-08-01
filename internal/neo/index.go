package neo

import (
	"bytes"
	"context"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/compute"
	"github.com/apache/arrow/go/v13/parquet"
	"github.com/apache/arrow/go/v13/parquet/file"
	"github.com/apache/arrow/go/v13/parquet/pqarrow"
	"github.com/cespare/xxhash/v2"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/blocks"
	"github.com/vinceanalytics/vince/pkg/entry"
)

// FindRowGroups returns a list of row groups containing indexed column values
func FindRowGroups(idx *blocks.Index, columns, values []string) (o []int) {
	must.Assert(len(columns) == len(values))(
		"mismatch column / value size",
	)
	b := roaring64.New()
	h := xxhash.New()
out:
	for idx, v := range idx.Bitmaps {
		b.Clear()
		must.One(b.UnmarshalBinary(v))("failed reading row bitmap")
		for i := range columns {
			h.Reset()
			h.WriteString(columns[i])
			h.WriteString(values[i])
			if !b.Contains(h.Sum64()) {
				continue out
			}
		}
		o = append(o, idx)
	}
	return
}

func Index(ctx context.Context, b []byte) (result blocks.Index) {
	ctx = entry.Context(ctx)
	f := must.Must(file.NewParquetReader(bytes.NewReader(b), file.WithReadProps(parquet.NewReaderProperties(
		entry.Pool,
	))))(
		"failed to open block for indexing",
	)
	r := must.Must(
		pqarrow.NewFileReader(f, pqarrow.ArrowReadProperties{
			Parallel: true,
		}, entry.Pool),
	)("failed to create arrow reader from parquet file ")

	// Indexes are build per row groups. This way we can skip groups that don't
	// contain relevant search.
	bitmap := roaring64.New()
	hash := xxhash.New()
	result.Bitmaps = make([][]byte, 0, f.NumRowGroups())
	for i := 0; i < f.NumRowGroups(); i++ {
		bitmap.Clear()
		hash.Reset()
		g := r.RowGroup(i)
		for k, v := range entry.IndexedColumns {
			hashColumn(ctx, v, hash, bitmap, g.Column(k))
		}
		result.Bitmaps = append(result.Bitmaps, must.Must(
			bitmap.MarshalBinary(),
		)(
			"failed marshalling bitmap",
		))
	}
	return
}

func hashColumn(ctx context.Context, name string, h *xxhash.Digest, b *roaring64.Bitmap, r pqarrow.ColumnChunkReader) {
	chunk := must.Must(
		r.Read(ctx),
	)("failed to read column for hashing")
	uniq := must.Must(
		compute.Unique(ctx, &compute.ChunkedDatum{Value: chunk}),
	)("failed to filter unique entries for hashed column")
	s := uniq.(*compute.ArrayDatum).MakeArray().(*array.String)

	for i := 0; i < s.Len(); i++ {
		h.Reset()
		h.WriteString(name)
		h.WriteString(s.Value(i))
		b.Add(h.Sum64())
	}
	chunk.Release()
	uniq.Release()
	s.Release()
}
