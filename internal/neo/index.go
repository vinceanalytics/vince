package neo

import (
	"os"

	"github.com/parquet-go/parquet-go"
	blockv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
)

func IndexBlockFile(f *os.File) (cols map[storev1.Column]*blockv1.ColumnIndex, err error) {
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

	cols = make(map[v1.Column]*blockv1.ColumnIndex)
	groups := r.RowGroups()
	for gi := range groups {
		g := groups[gi]
		chunks := g.ColumnChunks()
		{
			// index  timestamp column
			tidx, ok := cols[storev1.Column_timestamp]
			if !ok {
				tidx = &blockv1.ColumnIndex{}
				cols[storev1.Column_timestamp] = tidx
			}
			ts := chunks[m[storev1.Column_timestamp]]
			idx := ts.ColumnIndex()
			tg := &blockv1.ColumnIndex_RowGroup{}
			for i := 0; i < idx.NumPages(); i++ {
				lo, hi := idx.MinValue(i).Int64(), idx.MaxValue(i).Int64()
				if tg.Min == 0 {
					tg.Min = lo
				}
				if tidx.Min == 0 {
					tidx.Min = lo
				}
				tidx.Min, tidx.Max = min(tidx.Min, lo), max(tidx.Max, hi)
				tg.Min, tg.Max = min(tg.Min, lo), max(tg.Max, hi)
				tg.Pages = append(tg.Pages, &blockv1.ColumnIndex_Page{
					Min: lo,
					Max: hi,
				})
			}
			tidx.RowGroups = append(tidx.RowGroups, tg)
		}
		for i := storev1.Column_browser; i <= storev1.Column_utm_term; i++ {
			idx, ok := cols[i]
			if !ok {
				idx = &blockv1.ColumnIndex{}
				cols[i] = idx
			}
			rg := &blockv1.ColumnIndex_RowGroup{
				BloomFilter: readFilter(chunks[m[i]].BloomFilter()),
			}
			idx.RowGroups = append(idx.RowGroups, rg)
		}
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
