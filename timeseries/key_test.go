package timeseries

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/dgraph-io/badger/v4"
)

func TestKey(t *testing.T) {
	var id ID

	uid := uint64(rand.Int63())

	t.Run("sets user id", func(t *testing.T) {
		id.SetUserID(uid)
		if id.GetUserID() != uid {
			t.Fatalf("expected %d got %d", uid, id.GetUserID())
		}
	})
	sid := uint64(rand.Int63())

	t.Run("sets site id", func(t *testing.T) {
		id.SetSiteID(sid)

		if id.GetUserID() != uid {
			t.Fatalf("expected %d got %d", uid, id.GetUserID())
		}
		if id.GetSiteID() != sid {
			t.Fatalf("expected %d got %d", sid, id.GetSiteID())
		}
	})

	t.Run("sets date", func(t *testing.T) {

		if id.GetUserID() != uid {
			t.Fatalf("expected %d got %d", uid, id.GetUserID())
		}
		if id.GetSiteID() != sid {
			t.Fatalf("expected %d got %d", sid, id.GetSiteID())
		}
	})
	t.Run("sets entropy", func(t *testing.T) {

		if id.GetUserID() != uid {
			t.Fatalf("expected %d got %d", uid, id.GetUserID())
		}
		if id.GetSiteID() != sid {
			t.Fatalf("expected %d got %d", sid, id.GetSiteID())
		}
	})
}

func TestYay(t *testing.T) {
	o := badger.DefaultOptions("").WithInMemory(true).WithNumVersionsToKeep(0)
	db, err := badger.Open(o)
	if err != nil {
		t.Fatal(err)
	}
	k := []byte("key")
	for i := 0; i < 10; i += 1 {
		db.Update(func(txn *badger.Txn) error {
			return txn.Set(k, []byte(strconv.Itoa(i)))
		})
	}
	k = []byte("bar")
	for i := 0; i < 10; i += 1 {
		db.Update(func(txn *badger.Txn) error {
			return txn.Set(k, []byte(strconv.Itoa(i)))
		})
	}
	db.View(func(txn *badger.Txn) error {
		o := badger.DefaultIteratorOptions
		o.AllVersions = true
		o.Reverse = true
		o.Prefix = []byte("key")
		it := txn.NewIterator(o)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			x := it.Item()
			var v string
			x.Value(func(val []byte) error {
				v = string(val)
				return nil
			})
			println(string(x.Key()), x.Version(), v)
		}
		return nil
	})
	t.Error()
}
