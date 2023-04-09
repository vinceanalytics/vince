package config

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

type KeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

var SecurityKey *KeyPair

func setupKey(c *Config) error {
	if c.Ed25519KeyPair != nil {
		b, err := os.ReadFile(c.Ed25519KeyPair.PublicKeyPath)
		if err != nil {
			return err
		}
		data, _ := pem.Decode(b)
		pub, err := x509.ParsePKIXPublicKey(data.Bytes)
		if err != nil {
			return err
		}
		b, err = os.ReadFile(c.Ed25519KeyPair.PrivateKeyPath)
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
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	SecurityKey = &KeyPair{
		Public:  pub,
		Private: priv,
	}
	return nil
}

func GenKeyCMD() *cli.Command {
	return &cli.Command{
		Name:  "genkey",
		Usage: "generate Ed25519 and Write pem encoded data into vince_key(for private key) and vince_key.pub(for public key)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "directory to write the files",
			},
		},
		Action: func(ctx *cli.Context) error {
			return GenerateAndSaveEd25519(ctx.String("path"))
		},
	}
}

func GenerateAndSaveEd25519(dir string) error {
	var (
		err   error
		b     []byte
		block *pem.Block
		pub   ed25519.PublicKey
		priv  ed25519.PrivateKey
	)

	pub, priv, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	b, err = x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	block = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}

	name := filepath.Join(dir, "vince_key")

	println("write ", name)
	err = os.WriteFile(name, pem.EncodeToMemory(block), 0600)
	if err != nil {
		return err
	}

	// public key
	b, err = x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return err
	}

	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	name += ".pub"
	println("write ", name)
	return os.WriteFile(name, pem.EncodeToMemory(block), 0644)
}
