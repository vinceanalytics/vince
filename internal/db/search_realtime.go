package db

import (
	"context"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/bufbuild/protovalidate-go"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/bsi"
	"github.com/gernest/rows"
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
	result := new(realtimeQuery)
	err = db.Search(firstTime, now, []*v1.Filter{
		{Property: v1.Property_domain, Op: v1.Filter_equal, Value: req.SiteId},
	}, result)
	if err != nil {
		return nil, err
	}
	return &v1.Realtime_Response{Visitors: result.Visitors()}, nil
}

type realtimeQuery struct {
	roaring64.Bitmap
	fmt ViewFmt
}

func (r *realtimeQuery) View(_ time.Time) View {
	return r
}

func (r *realtimeQuery) Apply(tx *Tx, columns *rows.Row) error {
	view := r.fmt.Format(tx.View, "id")
	add := func(_, value uint64) error {
		r.Add(value)
		return nil
	}
	return tx.Cursor(view, func(c *rbf.Cursor) error {
		return bsi.Extract(c, tx.Shard, columns, add)
	})
}

func (r *realtimeQuery) Visitors() uint64 {
	return r.GetCardinality()
}
