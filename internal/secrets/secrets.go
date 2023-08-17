package secrets

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"filippo.io/age"
	"github.com/vinceanalytics/vince/internal/must"
)

const (
	API_KEY        = "vince_api"
	API_KEY_ENV    = "VINCE_BOOTSTRAP_KEY"
	AGE_KEY        = "vince_age"
	AGE_KEY_ENV    = "VINCE_SECRET_AGE"
	SECRET_KEY     = "vince_secret"
	SECRET_KEY_ENV = "VINCE_SECRET"
)

func APIKey() string {
	k := make([]byte, 32)
	rand.Read(k)
	return "vp_" + base64.StdEncoding.EncodeToString(k)
}

func AGE() string {
	a, err := age.GenerateX25519Identity()
	if err != nil {
		panic("failed to generate age key pair " + err.Error())
	}
	return a.String()
}

// ED25519 returns pem encoded base64 string of ed25519 key pair.
func ED25519() string {
	priv := ED25519Raw()
	b := must.Must(x509.MarshalPKCS8PrivateKey(priv))(
		"failed to  x509.MarshalPKCS8PrivateKey",
	)
	privBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}
	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(privBlock))
}

func ED25519Raw() ed25519.PrivateKey {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	must.One(err)("failed to generate ed25519 key pair")
	return priv
}
