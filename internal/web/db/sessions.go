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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/dchest/captcha"
	"github.com/google/uuid"
	"github.com/lestrrat-go/dataurl"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/ro2"
)

func newSession(path string) (*age.X25519Identity, error) {
	file := filepath.Join(path, "session")
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			secret, err := age.GenerateX25519Identity()
			if err != nil {
				return nil, err
			}
			return secret, os.WriteFile(file, []byte(secret.String()), 0600)
		}
		return nil, err
	}
	return age.ParseX25519Identity(string(data))
}

const MaxAge = 60 * 60 * 24 * 365 * 5

const cookie = "_vince"

type SessionContext struct {
	Data    Data
	captcha string
	user    *v1.User
	site    *v1.Site
	secret  *age.X25519Identity
}

func (s *SessionContext) clone() *SessionContext {
	return &SessionContext{secret: s.secret}
}
func (s *SessionContext) Context(base map[string]any) {
	if u := s.user; u != nil {
		base["current_user"] = map[string]any{
			"name":  u.Name,
			"id":    ro2.FormatID(u.Id),
			"email": u.Email,
			"admin": u.SuperUser,
		}
	}
	if s := s.site; s != nil {
		site := map[string]any{
			"domain": s.Domain,
			"id":     ro2.FormatID(s.Id),
			"public": s.Public,
		}
		base["site"] = site
	}

	if s.captcha != "" {
		base["captcha"] = template.HTMLAttr(fmt.Sprintf("src=%q", s.captcha))
	}
	if s.Data.Csrf != "" {
		base["csrf"] = template.HTML(s.Data.Csrf)
	}
	if f := s.Data.Flash; f != nil {
		base["flash"] = f
	}
}

type Data struct {
	TimeoutAt     time.Time `json:",omitempty"`
	CurrentUserID string    `json:",omitempty"`
	LastSeen      time.Time `json:",omitempty"`
	LoggedIn      bool      `json:",omitempty"`
	Captcha       string    `json:",omitempty"`
	Csrf          string    `json:",omitempty"`
	LoginDest     string    `json:",omitempty"`
	Flash         Flash     `json:",omitempty"`
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

func (c *Config) VerifyCaptchaSolution(r *http.Request) bool {
	r.ParseForm()
	digits := r.Form.Get("_captcha")
	digits = strings.TrimSpace(digits)
	if digits == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(digits), []byte(c.session.Data.Captcha)) == 1
}

func (c *Config) Wrap(f func(db *Config, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(c.clone(r), w, r)
	}
}

func (c *Config) Load(w http.ResponseWriter, r *http.Request) {
	c.load(r)
	if c.session.Data.CurrentUserID != "" {
		uid := uuid.MustParse(c.session.Data.CurrentUserID)
		usr := c.db.UserByID(uid)
		if usr == nil {
			c.session = c.session.clone()
			c.SaveSession(w)
		} else {
			c.session.user = usr
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
	now := time.Now().UTC()
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
	base["register"] = !c.disableRegistration
	c.session.Context(base)
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
		db:                  c.db,
		domains:             c.domains,
		cache:               c.cache,
		session:             c.session.clone(),
		disableRegistration: c.disableRegistration,
		logger:              c.logger.With(slog.String("path", r.URL.Path), "method", r.Method),
	}
}

func (s *SessionContext) Load(r *http.Request) error {
	cookie, err := r.Cookie(cookie)
	if err != nil {
		return nil
	}
	g, err := s.getSession(cookie.Value)
	if err != nil {
		return err
	}
	return json.Unmarshal(g, &s.Data)
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
	if c.session.user == nil {
		c.session.Data.LoginDest = r.URL.Path
		c.SaveSession(w)
		return false
	}
	return true
}

func (c *Config) Logout(w http.ResponseWriter) bool {
	if c.session.user != nil {
		c.session.Data.LoggedIn = true
		c.SaveSession(w)
		return false
	}
	return true
}

func (c *Config) CurrentUser() *v1.User {
	return c.session.user
}

func (c *Config) SetSite(site *v1.Site) {
	c.session.site = site
}

func (c *Config) CurrentSite() *v1.Site {
	return c.session.site
}

func (c *Config) Login(w http.ResponseWriter, uid uuid.UUID) string {
	c.session.Data.CurrentUserID = uid.String()
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
		Expires: time.Now().UTC().Add(time.Duration(MaxAge) * time.Second),
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

func (c *Config) SaveCaptcha(w http.ResponseWriter) {
	err := c.session.saveCaptcha(w)
	if err != nil {
		c.logger.Error("saving captcha", "err", err)
	}
}

func (s *SessionContext) saveCaptcha(w http.ResponseWriter) error {
	solution := captcha.RandomDigits(captcha.DefaultLen)
	img := captcha.NewImage("", solution, captcha.StdWidth, captcha.StdHeight)
	var b bytes.Buffer
	img.WriteTo(&b)
	data, err := dataurl.Encode(b.Bytes(), dataurl.WithBase64Encoding(true), dataurl.WithMediaType("image/png"))
	if err != nil {
		return err
	}
	s.Data.Captcha = formatCaptchaSolution(solution)
	s.save(w)
	s.captcha = string(data)
	return nil
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

func formatCaptchaSolution(sol []byte) string {
	var s strings.Builder
	s.Grow(len(sol))
	for _, b := range sol {
		s.WriteString(strconv.FormatInt(int64(b), 10))
	}
	return s.String()
}
