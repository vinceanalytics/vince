package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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
			interactive := ctx.Args().First() == "i"
			var o bytes.Buffer
			for _, f := range Flags() {
				var usage, name, value, env string
				switch e := f.(type) {
				case *cli.StringFlag:
					usage = e.Usage
					name = e.Name
					value = e.Value
					env = e.EnvVars[0]
				case *cli.IntFlag:
					usage = e.Usage
					name = e.Name
					value = strconv.FormatInt(int64(e.Value), 10)
					env = e.EnvVars[0]
				case *cli.DurationFlag:
					usage = e.Usage
					name = e.Name
					value = e.Value.String()
					env = e.EnvVars[0]
				case *cli.BoolFlag:
					usage = e.Usage
					name = e.Name
					value = strconv.FormatBool(e.Value)
					env = e.EnvVars[0]
				default:
					continue
				}
				if interactive {
					prompt := promptui.Prompt{
						Label:   name,
						Default: value,
					}
					var err error
					value, err = prompt.Run()
					if err != nil {
						return fmt.Errorf("failed to read prompt %v", err)
					}
				}
				fmt.Fprintf(&o, "# %s\n", usage)
				fmt.Fprintf(&o, "export  %s=%q\n", env, value)
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
