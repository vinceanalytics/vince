package config

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
)

//go:generate protoc -I=. --go_out=paths=source_relative:. config.proto

type configKey struct{}

func Set(ctx context.Context, conf *Config) context.Context {
	return context.WithValue(ctx, configKey{}, conf)
}

func Get(ctx context.Context) *Config {
	return ctx.Value(configKey{}).(*Config)
}

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Usage:   "configuration file in json format",
			Value:   "vince.json",
			EnvVars: []string{"VINCE_CONFIG"},
		},
		&cli.IntFlag{
			Name:    "port",
			Usage:   "bind the server to this port ",
			Value:   8080,
			EnvVars: []string{"VINCE_LISTEN_PORT"},
		},
		&cli.StringFlag{
			Name:    "data",
			Usage:   "path to data directory",
			Value:   ".vince",
			EnvVars: []string{"VINCE_DATA"},
		},
		&cli.DurationFlag{
			Name:    "flush-interval",
			Value:   30 * time.Minute,
			EnvVars: []string{"VINCE_FLUSH_INTERVAL"},
		},
		&cli.StringFlag{
			Name:    "env",
			Usage:   "environment on which vince is run (dev,staging,production)",
			Value:   "dev",
			EnvVars: []string{"VINCE_ENV"},
		},
		&cli.StringFlag{
			Name:    "url",
			Usage:   "url for the server on which vince is hosted(it shows up on emails)",
			Value:   "dev",
			EnvVars: []string{"VINCE_URL"},
		},
		&cli.StringFlag{
			Name:    "secret-key-base",
			Usage:   "secret with size 64 bytes",
			Value:   defaultSecret(),
			EnvVars: []string{"VINCE_SECRET_KEY_BASE"},
		},
		&cli.BoolFlag{
			Name:    "enable-email-verification",
			Usage:   "send emails for account verification",
			Value:   true,
			EnvVars: []string{"VINCE_ENABLE_EMAIL_VERIFICATION"},
		},
		&cli.BoolFlag{
			Name:    "self-host",
			Usage:   "self hosted version",
			Value:   true,
			EnvVars: []string{"VINCE_SELF_HOST"},
		},
		&cli.StringFlag{
			Name:    "mailer-address",
			Usage:   "email address used for the sender of outgoing emails ",
			EnvVars: []string{"VINCE_MAILER_ADDRESS"},
		},
		&cli.StringFlag{
			Name:    "mailer-address-name",
			Usage:   "email address name  used for the sender of outgoing emails ",
			EnvVars: []string{"VINCE_MAILER_ADDRESS_NAME"},
		},
		&cli.StringFlag{
			Name:    "mailer-smtp-host",
			Usage:   "host address of the smtp server used for outgoing emails",
			EnvVars: []string{"VINCE_MAILER_SMTP_HOST"},
		},
		&cli.IntFlag{
			Name:    "mailer-smtp-port",
			Usage:   "port address of the smtp server used for outgoing emails",
			EnvVars: []string{"VINCE_MAILER_SMTP_PORT"},
		},
		&cli.StringFlag{
			Name:    "mailer-smtp-anonymous",
			Usage:   "trace value for anonymous smtp auth",
			EnvVars: []string{"VINCE_MAILER_SMTP_ANONYMOUS"},
		},
		&cli.StringFlag{
			Name:    "mailer-smtp-plain-identity",
			Usage:   "identity value for plain smtp auth",
			EnvVars: []string{"VINCE_MAILER_SMTP_PLAIN_IDENTITY"},
		},
		&cli.StringFlag{
			Name:    "mailer-smtp-plain-username",
			Usage:   "username value for plain smtp auth",
			EnvVars: []string{"VINCE_MAILER_SMTP_PLAIN_USERNAME"},
		},
		&cli.StringFlag{
			Name:    "mailer-smtp-plain-password",
			Usage:   "password value for plain smtp auth",
			EnvVars: []string{"VINCE_MAILER_SMTP_PLAIN_PASSWORD"},
		},
		&cli.StringFlag{
			Name:    "mailer-smtp-oauth-username",
			Usage:   "username value for oauth bearer smtp auth",
			EnvVars: []string{"VINCE_MAILER_SMTP_OAUTH_USERNAME"},
		},
		&cli.StringFlag{
			Name:    "mailer-smtp-oauth-token",
			Usage:   "token value for oauth bearer smtp auth",
			EnvVars: []string{"VINCE_MAILER_SMTP_OAUTH_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "mailer-smtp-oauth-host",
			Usage:   "host value for oauth bearer smtp auth",
			EnvVars: []string{"VINCE_MAILER_SMTP_OAUTH_HOST"},
		},
		&cli.IntFlag{
			Name:    "mailer-smtp-oauth-port",
			Usage:   "port value for oauth bearer smtp auth",
			EnvVars: []string{"VINCE_MAILER_SMTP_OAUTH_PORT"},
		},
	}
}

