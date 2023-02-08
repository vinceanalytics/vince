package templates

import (
	"context"
	"embed"
	"html/template"
	"net/url"

	"github.com/belak/octicon"
	"github.com/gernest/vince/models"
)

//go:embed layouts auth error
var files embed.FS

var Login = template.Must(
	template.ParseFS(files,
		"layouts/focus.html",
		"auth/login_form.html",
	),
)

var Register = template.Must(
	template.Must(template.ParseFS(files,
		"layouts/focus.html",
	)).Funcs(funcs).ParseFS(files,
		"auth/register_form.html",
	),
)

var Error = template.Must(template.ParseFS(files,
	"error/error.html",
))

var funcs = template.FuncMap{
	"oAlertFill": wrap(octicon.AlertFill),
}

func wrap(f func(int, ...string) (string, bool)) func(int) template.HTML {
	return func(i int) template.HTML {
		v, _ := f(i)
		return template.HTML(v)
	}
}

type Context struct {
	Title       string
	CurrentUser *models.User
	Data        map[string]any
	CSRF        template.HTML
	Captcha     template.HTML
	Errors      map[string]string
	Form        url.Values
}

func New(ctx context.Context, f ...func(c *Context)) *Context {
	c := &Context{
		Data:    make(map[string]any),
		CSRF:    getCsrf(ctx),
		Captcha: getCaptcha(ctx),
	}
	if len(f) > 0 {
		f[0](c)
	}
	return c
}

type csrfTokenCtxKey struct{}

func getCsrf(ctx context.Context) template.HTML {
	if c := ctx.Value(csrfTokenCtxKey{}); c != nil {
		return c.(template.HTML)
	}
	return template.HTML("")
}

type captchaTokenKey struct{}

func getCaptcha(ctx context.Context) template.HTML {
	if c := ctx.Value(captchaTokenKey{}); c != nil {
		return c.(template.HTML)
	}
	return template.HTML("")
}

func SecureForm(ctx context.Context, csrf, captcha template.HTML) context.Context {
	ctx = context.WithValue(ctx, csrfTokenCtxKey{}, csrf)
	return context.WithValue(ctx, captchaTokenKey{}, captcha)
}
