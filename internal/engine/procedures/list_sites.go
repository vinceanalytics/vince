package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
)

func listSites(ctx *sql.Context) (sql.RowIter, error) {
	req := v1.ListSitesRequest{}
	err := valid.Validate(&req)
	if err != nil {
		return nil, err
	}
	base := session.Get(ctx)
	if err = base.Allow(scopes.ListSites); err != nil {
		return nil, err
	}
	ls, err := (&api.API{}).ListSites(base.Context(), &req)
	if err != nil {
		return nil, err
	}
	return sitesToRowIter(ls.List...), nil
}

func sitesToRowIter(sites ...*v1.Site) sql.RowIter {
	rows := make([]sql.Row, len(sites))
	for i := range sites {
		rows[i] = siteToRow(sites[i])
	}
	return sql.RowsToRowIter(rows...)
}

func siteToRow(site *v1.Site) (o sql.Row) {
	o = make(sql.Row, 7)
	o[0] = site.Domain
	o[1] = site.Description
	b := site.Stats.BaseStats
	o[2] = b.PageViews
	o[3] = b.Visitors
	o[4] = b.Visits
	o[5] = b.BounceRate
	o[7] = b.Duration
	return
}

func siteSchema() sql.Schema {
	return []*sql.Column{
		{Name: "domain", Type: types.Text},
		{Name: "description", Type: types.Text},
		{Name: "page_views", Type: types.Int64},
		{Name: "visitors", Type: types.Int64},
		{Name: "visits", Type: types.Int64},
		{Name: "bounce_rate", Type: types.Float64},
		{Name: "visit_duration", Type: types.Float64},
	}
}
