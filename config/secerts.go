package config

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func SecretsCMD() *cli.Command {
	return &cli.Command{
		Name:  "secrets",
		Usage: "generates all secrets used by vince",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path,p",
				Usage: "directory to save the secrets",
				Value: ".vince",
			},
		},
		Action: func(ctx *cli.Context) error {
			path, err := filepath.Abs(ctx.String("path"))
			if err != nil {
				return err
			}
			// secret base
			b := make([]byte, 64)
			rand.Read(b)
			basePath := filepath.Join(path, "base")
			err = os.WriteFile(basePath, []byte(base64.StdEncoding.EncodeToString(b)), 0600)
			if err != nil {
				return err
			}
			priv, pub, err := generateAndSaveEd25519(path)
			if err != nil {
				return err
			}
			b = make([]byte, 48)
			rand.Read(b)
			cookiePath := filepath.Join(path, "vince_session")
			err = os.WriteFile(cookiePath, []byte(base64.StdEncoding.EncodeToString(b)), 0600)
			if err != nil {
				return err
			}
			var o bytes.Buffer
			fmt.Fprintf(&o, "export  VINCE_SECRET_BASE=%q\n", basePath)
			fmt.Fprintf(&o, "export  VINCE_SECRET_ED25519_private=%q\n", priv)
			fmt.Fprintf(&o, "export  VINCE_SECRET_ED25519_public=%q\n", pub)
			fmt.Fprintf(&o, "export  VINCE_SECRET_SESSION=%q\n", cookiePath)
			return os.WriteFile(filepath.Join(path, "secrets"), o.Bytes(), 0600)
		},
	}
}
