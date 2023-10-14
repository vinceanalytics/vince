package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

func forceSave(ctx *sql.Context, domain string) (sql.RowIter, error) {
	req := v1.GetSiteRequest{Domain: domain}
	err := valid.Validate(&req)
	if err != nil {
		return nil, err
	}
	base := session.Get(ctx)
	if err = base.Allow(scopes.CreateSite); err != nil {
		return nil, err
	}
	bctx := base.Context()
	ok := timeseries.Block(bctx).ForceSave(bctx, domain)
	return rowToIter(ok), nil
}
