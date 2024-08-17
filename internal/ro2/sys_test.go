package ro2

import (
	"fmt"
	"testing"
	"time"

	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func TestSys(t *testing.T) {
	now := uint64(time.Now().UTC().UnixMilli())
	shard := now / ro.ShardWidth
	b := roaring64.New()
	ro.BSI(b, now, 400<<20)
	b.Each(func(key uint32, cKey uint16, value *roaring.Container) error {
		fmt.Println(value.Type(), key, cKey, value)
		return nil
	})
	t.Error(shard)

}
