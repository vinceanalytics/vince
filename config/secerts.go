package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gernest/vince/pkg/secrets"
	"github.com/manifoldco/promptui"
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
			interactive := ctx.Args().First() == "interactive"
			priv := base64.StdEncoding.EncodeToString(secrets.ED25519())
			age := base64.StdEncoding.EncodeToString(secrets.AGE())
			var o bytes.Buffer
			for _, f := range Flags() {
				switch e := f.(type) {
				case *cli.StringFlag:
					fmt.Fprintf(&o, "# %s\n", e.Usage)
					switch e.EnvVars[0] {
					case "VINCE_SECRET":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], priv)
					case "VINCE_SECRET_AGE":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], age)
					case "VINCE_BOOTSTRAP_KEY":
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], base64.StdEncoding.EncodeToString(secrets.APIKey()))
					default:
						value := e.Value
						if interactive {
							prompt := promptui.Prompt{
								Label:   e.Name,
								Default: e.Value,
							}
							var err error
							value, err = prompt.Run()
							if err != nil {
								return fmt.Errorf("failed to read prompt %v", err)
							}
						}
						fmt.Fprintf(&o, "export  %s=%q\n", e.EnvVars[0], value)
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
			path := ctx.String("path")
			_, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(path, 0755)
					if err != nil {
						return fmt.Errorf("failed creating data path:%v", err)
					}
				} else {
					return err
				}
			}
			path, err = filepath.Abs(path)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(path, "secrets"), o.Bytes(), 0600)
		},
	}
}
