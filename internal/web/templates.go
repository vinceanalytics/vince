package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/url"

	"github.com/vinceanalytics/vince/internal/config"
)

//go:embed templates
var templateData embed.FS

var (
	layouts = template.Must(template.ParseFS(
		templateData, "templates/layout/*",
	)).Funcs(funcMap())
	home          = template.Must(look("focus").ParseFS(templateData, "templates/page/index.html"))
	license       = template.Must(look("focus").ParseFS(templateData, "templates/page/license.html"))
	login         = template.Must(look("focus").ParseFS(templateData, "templates/auth/login.html"))
	statsLocked   = template.Must(look("focus").ParseFS(templateData, "templates/stats/site_locked.html"))
	sharePassword = template.Must(look("focus").ParseFS(templateData, "templates/stats/shared_link_password.html"))
	stats         = template.Must(look("app").ParseFS(templateData, "templates/stats/stats.html"))
	createSite    = template.Must(look("focus").ParseFS(templateData, "templates/site/new.html"))
	shared        = template.Must(look("focus").ParseFS(templateData, "templates/site/new_shared_link.html"))
	edit          = template.Must(look("focus").ParseFS(templateData, "templates/site/edit_shared_link.html"))
	sitesIndex    = template.Must(look("app").ParseFS(templateData, "templates/site/index.html"))
	siteSettings  = template.Must(look("app").ParseFS(templateData, "templates/site/settings.html"))
	userSettings  = template.Must(look("app").ParseFS(templateData, "templates/auth/settings.html"))
	newAPIKey     = template.Must(look("focus").ParseFS(templateData, "templates/auth/new_api_key.html"))
	addSnippet    = template.Must(look("focus").ParseFS(templateData, "templates/site/snippet.html"))

	e404 = template.Must(look("focus").ParseFS(templateData, "templates/error/404.html"))
	e500 = template.Must(look("focus").ParseFS(templateData, "templates/error/500.html"))
)

func look(name string) *template.Template {
	return template.Must(layouts.Lookup(name).Clone())
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"map":            mapStruct,
		"render_snippet": renderSnippet,
		"path_escape":    pathEscape,
	}
}

func pathEscape(value string) string {
	return url.PathEscape(value)
}

func renderSnippet(domain string) string {
	tracker := fmt.Sprintf("%s/js/script.js", config.C.Url)
	return fmt.Sprintf(`<script defer data-domain=%q src=%q></script>`, domain, tracker)
}

func mapStruct(values ...any) (o map[string]any) {
	o = make(map[string]any)
	for len(values) > 1 {
		o[fmt.Sprint(values[0])] = values[1]
		values = values[2:]
	}
	return
}
