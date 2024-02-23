package events

import (
	_ "embed"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/events/v1"
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

func Samples(opts ...SampleOption) *v1.List {
	o := &sampleOptions{
		now:  time.Now,
		step: time.Second,
	}
	for _, v := range opts {
		v(o)
	}
	var ls v1.List
	protojson.Unmarshal(sample, &ls)
	now := o.now().UTC()
	for i, e := range ls.GetItems() {
		e.Timestamp = now.Add(time.Duration(i) * o.step).UnixMilli()
	}
	return &ls
}
