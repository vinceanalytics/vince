package secrets

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"

	"filippo.io/age"
)

func APIKey() []byte {
	k := make([]byte, 64)
	rand.Read(k)
	return k
}

func AGE() []byte {
	a, err := age.GenerateX25519Identity()
	if err != nil {
		panic("failed to generate age key pair " + err.Error())
	}
	return []byte(a.String())
}

// ED25519 returns pem encoded base64 string of ed25519 key pair.
func ED25519() []byte {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
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
	return pem.EncodeToMemory(privBlock)
}
