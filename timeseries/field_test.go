package timeseries

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"
)

func TestCompression(t *testing.T) {
	e := &Aggr_Segment{
		Aggregates: map[string]*Aggr_Total{
			"/": {
				Visits:   rand.Uint64(),
				Visitors: rand.Uint64(),
				Events:   rand.Uint64(),
			},
			"/home": {
				Visits:   rand.Uint64(),
				Visitors: rand.Uint64(),
				Events:   rand.Uint64(),
			},
			"/about": {
				Visits:   rand.Uint64(),
				Visitors: rand.Uint64(),
				Events:   rand.Uint64(),
			},
		},
	}
	b, _ := proto.Marshal(e)
	var o bytes.Buffer
	enc, _ := zstd.NewWriter(&o)
	enc.Write(b)
	enc.Close()

	t.Error(len(b), o.Len(), 1<<10)
}
