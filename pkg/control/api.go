package control

import (
	"context"

	"github.com/gernest/vince/pkg/schema"
)

type VinceAPI interface {
	List(context.Context) ([]*schema.Site, error)
	Get(ctx context.Context, domain string) (*schema.Site, error)
	Create(ctx context.Context, domain string, public bool) (*schema.Site, error)
}
