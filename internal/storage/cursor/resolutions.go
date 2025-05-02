package cursor

import (
	"encoding/binary"
	"time"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/storage/bitmaps"
	"github.com/vinceanalytics/vince/internal/storage/date"
	"github.com/vinceanalytics/vince/internal/storage/fields"
)

// Resolutions maps resolved timestamp to matching columns. Both Timestamp and
// Columns are of the same length.
//
// Fields are separated because we process them  separately.
type Resolutions struct {
	Timestamp []time.Time
	Columns   []*roaring.Bitmap
}

type resolutionContext struct {
	field     v1.Field
	transform func(uint64) time.Time
	resolve   func(time.Time, bool) uint64
}

var resolutionMatrix = map[v1.Resolution]resolutionContext{
	v1.Resolution_Global: {
		field:     v1.Field_year,
		transform: date.Year,
		resolve:   date.FromYear,
	},
	v1.Resolution_Minute: {
		field:     v1.Field_minute,
		transform: date.Minute,
		resolve:   date.FromMinute,
	},
	v1.Resolution_Hour: {
		field:     v1.Field_hour,
		transform: date.Hour,
		resolve:   date.FromHour,
	},
	v1.Resolution_Day: {
		field:     v1.Field_day,
		transform: date.Day,
		resolve:   date.FromDay,
	},
	v1.Resolution_Week: {
		field:     v1.Field_week,
		transform: date.Week,
		resolve:   date.FromWeek,
	},
	v1.Resolution_Month: {
		field:     v1.Field_month,
		transform: date.Month,
		resolve:   date.FromMonth,
	},
}

// Resolve reurns all matching columns ids for the given time range. resolution determines which field is read
// for the given shard.
//
// We map global resolution to year resoltion.
func (cu *Cursor) Resolve(resolution v1.Resolution, shard uint64, from, to time.Time) (r Resolutions) {
	ctx, ok := resolutionMatrix[resolution]
	if !ok {
		return
	}

	if !cu.ResetData(ctx.field, v1.DataType_mutex, shard) {
		return
	}
	startTs := ctx.resolve(from, false)
	endTs := ctx.resolve(to, true)

	start := shardwidth.ShardWidth * startTs
	end := shardwidth.ShardWidth * endTs
	offset := shardwidth.ShardWidth * shard
	off := highbits(offset)
	hi0, hi1 := highbits(start), highbits(end)
	if !cu.Seek(hi0) {
		return
	}
	var prev uint64
	var ra *roaring.Bitmap
	for ; cu.Valid(); cu.it.Next() {
		key := cu.it.Key()
		ckey := binary.BigEndian.Uint64(key[fields.ContainerOffset:])
		if ckey >= hi1 {
			break
		}
		row := ckey >> bitmaps.ShardVsContainerExponent
		if row != prev {
			b := roaring.NewBitmap()
			r.Timestamp = append(r.Timestamp, ctx.transform(row))
			r.Columns = append(r.Columns, b)
			ra = b
			prev = row
			hi0 = ckey
		}
		ra.Containers.Put(off+(ckey-hi0), roaring.DecodeContainer(cu.it.Value()).Clone())
	}
	return
}
