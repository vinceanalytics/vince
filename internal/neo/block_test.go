package neo

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/events"
	"github.com/vinceanalytics/vince/internal/must"
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
	err = WriteBlock(context.Background(), txn, &buf, id.Bytes(), m.Record(now))
	if err != nil {
		t.Fatal(err)
	}
	txn.Commit()
	if buf.Len() == 0 {
		t.Fatal("expected data to be written")
	}
}
