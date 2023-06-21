package config

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"filippo.io/age"
)

func setupSecrets(c *Options) (ed25519.PrivateKey, *age.X25519Identity, error) {
	s := c.Secrets
	o, err := base64.StdEncoding.DecodeString(s.Secret)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid secret value %q", err)
	}
	data, _ := pem.Decode(o)
	if data == nil {
		return nil, nil, fmt.Errorf("invalid secret key. Make sure you provide a PEM encoded  ed25519 private key")
	}
	priv, err := x509.ParsePKCS8PrivateKey(data.Bytes)
	if err != nil {
		return nil, nil, err
	}
	privateKey := priv.(ed25519.PrivateKey)
	a, err := age.ParseX25519Identity(s.Age)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parsing age key %v", err)
	}
	return privateKey, a, nil
}

type securityKey struct{}

type ageKey struct{}

func GetSecuritySecret(ctx context.Context) ed25519.PrivateKey {
	return ctx.Value(securityKey{}).(ed25519.PrivateKey)
}

func GetAgeSecret(ctx context.Context) *age.X25519Identity {
	return ctx.Value(ageKey{}).(*age.X25519Identity)
}
