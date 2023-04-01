package pages

import (
	"html/template"
	"io"
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/docs"
	"github.com/gernest/vince/render"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func Mark(page string, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := docs.FS.Open(page)
		if err != nil {
			render.ERROR(r.Context(), w, http.StatusNotFound)
			return
		}
		b, _ := io.ReadAll(f)
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse(b)
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)
		b = markdown.Render(doc, renderer)
		render.HTML(r.Context(), w, templates.Markdown, http.StatusOK, func(ctx *templates.Context) {
			ctx.Title = title
			ctx.Content = template.HTML(b)
			ctx.ModTime = docs.ModTime(page)
		})
	}
}
