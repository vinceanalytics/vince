package oracle

import (
	"github.com/vinceanalytics/vince/internal/btx"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
)

func MinMax(c *rbf.Cursor, shard uint64) (int64, int64, error) {
	consider, err := cursor.Row(c, shard, bsiExistsBit)
	if err != nil {
		return 0, 0, err
	}
	min, _, err := btx.MinUnsigned(c, consider, shard, 64)
	if err != nil {
		return 0, 0, err
	}
	max, _, err := btx.MaxUnsigned(c, consider, shard, 64)
	if err != nil {
		return 0, 0, err
	}
	return int64(min), int64(max), nil

}
