package sessions

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dchest/captcha"
	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/flash"
	"github.com/gernest/vince/log"
	"github.com/lestrrat-go/dataurl"
)

type Session struct {
	name     string
	hashKey  [32]byte
	blockKey [16]byte
	block    cipher.Block
	macPool  *sync.Pool
	maxAge   int
}

type sessionsKey struct{}

func Set(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionsKey{}, s)
}

func Get(ctx context.Context) *Session {
	return ctx.Value(sessionsKey{}).(*Session)
}

func NewSession(name string) *Session {
	var s Session
	s.maxAge = 60 * 60 * 24 * 365 * 5
	s.name = name
	rand.Read(s.hashKey[:])
	rand.Read(s.blockKey[:])
	s.block, _ = aes.NewCipher(s.blockKey[:])
	s.macPool = &sync.Pool{
		New: func() any {
			return hmac.New(sha256.New, s.hashKey[:])
		},
	}
	return &s
}

func (s *Session) encrypt(value []byte) []byte {
	iv := make([]byte, s.block.BlockSize())
	rand.Read(iv)
	stream := cipher.NewCTR(s.block, iv)
	stream.XORKeyStream(value, value)
	return append(iv, value...)
}

func (s *Session) decrypt(value []byte) []byte {
	size := s.block.BlockSize()
	if len(value) > size {
		iv := value[:size]
		value = value[size:]
		stream := cipher.NewCTR(s.block, iv)
		stream.XORKeyStream(value, value)
		return value
	}
	return nil
}

func (s *Session) encode(value []byte) []byte {
	encoded := make([]byte, base64.URLEncoding.EncodedLen(len(value)))
	base64.URLEncoding.Encode(encoded, value)
	return encoded
}

func (s *Session) decode(ctx context.Context, value []byte) []byte {
	decoded := make([]byte, base64.URLEncoding.DecodedLen(len(value)))
	b, err := base64.URLEncoding.Decode(decoded, value)
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed to decode cookie value")
		return nil
	}
	return decoded[:b]
}

type SessionContext struct {
	Data Data
	s    *Session
}

type Data struct {
	TimeoutAt     time.Time    `json:",omitempty"`
	CurrentUserID uint64       `json:",omitempty"`
	LastSeen      time.Time    `json:",omitempty"`
	LoggedIn      bool         `json:",omitempty"`
	Captcha       string       `json:",omitempty"`
	Csrf          string       `json:",omitempty"`
	LoginDest     string       `json:",omitempty"`
	Flash         *flash.Flash `json:",omitempty"`
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

func (s *Session) Load(r *http.Request) (*SessionContext, *http.Request) {
	if c, ok := r.Context().Value(sessionContextKey{}).(*SessionContext); ok {
		return c, r
	}
	rs := &SessionContext{s: s}
	r = r.WithContext(context.WithValue(r.Context(), sessionContextKey{}, rs))
	cookie, err := r.Cookie(s.name)
	if err != nil {
		return rs, r
	}

	if g := s.Get(r.Context(), cookie.Value); g != nil {
		err := json.Unmarshal(g, &rs.Data)
		if err != nil {
			log.Get(r.Context()).Err(err).Msg("failed to decode session value")
		}
	}
	return rs, r
}

func (s *Session) Get(ctx context.Context, value string) []byte {
	if value == "" {
		return nil
	}
	b := s.decode(ctx, []byte(value))
	if b == nil {
		return nil
	}
	parts := bytes.SplitN(b, []byte("|"), 3)
	if len(parts) != 3 {
		log.Get(ctx).Error().Msg("invalid hmac")
		return nil
	}
	b = append([]byte(s.name+"|"), b[:len(b)-len(parts[2])-1]...)
	mac := s.macPool.Get().(hash.Hash)
	defer s.macPool.Put(mac)
	mac.Reset()
	mac.Write(b)
	sum := mac.Sum(nil)
	if subtle.ConstantTimeCompare(parts[2], sum) == 0 {
		return nil
	}
	var t1 int64
	var err error
	if t1, err = strconv.ParseInt(string(parts[0]), 10, 64); err != nil {
		return nil
	}
	t2 := time.Now().Unix()
	if t1 < t2-int64(s.maxAge) {
		return nil
	}
	if b = s.decode(ctx, parts[1]); b == nil {
		return nil
	}
	return s.decrypt(b)
}

func (s *SessionContext) Save(w http.ResponseWriter) {
	b, _ := json.Marshal(s.Data)
	value := s.s.Create(b)
	cookie := &http.Cookie{
		Name:    s.s.name,
		Value:   value,
		Expires: time.Now().Add(time.Duration(s.s.maxAge) * time.Second),
	}
	http.SetCookie(w, cookie)
}

func (s *Session) Create(b []byte) string {
	b = s.encrypt(b)
	b = s.encode(b)
	b = []byte(fmt.Sprintf("%s|%d|%s|", s.name, time.Now().Unix(), b))
	mac := s.macPool.Get().(hash.Hash)
	defer s.macPool.Put(mac)
	mac.Reset()
	mac.Write(b[:len(b)-1])
	b = append(b, mac.Sum(nil)...)[len(s.name)+1:]
	b = s.encode(b)
	return string(b)
}

func SaveCaptcha(w http.ResponseWriter, r *http.Request) *http.Request {
	session, r := Load(r)
	solution := captcha.RandomDigits(captcha.DefaultLen)
	img := captcha.NewImage("", solution, captcha.StdWidth, captcha.StdHeight)
	var b bytes.Buffer
	img.WriteTo(&b)
	data, err := dataurl.Encode(b.Bytes(), dataurl.WithBase64Encoding(true), dataurl.WithMediaType("image/png"))
	if err != nil {
		log.Get(r.Context()).Err(err).Msg("failed to encode captcha image")
		return r
	}
	session.Data.Captcha = formatCaptchaSolution(solution)
	session.Save(w)
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
	session.Save(w)
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
