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
			for _, f := range Flags() {
				switch e := f.(type) {
				case *cli.PathFlag:
					fmt.Fprintf(&o, "# %s\n", e.Usage)
					switch e.EnvVars[0] {
					case "VINCE_SECRET_BASE":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], basePath)
					case "VINCE_SECRET_ED25519_PRIVATE":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], priv)
					case "VINCE_SECRET_ED25519_PUBLIC":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], pub)
					case "VINCE_SECRET_SESSION":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], cookiePath)
					default:
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], e.Value)
					}
				case *cli.StringFlag:
					fmt.Fprintf(&o, "# %s\n", e.Usage)
					fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], e.Value)
				case *cli.IntFlag:
					fmt.Fprintf(&o, "# %s\n", e.Usage)
					fmt.Fprintf(&o, "export  %s=%d\n", e.EnvVars[0], e.Value)
				case *cli.DurationFlag:
					fmt.Fprintf(&o, "# %s\n", e.Usage)
					fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], e.Value)
				case *cli.BoolFlag:
					fmt.Fprintf(&o, "# %s\n", e.Usage)
					fmt.Fprintf(&o, "export  %s=%v\n", e.EnvVars[0], e.Value)
				}
			}
			return os.WriteFile(filepath.Join(path, "secrets"), o.Bytes(), 0600)
		},
	}
}
