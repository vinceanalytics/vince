package neo

import (
	"os"

	"github.com/parquet-go/parquet-go"
	blockv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
)

func IndexBlockFile(f *os.File) (b *blockv1.BlockIndex, err error) {
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	r, err := parquet.OpenFile(f, stat.Size())
	if err != nil {
		return nil, err
	}
	m := make(map[storev1.Column]int)
	schema := r.Schema()
	for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
		n, _ := schema.Lookup(i.String())
		m[i] = n.ColumnIndex
	}
	groups := r.RowGroups()
	b = &blockv1.BlockIndex{
		Bloom:     make([]*blockv1.BlockIndex_Bloom, 0, len(groups)),
		TimeRange: make([]*blockv1.BlockIndex_Range, 0, len(groups)),
	}

	for _, g := range r.RowGroups() {
		chunks := g.ColumnChunks()
		ts := chunks[m[storev1.Column_timestamp]]
		idx := ts.ColumnIndex()
		bf := &blockv1.BlockIndex_Bloom{
			Filters: make(map[string][]byte),
		}
		// we only save filters for string fields
		for i := storev1.Column_browser; i <= storev1.Column_utm_term; i++ {
			bf.Filters[i.String()] = readFilter(chunks[m[i]].BloomFilter())
		}
		b.Bloom = append(b.Bloom, bf)
		b.TimeRange = append(b.TimeRange, &blockv1.BlockIndex_Range{
			Min: idx.MinValue(0).Int64(),
			Max: idx.MaxValue(idx.NumPages() - 1).Int64(),
		})
	}
	return
}

func readFilter(b parquet.BloomFilter) []byte {
	if b == nil {
		return nil
	}
	o := make([]byte, b.Size())
	b.ReadAt(o, 0)
	return o
}
