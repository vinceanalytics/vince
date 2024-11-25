package timeseries

import (
	"time"

	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/timeseries/cursor"
)

const realtimeResolution = encoding.Minute

// Realtime computes total visitors in the last 5 minutes.
func (ts *Timeseries) Visitors(start, end time.Time, resolution encoding.Resolution, domain string) (visitors uint64) {
	var m FilterSet
	values := models.BitSet(0)
	values.Set(models.Field_id)
	r := ro2.NewBitmap()
	domainId := ts.Translate(models.Field_domain, []byte(domain))
	ts.Scan(domainId, resolution, start, end, m, values, func(cu *cursor.Cursor, field models.Field, view, shard uint64, columns *ro2.Bitmap) error {
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
	filterSet FilterSet,
	valueSet models.BitSet,
	cb ScanCall,
) error {
	if domainId == 0 {
		return nil
	}
	scan := config(filterSet, valueSet)

	filters := scan.Filter.Set.All()
	dataFields := scan.Data.Set.All()

	// use noFilters to preload existence bitmaps per shard.
	var noFilters []models.Field
	for _, f := range filters {
		if len(filterSet[f].No) > 0 {
			noFilters = append(noFilters, f)
		}
	}

	return ts.db.Iter(
		domainId,
		res, start, end, noFilters,
		func(cu *cursor.Cursor, shard, view uint64, m *ro2.Bitmap, exists map[models.Field]*ro2.Bitmap) error {

			var fs *ro2.Bitmap
			for _, field := range filters {
				if !cu.ResetData(res, field, view) {
					return nil
				}
				ex := exists[field]
				if ex != nil {
					ex = m.Intersect(ex)
				}
				b := filterSet[field].Apply(shard, cu, ex)
				if !b.Any() {
					return nil
				}
				if fs == nil {
					fs = b
					continue
				}
				fs = fs.Intersect(b)
				if !fs.Any() {
					return nil
				}
			}
			if fs != nil {
				m = m.Intersect(fs)
			}
			if !m.Any() {
				return nil
			}
			for _, field := range dataFields {
				if !cu.ResetData(res, field, view) {
					continue
				}
				err := cb(cu, field, view, shard, m)
				if err != nil {
					return err
				}
			}
			return nil
		})
}

type ScanConfig struct {
	All, Data, Filter struct {
		Set models.BitSet
	}
}

func config(
	filterSet FilterSet,
	valueSet models.BitSet) (co ScanConfig) {

	filterScanFields := filterSet.ScanFields()
	co.Filter.Set = filterScanFields

	fieldsToScan := filterScanFields.Or(valueSet)
	co.All.Set = fieldsToScan

	co.Data.Set = valueSet
	return
}