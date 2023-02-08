package vince

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
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Session struct {
	name     string
	hashKey  [32]byte
	blockKey [16]byte
	block    cipher.Block
	macPool  *sync.Pool
	maxAge   int
}

func NewSession(name string) *Session {
	var s Session
	s.maxAge = 86400
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

func (s *Session) decode(value []byte) []byte {
	decoded := make([]byte, base64.URLEncoding.DecodedLen(len(value)))
	b, err := base64.URLEncoding.Decode(decoded, value)
	if err != nil {
		xlg.Err(err).Msg("failed to decode cookie value")
		return nil
	}
	return decoded[:b]
}

type SessionContext struct {
	Data map[string]any
	s    *Session
}

func (s *SessionContext) VerifyCaptchaSolution(digits string) bool {
	digits = strings.TrimSpace(digits)
	if digits == "" {
		return false
	}
	if x, ok := s.Data[captchaKey]; ok {
		b := x.(string)
		return digits == b
	}
	return false
}

type sessionContextKey struct{}

func (s *Session) Load(r *http.Request) (*SessionContext, *http.Request) {
	if c, ok := r.Context().Value(sessionContextKey{}).(*SessionContext); ok {
		return c, r
	}
	rs := &SessionContext{
		Data: map[string]any{},
		s:    s,
	}
	r = r.WithContext(context.WithValue(r.Context(), sessionContextKey{}, rs))
	cookie, err := r.Cookie(s.name)
	if err != nil {
		return rs, r
	}

	if g := s.Get(cookie.Value); g != nil {
		err := json.Unmarshal(g, &rs.Data)
		if err != nil {
			xlg.Err(err).Msg("failed to decode session value")
		}
	}
	return rs, r
}

func (s *Session) Get(value string) []byte {
	if value == "" {
		return nil
	}
	b := s.decode([]byte(value))
	if b == nil {
		return nil
	}
	parts := bytes.SplitN(b, []byte("|"), 3)
	if len(parts) != 3 {
		xlg.Error().Msg("invalid hmac")
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
	if b = s.decode(parts[1]); b == nil {
		return nil
	}
	return s.decrypt(b)
}

func (s *SessionContext) Save(w http.ResponseWriter) string {
	b, _ := json.Marshal(s.Data)
	value := s.s.Create(b)
	cookie := &http.Cookie{
		Name:    s.s.name,
		Value:   value,
		Expires: time.Now().Add(time.Duration(s.s.maxAge) * time.Second),
	}
	http.SetCookie(w, cookie)
	return value
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
