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
	o := v1.Site_List_Response{}
	db.Get(r.Context()).Txn(false, func(txn db.Txn) error {
		key := keys.Site("")
		defer key.Release()

		it := txn.Iter(db.IterOpts{
			Prefix:         key.Bytes(),
			PrefetchSize:   64,
			PrefetchValues: true,
		})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			it.Value(func(val []byte) error {
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
	var b v1.Site_Create_Request
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
	response := v1.Site_Create_Response{
		Site: &v1.Site{
			Domain: b.Domain,
		},
	}
	err = db.Get(r.Context()).Txn(true, func(txn db.Txn) error {
		key := keys.Site(b.Domain)
		defer key.Release()
		return txn.Get(key.Bytes(), func(val []byte) error {
			return proto.Unmarshal(val, response.Site)
		}, func() error {
			return txn.Set(
				key.Bytes(),
				must.Must(proto.Marshal(response.Site))("failed encoding site object"),
			)
		})
	})
	if err != nil {
		render.ERROR(w, http.StatusInternalServerError)
		return
	}
	render.JSON(w, http.StatusOK, &response)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	var b v1.Site_Delete_Request
	err := pj.UnmarshalDefault(&b, r.Body)
	if err != nil || b.Domain == "" {
		render.ERROR(w, http.StatusBadRequest)
		return
	}
	err = db.Get(r.Context()).Txn(true, func(txn db.Txn) error {
		key := keys.Site(b.Domain)
		defer key.Release()
		return txn.Delete(key.Bytes())
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			render.ERROR(w, http.StatusNotFound)
			return
		}
		render.ERROR(w, http.StatusInternalServerError)
		return
	}
	render.JSON(w, http.StatusOK, &v1.Site_Delete_Response{})
}
