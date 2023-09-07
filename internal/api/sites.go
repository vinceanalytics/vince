package api

import (
	"context"

	sitesv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var _ sitesv1.SitesServer = (*API)(nil)

func (a *API) CreateSite(ctx context.Context, req *sitesv1.CreateSiteRequest) (*sitesv1.CreateSiteResponse, error) {
	response := sitesv1.CreateSiteResponse{
		Site: &sitesv1.Site{
			Domain: req.Domain,
		},
	}
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
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
		return nil, err
	}
	return &response, nil
}

func (a *API) GetSite(ctx context.Context, req *sitesv1.GetSiteRequest) (*sitesv1.GetSiteResponse, error) {
	var o sitesv1.Site
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		defer key.Release()
		return txn.Get(
			key.Bytes(), px.Decode(&o),
			func() error {
				return status.Error(codes.NotFound, "site not found")
			},
		)
	})
	if err != nil {
		return nil, err
	}
	return &sitesv1.GetSiteResponse{Site: &o}, nil
}

func (a *API) ListSites(ctx context.Context, req *sitesv1.ListSitesRequest) (*sitesv1.ListSitesResponse, error) {
	o := sitesv1.ListSitesResponse{}
	db.Get(ctx).Txn(false, func(txn db.Txn) error {
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
				var n sitesv1.Site
				must.One(proto.Unmarshal(val, &n))("failed decoding site object")
				o.List = append(o.List, &n)
				return nil
			})
		}
		return nil
	})
	return &o, nil
}

func (a *API) DeleteSite(ctx context.Context, req *sitesv1.DeleteSiteRequest) (*sitesv1.DeleteSiteResponse, error) {
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		defer key.Release()
		return txn.Delete(
			key.Bytes(),
			func() error {
				return status.Error(codes.NotFound, "site not found")
			},
		)
	})
	if err != nil {
		return nil, err
	}
	return &sitesv1.DeleteSiteResponse{}, nil
}
