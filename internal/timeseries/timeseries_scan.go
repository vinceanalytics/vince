package timeseries

import (
	"bytes"
	"time"

	"github.com/bits-and-blooms/bitset"
	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/xtime"
	"github.com/vinceanalytics/vince/internal/web/query"
)

type ScanData struct {
	Views   map[uint64]*FieldsData
	Columns *roaring.Bitmap
}

func (s *ScanData) Set(view uint64, field int, ra *roaring.Bitmap) {
	if s.Views == nil {
		s.Views = make(map[uint64]*FieldsData)
	}
	fd, ok := s.Views[view]
	if !ok {
		var fs FieldsData
		fd = &fs
		s.Views[view] = fd
	}
	fd[field] = ra.Clone()
}

type FieldsData [models.AllFields]*roaring.Bitmap

// Realtime computes total visitors in the last 5 minutes. We make a few assumptions
// to ensure this call is very fast and efficient.
//
//   - Only current shard is evaluated:
//     A shard comprises about 1Million events. We assume that a site will have
//     less than this unique visitors in a 5 minute span.
//
// We call this periodically but continuously when a user is on website dashboard.
// Covering only one shard strickes a balance to ensure responsiveness in UI and
// observation of useful insight.
//
// We can always adjust the number of shards we evaluate if we need to.
func (ts *Timeseries) Realtime(domain string) (visitors uint64) {
	end := xtime.Now()
	start := end.Add(-5 * time.Minute)
	ra := roaring.NewBitmap()
	for t := range query.Minute.Range(start, end) {
		ra.Set(uint64(t.UnixMilli()))
	}
	ts.views.mu.RLock()
	var shard uint64
	if len(ts.views.ra) > 0 {
		shard = uint64(len(ts.views.ra) - 1)
		ra.And(ts.views.ra[len(ts.views.ra)-1])
	} else {
		ra.Reset()
	}
	ts.views.mu.RUnlock()
	if ra.IsEmpty() {
		return
	}
	id := ts.Translate(models.Field_domain, []byte(domain))
	domainStart := encoding.Bitmap(shard, ra.Minimum(), models.Field_domain)
	domainEnd := encoding.Bitmap(shard, ra.Maximum(), models.Field_domain)
	it, err := ts.db.NewIter(nil)
	assert.Nil(err, "initializing iterator for global scan")
	defer it.Close()
	match := map[uint64]*roaring.Bitmap{}
	matcView := roaring.NewBitmap()
	for it.SeekGE(domainStart); it.Valid(); it.Next() {
		key := it.Key()
		if bytes.Compare(key, domainEnd) == 1 {
			break
		}
		if key[len(key)-1] != byte(models.Field_domain) {
			break
		}
		_, view := encoding.Component(key)
		m := roaring.FromBuffer(it.Value()).Row(shard, id)
		if m.IsEmpty() {
			continue
		}
		match[view] = m
		matcView.Set(view)
	}
	field := models.Field_id
	startField := encoding.Bitmap(shard, matcView.Minimum(), field)
	endField := encoding.Bitmap(shard, matcView.Maximum(), field)
	r := roaring.NewBitmap()
	for it.SeekGE(startField); it.Valid(); it.Next() {
		key := it.Key()
		if bytes.Compare(key, endField) == 1 {
			break
		}
		if key[len(key)-1] != byte(field) {
			continue
		}
		_, view := encoding.Component(key)
		if !matcView.Contains(view) {
			continue
		}
		roaring.FromBuffer(it.Value()).ExtractBSI(shard, match[view], func(id uint64, value int64) {
			r.Set(uint64(value))
		})
	}
	visitors = uint64(r.GetCardinality())
	return
}

// ScanGlobal reads bitmaps for the field that belongs to domain in a global key space
// Global key space has time resolution of 0.
//
// Useful to compute aggregates across all shards like total visitors to a website
// since we received first event.
//
// f is called with the shard and ra which should not be modified as its memory
// is borrowed and will be invalidated in the next call to f. If you want to own
// ra please call ra.Clone().
//
// We store global bitmaps for all fields. Right now we only use this to display
// website's visitors on sites home dashboard.
func (ts *Timeseries) ScanGlobal(field models.Field, domain string, f func(shard uint64, columns, ra *roaring.Bitmap)) {
	ts.views.mu.RLock()
	shards := len(ts.views.ra) + 1
	ts.views.mu.RUnlock()
	id := ts.Translate(models.Field_domain, []byte(domain))
	domainStaet := encoding.Bitmap(0, 0, models.Field_domain)
	domainEnd := encoding.Bitmap(uint64(shards), 0, models.Field_domain)
	it, err := ts.db.NewIter(nil)
	assert.Nil(err, "initializing iterator for global scan")
	defer it.Close()
	match := map[uint64]*roaring.Bitmap{}
	matchShards := roaring.NewBitmap()
	for it.SeekGE(domainStaet); it.Valid(); it.Next() {
		key := it.Key()
		if bytes.Compare(key, domainEnd) == 1 {
			break
		}
		if key[len(key)-1] != byte(models.Field_domain) {
			break
		}
		shard, _ := encoding.Component(key)
		m := roaring.FromBuffer(it.Value()).Row(shard, id)
		if m.IsEmpty() {
			continue
		}
		match[shard] = m
		matchShards.Set(shard)
	}

	startField := encoding.Bitmap(matchShards.Minimum(), 0, field)
	endField := encoding.Bitmap(matchShards.Maximum(), 0, field)
	for it.SeekGE(startField); it.Valid(); it.Next() {
		key := it.Key()
		if bytes.Compare(key, endField) == 1 {
			break
		}
		if key[len(key)-1] != byte(field) {
			continue
		}
		shard, _ := encoding.Component(key)
		if !matchShards.Contains(shard) {
			continue
		}
		f(shard, match[shard], roaring.FromBuffer(it.Value()))
	}
}

