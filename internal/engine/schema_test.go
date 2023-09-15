package engine

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/parquet-go/parquet-go"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/entry"
)

func TestSchema(t *testing.T) {

	t.Run("creates a sql schema and arrow schema", func(t *testing.T) {
		ts := testSchema()
		expectSql := sql.Schema{
			{
				Name:   v1.Column_id.String(),
				Source: testSource,
				Type:   types.Int64,
			},
			{
				Name:   v1.Column_timestamp.String(),
				Source: testSource,
				Type:   types.Timestamp,
			},
			{
				Name:   v1.Column_duration.String(),
				Source: testSource,
				Type:   types.Int64,
			},
			{
				Name:   v1.Column_event.String(),
				Source: testSource,
				Type:   types.Text,
			},
		}
		if !expectSql.Equals(ts.sql) {
			t.Error("mismatch sal schema")
		}
		expectedArrow := arrow.NewSchema([]arrow.Field{
			{
				Name: v1.Column_id.String(),
				Type: arrow.PrimitiveTypes.Int64,
			},
			{
				Name: v1.Column_timestamp.String(),
				Type: arrow.FixedWidthTypes.Timestamp_ms,
			},
			{
				Name: v1.Column_duration.String(),
				Type: arrow.FixedWidthTypes.Duration_ms,
			},
			{
				Name: v1.Column_event.String(),
				Type: arrow.BinaryTypes.String,
			},
		}, nil)
		if !expectedArrow.Equal(ts.arrow) {
			t.Errorf("mismatch arrow schema \n %s \n %s ", expectedArrow, ts.arrow)
		}
	})

	t.Run("read", func(t *testing.T) {
		ts := createSchema(testSource, []v1.Column{
			v1.Column_id,        // int64
			v1.Column_timestamp, // time.Time
			v1.Column_duration,  // time.Duration
			v1.Column_event,     // time.Duration
		})

		f, err := testParquetFile()
		if err != nil {
			t.Fatal(err)
		}
		r, err := ts.read(context.TODO(), f)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := 1, r.NumRows(); want != int(got) {
			t.Errorf("expected %d got %d", want, got)
		}
		cols := r.Columns()
		if want, got := int64(1), cols[0].(*array.Int64).Value(0); want != got {
			t.Errorf("expected %d got %d", want, got)
		}
		stamp, err := time.Parse(time.RFC822Z, time.RFC822Z)
		if err != nil {
			t.Fatal(err)
		}
		if want, got := arrow.Timestamp(stamp.UTC().UnixMilli()), cols[1].(*array.Timestamp).Value(0); want != got {
			t.Errorf("expected %d got %d", want, got)
		}
		if want, got := float64(1), cols[2].(*array.Float64).Value(0); want != got {
			t.Errorf("expected %v got %v", want, got)
		}
		if want, got := "pageview", cols[3].(*array.String).Value(0); want != got {
			t.Errorf("expected %v got %v", want, got)
		}
	})
}

const testSource = "vince"

func testSchema() tableSchema {
	return createSchema(testSource, []v1.Column{
		v1.Column_id,        // int64
		v1.Column_timestamp, // time.Time
		v1.Column_duration,  // time.Duration
		v1.Column_event,     // time.Duration
	})
}

func testParquetFile() (*parquet.File, error) {
	var buf bytes.Buffer
	writeTestParquetFile(&buf)
	return parquet.OpenFile(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
}

func writeTestParquetFile(buf *bytes.Buffer) error {
	w := entry.NewWriter(buf)
	stamp, err := time.Parse(time.RFC822Z, time.RFC822Z)
	if err != nil {
		return err
	}
	_, err = w.Write([]*entry.Entry{
		{ID: 1, Timestamp: stamp, Duration: time.Second, Event: "pageview"},
	})
	if err != nil {
		return err
	}
	return w.Close()
}

func BenchmarkCreateSchema(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		createSchema(testSource, Columns)
	}
}

func BenchmarkSchema_read(b *testing.B) {
	ts := testSchema()
	ctx := context.TODO()
	var buf bytes.Buffer
	writeTestParquetFile(&buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		f, err := parquet.OpenFile(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		r, err := ts.read(ctx, f)
		if err != nil {
			b.Fatal(err)
		}
		r.Release()
	}
}