package secrets

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"

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

type privateKey struct{}

func Open(ctx context.Context, log *slog.Logger, key string) context.Context {
	block, _ := pem.Decode([]byte(key))
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Error("failed to decode pem block containing private key")
		os.Exit(1)
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Error("failed parsing private key", "err", err)
		os.Exit(1)
	}
	a, ok := priv.(ed25519.PrivateKey)
	if !ok {
		log.Error("only ed25519.PrivateKey is supported ", "key", fmt.Sprintf("%#T", priv))
		os.Exit(1)
	}
	return context.WithValue(ctx, privateKey{}, a)
}

func Get(ctx context.Context) ed25519.PrivateKey {
	return ctx.Value(privateKey{}).(ed25519.PrivateKey)
}
