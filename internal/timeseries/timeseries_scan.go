package timeseries

import (
	"time"

	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/data"
)

const realtimeResolution = encoding.Minute

// Realtime computes total visitors in the last 5 minutes.
func (ts *Timeseries) Visitors(start, end time.Time, resolution encoding.Resolution, domain string) (visitors uint64) {
	var m FilterSet
	m.Set(true, models.Field_domain, ts.Translate(models.Field_domain, []byte(domain)))
	values := models.BitSet(0)
	values.Set(models.Field_id)
	r := roaring.NewBitmap()
	ts.Scan(resolution, start, end, m, values, func(field models.Field, view, shard uint64, columns, ra *roaring.Bitmap) bool {
		ra.ExtractBSI(shard, columns, func(id uint64, value int64) {
			r.Set(uint64(value))
		})
		return true
	})
	visitors = uint64(r.GetCardinality())
	return
}

func (ts *Timeseries) Scan(
	res encoding.Resolution,
	start, end time.Time,
	filterSet FilterSet,
	valueSet models.BitSet,
	cb func(field models.Field, view, shard uint64, columns, ra *roaring.Bitmap) bool,
) {
	scan := config(filterSet, valueSet)
	from := toView(start)
	to := toView(end)

	it, err := ts.db.NewIter(nil)
	assert.Nil(err, "initializing iterator for scan")
	defer it.Close()
	// apply filters
	match := make(filterMatchSet)
	data.Range(it,
		encoding.Min(models.Field(scan.Filter.Min), res, from),
		encoding.Max(models.Field(scan.Filter.Max), res, to),
		func(key []byte, ra *roaring.Bitmap) bool {
			field, _, view, shard := encoding.Component(key)
			if !scan.Filter.Set.Test(field) {
				return true
			}
			rs := filterSet[field].Apply(shard, ra)
			match.Add(field, view, shard, rs)
			return true
		},
	)
	matchSet := match.Apply()
	if len(matchSet) == 0 {
		return
	}
	data.Range(it,
		encoding.Min(models.Field(scan.Data.Min), res, from),
		encoding.Max(models.Field(scan.Data.Max), res, to),
		func(key []byte, ra *roaring.Bitmap) bool {
			field, _, view, shard := encoding.Component(key)
			if !scan.Data.Set.Test(field) {
				return true
			}
			vs, ok := matchSet[view]
			if !ok {
				return true
			}
			sv, ok := vs[shard]
			if !ok {
				return true
			}
			return cb(field, view, shard, sv, ra)
		},
	)
	return
}

var zero = time.Time{}

func toView(ts time.Time) uint64 {
	if ts == zero {
		return 0
	}
	return uint64(ts.UnixMilli())
}

// Tracks  [view][shard][filter_field_bitmap]. Because we perform sequentiial scans
type filterMatchSet map[uint64]map[uint64]map[models.Field]*roaring.Bitmap

func (fs filterMatchSet) Apply() map[uint64]map[uint64]*roaring.Bitmap {
	result := map[uint64]map[uint64]*roaring.Bitmap{}
	add := func(view, shard uint64, ra *roaring.Bitmap) {
		vs, ok := result[view]
		if !ok {
			vs = make(map[uint64]*roaring.Bitmap)
			result[view] = vs
		}
		vs[shard] = ra
	}
	for k, v := range fs {
		for s, sv := range v {
			r := reduce(sv)
			if !r.IsEmpty() {
				add(k, s, r)
			}
		}

	}
	return result
}

func reduce(ms map[models.Field]*roaring.Bitmap) (r *roaring.Bitmap) {
	for _, v := range ms {
		if r == nil {
			r = v
			continue
		}
		r.And(v)
		if r.IsEmpty() {
			return
		}
	}
	return
}

func (fs filterMatchSet) Add(field models.Field, view, shard uint64, ra *roaring.Bitmap) {
	if ra.IsEmpty() {
		return
	}
	vs, ok := fs[view]
	if !ok {
		vs = make(map[uint64]map[models.Field]*roaring.Bitmap)
		fs[view] = vs
	}
	ss, ok := vs[shard]
	if !ok {
		ss = make(map[models.Field]*roaring.Bitmap)
		vs[shard] = ss
	}
	bs, ok := ss[field]
	if !ok {
		ss[field] = ra
	} else {
		bs.Or(ra)
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
