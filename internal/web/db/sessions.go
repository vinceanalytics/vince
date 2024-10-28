package db

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/trace"
	"strconv"
	"time"

	"filippo.io/age"
	"github.com/google/uuid"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/util/oracle"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

const MaxAge = 60 * 60 * 24 * 365 * 5

const cookie = "_vince"

type SessionContext struct {
	site   *v1.Site
	secret *age.X25519Identity
	user   string
	Data   Data
}

func (s *SessionContext) clone() *SessionContext {
	return &SessionContext{secret: s.secret}
}

type Data struct {
	TimeoutAt     time.Time `json:",omitempty"`
	LastSeen      time.Time `json:",omitempty"`
	Flash         Flash     `json:",omitempty"`
	CurrentUserID string    `json:",omitempty"`
	Csrf          string    `json:",omitempty"`
	LoginDest     string    `json:",omitempty"`
	LoggedIn      bool      `json:",omitempty"`
}

func (s *SessionContext) SuccessFlash(m string) *SessionContext {
	if s.Data.Flash == nil {
		s.Data.Flash = make(Flash)
	}
	s.Data.Flash.Success(m)
	return s
}

func (s *SessionContext) FailFlash(m string) *SessionContext {
	if s.Data.Flash == nil {
		s.Data.Flash = make(Flash)
	}
	s.Data.Flash.Error(m)
	return s
}

func (c *Config) Wrap(kind string) func(f func(db *Config, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(f func(db *Config, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, task := trace.NewTask(r.Context(), kind)
			defer task.End()
			f(c.clone(r), w, r.WithContext(ctx))
		}
	}
}

func (c *Config) Load(w http.ResponseWriter, r *http.Request) {
	c.load(r)
	if c.session.Data.CurrentUserID != "" {
		if c.ops.Admin() != c.session.Data.CurrentUserID {
			c.session = c.session.clone()
			c.SaveSession(w)
		} else {
			c.session.user = c.session.Data.CurrentUserID
		}
	}
}

func (c *Config) Flash(w http.ResponseWriter) {
	if c.session.Data.Flash != nil {
		flash := c.session.Data.Flash
		c.session.Data.Flash = nil
		// we update session without flash
		c.SaveSession(w)
		// bring them back so they can be available in templates
		c.session.Data.Flash = flash
	}
}

func (c *Config) SessionTimeout(w http.ResponseWriter) {
	now := xtime.Now().UTC()
	switch {
	case c.session.Data.CurrentUserID != "" && !c.session.Data.TimeoutAt.IsZero() && now.After(c.session.Data.TimeoutAt):
		c.session.Data = Data{}
		c.SaveSession(w)
	case c.session.Data.CurrentUserID != "":
		c.session.Data.TimeoutAt = now.Add(24 * 7 * 2 * time.Hour)
		c.SaveSession(w)
	}
}

func (c *Config) Context(base map[string]any) map[string]any {
	if base == nil {
		base = make(map[string]any)
	}
	s := c.session
	if u := s.user; u != "" {
		base["current_user"] = map[string]any{
			"name":  u,
			"email": u,
			"admin": true,
		}
	}
	if s := s.site; s != nil {
		site := map[string]any{
			"domain": s.Domain,
			"public": s.Public,
			"locked": s.Locked,
		}
		share := make([]map[string]any, 0, len(s.Shares))
		u := oracle.Endpoint
		p := u + fmt.Sprintf("/v1/share/%s?auth=", url.PathEscape(s.Domain))

		for _, r := range s.Shares {
			m := map[string]any{
				"slug":     r.Id,
				"name":     r.Name,
				"dest":     p + r.Id,
				"password": len(r.Password) > 0,
			}
			share = append(share, m)
		}
		if len(share) > 0 {
			site["share"] = share
		}
		goals := make([]string, 0, len(s.Goals))
		for _, g := range s.Goals {
			if g.Path != "" {
				goals = append(goals, "Visit "+g.Path)
			} else {
				goals = append(goals, g.Name)
			}
		}
		if len(goals) > 0 {
			site["goals"] = goals
		}
		base["site"] = site
	}

	if s.Data.Csrf != "" {
		base["csrf"] = template.HTML(s.Data.Csrf)
	}
	if f := s.Data.Flash; f != nil {
		base["flash"] = f
	}
	return base
}

func (c *Config) load(r *http.Request) {
	err := c.session.Load(r)
	if err != nil {
		c.logger.Error("loading session", "err", err)
	}
}

func (c *Config) clone(r *http.Request) *Config {
	return &Config{
		db:      c.db,
		ts:      c.ts,
		ops:     c.ops,
		cache:   c.cache,
		session: c.session.clone(),
		buffer:  c.buffer,
		logger:  c.logger.With(slog.String("path", r.URL.Path), "method", r.Method),
	}
}

func (s *SessionContext) Load(r *http.Request) error {
	g, err := s.LoadBase(r, cookie)
	if err != nil {
		return err
	}
	if len(g) == 0 {
		return nil
	}
	return json.Unmarshal(g, &s.Data)
}

func (s *SessionContext) LoadBase(r *http.Request, name string) ([]byte, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil, nil
	}
	return s.getSession(cookie.Value)
}

