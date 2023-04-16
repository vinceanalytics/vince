package templates

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/belak/octicon"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/flash"
	"github.com/gernest/vince/internal/plans"
	"github.com/gernest/vince/models"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:embed layout pages plot docs site stats auth error email
var Files embed.FS

var LoginForm = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"layout/flash.html",
		"layout/csrf.html",
		"layout/captcha.html",
		"auth/login_form.html",
	),
)

var RegisterForm = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"layout/flash.html",
		"layout/csrf.html",
		"layout/captcha.html",
		"auth/register_form.html",
	),
)

var Error = template.Must(template.ParseFS(Files,
	"error/error.html",
))

var ActivationEmail = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"email/activation_code.html",
	),
)

var Activate = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"layout/flash.html",
		"layout/csrf.html",
		"auth/activate.html",
	),
)

var Home = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
	),
)

var Sites = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"plot/plot.html",
		"site/index.html",
		"layout/footer.html",
	),
)

var SiteNew = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"site/new.html",
	),
)

var Pricing = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"pages/pricing.html",
	),
)

var Markdown = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"pages/markdown.html",
	),
)

var DocsPage = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"docs/side_nav.html",
		"docs/page.html",
	),
)

var AddSnippet = template.Must(
	template.ParseFS(Files,
		"layout/focus.html",
		"layout/flash.html",
		"layout/csrf.html",
		"site/snippet.html",
	),
)

var Stats = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"stats/stats.html",
	),
)

var WaitingFirstPageView = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"stats/stats.html",
	),
)

var SiteLocked = template.Must(
	template.ParseFS(Files,
		"layout/app.html",
		"layout/csrf.html",
		"layout/header.html",
		"layout/flash.html",
		"layout/notice.html",
		"layout/footer.html",
		"stats/site_locked.html",
	),
)

type NewSite struct {
	IsFirstSite bool
	IsAtLimit   bool
	SiteLimit   int
}

type Errors struct {
	Status     int
	StatusText string
}

// For our logo the font used is Contrail One face 700-bold italic with size 150

type Context struct {
	Title         string
	CurrentUser   *models.User
	Data          map[string]any
	CSRF          template.HTML
	Captcha       template.HTMLAttr
	Errors        map[string]string
	Form          url.Values
	Code          uint64
	Config        *config.Config
	HasInvitation bool
	HasPin        bool
	Flash         *flash.Flash
	NewSite       *NewSite
	Error         *Errors
	Page          string
	Content       template.HTML
	ModTime       time.Time
	Site          *models.Site
	IsFIrstSite   bool
	SitesOverview []models.SiteOverView
	EmailReport   bool
	HasGoals      bool
	Owner         *models.User
	Docs          bool
	Pages         []*Page
	DocPage       *Page
}

type Page struct {
	Meta       Meta
	Next, Prev *Page
	Data       []byte
	Path       string
	UpdatedAt  time.Time
}

func (p *Page) Render(w io.Writer, pages Pages) error {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	x := parser.NewWithExtensions(extensions)
	doc := x.Parse(p.Data)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	b := markdown.Render(doc, renderer)
	return DocsPage.Execute(w, &Context{
		Title:   p.Meta.Title,
		Content: template.HTML(b),
		ModTime: p.UpdatedAt,
		Docs:    true,
		Pages:   pages,
		DocPage: p,
	})
}

func (p *Page) Read(path string, b []byte, modTime time.Time) {
	p.Path = path
	p.Data = p.Meta.Read(b)
	p.UpdatedAt = modTime
}

type Meta struct {
	Weight  int    `toml:"weight,omitempty"`
	Title   string `toml:"title,omitempty"`
	Section string `toml:"section,omitempty"`
	Layout  string `toml:"layout,omitempty"`
}

var marker = []byte("--- mark ---")

