package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"

	"filippo.io/age"
)

type KeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

var SECURITY *KeyPair

var AGE *age.X25519Identity

func setupSecrets(c *Config) (err error) {
	s := c.Secrets

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
	SECURITY = &KeyPair{
		Public:  pub.(ed25519.PublicKey),
		Private: priv.(ed25519.PrivateKey),
	}
	ageFile, err := os.ReadFile(s.Age.PrivateKey)
	if err != nil {
		return err
	}
	AGE, err = age.ParseX25519Identity(string(ageFile))
	return
}

func generateAndSaveEd25519(dir string) (privPath, pubPath string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}

	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", err
	}

	privBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}

	// public key
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	name := "vince_ed25519_key"
	privPath = filepath.Join(dir, name)
	pubPath = filepath.Join(dir, name+".pub")

	err = errors.Join(
		os.WriteFile(privPath, pem.EncodeToMemory(privBlock), 0600),
		os.WriteFile(pubPath, pem.EncodeToMemory(pubBlock), 0644),
	)
	return
}

func generateAndSaveAge(dir string) (private, public string, err error) {
	a, err := age.GenerateX25519Identity()
	if err != nil {
		return "", "", err
	}
	name := "vince_age_key"
	private = filepath.Join(dir, name)
	public = filepath.Join(dir, name+".pub")
	err = errors.Join(
		os.WriteFile(private, []byte(a.String()), 0600),
		os.WriteFile(public, []byte(a.Recipient().String()), 0600),
	)
	return
}
