package db

import (
	"context"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/bufbuild/protovalidate-go"
	"github.com/gernest/rbf/dsl/bsi"
	"github.com/gernest/rbf/dsl/mutex"
	"github.com/gernest/rbf/dsl/query"
	"github.com/gernest/rbf/dsl/tx"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/defaults"
	"github.com/vinceanalytics/vince/internal/logger"
)

var validate *protovalidate.Validator

func init() {
	var err error
	validate, err = protovalidate.New(protovalidate.WithFailFast(true))
	if err != nil {
		logger.Fail("Failed setting up validator", "err", err)
	}
}

func (db *DB) Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error) {
	defaults.Set(req)
	err := validate.Validate(req)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	firstTime := now.Add(-5 * time.Minute)

	r, err := db.db.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Release()
	o := roaring64.New()
	fs := append(query.And{
		bsi.Filter("timestamp", bsi.RANGE, firstTime.UnixMilli(), now.UnixMilli()),
	}, filterProperties(
		&v1.Filter{Property: v1.Property_domain, Op: v1.Filter_equal, Value: req.SiteId},
	))
	for _, shard := range r.Range(now, now) {
		err = r.View(shard, func(txn *tx.Tx) error {
			r, err := fs.Apply(txn, nil)
			if err != nil {
				return err
			}
			if r.IsEmpty() {
				return nil
			}
			return mutex.Distinct(txn, "id", o, r)
		})
		if err != nil {
			return nil, err
		}
	}
	return &v1.Realtime_Response{Visitors: o.GetCardinality()}, nil
}
