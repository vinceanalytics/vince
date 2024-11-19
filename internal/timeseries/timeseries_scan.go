package timeseries

import (
	"fmt"
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

	return ts.db.Iter(res, start, end, func(db *pebble.DB, shard uint64, views *ro2.Bitmap) error {
		it, err := db.NewIter(nil)
		if err != nil {
			return fmt.Errorf("initializing iterator for scan %w", err)
		}
		defer it.Close()
		cu.SetIter(it)

		filters := scan.Filter.Set.All()
		dataFields := scan.Data.Set.All()

		return views.ForEach(func(view uint64) error {
			var m *ro2.Bitmap
			for _, field := range filters {
				if !cu.Reset(field, res, view) {
					return nil
				}
				b, err := filterSet[field].Apply(shard, cu)
				if err != nil {
					return err
				}
				if !b.Any() {
					return nil
				}
				if m == nil {
					m = b
					continue
				}
				m.IntersectInPlace(b)
				if !m.Any() {
					return nil
				}
			}
			if m == nil || !m.Any() {
				return nil
			}

			// we have matching filter for this view

			for _, field := range dataFields {
				if !cu.Reset(field, res, view) {
					continue
				}
				err := cb(cu, field, view, shard, m)
				if err != nil {
					return err
				}
			}
			return nil
		})

	})

}

var zero = time.Time{}

func toView(ts time.Time) uint64 {
	if ts == zero {
		return 0
	}
	return uint64(ts.UnixMilli())
}

// Tracks  [view][shard][filter_field_bitmap]. Because we perform sequentiial scans
type filterMatchSet map[uint64]map[uint64]map[models.Field]*ro2.Bitmap

func (fs filterMatchSet) Apply() map[uint64]map[uint64]*ro2.Bitmap {
	result := map[uint64]map[uint64]*ro2.Bitmap{}
	add := func(view, shard uint64, ra *ro2.Bitmap) {
		vs, ok := result[view]
		if !ok {
			vs = make(map[uint64]*ro2.Bitmap)
			result[view] = vs
		}
		vs[shard] = ra
	}
	for k, v := range fs {
		for s, sv := range v {
			r := reduce(sv)
			if r.Any() {
				add(k, s, r)
			}
		}

	}
	return result
}

func reduce(ms map[models.Field]*ro2.Bitmap) (r *ro2.Bitmap) {
	for _, v := range ms {
		if r == nil {
			r = v
			continue
		}
		r.IntersectInPlace(v)
		if !r.Any() {
			return
		}
	}
	return
}

func (fs filterMatchSet) Add(field models.Field, view, shard uint64, ra *ro2.Bitmap) {
	if !ra.Any() {
		return
	}
	vs, ok := fs[view]
	if !ok {
		vs = make(map[uint64]map[models.Field]*ro2.Bitmap)
		fs[view] = vs
	}
	ss, ok := vs[shard]
	if !ok {
		ss = make(map[models.Field]*ro2.Bitmap)
		vs[shard] = ss
	}
	bs, ok := ss[field]
	if !ok {
		ss[field] = ra
	} else {
		bs.UnionInPlace(ra)
	}
}

type ScanConfig struct {
	All, Data, Filter struct {
		Set      models.BitSet
		Min, Max models.Field
	}
}

func config(
	filterSet FilterSet,
	valueSet models.BitSet) (co ScanConfig) {

	filterScanFields := filterSet.ScanFields()
	lo, hi := minmax(filterScanFields)
	co.Filter.Set = filterScanFields
	co.Filter.Min = lo
	co.Filter.Max = hi

	fieldsToScan := filterScanFields.Or(valueSet)
	lo, hi = minmax(fieldsToScan)
	co.All.Min = lo
	co.All.Max = hi
	co.All.Set = fieldsToScan

	lo, hi = minmax(valueSet)
	co.Data.Min = lo
	co.Data.Max = hi
	co.Data.Set = valueSet
	return
}

func minmax(bs models.BitSet) (min, max models.Field) {
	if bs.Len() == 0 {
		return
	}
	set := bs.All()
	min = set[0]
	max = set[len(set)-1]
	return
}
