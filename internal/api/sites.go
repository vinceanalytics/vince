package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/timeseries"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

func ListSites(w http.ResponseWriter, r *http.Request) {
	o := []*v1.Site{}
	timeseries.Store(r.Context()).View(func(txn *badger.Txn) error {
		itO := badger.DefaultIteratorOptions
		prefix := (&v1.StoreKey{
			Prefix: v1.StorePrefix_SITES,
		}).Badger()
		itO.Prefix = prefix
		it := txn.NewIterator(itO)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			it.Item().Value(func(val []byte) error {
				var n v1.Site
				must.One(proto.Unmarshal(val, &n))()
				o = append(o, &n)
				return nil
			})
		}
		return nil
	})
	render.JSON(w, http.StatusOK, o)
}

func Create(w http.ResponseWriter, r *http.Request) {
	var b v1.Site
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil || b.Domain == "" {
		render.ERROR(w, http.StatusBadRequest)
		return
	}
	timeseries.Store(r.Context()).Update(func(txn *badger.Txn) error {
		key := (&v1.StoreKey{Prefix: v1.StorePrefix_SITES, Key: b.Domain}).Badger()
		it, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return txn.Set(
					key,
					must.Must(proto.Marshal(&b))(),
				)
			}
			return err
		}
		return it.Value(func(val []byte) error {
			return proto.Unmarshal(val, &b)
		})
	})
	render.JSON(w, http.StatusOK, &b)
}
