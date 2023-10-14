package control

import (
	"context"

	v1 "github.com/vinceanalytics/proto/gen/go/vince/sites/v1"
)

type VinceAPI interface {
	List(context.Context) ([]*v1.Site, error)
	Get(ctx context.Context, domain string) (*v1.Site, error)
	Create(ctx context.Context, domain string, public bool) (*v1.Site, error)
}
