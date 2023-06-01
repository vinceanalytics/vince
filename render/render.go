package render

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"net/http"

	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/templates"
)

func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func HTML(ctx context.Context, w http.ResponseWriter, tpl *template.Template, code int, f ...func(*templates.Context)) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(code)
	err := EXEC(ctx, w, tpl, f...)
	if err != nil {
		log.Get().Err(err).Str("template", tpl.Name()).Msg("Failed to render")
	}
}

func EXEC(ctx context.Context, w io.Writer, tpl *template.Template, f ...func(*templates.Context)) error {
	data := templates.New(ctx)
	if len(f) > 0 {
		f[0](data)
	}
	return tpl.Execute(w, data)
}

func ERROR(ctx context.Context, w http.ResponseWriter, code int, f ...func(*templates.Context)) {
	HTML(ctx, w, templates.Error, code, func(ctx *templates.Context) {
		ctx.Error = &templates.Errors{
			Status:     code,
			StatusText: http.StatusText(code),
		}
		if len(f) > 0 {
			f[0](ctx)
		}
	})
}
