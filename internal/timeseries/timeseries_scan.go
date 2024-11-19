package timeseries

import (
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
)

const realtimeResolution = encoding.Minute

// Realtime computes total visitors in the last 5 minutes.
func (ts *Timeseries) Visitors(start, end time.Time, resolution encoding.Resolution, domain string) (visitors uint64) {
	var m FilterSet
	m.Set(true, models.Field_domain, ts.Translate(models.Field_domain, []byte(domain)))
	values := models.BitSet(0)
	values.Set(models.Field_id)
	r := ro2.NewBitmap()
	ts.Scan(resolution, start, end, m, values, func(cu *Cursor, field models.Field, view, shard uint64, columns *ro2.Bitmap) error {
		uniq := ro2.ReadDistinctBSI(cu, shard, columns)
		r.UnionInPlace(uniq)
		return nil
	})
	visitors = r.Count()
	return
}

func (ts *Timeseries) Scan(
	res encoding.Resolution,
	start, end time.Time,
	filterSet FilterSet,
	valueSet models.BitSet,
	cb ScanCall,
) error {
	scan := config(filterSet, valueSet)

	cu := new(Cursor)
	defer cu.Release()

	filters := scan.Filter.Set.All()
	dataFields := scan.Data.Set.All()

	return ts.db.Iter(res, start, end, func(it *pebble.Iterator, shard uint64, from, to int64) error {
		cu.SetIter(it)

		cu.Reset(models.Field_timestamp)

		var m *ro2.Bitmap
		if from == 0 && to == 0 {
			// special global resolution
			m = ro2.Existence(cu, shard)
		} else {
			m = ro2.Range(cu, ro2.BETWEEN, shard, cu.BitLen(), from, to)
		}
		if !m.Any() {
			return nil
		}

		var fs *ro2.Bitmap
		for _, field := range filters {
			if !cu.Reset(field) {
				return nil
			}
			b, err := filterSet[field].Apply(shard, cu)
			if err != nil {
				return err
			}
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
		m = m.Intersect(fs)
		if !m.Any() {
			return nil
		}
		for _, field := range dataFields {
			if !cu.Reset(field) {
				continue
			}
			err := cb(cu, field, uint64(from), shard, m)
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
