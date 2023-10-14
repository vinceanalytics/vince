package neo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/compute"
	"github.com/apache/arrow/go/v14/arrow/math"
	"github.com/apache/arrow/go/v14/arrow/scalar"
	"github.com/parquet-go/parquet-go"
	blockv1 "github.com/vinceanalytics/proto/gen/go/vince/blocks/v1"
	storev1 "github.com/vinceanalytics/proto/gen/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/entry"
	"golang.org/x/sync/errgroup"
)

func IndexBlockFile(ctx context.Context, f *os.File) (map[storev1.Column]*blockv1.ColumnIndex, *blockv1.BaseStats, error) {
	stat, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	r, err := parquet.OpenFile(f, stat.Size())
	if err != nil {
		return nil, nil, err
	}
	m := make(map[storev1.Column]int)
	schema := r.Schema()
	for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
		n, _ := schema.Lookup(i.String())
		m[i] = n.ColumnIndex
	}

	cols := make(map[storev1.Column]*blockv1.ColumnIndex)
	groups := r.RowGroups()
	for gi := range groups {
		g := groups[gi]
		chunks := g.ColumnChunks()
		{
			// index  timestamp column
			tidx, ok := cols[storev1.Column_timestamp]
			if !ok {
				tidx = &blockv1.ColumnIndex{}
				cols[storev1.Column_timestamp] = tidx
			}
			ts := chunks[m[storev1.Column_timestamp]]
			idx := ts.ColumnIndex()
			tg := &blockv1.ColumnIndex_RowGroup{}
			for i := 0; i < idx.NumPages(); i++ {
				lo, hi := idx.MinValue(i).Int64(), idx.MaxValue(i).Int64()
				if tg.Min == 0 {
					tg.Min = lo
				}
				if tidx.Min == 0 {
					tidx.Min = lo
				}
				tidx.Min, tidx.Max = min(tidx.Min, lo), max(tidx.Max, hi)
				tg.Min, tg.Max = min(tg.Min, lo), max(tg.Max, hi)
				tg.Pages = append(tg.Pages, &blockv1.ColumnIndex_Page{
					Min: lo,
					Max: hi,
				})
			}
			tidx.RowGroups = append(tidx.RowGroups, tg)
		}
		for i := storev1.Column_browser; i <= storev1.Column_utm_term; i++ {
			idx, ok := cols[i]
			if !ok {
				idx = &blockv1.ColumnIndex{}
				cols[i] = idx
			}
			rg := &blockv1.ColumnIndex_RowGroup{
				BloomFilter: readFilter(chunks[m[i]].BloomFilter()),
			}
			idx.RowGroups = append(idx.RowGroups, rg)
		}
	}
	views, visitors, visits, sessionDuration, bounceRate, err := baseStats(ctx, r)
	if err != nil {
		return nil, nil, err
	}
	return cols, &blockv1.BaseStats{
		PageViews: views, Visitors: visitors, Visits: visits,
		Duration: sessionDuration, BounceRate: bounceRate,
	}, nil
}

func readFilter(b parquet.BloomFilter) []byte {
	if b == nil {
		return nil
	}
	o := make([]byte, b.Size())
	b.ReadAt(o, 0)
	return o
}