func (m *Meta) Read(src []byte) []byte {
	b := src
	start := bytes.Index(b, marker)
	if start == -1 {
		return src
	}
	b = b[start+len(marker):]
	last := bytes.Index(b, marker)
	if last == -1 {
		return src
	}
	chunk := b[:last]
	b = b[last+len(marker):]
	_, err := toml.Decode(string(chunk), m)
	if err != nil {
		log.Println("failed decoding meta "+err.Error(), string(chunk))
	}
	return b
}

type Pages []*Page

func (p Pages) Sort() Pages {
	sort.SliceStable(p, func(i, j int) bool {
		n := strings.Compare(p[i].Meta.Section, p[j].Meta.Section)
		if n == 0 {
			return p[i].Meta.Weight < p[j].Meta.Weight
		}
		return n == -1
	})
	var prev *Page
	for i, x := range p {
		x.Prev = prev
		if i+1 < len(p) {
			x.Next = p[i+1]
		}
		prev = x
	}
	return p
}

func (Context) Section(p Pages) string {
	if len(p) > 0 {
		return p[0].Meta.Section
	}
	return ""
}

func (Context) Sections(p Pages) (o []Pages) {
	var ls []*Page
	for _, v := range p {
		if v.Meta.Section == "" {
			continue
		}
		if len(ls) == 0 {
			ls = append(ls, v)
		} else {
			if ls[0].Meta.Section == v.Meta.Section {
				ls = append(ls, v)
			} else {
				o = append(o, ls)
				ls = []*Page{v}
			}
		}
	}
	if len(ls) > 0 {
		o = append(o, ls)
	}
	return
}

func New(ctx context.Context, f ...func(c *Context)) *Context {
	c := &Context{
		Data:        make(map[string]any),
		CSRF:        getCsrf(ctx),
		Captcha:     getCaptcha(ctx),
		CurrentUser: models.GetUser(ctx),
		Config:      config.Get(ctx),
		Code:        GetActivationCode(ctx),
		Flash:       flash.Get(ctx),
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

func (t *Context) VinceURL() template.HTML {
	return template.HTML("http://localhost:8080")
}

func (t *Context) Logo(width, height int) template.HTML {
	return template.HTML(fmt.Sprintf(
		`<img alt="Vince Analytics logo" width=%d height=%d src=%q>`,
		width, height, "/image/logo.svg",
	))
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

func (t *Context) Icon(name string, height int, class ...string) (template.HTML, error) {
	return octicon.IconTemplateFunc(name, height, class...)
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

type activationCodeKey struct{}

func SetActivationCode(ctx context.Context, code uint64) context.Context {
	return context.WithValue(ctx, activationCodeKey{}, code)
}

func GetActivationCode(ctx context.Context) uint64 {
	if v := ctx.Value(activationCodeKey{}); v != nil {
		return v.(uint64)
	}
	return 0
}

func (t *Context) Format(n uint64) string {
	switch {
	case n >= 1_000 && n < 1_000_000:
		thousands := (n / 100) / 10
		return fmt.Sprintf("%dK", thousands)
	case n >= 1_000_000 && n < 1_000_000_000:
		millions := (n / 100_000) / 10
		return fmt.Sprintf("%dM", millions)
	case n >= 1_000_000_000 && n < 1_000_000_000_000:
		billions := (n / 100_000_000) / 10
		return fmt.Sprintf("%dB", billions)
	default:
		return strconv.FormatUint(n, 10)
	}
}

func (Context) Plans() []plans.Plan {
	return plans.All
}

func (Context) Slogan() string {
	return "Self hosted, single file ,privacy friendly web analytics platform"
}

func (t *Context) Snippet() string {
	track := fmt.Sprintf("https://%s/js/script.js", t.Config.Url)
	if t.Site.CustomDomain != nil {
		track = fmt.Sprintf("https://%s/js/index.js", t.Site.CustomDomain.Domain)
	}
	src := fmt.Sprintf("<script defer data-domain=%q src=%q></script>", t.Site.Domain, track)
	return src
}
