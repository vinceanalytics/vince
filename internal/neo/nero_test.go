package neo

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/entry"
)

func TestQuery(t *testing.T) {
	ts := time.Now().UTC()
	start := ts.Add(-24 * time.Hour)
	var b bytes.Buffer
	w := entry.Writer(&b)
	w.Write([]entry.Entry{
		{Path: "/", Timestamp: start.Add(time.Hour)},
		{Path: "/hello", Timestamp: start.Add(2 * time.Hour)},
	})
	w.Close()
	t.Run("select everything when no filters ", func(t *testing.T) {
		f, err := parquet.OpenFile(bytes.NewReader(b.Bytes()), int64(b.Len()))
		if err != nil {
			t.Fatal(err)
		}
		r, err := Exec(context.Background(), Options{
			Start:  start.UnixMilli(),
			End:    ts.UnixMilli(),
			Select: []string{"path"},
		}, func(gp GroupProcess) error {
			for _, g := range f.RowGroups() {
				err := gp(g)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		schema := r.Schema()
		if got, want := len(schema.Fields()), 1; got != want {
			t.Errorf("expected %d fields got %d", want, got)
		}
		if got, want := r.NumCols(), int64(1); got != want {
			t.Errorf("expected %d columns got %d", want, got)
		}
		if got, want := r.NumRows(), int64(2); got != want {
			t.Errorf("expected %d rows got %d", want, got)
		}
		if got, want := r.ColumnName(0), "path"; got != want {
			t.Errorf("expected %q column name got %q", want, got)
		}
		a := r.Column(0).(*array.String)
		if got, want := a.Value(0), "/"; got != want {
			t.Errorf("expected %q row value got %q", want, got)
		}
		if got, want := a.Value(1), "/hello"; got != want {
			t.Errorf("expected %q row value got %q", want, got)
		}
	})
	t.Run("select everything when no filters and late start ", func(t *testing.T) {
		f, err := parquet.OpenFile(bytes.NewReader(b.Bytes()), int64(b.Len()))
		if err != nil {
			t.Fatal(err)
		}
		r, err := Exec(context.Background(), Options{
			Start:  start.Add(time.Hour + time.Millisecond).UnixMilli(),
			End:    ts.UnixMilli(),
			Select: []string{"path"},
		}, func(gp GroupProcess) error {
			for _, g := range f.RowGroups() {
				err := gp(g)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		schema := r.Schema()
		if got, want := len(schema.Fields()), 1; got != want {
			t.Errorf("expected %d fields got %d", want, got)
		}
		if got, want := r.NumCols(), int64(1); got != want {
			t.Errorf("expected %d columns got %d", want, got)
		}
		if got, want := r.NumRows(), int64(2); got != want {
			t.Errorf("expected %d rows got %d", want, got)
		}
		if got, want := r.ColumnName(0), "path"; got != want {
			t.Errorf("expected %q column name got %q", want, got)
		}
		if got, want := must.Must(json.Marshal(r)), must.Must(os.ReadFile("testdata/simple_result.json")); !bytes.Equal(want, got) {
			t.Error("failed expectation")
		}
	})
}
