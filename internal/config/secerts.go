package config

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/secrets"
)

func ConfigCMD() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "generates configurations for vince",
		Action: func(ctx *cli.Context) error {
			var o bytes.Buffer
			for _, f := range ctx.App.Flags {
				var usage, env, value string
				switch e := f.(type) {
				case *cli.StringFlag:
					switch e.Name {
					case "secret":
						e.Value = secrets.ED25519()
					case "secret-age":
						e.Value = secrets.AGE()
					case "bootstrap-key":
						e.Value = secrets.APIKey()
					}
					usage, env, value = e.GetUsage(), e.GetEnvVars()[0], e.GetValue()
				case *cli.BoolFlag:
					if e.Name == "help" {
						continue
					}
					usage, env, value = e.GetUsage(), e.GetEnvVars()[0], strconv.FormatBool(e.Value)
				case *cli.IntFlag:
					usage, env, value = e.GetUsage(), e.GetEnvVars()[0], e.GetValue()
				case *cli.DurationFlag:
					usage, env, value = e.GetUsage(), e.GetEnvVars()[0], e.GetValue()
				case *cli.StringSliceFlag:
					usage, env, value = e.GetUsage(), e.GetEnvVars()[0], strings.Join(e.Value, ",")
				case *cli.Uint64SliceFlag:
					for k, v := range e.Value {
						if k != 0 {
							value += ","
						}
						value += strconv.FormatUint(v, 10)
					}
					usage, env = e.GetUsage(), e.GetEnvVars()[0]
				}
				fmt.Fprintf(&o, "# %s\n", usage)
				fmt.Fprintf(&o, "export  %s=%q\n", env, value)
			}
			os.Stdout.Write(o.Bytes())
			return nil
		},
	}
}
