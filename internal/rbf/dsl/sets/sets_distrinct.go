package sets

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/mutex"
)

func Distinct(c *rbf.Cursor, o *roaring64.Bitmap, filters *rows.Row) error {
	return mutex.Distinct(c, o, filters)
}
