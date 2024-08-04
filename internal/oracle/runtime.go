package oracle

import (
	"context"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
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
			name := string(fd.Name())
			switch fd.Kind() {
			case protoreflect.StringKind:
				idx.String(name, v.String())
			case protoreflect.BoolKind:
				idx.Int64(name, 1)
			case protoreflect.Int64Kind:
				if name == "timestamp" {
					ts := v.Int()
					idx.Int64(name, ts)
					idx.String("date", date(ts))
				} else {
					idx.Int64(name, v.Int())
				}

			case protoreflect.Uint64Kind, protoreflect.Uint32Kind:
				idx.Int64(name, int64(v.Uint()))
			default:
				if fd.IsMap() {
					prefix := name + "."
					v.Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
						idx.String(prefix+mk.String(), v.String())
						return true
					})
				}
			}
			return true
		})

		// For bounce, session, view and duration fields we only perform sum on the
		// bsi. To save space we don't bother storing zero values.
		//
		// null bounce act as a clear signal , we set it to -1 so that when
		// we a user stay on the site and navigated to a different page during
		// a live session the result will be 0..
		if e.Bounce == nil {
			idx.Int64("bounce", -1)
		}
		return nil
	})
}

func date(ts int64) string {
	yy, mm, dd := time.UnixMilli(ts).Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
}
