package db

import (
	"context"
	"time"

	"github.com/bufbuild/protovalidate-go"
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

	count := make(map[uint64]struct{})
	err = db.view(firstTime, now, req.SiteId, func(tx *view, r *rows.Row) error {
		return tx.uidCount(r, count)
	})
	if err != nil {
		return nil, err
	}
	return &v1.Realtime_Response{Visitors: uint64(len(count))}, nil
}
