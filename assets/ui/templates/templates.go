package templates

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/belak/octicon"
	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/models"
)

//go:embed layouts auth error email
var files embed.FS

var Login = template.Must(
	template.ParseFS(files,
		"layouts/focus.html",
		"auth/login_form.html",
	),
)

var Register = template.Must(
	template.ParseFS(files,
		"layouts/focus.html",
		"auth/register_form.html",
	),
)

var Error = template.Must(template.ParseFS(files,
	"error/error.html",
))

var ActivationEmail = template.Must(
	template.ParseFS(files,
		"layouts/focus.html",
		"email/activation_code.html",
	),
)

type Context struct {
	Title       string
	CurrentUser *models.User
	Data        map[string]any
	CSRF        template.HTML
	Captcha     template.HTML
	Errors      map[string]string
	Form        url.Values
	Code        uint64
	Config      *config.Config
}

func New(ctx context.Context, f ...func(c *Context)) *Context {
	c := &Context{
		Data:        make(map[string]any),
		CSRF:        getCsrf(ctx),
		Captcha:     getCaptcha(ctx),
		CurrentUser: models.GetCurrentUser(ctx),
		Config:      config.Get(ctx),
		Code:        auth.GetActivationCode(ctx),
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

func (t *Context) VinceURL() template.HTML {
	return template.HTML("http://localhost:8080")
}

func (t *Context) Validate(name string) template.HTML {
	if t.Errors != nil {
		o, _ := octicon.Icon("alert-fill", 12)
		if v, ok := t.Errors[name]; ok {
			return template.HTML(fmt.Sprintf(`
<div class="FormControl-inlineValidation">
    %s
    <span>%s</span>
</div>
		`, o, v))
		}
	}
	return template.HTML("")
}

func (t *Context) InputField(name string) template.HTMLAttr {
	var s strings.Builder
	if t.Errors != nil && t.Errors[name] != "" {
		s.WriteString(`invalid="true"`)
	}
	if t.Form != nil && t.Form.Get(name) != "" {
		s.WriteString(fmt.Sprintf("value=%q", t.Form.Get(name)))
	}
	return template.HTMLAttr(s.String())
}
