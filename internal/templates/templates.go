package templates

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/flash"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/pkg/octicon"
	"github.com/vinceanalytics/vince/pkg/property"
	"github.com/vinceanalytics/vince/pkg/spec"
	"github.com/vinceanalytics/vince/pkg/timex"
)

//go:embed layout  plot site  auth error email user
var Files embed.FS

var Layouts = template.Must(
	base().ParseFS(Files,
		"layout/*.html",
	),
)

func Must(name, layout_ string) *template.Template {
	return template.Must(layout().ParseFS(Files, name)).Lookup(layout_)
}

func Focus(name string) *template.Template {
	return Must(name, "focus")
}

func App(name string) *template.Template {
	return Must(name, "app")
}

func Email(name string) *template.Template {
	return Must(name, "base_email")
}

func base() *template.Template {
	m := template.FuncMap{
		"Icon":       octicon.IconTemplateFunc,
		"Avatar":     Avatar,
		"Logo":       LogoText,
		"SafeDomain": models.SafeDomain,
		"JSON":       JSON,
	}
	return template.New("root").Funcs(m)
}

func JSON(a any) (string, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func layout() *template.Template {
	return template.Must(Layouts.Clone())
}

var Error = template.Must(
	template.ParseFS(Files,
		"error/error.html",
	))

type Errors struct {
	Status     int
	StatusText string
}

// For our logo the font used is Contrail One face 700-bold italic with size 150

type Context struct {
	Title     string
	Header    Header
	USER      *models.User
	Data      map[string]any
	CSRF      template.HTML
	Captcha   template.HTMLAttr
	Errors    map[string]string
	Form      url.Values
	Code      uint64
	ResetLink string
	Token     string
	Email     string
	Config    *config.Options
	HasPin    bool
	Flash     *flash.Flash
	Error     *Errors
	Site      *models.Site
	Recipient string
	Overview  *Overview
	Stats     *SiteStats
}

func (t *Context) ProfileOverview() string {
	return "/" + t.USER.Name
}

func (t *Context) ProfileSites() string {
	q := make(url.Values)
	q.Set("panel", "sites")
	return "/" + t.USER.Name + "?" + q.Encode()
}

type Header struct {
	Context    string
	Mode       string
	ContextRef string
}

type Overview struct {
	Global spec.Global[spec.Metrics]
	Sites  []SiteOverView
	Panel  string
}

type SiteOverView struct {
	Site   *models.Site
	Owner  string
	Global spec.Global[spec.Metrics]
}

type SiteStats struct {
	Site   *models.Site
	Owner  string
	Metric property.Metric
	Window timex.Duration
	Global spec.Global[spec.Metrics]
	Series spec.Series[[]uint64]
}

type Period struct {
	Name     string
	Selected bool
	Query    string
}

func (s *SiteStats) Periods() []Period {
	return []Period{
		s.period(timex.Today),
		s.period(timex.ThisWeek),
		s.period(timex.ThisMonth),
		s.period(timex.ThisYear),
	}
}

func (s *SiteStats) Metrics() []Metric {
	return []Metric{
		s.metric(property.Visitors),
		s.metric(property.Visits),
		s.metric(property.Events),
		s.metric(property.Views),
	}
}

func (s *SiteStats) period(d timex.Duration) Period {
	q := s.query()
	q.Set("w", d.String())
	return Period{
		Name:     d.String(),
		Selected: d == s.Window,
		Query:    fmt.Sprintf("/%s/%s?%s", s.Owner, s.Site.Domain, q.Encode()),
	}
}

type Metric struct {
	Name     string
	Query    string
	Selected bool
	Value    uint64
}

func (m *Metric) Select() string {
	if m.Selected {
		return "color-bg-accent"
	}
	return ""
}

func (s *SiteStats) metric(m property.Metric) Metric {
	q := s.query()
	q.Set("m", m.String())
	var value uint64
	switch m {
	case property.Visitors:
		value = s.Global.Result.Visitors
	case property.Views:
		value = s.Global.Result.Views
	case property.Events:
		value = s.Global.Result.Events
	case property.Visits:
		value = s.Global.Result.Visits
	}
	return Metric{
		Name:     m.Label(),
		Selected: s.Metric == m,
		Query:    fmt.Sprintf("/%s/%s?%s", s.Owner, s.Site.Domain, q.Encode()),
		Value:    value,
	}
}

func (s *SiteStats) query() url.Values {
	m := make(url.Values)
	m.Set("m", s.Metric.String())
	m.Set("w", s.Window.String())
	return m
}

func (t *Context) GreetRecipient() string {
	if t.Recipient == "" {
		return "Hey"
	}
	return "Hey " + strings.Split(t.Recipient, " ")[0]
}

func New(ctx context.Context, f ...func(c *Context)) *Context {
	c := &Context{
		Data:    make(map[string]any),
		CSRF:    getCsrf(ctx),
		Captcha: getCaptcha(ctx),
		USER:    models.GetUser(ctx),
		Config:  config.Get(ctx),
		Flash:   flash.Get(ctx),
		Errors:  make(map[string]string),
		Form:    make(url.Values),
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

func SetCaptcha(ctx context.Context, x template.HTMLAttr) context.Context {
	return context.WithValue(ctx, captchaTokenKey{}, x)
}

func SetCSRF(ctx context.Context, x template.HTML) context.Context {
	return context.WithValue(ctx, csrfTokenCtxKey{}, x)
}

func getCaptcha(ctx context.Context) template.HTMLAttr {
	if c := ctx.Value(captchaTokenKey{}); c != nil {
		return c.(template.HTMLAttr)
	}
	return template.HTMLAttr("")
}

func (t *Context) Home() string {
	return t.Config.URL + "/"
}

func Avatar(uid string, size uint, class ...string) template.HTML {
	q := make(url.Values)
	q.Set("u", uid)
	q.Set("s", strconv.Itoa(int(size)))
	u := "/avatar?" + q.Encode()
	return template.HTML(fmt.Sprintf(`<img class=%q width="%d" height="%d" src=%q>`,
		strings.Join(class, " "), size, size, u,
	))
}