func Load(ctx *cli.Context) (*Config, error) {
	base := fromCli(ctx)
	conf, err := fromFile(ctx)
	if err != nil {
		return nil, err
	}
	proto.Merge(base, conf)
	return base, nil
}

func defaultSecret() string {
	b := make([]byte, 64)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func fromFile(ctx *cli.Context) (*Config, error) {
	path := ctx.String("config")
	if path == "" {
		return &Config{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	err = protojson.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func fromCli(ctx *cli.Context) *Config {
	x := &Config{
		Port:                    int32(ctx.Int("port")),
		Url:                     ctx.String("url"),
		DataPath:                ctx.String("data"),
		FlushInterval:           durationpb.New(ctx.Duration("flush-interval")),
		EnableEmailVerification: ctx.Bool("enable-email-verification"),
		IsSelfHost:              ctx.Bool("self-host"),
		SecretKeyBase:           ctx.String("secret-key-base"),
		Mailer: &Config_Mailer{
			Address: &Config_Address{
				Name:  ctx.String("mailer-address-name"),
				Email: ctx.String("mailer-address"),
			},
			Smtp: &Config_Mailer_Smtp{
				Host: ctx.String("mailer-smtp-host"),
				Port: int32(ctx.Int("mailer-smtp-port")),
			},
		},
	}
	switch ctx.String("env") {
	case "dev":
		x.Env = Config_DEVELOPMENT
	case "staging":
		x.Env = Config_STAGING
	case "prod":
		x.Env = Config_PRODUCTION
	}

	var anon *Config_Mailer_Anonymous
	var plain *Config_Mailer_Plain
	var oauth *Config_Mailer_OauthBearer
	if v := ctx.String("mailer-smtp-anonymous"); v != "" {
		anon = &Config_Mailer_Anonymous{Trace: v}
	}

	if v := ctx.String("mailer-smtp-plain-identity"); v != "" {
		if plain == nil {
			plain = &Config_Mailer_Plain{}
		}
		plain.Identity = v
	}
	if v := ctx.String("mailer-smtp-plain-username"); v != "" {
		if plain == nil {
			plain = &Config_Mailer_Plain{}
		}
		plain.Username = v
	}
	if v := ctx.String("mailer-smtp-plain-password"); v != "" {
		if plain == nil {
			plain = &Config_Mailer_Plain{}
		}
		plain.Password = v
	}
	if v := ctx.String("mailer-smtp-oauth-username"); v != "" {
		if oauth == nil {
			oauth = &Config_Mailer_OauthBearer{}
		}
		oauth.Username = v
	}
	if v := ctx.String("mailer-smtp-oauth-token"); v != "" {
		if oauth == nil {
			oauth = &Config_Mailer_OauthBearer{}
		}
		oauth.Token = v
	}
	if v := ctx.String("mailer-smtp-oauth-host"); v != "" {
		if oauth == nil {
			oauth = &Config_Mailer_OauthBearer{}
		}
		oauth.Host = v
	}
	if v := ctx.Int("mailer-smtp-oauth-port"); v != 0 {
		if oauth == nil {
			oauth = &Config_Mailer_OauthBearer{}
		}
		oauth.Port = int32(v)
	}
	switch {
	case anon != nil:
		x.Mailer.Smtp.Auth = &Config_Mailer_Smtp_Anonymous{Anonymous: anon}
	case plain != nil:
		x.Mailer.Smtp.Auth = &Config_Mailer_Smtp_Plain{Plain: plain}
	case oauth != nil:
		x.Mailer.Smtp.Auth = &Config_Mailer_Smtp_OauthBearer{OauthBearer: oauth}
	}
	return x
}

func (c *Config) WriteToFile(name string) error {
	b, err := protojson.Marshal(c.Scrub())
	if err != nil {
		return err
	}
	return os.WriteFile(name, b, 0600)
}

func (c *Config) Scrub() *Config {
	n := proto.Clone(c).(*Config)
	base := make([]byte, len(c.SecretKeyBase))
	for i := range base {
		if i < 6 {
			base[i] = c.SecretKeyBase[i]
		} else {
			base[i] = '*'
		}
	}
	n.SecretKeyBase = string(base)
	return n
}
