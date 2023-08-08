package api

import (
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
		prefix := []byte(v1.StorePrefix_SITES.String())
		itO.Prefix = prefix
		it := txn.NewIterator(itO)
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
