package neo

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/cespare/xxhash/v2"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

// FindRowGroups returns a list of row groups containing indexed column values
func FindRowGroups(idx *v1.Block_Index, lo, hi int64, columns, values []string) (o []int) {
	must.Assert(len(columns) == len(values))(
		"mismatch column / value size",
	)
	if lo > idx.Max || hi < idx.Min {
		return
	}
	b := roaring64.New()
	h := xxhash.New()
out:
	for idx, v := range idx.Groups {
		if v.Min > hi {
			// Found a row group higher than our upper bound. Stop looking for further
			// groups
			break
		}
		b.Clear()
		must.One(b.UnmarshalBinary(v.Bitmap))("failed reading row bitmap")
		for i := range columns {
			h.Reset()
			h.WriteString(columns[i])
			h.WriteString(values[i])
			if !b.Contains(h.Sum64()) {
				continue out
			}
		}
		o = append(o, int(idx))
	}
	return
}