// Computes base stats per parquet file. Row groups are read concurrently. Base
// stats are calculated as follows.
//
//	pageViews: Counts name column with value "pageview"
//	visitors: counts unique id column values
//	visits: counts unique session column
//	sessionDuration: sum of duration column divide vy visits
//	bounceRate: percentage of sum of bounce column to visits
func baseStats(ctx context.Context, r *parquet.File) (
	pageViews, visitors, visits int64,
	sessionDuration, bounceRate float64, err error) {
	ctx = entry.Context(ctx)
	schema := r.Schema()
	c, ok := schema.Lookup(storev1.Column_event.String())
	if !ok {
		err = errors.New("parquet file missing event column")
		return
	}
	event := c.ColumnIndex
	c, ok = schema.Lookup(storev1.Column_id.String())
	if !ok {
		err = errors.New("parquet file missing id column")
		return
	}
	id := c.ColumnIndex
	c, ok = schema.Lookup(storev1.Column_session.String())
	if !ok {
		err = errors.New("parquet file missing session column")
		return
	}
	session := c.ColumnIndex
	d, ok := schema.Lookup(storev1.Column_duration.String())
	if !ok {
		err = errors.New("parquet file missing duration column")
		return
	}
	duration := d.ColumnIndex
	b, ok := schema.Lookup(storev1.Column_bounce.String())
	if !ok {
		err = errors.New("parquet file missing duration column")
		return
	}
	bounce := b.ColumnIndex

	pageviewBuilder := array.NewStringBuilder(entry.Pool)
	idBuilder := array.NewInt64Builder(entry.Pool)
	sessionBuilder := array.NewInt64Builder(entry.Pool)
	durationBuilder := array.NewInt64Builder(entry.Pool)
	bounceBuilder := array.NewInt64Builder(entry.Pool)
	defer func() {
		pageviewBuilder.Release()
		idBuilder.Release()
		sessionBuilder.Release()
		durationBuilder.Release()
		bounceBuilder.Release()
	}()
	readPages := readString(pageviewBuilder)
	readId := readInt64(idBuilder)
	readSession := readInt64(sessionBuilder)
	readDuration := readInt64(durationBuilder)
	readBounce := readInt64(bounceBuilder)
	var eg errgroup.Group
	for _, rg := range r.RowGroups() {
		chunks := rg.ColumnChunks()
		eg.Go(readColumn(readPages, chunks[event]))
		eg.Go(readColumn(readId, chunks[id]))
		eg.Go(readColumn(readSession, chunks[session]))
		eg.Go(readColumn(readDuration, chunks[duration]))
		eg.Go(readColumn(readBounce, chunks[bounce]))
	}
	err = eg.Wait()
	if err != nil {
		err = fmt.Errorf("failed reading base stat columns from parquet file:%v", err)
		return
	}
	pages := pageviewBuilder.NewArray()
	// select pageview
	p, err := entry.Call(ctx, "equal", nil, pages, scalar.MakeScalar("pageview"))
	if err != nil {
		err = fmt.Errorf("failed applying equal filter to page view array :%v", err)
		return
	}
	f, err := entry.Call(ctx, "filter", nil, pages, p, pages.Release, p.Release)
	if err != nil {
		err = fmt.Errorf("failed applying  filter to page view array: %v2", err)
		return
	}
	pageViews = int64(f.Len())
	f.Release()
	// calculate visitors
	va := idBuilder.NewArray()
	vb, err := compute.UniqueArray(ctx, va)
	if err != nil {
		err = fmt.Errorf("failed computing unique id: %v", err)
		return
	}
	visitors = int64(vb.Len())
	vb.Release()
	// calculate visits
	sa := sessionBuilder.NewArray()
	sb, err := compute.UniqueArray(ctx, sa)
	if err != nil {
		err = fmt.Errorf("failed computing unique sessions: %v", err)
		return
	}
	visits = int64(sb.Len())
	sb.Release()

	da := durationBuilder.NewInt64Array()
	durationSum := math.Int64.Sum(da)
	da.Release()
	sessionDuration = time.Duration(durationSum).Seconds() / float64(visits)

	ba := bounceBuilder.NewInt64Array()
	bounceSum := math.Int64.Sum(ba)
	ba.Release()
	bounceRate = float64(bounceSum) / float64(visits) * 100
	return
}

func readString(b *array.StringBuilder) func(*entry.ValuesBuf) {
	var lock sync.Mutex
	return func(vb *entry.ValuesBuf) {
		lock.Lock()
		b.Reserve(len(vb.Values))
		for i := range vb.Values {
			b.Append(vb.Values[i].String())
		}
		lock.Unlock()
	}
}

func readInt64(b *array.Int64Builder) func(*entry.ValuesBuf) {
	var lock sync.Mutex
	return func(vb *entry.ValuesBuf) {
		lock.Lock()
		b.Reserve(len(vb.Values))
		for i := range vb.Values {
			b.UnsafeAppend(vb.Values[i].Int64())
		}
		lock.Unlock()
	}
}

func readColumn(b func(*entry.ValuesBuf), chunk parquet.ColumnChunk) func() error {
	return func() error {
		buf := entry.NewValuesBuf()
		defer buf.Release()
		err := readValuesPages(buf, chunk.Pages())
		if err != nil {
			return err
		}
		b(buf)
		return nil
	}
}

func readValuesPages(buf *entry.ValuesBuf, pages parquet.Pages) error {
	defer pages.Close()
	for {
		page, err := pages.ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			buf.Release()
			return err
		}
		size := page.NumValues()
		o := buf.Get(int(size))
		page.Values().ReadValues(o)
	}
}
