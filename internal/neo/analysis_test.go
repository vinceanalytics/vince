package neo

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/entry"
)

func TestComputedPartition(t *testing.T) {
	t.Run("defaults to only metrics and timestamp", func(t *testing.T) {
		got := []byte(computedPartition().String())
		os.WriteFile("testdata/computed_partition_schema_default.txt", got, 0600)
		want := must.
			Must(os.ReadFile("testdata/computed_partition_schema_default.txt"))()
		if !bytes.Equal(got, want) {
			t.Error("schema changed")
		}
	})
	t.Run("with partition key", func(t *testing.T) {
		got := []byte(computedPartition("path").String())
		os.WriteFile("testdata/computed_partition_schema_with_keys.txt", got, 0600)
		want := must.
			Must(os.ReadFile("testdata/computed_partition_schema_with_keys.txt"))()
		if !bytes.Equal(got, want) {
			t.Error("schema changed")
		}
	})
	t.Run("check computed partition record ", func(t *testing.T) {
		b := array.NewRecordBuilder(entry.Pool, computedPartition("path"))
		defer b.Release()
		now := time.Now().UTC().UnixMilli()
		b.Field(0).(*array.TimestampBuilder).Append(arrow.Timestamp(now))
		ls := b.Field(1).(*array.ListBuilder)
		ls.Append(true)
		s := ls.ValueBuilder().(*array.StructBuilder)
		s.Append(true)
		// first is the field is the value
		s.FieldBuilder(0).(*array.StringBuilder).Append("/")
		x := s.FieldBuilder(1).(*array.StructBuilder)
		x.Append(true)
		for i := range computedFields {
			x.FieldBuilder(i).(*array.Float64Builder).Append(float64(i))
		}
		r := b.NewRecord()
		defer r.Release()
		got := must.Must(json.MarshalIndent(r, "", " "))()
		os.WriteFile("testdata/computed_partition_schema_record.json", got, 0600)
		want := must.Must(os.ReadFile("testdata/computed_partition_schema_record.json"))()
		if !bytes.Equal(want, got) {
			t.Error("record schema changed")
		}
	})
}
