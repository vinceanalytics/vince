package api

import (
	"errors"
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/dlclark/regexp2"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/render"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

func ListSites(w http.ResponseWriter, r *http.Request) {
	o := v1.Site_List{}
	db.Get(r.Context()).View(func(txn *badger.Txn) error {
		itO := badger.DefaultIteratorOptions
		prefix := keys.Site("")
		itO.Prefix = []byte(prefix)
		it := txn.NewIterator(itO)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			it.Item().Value(func(val []byte) error {
				var n v1.Site
				must.One(proto.Unmarshal(val, &n))("failed decoding site object")
				o.List = append(o.List, &n)
				return nil
			})
		}
		return nil
	})
	render.JSON(w, http.StatusOK, &o)
}

var domain = regexp2.MustCompile(`^(?!-)[A-Za-z0-9-]+([-.]{1}[a-z0-9]+)*.[A-Za-z]{2,6}$`, regexp2.ECMAScript)

func Create(w http.ResponseWriter, r *http.Request) {
	var b v1.Site_CreateOptions
	err := pj.UnmarshalDefault(&b, r.Body)
	if err != nil || b.Domain == "" {
		render.ERROR(w, http.StatusBadRequest)
		return
	}
	ok, _ := domain.MatchString(b.Domain)
	if !ok {
		render.ERROR(w, http.StatusBadRequest,
			"invalid domain name",
		)
		return
	}
	site := v1.Site{
		Domain: b.Domain,
	}
	err = db.Get(r.Context()).Update(func(txn *badger.Txn) error {
		key := keys.Site(b.Domain)
		it, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return txn.Set(
					[]byte(key),
					must.Must(proto.Marshal(&site))("failed encoding site object"),
				)
			}
			return err
		}
		return it.Value(func(val []byte) error {
			return proto.Unmarshal(val, &site)
		})
	})
	if err != nil {
		render.ERROR(w, http.StatusInternalServerError)
		return
	}
	render.JSON(w, http.StatusOK, &site)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	var b v1.Site_DeleteOptions
	err := pj.UnmarshalDefault(&b, r.Body)
	if err != nil || b.Domain == "" {
		render.ERROR(w, http.StatusBadRequest)
		return
	}
	var site v1.Site
	err = db.Get(r.Context()).Update(func(txn *badger.Txn) error {
		key := keys.Site(b.Domain)
		it, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return errors.Join(
			it.Value(func(val []byte) error {
				return proto.Unmarshal(val, &site)
			}),
			txn.Delete([]byte(key)),
		)
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			render.ERROR(w, http.StatusNotFound)
			return
		}
		render.ERROR(w, http.StatusInternalServerError)
		return
	}
	render.JSON(w, http.StatusOK, &site)
}
