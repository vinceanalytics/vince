package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
)

func seenFirstEvent(ctx *sql.Context, domain string) (sql.RowIter, error) {
	req := v1.GetSiteRequest{Domain: domain}
	err := valid.Validate(&req)
	if err != nil {
		return nil, err
	}
	base := session.Get(ctx)
	if err = base.Allow(scopes.GetSite); err != nil {
		return nil, err
	}
	site, err := (&api.API{}).GetSite(base.Context(), &req)
	if err != nil {
		return nil, err
	}
	return rowToIter(site.SeenFirstEvent), nil
}