func (s *SessionContext) getSession(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}
	base, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	r, err := age.Decrypt(bytes.NewReader(base), s.secret)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

func (c *Config) Authorize(w http.ResponseWriter, r *http.Request) bool {
	if c.session.user == "" {
		c.session.Data.LoginDest = r.URL.Path
		c.SaveSession(w)
		return false
	}
	return true
}

func (c *Config) IsLoggedOut(w http.ResponseWriter) bool {
	if c.session.user != "" {
		c.session.Data.LoggedIn = true
		c.SaveSession(w)
		return false
	}
	return true
}

func (c *Config) Logout(w http.ResponseWriter) {
	c.session.Data = Data{}
	c.SaveSession(w)
}

func (c *Config) CurrentUser() string {
	return c.session.user
}

func (c *Config) SetSite(site *v1.Site) {
	c.session.site = site
}

func (c *Config) CurrentSite() *v1.Site {
	return c.session.site
}

func (c *Config) Login(w http.ResponseWriter) string {
	c.session.Data.CurrentUserID = c.ops.Admin()
	c.session.Data.LoggedIn = true
	dest := c.session.Data.LoginDest
	c.session.Data.LoginDest = ""
	c.SaveSession(w)
	if dest == "" {
		return "/sites"
	}
	return dest
}

func (c *Config) SaveSuccessRegister(w http.ResponseWriter, uid uuid.UUID) {
	c.session.Data.CurrentUserID = uid.String()
	c.session.Data.LoggedIn = true
	c.SaveSession(w)
}

func (c *Config) Success(msg string) {
	c.session.SuccessFlash(msg)
}

func (c *Config) SaveSession(w http.ResponseWriter) {
	err := c.session.save(w)
	if err != nil {
		c.logger.Error("saving session", "err", err)
	}
}

func (s *SessionContext) save(w http.ResponseWriter) error {
	b, _ := json.Marshal(s.Data)
	value, err := s.create(b)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Path:    "/",
		Name:    cookie,
		Value:   value,
		Expires: xtime.Now().UTC().Add(time.Duration(MaxAge) * time.Second),
	}
	http.SetCookie(w, cookie)
	return nil
}

func (db *Config) SaveSharedLinkSession(w http.ResponseWriter, name string) {
	err := db.session.SaveSharedLink(w, name)
	if err != nil {
		db.logger.Error("saving shared link session", "err", err)
	}
}

func (db *Config) LoadSharedLinkSession(r *http.Request, name string) time.Time {
	data, err := db.session.LoadBase(r, name)
	if err != nil {
		db.logger.Error("reading shared link session", "err", err)
		return time.Time{}
	}
	ts, _ := strconv.ParseInt(string(data), 10, 64)
	return xtime.UnixMilli(ts).UTC()
}

func (s *SessionContext) SaveSharedLink(w http.ResponseWriter, name string) error {
	vs := strconv.FormatInt(xtime.Now().Add(24*time.Hour).UTC().UnixMilli(), 10)
	value, err := s.create([]byte(vs))
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Path:    "/",
		Name:    name,
		Value:   value,
		Expires: xtime.Now().UTC().Add(time.Duration(60*60*24) * time.Second),
	}
	http.SetCookie(w, cookie)
	return nil
}

func (s *SessionContext) create(b []byte) (string, error) {
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, s.secret.Recipient())
	if err != nil {
		return "", err
	}
	_, err = w.Write(b)
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (c *Config) SaveCsrf(w http.ResponseWriter) {
	err := c.session.saveCsrf(w)
	if err != nil {
		c.logger.Error("saving csrf", "err", err)
	}
}

func (s *SessionContext) saveCsrf(w http.ResponseWriter) error {
	token := make([]byte, 32)
	rand.Read(token)
	csrf := base64.StdEncoding.EncodeToString(token)
	s.Data.Csrf = csrf
	return s.save(w)
}

func (c *Config) IsValidCsrf(r *http.Request) bool {
	r.ParseForm()
	value := r.Form.Get("_csrf")
	return subtle.ConstantTimeCompare([]byte(value), []byte(c.session.Data.Csrf)) == 1
}
