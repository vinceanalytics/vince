package sessions

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/dchest/captcha"
	"github.com/lestrrat-go/dataurl"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/flash"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/pkg/log"
)

const MaxAge = 60 * 60 * 24 * 365 * 5

type Session string

type sessionsKey struct{}

func Set(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionsKey{}, s)
}

func Get(ctx context.Context) Session {
	return ctx.Value(sessionsKey{}).(Session)
}

func NewSession(name string) Session {
	return Session(name)
}

type SessionContext struct {
	Data Data
	s    Session
}

type Data struct {
	TimeoutAt     time.Time       `json:",omitempty"`
	CurrentUserID uint64          `json:",omitempty"`
	LastSeen      time.Time       `json:",omitempty"`
	LoggedIn      bool            `json:",omitempty"`
	Captcha       string          `json:",omitempty"`
	Csrf          string          `json:",omitempty"`
	LoginDest     string          `json:",omitempty"`
	Flash         *flash.Flash    `json:",omitempty"`
	EmailReport   map[string]bool `json:",omitempty"`
}

func (s *SessionContext) SuccessFlash(m string) *SessionContext {
	if s.Data.Flash == nil {
		s.Data.Flash = &flash.Flash{}
	}
	s.Data.Flash.Success = append(s.Data.Flash.Success, m)
	return s
}

func (s *SessionContext) FailFlash(m string) *SessionContext {
	if s.Data.Flash == nil {
		s.Data.Flash = &flash.Flash{}
	}
	s.Data.Flash.Error = append(s.Data.Flash.Error, m)
	return s
}

func (s *SessionContext) VerifyCaptchaSolution(r *http.Request) bool {
	r.ParseForm()
	digits := r.Form.Get("_captcha")
	digits = strings.TrimSpace(digits)
	if digits == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(digits), []byte(s.Data.Captcha)) == 1
}

type sessionContextKey struct{}

func Load(r *http.Request) (*SessionContext, *http.Request) {
	return Get(r.Context()).Load(r)
}

func (s Session) Load(r *http.Request) (*SessionContext, *http.Request) {
	if c, ok := r.Context().Value(sessionContextKey{}).(*SessionContext); ok {
		return c, r
	}
	rs := &SessionContext{s: s}
	r = r.WithContext(context.WithValue(r.Context(), sessionContextKey{}, rs))
	cookie, err := r.Cookie(string(s))
	if err != nil {
		return rs, r
	}

	if g := s.Get(r.Context(), cookie.Value); g != nil {
		err := json.Unmarshal(g, &rs.Data)
		if err != nil {
			log.Get().Err(err).Msg("failed to decode session value")
		}
	}
	return rs, r
}

func (s *Session) Get(ctx context.Context, value string) []byte {
	if value == "" {
		return nil
	}
	base, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil
	}
	r, err := age.Decrypt(bytes.NewReader(base), config.GetAgeSecret(ctx))
	if err != nil {
		// No need to log here. Bad Sessions will simply be ignored
		return nil
	}
	b, err := io.ReadAll(r)
	if err != nil {
		panic("failed to read decrypted age data " + err.Error())
	}
	return b
}

func (s *SessionContext) Save(ctx context.Context, w http.ResponseWriter) {
	b, _ := json.Marshal(s.Data)
	value := s.s.Create(ctx, b)
	cookie := &http.Cookie{
		Path:    "/",
		Name:    string(s.s),
		Value:   value,
		Expires: core.Now(ctx).Add(time.Duration(MaxAge) * time.Second),
	}
	http.SetCookie(w, cookie)
}

func (s *Session) Create(ctx context.Context, b []byte) string {
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, config.GetAgeSecret(ctx).Recipient())
	if err != nil {
		panic("failed to encrypt session " + err.Error())
	}
	_, err = w.Write(b)
	if err != nil {
		panic("failed to encrypt session " + err.Error())
	}
	err = w.Close()
	if err != nil {
		panic("failed to encrypt session " + err.Error())
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func SaveCaptcha(w http.ResponseWriter, r *http.Request) *http.Request {
	session, r := Load(r)
	solution := captcha.RandomDigits(captcha.DefaultLen)
	img := captcha.NewImage("", solution, captcha.StdWidth, captcha.StdHeight)
	var b bytes.Buffer
	img.WriteTo(&b)
	data, err := dataurl.Encode(b.Bytes(), dataurl.WithBase64Encoding(true), dataurl.WithMediaType("image/png"))
	if err != nil {
		log.Get().Err(err).Msg("failed to encode captcha image")
		return r
	}
	session.Data.Captcha = formatCaptchaSolution(solution)
	session.Save(r.Context(), w)
	return r.WithContext(templates.SetCaptcha(r.Context(),
		template.HTMLAttr(fmt.Sprintf("src=%q", string(data))),
	))
}

func SaveCsrf(w http.ResponseWriter, r *http.Request) *http.Request {
	session, r := Load(r)
	token := make([]byte, 32)
	rand.Read(token)
	csrf := base64.StdEncoding.EncodeToString(token)
	session.Data.Csrf = csrf
	session.Save(r.Context(), w)
	return r.WithContext(templates.SetCSRF(r.Context(), template.HTML(csrf)))
}

func IsValidCSRF(r *http.Request) bool {
	r.ParseForm()
	session, _ := Load(r)
	value := r.Form.Get("_csrf")
	return subtle.ConstantTimeCompare([]byte(value), []byte(session.Data.Csrf)) == 1
}

func formatCaptchaSolution(sol []byte) string {
	var s strings.Builder
	s.Grow(len(sol))
	for _, b := range sol {
		s.WriteString(strconv.FormatInt(int64(b), 10))
	}
	return s.String()
}
