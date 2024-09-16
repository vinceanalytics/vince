package ro2

import (
	"errors"
	"time"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func (tx *Tx) quatum(b *roaring64.Bitmap, shard, id uint64, tsMs int64) error {
	ts := time.UnixMilli(tsMs)
	return errors.Join(
		tx.time(b, alicia.MINUTE, shard, id, minute(ts)),
		tx.time(b, alicia.HOUR, shard, id, hour(ts)),
		tx.time(b, alicia.DAY, shard, id, day(ts)),
		tx.time(b, alicia.WEEK, shard, id, week(ts)),
		tx.time(b, alicia.MONTH, shard, id, month(ts)),
	)
}

func (tx *Tx) time(b *roaring64.Bitmap, field alicia.Field, shard, id uint64, ts time.Time) error {
	b.Clear()
	ro.BSI(b, id, ts.UnixMilli())
	return tx.Add(shard, uint64(field), b)
}

func minute(ts time.Time) time.Time {
	return ts.Truncate(time.Minute)
}

func hour(ts time.Time) time.Time {
	return ts.Truncate(time.Hour)
}

func day(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
}

func week(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	day := time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
	weekday := int(day.Weekday())
	return day.AddDate(0, 0, -weekday)
}

func month(ts time.Time) time.Time {
	yy, mm, _ := ts.Date()
	return time.Date(yy, mm, 1, 0, 0, 0, 0, time.UTC)
}
