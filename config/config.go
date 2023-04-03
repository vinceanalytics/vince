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
		&cli.StringFlag{
			Name:    "listen-address",
			Usage:   "bind the server to this port",
			Value:   ":8080",
			EnvVars: []string{"VINCE_LISTEN_ADDRESS"},
		},
		&cli.StringFlag{
			Name:    "data",
			Usage:   "path to data directory",
			Value:   ".vince",
			EnvVars: []string{"VINCE_DATA"},
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
			EnvVars: []string{"VINCE_URL"},
		},
		&cli.StringFlag{
			Name:    "secret-key-base",
			Usage:   "secret with size 64 bytes",
			Value:   defaultSecret(),
			EnvVars: []string{"VINCE_SECRET_KEY_BASE"},
		},
		&cli.StringFlag{
			Name:    "cookie-store-secret",
			Usage:   "48 bytes base64 encoded cookie encryption key",
			Value:   defaultSecret(),
			EnvVars: []string{"VINCE_COOKIE_STORE_SECRET"},
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
			Name:    "log",
			Usage:   "level of logging",
			Value:   "info",
			EnvVars: []string{"VINCE_LOG_LEVEL"},
		},
		&cli.StringFlag{
			Name:    "backup-dir",
			Usage:   "directory where backups are stored",
			Value:   "info",
			EnvVars: []string{"VINCE_BACKUP_DIR"},
		},
		&cli.IntFlag{
			Name:    "site-limit",
			Usage:   "maximum number os sites per user",
			Value:   50,
			EnvVars: []string{"VINCE_SITE_LIMIT"},
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
		&cli.DurationFlag{
			Name:    "cache-refresh",
			Usage:   "window for refreshing sites cache",
			Value:   15 * time.Minute,
			EnvVars: []string{"VINCE_SITE_CACHE_REFRESH_INTERVAL"},
		},
		&cli.DurationFlag{
			Name:    "rotation-check",
			Usage:   "window for checking log rotation",
			Value:   time.Hour,
			EnvVars: []string{"VINCE_LOG_ROTATION_CHECK_INTERVAL"},
		},
		&cli.DurationFlag{
			Name:  "ts-buffer",
			Usage: "window for buffering timeseries in memory",
			// This seems reasonable to avoid users to wait for a long time between
			// creating the site and seeing something on the dashboard. A bigger
			// duration is better though, to reduce pressure on our kv store
			Value:   15 * time.Minute,
			EnvVars: []string{"VINCE_TS_BUFFER_INTERVAL"},
		},
		&cli.DurationFlag{
			Name:  "ts-merge",
			Usage: "window for compacting daily timeseries data",
			// We pick this value to balance the number of items processed per
			// iteration. This will ensure we have al most 3 files per user.
			//
			// Merge operation is slow, and gets even slower with the number of active
			// users in the system growing. This is by design to conserve amount of memory used
			// during the operation.
			Value:   30 * time.Minute,
			EnvVars: []string{"VINCE_TS_MERGE_INTERVAL"},
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
	return base, setupKey(base)
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
		ListenAddress:           ctx.String("listen-address"),
		Url:                     ctx.String("url"),
		DataPath:                ctx.String("data"),
		EnableEmailVerification: ctx.Bool("enable-email-verification"),
		IsSelfHost:              ctx.Bool("self-host"),
		SecretKeyBase:           ctx.String("secret-key-base"),
		CookieStoreSecret:       ctx.String("cookie-store-secret"),
		BackupDir:               ctx.String("backup-dir"),
		SiteLimit:               uint32(ctx.Int("site-limit")),
		Intervals: &Intervals{
			SitesByDomainCacheRefreshInterval: durationpb.New(ctx.Duration("cache-refresh")),
			LogRotationCheckInterval:          durationpb.New(ctx.Duration("rotation-check")),
			SaveTimeseriesBufferInterval:      durationpb.New(ctx.Duration("ts-buffer")),
			MergeTimeseriesInterval:           durationpb.New(ctx.Duration("ts-merge")),
		},
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
		Cors: &Cors{
			Origin:      "*",
			Credentials: true,
			MaxAge:      1_728_000,
			Headers: []string{
				"Authorization",
				"Content-Type",
				"Accept",
				"Origin",
				"User-Agent",
				"DNT",
				"Cache-Control",
				"X-Mx-ReqToken",
				"Keep-Alive",
				"X-Requested-With",
				"If-Modified-Since",
				"X-CSRF-Token",
			},
			Methods:               []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			SendPreflightResponse: true,
		},
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
	n.SecretKeyBase = n.SecretKeyBase[0:6] + "***"
	n.SuperUserId = nil
	return n
}

func (c *Config) IsSuperUser(id uint64) bool {
	for _, u := range c.SuperUserId {
		if u == id {
			return true
		}
	}
	return false
}

func (c *Config) IsExempt(email string) bool {
	for _, s := range c.SiteLimitExempt {
		if s == email {
			return true
		}
	}
	return false
}
