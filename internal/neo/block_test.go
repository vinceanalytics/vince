package neo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/events"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/entry"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func TestWriteBlock_basic(t *testing.T) {
	var requests []*entry.Request
	b, err := os.ReadFile("testdata/basic.json")
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(b, &requests)
	if err != nil {
		t.Fatal(err)
	}
	start := must.Must(time.Parse(time.RFC822, time.RFC822))().UTC()
	m := entry.NewMulti()
	for i, r := range requests {
		e, err := events.Parse(r, start.Add(time.Duration(i)*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		m.Append(e)
	}
	db, err := badger.Open(badger.DefaultOptions("").
		WithInMemory(true).WithLoggingLevel(10))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	txn := db.NewTransaction(true)
	id := ulid.Make()
	var buf bytes.Buffer
	now := start.Add(4 * time.Hour).UnixMilli()
	t.Run("can write parquet file from a record", func(t *testing.T) {
		err = WriteBlock(context.Background(), txn, &buf, id.Bytes(), m.Record(now))
		if err != nil {
			t.Fatal(err)
		}
		txn.Commit()
		if buf.Len() == 0 {
			t.Fatal("expected data to be written")
		}
	})

	t.Run("can read written data", func(t *testing.T) {
		r, err := ReadRecord(context.Background(), bytes.NewReader(buf.Bytes()), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		got := must.Must(json.MarshalIndent(r, "", " "))()
		want := must.Must(os.ReadFile("testdata/basic_write.json"))()
		if !bytes.Equal(got, want) {
			t.Error("failed roundtrip")
		}
	})

	t.Run("written data is committed to badger", func(t *testing.T) {
		err := db.View(func(txn *badger.Txn) error {
			it, err := txn.Get(id.Bytes())
			if err != nil {
				return err
			}
			return it.Value(func(val []byte) error {
				if !bytes.Equal(buf.Bytes(), val) {
					return errors.New("wrong value in badger")
				}
				return nil
			})
		})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("can read base fields", func(t *testing.T) {
		base := NewBase([]string{})
		defer base.Release()
		err := ReadBlock(context.Background(), db, id.Bytes(), base)
		if err != nil {
			t.Fatal(err)
		}
		got := must.Must(json.MarshalIndent(base.Record(), "", " "))()
		want := must.Must(os.ReadFile("testdata/basic_write_base.json"))()
		if !bytes.Equal(got, want) {
			t.Error("failed roundtrip")
		}
	})

	t.Run("can read base fields with pick", func(t *testing.T) {
		base := NewBase([]string{"path"})
		err := ReadBlock(context.Background(), db, id.Bytes(), base)
		if err != nil {
			t.Fatal(err)
		}
		got := must.Must(json.MarshalIndent(base.Record(), "", " "))()
		want := must.Must(os.ReadFile("testdata/basic_write_base_pick.json"))()
		if !bytes.Equal(got, want) {
			t.Error("failed roundtrip")
		}
	})
	t.Run("can read base fields with pick and filter", func(t *testing.T) {
		base := NewBase([]string{"path"}, &v1.Filter{
			Column: "path",
			Op:     v1.Op_equal,
			Value: &v1.Filter_Str{
				Str: "/home",
			},
		})
		defer base.Release()
		err := ReadBlock(context.Background(), db, id.Bytes(), base)
		if err != nil {
			t.Fatal(err)
		}
		got := must.Must(json.MarshalIndent(base.Record(), "", " "))()
		want := must.Must(os.ReadFile("testdata/basic_write_base_pick_filter.json"))()
		if !bytes.Equal(got, want) {
			t.Error("failed roundtrip")
		}
	})

	t.Run("written block can be indexed", func(t *testing.T) {
		idx := Index(context.Background(), buf.Bytes())
		got := FindRowGroups(&idx, []string{"path"}, []string{"/home"})[0]
		want := 0
		if got != want {
			t.Errorf("expected row group %d got %d", want, got)
		}
	})
}

func TestTransform(t *testing.T) {
	ctx := entry.Context(context.Background())
	t.Run("full file transform no partitions", func(t *testing.T) {
		f := must.Must(os.Open("testdata/block.parquet"))()
		defer f.Close()

		base := NewBase(nil)
		defer base.Release()

		r := must.Must(ReadRecord(ctx, f, base.ColumnIndices(), nil))()
		base.Analyze(ctx, r)
		r = base.Record()
		defer r.Release()

		// consistent timestamp.
		ctx = core.SetNow(ctx, func() time.Time {
			return must.Must(time.Parse(time.RFC822, time.RFC822))()
		})

		r = Transform(ctx, r, 0, 0)
		got := must.Must(json.MarshalIndent(r, "", " "))()
		// os.WriteFile("testdata/block_full_no_partitions.json", got, 0600)
		want := must.Must(os.ReadFile("testdata/block_full_no_partitions.json"))
		if !bytes.Equal(got, want()) {
			t.Error("wrong full computation")
		}
	})
}

func TestMetrics(t *testing.T) {
	ctx := entry.Context(context.Background())

	t.Run("compute", func(t *testing.T) {
		f := must.Must(os.Open("testdata/block.parquet"))()
		defer f.Close()

		base := NewBase(nil)
		defer base.Release()

		r := must.Must(ReadRecord(ctx, f, base.ColumnIndices(), nil))()
		base.Analyze(ctx, r)
		r = base.Record()
		defer r.Release()

		var m Metrics
		m.Compute(ctx, r)
		got := must.Must(json.MarshalIndent(m, "", " "))()
		os.WriteFile("testdata/metrics_compute.json", got, 0600)
		want := must.Must(os.ReadFile("testdata/metrics_compute.json"))
		if !bytes.Equal(got, want()) {
			t.Error("wrong full computation")
		}
	})
}
