package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gernest/vince/pkg/secrets"
	"github.com/urfave/cli/v3"
)

func ConfigCMD() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "generates configurations for vince",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path,p",
				Usage: "directory to save configurations (including secrets)",
				Value: ".vince",
			},
		},
		Action: func(ctx *cli.Context) error {
			path, err := filepath.Abs(ctx.String("path"))
			if err != nil {
				return err
			}
			priv, pub := secrets.ED25519()
			if err != nil {
				return err
			}
			age := secrets.AGE()
			var o bytes.Buffer
			for _, f := range Flags() {
				switch e := f.(type) {
				case *cli.StringFlag:
					fmt.Fprintf(&o, "# %s\n", e.Usage)
					switch e.EnvVars[0] {
					case "VINCE_SECRET_ED25519_PRIVATE":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], priv)
					case "VINCE_SECRET_ED25519_PUBLIC":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], pub)
					case "VINCE_SECRET_AGE":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], age)
					default:
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], e.Value)
					}
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
