package vince

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func CSRF() func(w http.ResponseWriter, r *http.Request) bool {
	cookieName := "_csrf"
	cookieMaxAge := 86400
	headerName := "X-CSRF-Token"
	hashKey := make([]byte, 32)
	rand.Read(hashKey)
	blockKey := make([]byte, 16)
	rand.Read(blockKey)
	block, _ := aes.NewCipher(blockKey)

	encrypt := func(value []byte) []byte {
		iv := make([]byte, block.BlockSize())
		rand.Read(iv)
		stream := cipher.NewCTR(block, iv)
		stream.XORKeyStream(value, value)
		return append(iv, value...)
	}

	decrypt := func(value []byte) []byte {
		size := block.BlockSize()
		if len(value) > size {
			// Extract iv.
			iv := value[:size]
			// Extract ciphertext.
			value = value[size:]
			// Decrypt it.
			stream := cipher.NewCTR(block, iv)
			stream.XORKeyStream(value, value)
			return value
		}
		return nil
	}

	encode := func(value []byte) []byte {
		encoded := make([]byte, base64.URLEncoding.EncodedLen(len(value)))
		base64.URLEncoding.Encode(encoded, value)
		return encoded
	}

	decode := func(value []byte) []byte {
		decoded := make([]byte, base64.URLEncoding.DecodedLen(len(value)))
		b, err := base64.URLEncoding.Decode(decoded, value)
		if err != nil {
			return nil
		}
		return decoded[:b]
	}
	macPool := &sync.Pool{
		New: func() any {
			return hmac.New(sha256.New, hashKey)
		},
	}

	newToken := func(b []byte) string {
		b = encrypt(b)
		b = encode(b)
		b = []byte(fmt.Sprintf("%s|%d|%s|", cookieName, time.Now().Unix(), b))
		mac := macPool.Get().(hash.Hash)
		defer macPool.Put(mac)
		mac.Reset()
		mac.Write(b[:len(b)-1])
		b = append(b, mac.Sum(nil)...)
		b = encode(b)
		return string(b)
	}

	oldToken := func(value string) []byte {
		b := decode([]byte(value))
		if b == nil {
			xlg.Error().Str("value", value).Msg("Failed to base64 decode ")
			return nil
		}
		parts := bytes.SplitN(b, []byte("|"), 3)
		if len(parts) != 3 {
			xlg.Error().Str("value", value).Msg("invalid hmac ")
			return nil
		}
		b = append([]byte(cookieName+"|"), b[:len(b)-len(parts[2])-1]...)
		mac := macPool.Get().(hash.Hash)
		defer macPool.Put(mac)
		mac.Reset()
		mac.Write(b)
		if subtle.ConstantTimeCompare(parts[2], mac.Sum(nil)) != 1 {
			return nil
		}
		var t1 int64
		var err error
		if t1, err = strconv.ParseInt(string(parts[0]), 10, 64); err != nil {
			return nil
		}
		t2 := time.Now().Unix()
		if t1 < t2-int64(cookieMaxAge) {
			return nil
		}
		if b = decode(parts[1]); b == nil {
			return nil
		}
		return decrypt(parts[1])
	}

	return func(w http.ResponseWriter, r *http.Request) bool {
		var token string
		var tokenValue []byte
		if c, err := r.Cookie(cookieName); err != nil {
			tokenValue = make([]byte, 32)
			rand.Read(tokenValue)
			token = newToken(tokenValue)
		} else {
			token = c.Value
			tokenValue = oldToken(c.Value)
			if tokenValue == nil {
				tokenValue = make([]byte, 32)
				rand.Read(tokenValue)
				token = newToken(tokenValue)
			}
		}
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			value := r.Header.Get(headerName)
			b := oldToken(value)
			if value == "" || b == nil || subtle.ConstantTimeCompare(tokenValue, b) != 1 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return false
			}
		}
		cookie := &http.Cookie{
			Name:    cookieName,
			Value:   token,
			Expires: time.Now().Add(time.Duration(cookieMaxAge) * time.Second),
		}
		http.SetCookie(w, cookie)
		w.Header().Set("Vary", "Cookie")
		w.Header().Set(headerName, token)
		return true
	}
}
