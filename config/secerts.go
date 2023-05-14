package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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
			&cli.BoolFlag{
				Name:    "interactive",
				Usage:   "Interactive configuration",
				Aliases: []string{"i"},
			},
		},
		Action: func(ctx *cli.Context) error {
			var o bytes.Buffer
			for _, p := range build(ctx.App) {
				a, err := p()
				if err != nil {
					return err
				}
				for _, v := range a {
					fmt.Fprintf(&o, "# %s\n", v.usage)
					fmt.Fprintf(&o, "export  %s=%q\n", v.env, v.value)
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

type Prompt = struct {
	usage, name, value, env string
}

type promptCall func() ([]*Prompt, error)

func build(a *cli.App) []promptCall {
	stringFlags := make(map[string]*cli.StringFlag)
	boolFlags := make(map[string]*cli.BoolFlag)
	durationFlags := make(map[string]*cli.DurationFlag)
	intFlags := make(map[string]*cli.IntFlag)
	for _, f := range a.Flags {
		switch e := f.(type) {
		case *cli.StringFlag:
			stringFlags[e.Name] = e
			switch e.Name {
			case "secret":
				e.Value = base64.StdEncoding.EncodeToString(secrets.ED25519())
			case "secret-age":
				e.Value = base64.StdEncoding.EncodeToString(secrets.AGE())
			case "bootstrap-key":
				e.Value = base64.StdEncoding.EncodeToString(secrets.APIKey())
			}
		case *cli.BoolFlag:
			boolFlags[e.Name] = e
		case *cli.IntFlag:
			intFlags[e.Name] = e
		case *cli.DurationFlag:
			durationFlags[e.Name] = e
		}
	}
	return []promptCall{
		buildBool(boolFlags["enable-bootstrap"],
			buildString(stringFlags["bootstrap-name"]),
			buildString(stringFlags["bootstrap-email"]),
			buildString(stringFlags["bootstrap-password"], true),
			buildStringNoPrompt(stringFlags["bootstrap-key"]),
		),
		buildString(stringFlags["config"]),
		buildString(stringFlags["data"]),
		buildString(stringFlags["listen"]),
		buildString(stringFlags["env"]),
		buildDuration(durationFlags["ts-buffer"]),
		buildString(stringFlags["url"]),
		buildBool(boolFlags["enable-profile"]),
		buildBool(boolFlags["enable-alerts"]),
		buildStringNoPrompt(stringFlags["secret"]),
		buildStringNoPrompt(stringFlags["secret-age"]),
		buildBool(boolFlags["enable-email"],
			buildString(stringFlags["mailer-address"]),
			buildString(stringFlags["mailer-address-name"]),
			buildString(stringFlags["mailer-smtp-anonymous"]),
			buildString(stringFlags["mailer-smtp-host"]),
			buildInt(intFlags["mailer-smtp-port"]),
			buildString(stringFlags["mailer-smtp-oauth-host"]),
			buildInt(intFlags["mailer-smtp-oauth-port"]),
			buildString(stringFlags["mailer-smtp-oauth-token"], true),
			buildString(stringFlags["mailer-smtp-oauth-username"]),
			buildString(stringFlags["mailer-smtp-plain-identity"]),
			buildString(stringFlags["mailer-smtp-plain-username"]),
			buildString(stringFlags["mailer-smtp-plain-password"], true),
		),
		buildBool(boolFlags["enable-tls"],
			buildString(stringFlags["tls-address"]),
			buildString(stringFlags["tls-cert"]),
			buildString(stringFlags["tls-key"]),
			buildBool(boolFlags["enable-auto-tls"],
				buildString(stringFlags["acme-domain"]),
				buildString(stringFlags["acme-email"]),
			),
		),
	}

}
func buildString(f *cli.StringFlag, mask ...bool) promptCall {
	return func() ([]*Prompt, error) {
		prompt := promptui.Prompt{
			Label:   f.Name,
			Default: f.Value,
		}
		if len(mask) > 0 {
			prompt.Mask = 'x'
		}
		value, err := prompt.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to read prompt %v", err)
		}
		return []*Prompt{
			{
				usage: f.Usage,
				name:  f.Name,
				value: value,
				env:   f.EnvVars[0],
			},
		}, nil
	}
}

func buildStringNoPrompt(f *cli.StringFlag) promptCall {
	return func() ([]*Prompt, error) {
		return []*Prompt{
			{
				usage: f.Usage,
				name:  f.Name,
				value: f.Value,
				env:   f.EnvVars[0],
			},
		}, nil
	}
}

func buildDuration(f *cli.DurationFlag) promptCall {
	return func() ([]*Prompt, error) {
		prompt := promptui.Prompt{
			Label:   f.Name,
			Default: f.Value.String(),
		}
		value, err := prompt.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to read prompt %v", err)
		}
		return []*Prompt{
			{
				usage: f.Usage,
				name:  f.Name,
				value: value,
				env:   f.EnvVars[0],
			},
		}, nil
	}
}
func buildInt(f *cli.IntFlag) promptCall {
	return func() ([]*Prompt, error) {
		prompt := promptui.Prompt{
			Label:   f.Name,
			Default: strconv.Itoa(f.Value),
		}
		value, err := prompt.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to read prompt %v", err)
		}
		return []*Prompt{
			{
				usage: f.Usage,
				name:  f.Name,
				value: value,
				env:   f.EnvVars[0],
			},
		}, nil
	}
}

func buildBool(f *cli.BoolFlag, next ...promptCall) promptCall {
	return func() ([]*Prompt, error) {
		prompt := promptui.Prompt{
			Label:   f.Name,
			Default: strconv.FormatBool(f.Value),
		}
		value, err := prompt.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to read prompt %v", err)
		}
		o := []*Prompt{
			{
				usage: f.Usage,
				name:  f.Name,
				value: value,
				env:   f.EnvVars[0],
			},
		}
		if value == "true" {
			for _, n := range next {
				x, err := n()
				if err != nil {
					return nil, err
				}
				o = append(o, x...)
			}
		}
		return o, nil
	}
}
