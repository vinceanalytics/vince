package api

import (
	"context"

	sitesv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ sitesv1.SitesServer = (*API)(nil)

var sites404 = status.Error(codes.NotFound, "site does not exist")

func Sites404() error {
	return sites404
}

func (a *API) CreateSite(ctx context.Context, req *sitesv1.CreateSiteRequest) (*sitesv1.CreateSiteResponse, error) {
	response := sitesv1.CreateSiteResponse{
		Site: &sitesv1.Site{
			Domain:      req.Domain,
			Description: req.Description,
		},
	}
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		return txn.Get(key, px.Decode(response.Site), func() error {
			return txn.Set(key, px.Encode(response.Site))
		})
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (a *API) GetSite(ctx context.Context, req *sitesv1.GetSiteRequest) (*sitesv1.Site, error) {
	var o sitesv1.Site
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		return txn.Get(key, px.Decode(&o), Sites404)
	})
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (a *API) ListSites(ctx context.Context, req *sitesv1.ListSitesRequest) (*sitesv1.ListSitesResponse, error) {
	o := sitesv1.ListSitesResponse{}
	db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Site("")
		it := txn.Iter(db.IterOpts{
			Prefix:         key,
			PrefetchSize:   64,
			PrefetchValues: true,
		})
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			var m sitesv1.Site
			err := it.Value(px.Decode(&m))
			if err != nil {
				return err
			}
			o.List = append(o.List, &m)
		}
		return nil
	})
	return &o, nil
}

func (a *API) DeleteSite(ctx context.Context, req *sitesv1.DeleteSiteRequest) (*sitesv1.DeleteSiteResponse, error) {
	err := db.Get(ctx).Txn(true, func(txn db.Txn) error {
		key := keys.Site(req.Domain)
		return txn.Delete(key, Sites404)
	})
	if err != nil {
		return nil, err
	}
	return &sitesv1.DeleteSiteResponse{}, nil
}
