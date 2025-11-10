package timeseries

import (
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/timeseries/cursor"
)

const realtimeResolution = encoding.Minute

// Realtime computes total visitors in the last 5 minutes.
func (ts *Timeseries) Visitors(start, end time.Time, resolution encoding.Resolution, domain string) (visitors uint64) {
	values := models.BitSet(0)
	values.Set(v1.Field_id)
	r := ro2.NewBitmap()
	domainId := ts.Translate(v1.Field_domain, []byte(domain))
	ts.Scan(domainId, resolution, start, end, nil, values, func(cu *cursor.Cursor, field models.Field, view, shard uint64, columns *ro2.Bitmap) error {
		uniq := ro2.ReadDistinctBSI(cu, shard, columns)
		r.UnionInPlace(uniq)
		return nil
	})
	visitors = r.Count()
	return
}

func (ts *Timeseries) Scan(
	domainId uint64,
	res encoding.Resolution,
	start, end time.Time,
	filter Filter,
	valueSet models.BitSet,
	cb ScanCall,
) error {
	if domainId == 0 {
		return nil
	}
	dataFields := valueSet.All()

	return ts.db.Iter(
		domainId,
		res, start, end,
		func(cu *cursor.Cursor, shard, view uint64, match *ro2.Bitmap) error {

			if filter != nil {
				match = match.Intersect(filter.Apply(cu, res, shard, view))
			}
			if !match.Any() {
				return nil
			}
			for _, field := range dataFields {
				if !cu.ResetData(res, field, view) {
					continue
				}
				err := cb(cu, field, view, shard, match)
				if err != nil {
					return err
				}
			}
			return nil
		})
}
