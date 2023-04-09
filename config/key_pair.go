package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
)

type KeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

var SecurityKey *KeyPair

var SessionKey []byte

var BaseKey []byte

func readSecret(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(string(b))
}

func setupSecrets(c *Config) (err error) {
	s := c.Secrets
	BaseKey, err = readSecret(s.SecretKeyBase)
	if err != nil {
		return err
	}
	SessionKey, err = readSecret(s.Session)
	if err != nil {
		return err
	}

	b, err := os.ReadFile(s.Ed25519KeyPair.PublicKey)
	if err != nil {
		return err
	}
	data, _ := pem.Decode(b)
	pub, err := x509.ParsePKIXPublicKey(data.Bytes)
	if err != nil {
		return err
	}
	b, err = os.ReadFile(s.Ed25519KeyPair.PrivateKey)
	if err != nil {
		return err
	}
	data, _ = pem.Decode(b)
	priv, err := x509.ParsePKCS8PrivateKey(data.Bytes)
	if err != nil {
		return err
	}
	SecurityKey = &KeyPair{
		Public:  pub.(ed25519.PublicKey),
		Private: priv.(ed25519.PrivateKey),
	}
	return nil
}

func generateAndSaveEd25519(dir string) (privPath, pubPath string, err error) {
	var (
		b     []byte
		block *pem.Block
		pub   ed25519.PublicKey
		priv  ed25519.PrivateKey
	)

	pub, priv, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	b, err = x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return
	}

	block = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}

	name := filepath.Join(dir, "vince_key")

	err = os.WriteFile(name, pem.EncodeToMemory(block), 0600)
	if err != nil {
		return
	}

	privPath = name

	// public key
	b, err = x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return
	}

	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	name += ".pub"
	err = os.WriteFile(name, pem.EncodeToMemory(block), 0644)
	pubPath = name
	return
}
