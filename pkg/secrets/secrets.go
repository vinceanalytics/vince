package secrets

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"filippo.io/age"
)

func APIKey() string {
	k := make([]byte, 64)
	rand.Read(k)
	return base64.URLEncoding.EncodeToString(k)
}

// AGE returns base64 encoded *age.X25519Identity .
func AGE() string {
	a, err := age.GenerateX25519Identity()
	if err != nil {
		panic("failed to generate age key pair " + err.Error())
	}
	return base64.StdEncoding.EncodeToString([]byte(a.String()))
}

// ED25519 returns pem encoded base64 string of ed25519 key pair.
func ED25519() (privateKey, publicKey string) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic("failed to generate ed25519 key pair " + err.Error())
	}

	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		panic("failed to  x509.MarshalPKCS8PrivateKey " + err.Error())
	}

	privBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}

	// public key
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		panic("failed to  x509.MarshalPKIXPublicKey " + err.Error())
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	privateKey = base64.StdEncoding.EncodeToString(pem.EncodeToMemory(privBlock))
	publicKey = base64.StdEncoding.EncodeToString(pem.EncodeToMemory(pubBlock))
	return
}
