package mutex

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
)

const (
	// shardVsContainerExponent is the power of 2 of ShardWith minus the power
	// of two of roaring container width (which is 16).
	// 2^shardVsContainerExponent is the number of containers in a shard row.
	//
	// It is represented in this rather awkward way because calculating the row
	// which a given container is in means dividing by the number of rows per
	// container which is performantly expressed as a right shift by this
	// exponent.
	shardVsContainerExponent = shardwidth.Exponent - 16
)

func Distinct(c *rbf.Cursor, o *roaring64.Bitmap, filters *rows.Row) error {
	fragData := c.Iterator()

	var filterBitmap *roaring.Bitmap
	if filters != nil && len(filters.Segments) > 0 {
		filterBitmap = filters.Segments[0].Data()
	}
	// We can't grab the containers "for each row" from the set-type field,
	// because we don't know how many rows there are, and some of them
	// might be empty, so really, we're going to iterate through the
	// containers, and then intersect them with the filter if present.
	var filter []*roaring.Container
	if filterBitmap != nil {
		filter = make([]*roaring.Container, 1<<shardVsContainerExponent)
		filterIterator, _ := filterBitmap.Containers.Iterator(0)
		// So let's get these all with a nice convenient 0 offset...
		for filterIterator.Next() {
			k, c := filterIterator.Value()
			if c.N() == 0 {
				continue
			}
			filter[k%(1<<shardVsContainerExponent)] = c
		}
	}
	prevRow := ^uint64(0)
	seenThisRow := false
	for fragData.Next() {
		k, c := fragData.Value()
		row := k >> shardVsContainerExponent
		if row == prevRow {
			if seenThisRow {
				continue
			}
		} else {
			seenThisRow = false
			prevRow = row
		}
		if filterBitmap != nil {
			if roaring.IntersectionAny(c, filter[k%(1<<shardVsContainerExponent)]) {
				o.Add(row)
				seenThisRow = true
			}
		} else if c.N() != 0 {
			o.Add(row)
			seenThisRow = true
		}
	}
	return nil
}
