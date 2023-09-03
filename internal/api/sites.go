package api

import (
	"context"

	apiv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/px"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (a *API) CreateSite(ctx context.Context, req *apiv1.CreateSiteRequest) (*apiv1.CreateSiteResponse, error) {
	response := v1.CreateSiteResponse{
		Site: &v1.Site{
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

func (a *API) GetSite(ctx context.Context, req *apiv1.GetSiteRequest) (*apiv1.GetSiteResponse, error) {
	var o apiv1.Site
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
	return &apiv1.GetSiteResponse{Site: &o}, nil
}

func (a *API) ListSites(ctx context.Context, req *apiv1.ListSitesRequest) (*apiv1.ListSitesResponse, error) {
	o := v1.ListSitesResponse{}
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
				var n v1.Site
				must.One(proto.Unmarshal(val, &n))("failed decoding site object")
				o.List = append(o.List, &n)
				return nil
			})
		}
		return nil
	})
	return &o, nil
}

func (a *API) DeleteSite(ctx context.Context, req *apiv1.DeleteSiteRequest) (*apiv1.DeleteSiteResponse, error) {
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
	return &apiv1.DeleteSiteResponse{}, nil
}
