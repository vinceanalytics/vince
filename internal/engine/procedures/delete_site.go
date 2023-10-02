package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
)

func deleteSite(ctx *sql.Context, domain string) (sql.RowIter, error) {
	req := v1.DeleteSiteRequest{
		Domain: domain,
	}
	err := valid.Validate(&req)
	if err != nil {
		return nil, err
	}
	base := session.Get(ctx)
	if err = base.Allow(scopes.DeleteSite); err != nil {
		return nil, err
	}
	_, err = (&api.API{}).DeleteSite(base.Context(), &req)
	if err != nil {
		return nil, err
	}
	return rowToIter("ok"), nil
}
