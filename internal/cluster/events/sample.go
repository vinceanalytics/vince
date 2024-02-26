package events

import (
	_ "embed"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/memory"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:embed testdata/sample.json
var sample []byte

type sampleOptions struct {
	now  func() time.Time
	step time.Duration
}

type SampleOption func(*sampleOptions)

func WithNow(now func() time.Time) SampleOption {
	return func(so *sampleOptions) {
		so.now = now
	}
}

func WithStep(step time.Duration) SampleOption {
	return func(so *sampleOptions) {
		so.step = step
	}
}

func SampleRecord(opts ...SampleOption) arrow.Record {
	ls := Samples(opts...)
	b := New(memory.DefaultAllocator)
	defer b.Release()
	return b.Write(ls)
}

func Samples(opts ...SampleOption) *v1.Data_List {
	o := &sampleOptions{
		now:  Now(),
		step: time.Second,
	}
	for _, v := range opts {
		v(o)
	}
	var ls v1.Data_List
	protojson.Unmarshal(sample, &ls)
	now := o.now().UTC()
	for i, e := range ls.GetItems() {
		e.Timestamp = now.Add(time.Duration(i) * o.step).UnixMilli()
	}
	return &ls
}

func Now() func() time.Time {
	ts, _ := time.Parse(time.RFC822, time.RFC822)
	ts = ts.UTC()
	return func() time.Time { return ts }
}
