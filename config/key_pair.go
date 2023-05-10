package config

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"filippo.io/age"
)

type KeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

func setupSecrets(c *Config) (*KeyPair, *age.X25519Identity, error) {
	s := c.Secrets
	b, err := readSecret(s.Secret)
	if err != nil {
		return nil, nil, fmt.Errorf("reading private key  %v", err)
	}
	data, _ := pem.Decode(b)
	priv, err := x509.ParsePKCS8PrivateKey(data.Bytes)
	if err != nil {
		return nil, nil, err
	}
	privateKey := priv.(ed25519.PrivateKey)
	sec := &KeyPair{
		Public:  privateKey.Public().(ed25519.PublicKey),
		Private: privateKey,
	}
	ageFile, err := readSecret(s.Age)
	if err != nil {
		return nil, nil, fmt.Errorf("reading age secret  %v", err)
	}
	a, err := age.ParseX25519Identity(string(ageFile))
	if err != nil {
		return nil, nil, err
	}
	return sec, a, nil
}

type securityKey struct{}

type ageKey struct{}

func GetSecuritySecret(ctx context.Context) *KeyPair {
	return ctx.Value(securityKey{}).(*KeyPair)
}

func GetAgeSecret(ctx context.Context) *age.X25519Identity {
	return ctx.Value(ageKey{}).(*age.X25519Identity)
}

// Handles base64 encoded files or strings(coming from env vars)
func readSecret(path string) ([]byte, error) {
	// First we check if it is a file.
	f, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && path != "" {
			// It is not a file. We Try to decode it as base64
			b, e := base64.StdEncoding.DecodeString(path)
			if e != nil {
				return nil, err
			}
			return b, nil
		}
		return nil, err
	}
	if len(f) > 0 {
		// If we can decode as bas64 we do so else we return raw bytes.
		b, err := base64.StdEncoding.DecodeString(string(f))
		if err != nil {
			return f, nil
		}
		return b, nil
	}
	return f, nil
}
