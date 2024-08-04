package oracle

import (
	"context"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/len64/v1"
	"github.com/vinceanalytics/vince/internal/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// critical path, any failure here results in a program exit because it will
// result in corrupt database state
func (db *dbShard) process(ctx context.Context, events chan *v1.Model) {
	ts := time.NewTicker(time.Minute)
	defer ts.Stop()
	w, err := db.Write()
	assert.Assert(err == nil, "opening events writer", "err", err)

	defer func() {
		w.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-events:
			w.event(e)
		case <-ts.C:
			err := w.Close()
			assert.Assert(err == nil, "closing events writer", "err", err)
			w, err = db.Write()
			assert.Assert(err == nil, "opening events writer", "err", err)
		}
	}
}

func (w *write) event(e *v1.Model) {
	w.Write(e.Timestamp, func(idx Columns) error {
		e.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			if fd.Kind() == protoreflect.StringKind {
				idx.String(string(fd.Name()), v.String())
				return true
			}
			if fd.IsMap() {
				prefix := string(fd.Name()) + "."
				v.Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
					idx.String(prefix+mk.String(), v.String())
					return true
				})
				idx.String(string(fd.Name()), v.String())
			}
			return true
		})
		idx.Int64("timestamp", int64(e.Timestamp))
		idx.String("date", date(e.Timestamp))
		idx.Int64("uid", int64(e.Id))
		if e.Bounce != nil {
			idx.Bool("bounce", e.GetBounce())
		} else {
			// null bounce means we clear bounce status
			idx.Int64("bounce", -1)
		}
		idx.Bool("session", e.Session)
		idx.Bool("view", e.View)
		idx.Int64("duration", int64(e.Duration))
		if e.City != 0 {
			idx.Int64("city_geoname_id", int64(e.City))
		}
		return nil
	})
}

func date(ts int64) string {
	yy, mm, dd := time.UnixMilli(ts).Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
}
