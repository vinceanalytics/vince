package db

import (
	"context"
	"time"

	"github.com/bufbuild/protovalidate-go"
	"github.com/gernest/rbf"
	"github.com/gernest/roaring"
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

	var count uint64
	err = db.view(firstTime, now, req.SiteId, func(tx *view, r *rows.Row) error {
		count, err = uniqueUID(tx, r)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &v1.Realtime_Response{Visitors: count}, nil
}

func uniqueUID(txn *view, filters *rows.Row) (uint64, error) {
	c, err := txn.get("uid")
	if err != nil {
		return 0, err
	}
	count, _, err := sumCount(c, filters)
	return uint64(count), err
}
func sumCount(c *rbf.Cursor, filters *rows.Row) (int32, int64, error) {
	var filterData *roaring.Bitmap
	if filters != nil && len(filters.Segments) > 0 {
		filterData = filters.Segments[0].Data()
	}
	bsi := roaring.NewBitmapBSICountFilter(filterData)
	err := c.ApplyFilter(0, bsi)
	if err != nil {
		return 0, 0, err
	}
	// Sum is undefined
	count, sum := bsi.Total()
	return count, sum, nil
}
