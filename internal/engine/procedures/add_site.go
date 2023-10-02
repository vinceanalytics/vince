package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
)

func addSite(ctx *sql.Context, domain string) (sql.RowIter, error) {
	err := doAddSite(ctx, &v1.CreateSiteRequest{
		Domain: domain,
	})
	if err != nil {
		return nil, err
	}
	return rowToIter("ok"), nil
}

func addSiteWithDescription(ctx *sql.Context, domain, desc string) (sql.RowIter, error) {
	err := doAddSite(ctx, &v1.CreateSiteRequest{
		Domain:      domain,
		Description: desc,
	})
	if err != nil {
		return nil, err
	}
	return rowToIter("ok"), nil
}

func doAddSite(ctx *sql.Context, req *v1.CreateSiteRequest) error {
	err := valid.Validate(req)
	if err != nil {
		return err
	}
	base := session.Get(ctx)
	if err = base.Allow(scopes.CreateSite); err != nil {
		return err
	}
	_, err = (&api.API{}).CreateSite(base.Context(), req)
	return err
}