func (ts *Timeseries) Scan(
	views []*roaring.Bitmap,
	filterSet FilterSet,
	valueSet *bitset.BitSet,
) (data []ScanData) {

	data = make([]ScanData, len(views))
	scan := config(views, filterSet, valueSet)

	if !scan.NotEmpty {
		// no shard/view matched
		return
	}
	it, err := ts.db.NewIter(&pebble.IterOptions{
		// We assume min view belongs to min shard and max view belong to max shard
		LowerBound: encoding.Bitmap(scan.Shards.Min, scan.Views.Min, models.Field(scan.All.Min)),
	})
	assert.Nil(err, "initializing iterator for scan")
	defer it.Close()
	for shard := range views {
		view := views[shard]
		if view.IsEmpty() {
			continue
		}
		match := roaring.NewBitmap()
		doScan(it, scan.Filter.Set, uint64(shard),
			view.Minimum(), view.Maximum(), scan.Filter.Min, scan.Filter.Max,
			func(shard, view uint64, field int, b *roaring.Bitmap) bool {
				ra := filterSet[field].Apply(shard, b)
				if field == scan.Filter.Min {
					match.Or(ra)
				} else {
					match.And(ra)
				}
				return !match.IsEmpty()
			})
		if match.IsEmpty() {
			continue
		}
		data[shard].Columns = match
		// scan data bitmaps
		doScan(it, scan.Data.Set, uint64(shard),
			view.Minimum(), view.Maximum(), scan.Data.Min, scan.Data.Max,
			func(shard, view uint64, field int, b *roaring.Bitmap) bool {
				data[shard].Set(view, field, b)
				return true
			})
	}
	return
}

func doScan(it *pebble.Iterator,
	chhose *bitset.BitSet,
	shard uint64, startView, endView uint64,
	startField, endField int,
	f func(shard, view uint64, field int, b *roaring.Bitmap) bool) {
	from := encoding.Bitmap(shard, startView, models.Field(startField))
	to := encoding.Bitmap(shard, endView, models.Field(endField))
	for it.SeekGE(from); it.Valid(); it.Next() {
		key := it.Key()
		if bytes.Compare(key, to) == 1 {
			break
		}
		fd := key[len(key)-1]
		if !chhose.Test(uint(fd)) {
			continue
		}
		shard, view := encoding.Component(key)
		if !f(shard, view, int(fd), roaring.FromBuffer(it.Value())) {
			break
		}
	}
}

type ScanConfig struct {
	NotEmpty bool
	Shards   struct {
		Min, Max uint64
	}
	Views struct {
		Min, Max uint64
	}

	All, Data, Filter struct {
		Set      *bitset.BitSet
		Min, Max int
	}
}

func config(views []*roaring.Bitmap,
	filterSet FilterSet,
	valueSet *bitset.BitSet) (co ScanConfig) {
	for shard, view := range views {
		if view.IsEmpty() {
			continue
		}
		co.Shards.Min = min(co.Shards.Min, uint64(shard))
		co.Shards.Max = min(co.Shards.Max, uint64(shard))

		co.Views.Min = min(co.Views.Min, view.Minimum())
		co.Views.Max = min(co.Views.Max, view.Maximum())
		co.NotEmpty = true
	}
	filterScanFields := filterSet.ScanFields()
	lo, hi := minmax(filterScanFields)
	co.Filter.Set = filterScanFields
	co.Filter.Min = lo
	co.Filter.Max = hi

	fieldsToScan := filterScanFields.Union(valueSet)
	lo, hi = minmax(fieldsToScan)
	co.All.Min = lo
	co.All.Max = hi
	co.All.Set = fieldsToScan

	lo, hi = minmax(valueSet)
	co.Data.Min = lo
	co.Data.Max = hi
	co.Data.Set = valueSet.Clone()
	return
}

func minmax(bs *bitset.BitSet) (min, max int) {
	if bs.Len() == 0 {
		return
	}
	a := make([]uint, models.AllFields)
	_, set := bs.NextSetMany(0, a)
	min = int(set[0])
	max = int(set[len(set)-1])
	return
}
