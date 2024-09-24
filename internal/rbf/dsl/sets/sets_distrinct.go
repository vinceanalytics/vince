package sets

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/mutex"
	"github.com/gernest/rows"
)

func Distinct(c *rbf.Cursor, o *roaring64.Bitmap, filters *rows.Row) error {
	return mutex.Distinct(c, o, filters)
}
