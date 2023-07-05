package ops

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/pkg/spec"
)

var systemTpl = templates.App("user/system.html")

func System(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	q := r.URL.Query()
	tab := q.Get("tab")
	if tab == "" {
		tab = "allocation_total"
	}
	sys := spec.System{
		Name: tab,
	}
	render.HTML(ctx, w, systemTpl, http.StatusOK, func(ctx *templates.Context) {
		ctx.USER = usr
		ctx.Header.Context = "system"
		ctx.Header.ContextRef = "/system"
		ctx.System = &sys
	})
}
