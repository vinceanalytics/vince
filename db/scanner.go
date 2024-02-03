package db

import (
	"context"

	"github.com/apache/arrow/go/v15/arrow"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
)

type Scanner interface {
	Scan(
		ctx context.Context,
		start, end int64,
		fs *v1.Filters,
	) (arrow.Record, error)
}
