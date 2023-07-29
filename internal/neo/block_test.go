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
	"github.com/vinceanalytics/vince/internal/events"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/blocks"
	"github.com/vinceanalytics/vince/pkg/entry"
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
	start := must.Must(time.Parse(time.RFC822, time.RFC822)).UTC()
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
		r, err := ReadRecord(context.Background(), buf.Bytes(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		// os.WriteFile("testdata/basic_write.json",
		// 	must.Must(json.MarshalIndent(r, "", " ")), 0600)
		got := must.Must(json.MarshalIndent(r, "", " "))
		want := must.Must(os.ReadFile("testdata/basic_write.json"))
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
		got := must.Must(json.MarshalIndent(base.Record(), "", " "))
		want := must.Must(os.ReadFile("testdata/basic_write_base.json"))
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
		got := must.Must(json.MarshalIndent(base.Record(), "", " "))
		want := must.Must(os.ReadFile("testdata/basic_write_base_pick.json"))
		if !bytes.Equal(got, want) {
			t.Error("failed roundtrip")
		}
	})
	t.Run("can read base fields with pick and filter", func(t *testing.T) {
		base := NewBase([]string{"path"}, &blocks.Filter{
			Column: "path",
			Op:     blocks.Op_equal,
			Value: &blocks.Filter_Str{
				Str: "/home",
			},
		})
		defer base.Release()
		err := ReadBlock(context.Background(), db, id.Bytes(), base)
		if err != nil {
			t.Fatal(err)
		}
		got := must.Must(json.MarshalIndent(base.Record(), "", " "))
		want := must.Must(os.ReadFile("testdata/basic_write_base_pick_filter.json"))
		if !bytes.Equal(got, want) {
			t.Error("failed roundtrip")
		}
	})

}
