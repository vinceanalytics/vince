package config

import (
	"encoding/base64"
	"flag"
	"net"
	"path/filepath"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/mholt/acmez/acme"
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/secrets"
)

const (
	TestName     = "Jane Done"
	TestEmail    = "jane@example.com"
	TestPassword = "1234"
)

type Options struct {
	Listen string
	URL    string
	Mailer struct {
		Enabled       bool
		Name, Address string
		SMTP          struct {
			Address       string
			AuthAnonymous struct {
				Enabled bool
				Trace   string
			}

			AuthOAUTHBearer struct {
				Enabled               bool
				Username, Token, Host string
				Port                  int
			}
			AuthPlain struct {
				Enabled                      bool
				Identity, Username, Password string
			}
			EnableMailHog bool
		}
	}
	DataPath string
	LogLevel string
	Secrets  struct {
		Age, Secret string
	}
	Cors struct {
		Origin                string
		Credentials           bool
		MaxAge                int
		Headers               []string
		Expose                []string
		Methods               []string
		SendPreflightResponse bool
	}
	SuperUserId []uint64
	Firewall    struct {
		Enabled bool
		BlockIP []string
		AllowIP []string
	}
	Intervals struct {
		SiteCache, TSSync time.Duration
	}
	Backup struct {
		Enabled bool
		Dir     string
	}
	Acme struct {
		Enabled         bool
		Domain          string
		CertsPath       string
		Issuer          certmagic.ACMEIssuer
		ExternalAccount acme.EAB
	}
	TLS struct {
		Enabled            bool
		Address, Key, Cert string
	}
	Bootstrap struct {
		Enabled                    bool
		Name, Email, Password, Key string
	}
	Alerts struct {
		Enabled bool
		Source  string
	}
	EnableProfile bool
	// If true don't listen on os.Interrupt signal.
	NoSignal bool
}

func (o *Options) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category:    "core",
			Name:        "listen",
			Usage:       "http address to listen to",
			Value:       ":8080",
			Destination: &o.Listen,
			EnvVars:     []string{"VINCE_LISTEN"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "log-level",
			Usage:       "log level, values are (trace,debug,info,warn,error,fatal,panic)",
			Value:       "debug",
			Destination: &o.LogLevel,
			EnvVars:     []string{"VINCE_LOG_LEVEL"},
		},
		&cli.BoolFlag{
			Category:    "tls",
			Name:        "enable-tls",
			Usage:       "Enables serving https traffic.",
			Destination: &o.TLS.Enabled,
			EnvVars:     []string{"VINCE_ENABLE_TLS"},
		},
		&cli.StringFlag{
			Category:    "tls",
			Name:        "tls-address",
			Usage:       "https address to listen to. You must provide tls-key and tls-cert or configure auto-tls",
			Value:       ":8443",
			Destination: &o.TLS.Address,
			EnvVars:     []string{"VINCE_TLS_LISTEN"},
		},
		&cli.StringFlag{
			Category:    "tls",
			Name:        "tls-key",
			Usage:       "Path to key file used for https",
			Destination: &o.TLS.Key,
			EnvVars:     []string{"VINCE_TLS_KEY"},
		},
		&cli.StringFlag{
			Category:    "tls",
			Name:        "tls-cert",
			Usage:       "Path to certificate file used for https",
			Destination: &o.TLS.Cert,
			EnvVars:     []string{"VINCE_TLS_CERT"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "data",
			Usage:       "path to data directory",
			Value:       ".vince",
			Destination: &o.DataPath,
			EnvVars:     []string{"VINCE_DATA"},
		},
		&cli.StringFlag{
			Category:    "core",
			Name:        "url",
			Usage:       "url for the server on which vince is hosted(it shows up on emails)",
			Value:       "http://localhost:8080",
			Destination: &o.URL,
			EnvVars:     []string{"VINCE_URL"},
		},
		&cli.BoolFlag{
			Category:    "backup",
			Name:        "enable-backup",
			Usage:       "Allows backing up and restoring",
			Destination: &o.Backup.Enabled,
			EnvVars:     []string{"VINCE_BACKUP_ENABLED"},
		},
		&cli.StringFlag{
			Category:    "backup",
			Name:        "backup-dir",
			Usage:       "directory where backups are stored",
			Destination: &o.Backup.Dir,
			EnvVars:     []string{"VINCE_BACKUP_DIR"},
		},

		&cli.BoolFlag{
			Category:    "mailer",
			Name:        "enable-email",
			Usage:       "allows sending emails",
			Destination: &o.Mailer.Enabled,
			EnvVars:     []string{"VINCE_ENABLE_EMAIL"},
		},

		&cli.StringFlag{
			Category:    "mailer",
			Name:        "mailer-address",
			Usage:       "email address used for the sender of outgoing emails ",
			Value:       "vince@mailhog.example",
			Destination: &o.Mailer.Address,
			EnvVars:     []string{"VINCE_MAILER_ADDRESS"},
		},
		&cli.StringFlag{
			Category:    "mailer",
			Name:        "mailer-address-name",
			Usage:       "email address name  used for the sender of outgoing emails ",
			Value:       "gernest from vince analytics",
			Destination: &o.Mailer.Name,
			EnvVars:     []string{"VINCE_MAILER_ADDRESS_NAME"},
		},
		&cli.StringFlag{
			Category:    "mailer.smtp",
			Name:        "mailer-smtp-address",
			Usage:       "host:port address of the smtp server used for outgoing emails",
			Value:       "localhost:1025",
			Destination: &o.Mailer.SMTP.Address,
			EnvVars:     []string{"VINCE_MAILER_SMTP_ADDRESS"},
		},
		&cli.BoolFlag{
			Category:    "mailer.smtp",
			Name:        "mailer-smtp-enable-mailhog",
			Usage:       "port address of the smtp server used for outgoing emails",
			Destination: &o.Mailer.SMTP.EnableMailHog,
			EnvVars:     []string{"VINCE_MAILER_SMTP_ENABLE_MAILHOG"},
		},
		&cli.BoolFlag{
			Category:    "mailer.smtp.auth.anonymous",
			Name:        "mailer-smtp-anonymous-enable",
			Usage:       "enables anonymous authenticating smtp client",
			Destination: &o.Mailer.SMTP.AuthAnonymous.Enabled,
			EnvVars:     []string{"VINCE_MAILER_SMTP_ANONYMOUS_ENABLED"},
		},
		&cli.StringFlag{
			Category:    "mailer.smtp.auth.anonymous",
			Name:        "mailer-smtp-anonymous-trace",
			Usage:       "trace value for anonymous smtp auth",
			Destination: &o.Mailer.SMTP.AuthAnonymous.Trace,
			EnvVars:     []string{"VINCE_MAILER_SMTP_ANONYMOUS_TRACE"},
		},
		&cli.BoolFlag{
			Category:    "mailer.smtp.auth.plain",
			Name:        "mailer-smtp-plain-enabled",
			Usage:       "enables PLAIN authentication of smtp client",
			Destination: &o.Mailer.SMTP.AuthPlain.Enabled,
			EnvVars:     []string{"VINCE_MAILER_SMTP_PLAIN_ENABLED"},
		},
		&cli.StringFlag{
			Category:    "mailer.smtp.auth.plain",
			Name:        "mailer-smtp-plain-identity",
			Usage:       "identity value for plain smtp auth",
			Destination: &o.Mailer.SMTP.AuthPlain.Identity,
			EnvVars:     []string{"VINCE_MAILER_SMTP_PLAIN_IDENTITY"},
		},
		&cli.StringFlag{
			Category:    "mailer.smtp.auth.plain",
			Name:        "mailer-smtp-plain-username",
			Usage:       "username value for plain smtp auth",
			Destination: &o.Mailer.SMTP.AuthPlain.Username,
			EnvVars:     []string{"VINCE_MAILER_SMTP_PLAIN_USERNAME"},
		},
		&cli.StringFlag{
			Category:    "mailer.smtp.auth.plain",
			Name:        "mailer-smtp-plain-password",
			Usage:       "password value for plain smtp auth",
			Destination: &o.Mailer.SMTP.AuthPlain.Password,
			EnvVars:     []string{"VINCE_MAILER_SMTP_PLAIN_PASSWORD"},
		},
		&cli.BoolFlag{
			Category:    "mailer.smtp.auth.oauth",
			Name:        "mailer-smtp-oauth-username",
			Usage:       "allows oauth authentication on smtp client",
			Destination: &o.Mailer.SMTP.AuthOAUTHBearer.Enabled,
			EnvVars:     []string{"VINCE_MAILER_SMTP_OAUTH_USERNAME"},
		},
		&cli.StringFlag{
			Category:    "mailer.smtp.auth.oauth",
			Name:        "mailer-smtp-oauth-token",
			Usage:       "token value for oauth bearer smtp auth",
			Destination: &o.Mailer.SMTP.AuthOAUTHBearer.Token,
			EnvVars:     []string{"VINCE_MAILER_SMTP_OAUTH_TOKEN"},
		},
		&cli.StringFlag{
			Category:    "mailer.smtp.auth.oauth",
			Name:        "mailer-smtp-oauth-host",
			Usage:       "host value for oauth bearer smtp auth",
			Destination: &o.Mailer.SMTP.AuthOAUTHBearer.Host,
			EnvVars:     []string{"VINCE_MAILER_SMTP_OAUTH_HOST"},
		},
		&cli.IntFlag{
			Category:    "mailer.smtp.auth.oauth",
			Name:        "mailer-smtp-oauth-port",
			Usage:       "port value for oauth bearer smtp auth",
			Destination: &o.Mailer.SMTP.AuthOAUTHBearer.Port,
			EnvVars:     []string{"VINCE_MAILER_SMTP_OAUTH_PORT"},
		},
		&cli.DurationFlag{
			Category:    "intervals",
			Name:        "cache-refresh-interval",
			Usage:       "window for refreshing sites cache",
			Value:       15 * time.Minute,
			Destination: &o.Intervals.SiteCache,
			EnvVars:     []string{"VINCE_SITE_CACHE_REFRESH_INTERVAL"},
		},
		&cli.DurationFlag{
			Category: "intervals",
			Name:     "ts-buffer-sync-interval",
			Usage:    "window for buffering timeseries in memory before savin them",
			// This seems reasonable to avoid users to wait for a long time between
			// creating the site and seeing something on the dashboard. A bigger
			// duration is better though, to reduce pressure on our kv store
			Value:       time.Minute,
			Destination: &o.Intervals.TSSync,
			EnvVars:     []string{"VINCE_TS_BUFFER_INTERVAL"},
		},
		// secrets
		&cli.StringFlag{
			Category:    "secrets",
			Name:        "secret",
			Usage:       "path to a file with  ed25519 private key",
			Destination: &o.Secrets.Secret,
			EnvVars:     []string{"VINCE_SECRET"},
		},
		&cli.StringFlag{
			Category:    "secrets",
			Name:        "secret-age",
			Usage:       "path to file with age.X25519Identity",
			Destination: &o.Secrets.Age,
			EnvVars:     []string{"VINCE_SECRET_AGE"},
		},
		&cli.BoolFlag{
			Category:    "tls.acme",
			Name:        "enable-auto-tls",
			Usage:       "Enables using acme for automatic https.",
			Destination: &o.Acme.Enabled,
			EnvVars:     []string{"VINCE_AUTO_TLS"},
		},
		&cli.StringFlag{
			Category:    "tls.acme",
			Name:        "acme-domain",
			Usage:       "Domain to use with letsencrypt.",
			Destination: &o.Acme.Domain,
			EnvVars:     []string{"VINCE_ACME_DOMAIN"},
		},
		&cli.StringFlag{
			Category:    "tls.acme",
			Name:        "acme-certs-path",
			Usage:       "Patch where issued certs will be stored",
			Destination: &o.Acme.CertsPath,
			EnvVars:     []string{"VINCE_ACME_CERTS_PATH"},
		},
		&cli.StringFlag{
			Category:    "tls.acme.issuer",
			Name:        "acme-issuer-ca",
			Usage:       "The endpoint of the directory for the ACME  CA",
			Destination: &o.Acme.Issuer.CA,
			Value:       certmagic.LetsEncryptProductionCA,
			EnvVars:     []string{"VINCE_ACME_ISSUER_CA"},
		},
		&cli.StringFlag{
			Category:    "tls.acme.issuer",
			Name:        "acme-issuer-test-ca",
			Usage:       "The endpoint of the directory for the ACME  CA to use to test domain validation",
			Destination: &o.Acme.Issuer.TestCA,
			Value:       certmagic.LetsEncryptStagingCA,
			EnvVars:     []string{"VINCE_ACME_ISSUER_TEST_CA"},
		},
		&cli.StringFlag{
			Category:    "tls.acme.issuer",
			Name:        "acme-issuer-email",
			Usage:       "The email address to use when creating or selecting an existing ACME server account",
			Destination: &o.Acme.Issuer.Email,
			EnvVars:     []string{"VINCE_ACME_ISSUER_EMAIL"},
		},
		&cli.StringFlag{
			Category:    "tls.acme.issuer",
			Name:        "acme-issuer-account-key-pem",
			Usage:       "The PEM-encoded private key of the ACME account to use",
			Destination: &o.Acme.Issuer.AccountKeyPEM,
			EnvVars:     []string{"VINCE_ACME_ISSUER_ACCOUNT_KEY_PEM"},
		},
		&cli.BoolFlag{
			Category:    "tls.acme.issuer",
			Name:        "acme-issuer-agreed",
			Usage:       "Agree to CA's subscriber agreement",
			Destination: &o.Acme.Issuer.Agreed,
			Value:       true,
			EnvVars:     []string{"VINCE_ACME_ISSUER_AGREED"},
		},
		&cli.StringFlag{
			Category:    "tls.acme.issuer.external-account",
			Name:        "acme-issuer-external-account-key-id",
			Destination: &o.Acme.ExternalAccount.KeyID,
			EnvVars:     []string{"VINCE_ACME_ISSUER_EXTERNAL_ACCOUNT_KEY_ID"},
		},
		&cli.StringFlag{
			Category:    "tls.acme.issuer.external-account",
			Name:        "acme-issuer-external-account-mac-key",
			Destination: &o.Acme.ExternalAccount.MACKey,
			EnvVars:     []string{"VINCE_ACME_ISSUER_EXTERNAL_ACCOUNT_MAC_KEY"},
		},
		&cli.BoolFlag{
			Category:    "tls.acme.issuer",
			Name:        "acme-issuer-disable-http-challenge",
			Destination: &o.Acme.Issuer.DisableHTTPChallenge,
			EnvVars:     []string{"VINCE_ACME_ISSUER_DISABLE_HTTP_CHALLENGE"},
		},
		&cli.BoolFlag{
			Category:    "tls.acme.issuer",
			Name:        "acme-issuer-disable-tls-alpn-challenge",
			Destination: &o.Acme.Issuer.DisableTLSALPNChallenge,
			EnvVars:     []string{"VINCE_ACME_ISSUER_DISABLE_TLS_ALPN_CHALLENGE"},
		},
		&cli.BoolFlag{
			Category:    "bootstrap",
			Name:        "enable-bootstrap",
			Usage:       "allows creating a user and api key on startup.",
			Destination: &o.Bootstrap.Enabled,
			EnvVars:     []string{"VINCE_ENABLE_BOOTSTRAP"},
		},
		&cli.StringFlag{
			Category:    "bootstrap",
			Name:        "bootstrap-name",
			Usage:       "Full name of the user to bootstrap.",
			Destination: &o.Bootstrap.Name,
			EnvVars:     []string{"VINCE_BOOTSTRAP_NAME"},
		},
		&cli.StringFlag{
			Category:    "bootstrap",
			Name:        "bootstrap-email",
			Usage:       "Email address of the user to bootstrap.",
			Destination: &o.Bootstrap.Email,
			EnvVars:     []string{"VINCE_BOOTSTRAP_EMAIL"},
		},
		&cli.StringFlag{
			Category:    "bootstrap",
			Name:        "bootstrap-password",
			Usage:       "Password of the user to bootstrap.",
			Destination: &o.Bootstrap.Password,
			EnvVars:     []string{"VINCE_BOOTSTRAP_PASSWORD"},
		},
		&cli.StringFlag{
			Category:    "bootstrap",
			Name:        "bootstrap-key",
			Usage:       "API Key of the user to bootstrap.",
			Destination: &o.Bootstrap.Key,
			EnvVars:     []string{"VINCE_BOOTSTRAP_KEY"},
		},
		&cli.BoolFlag{
			Category:    "core",
			Name:        "enable-profile",
			Usage:       "Expose /debug/pprof endpoint",
			Destination: &o.EnableProfile,
			EnvVars:     []string{"VINCE_ENABLE_PROFILE"},
		},
		&cli.BoolFlag{
			Category:    "alerts",
			Name:        "enable-alerts",
			Usage:       "allows loading and executing alerts",
			Destination: &o.Alerts.Enabled,
			EnvVars:     []string{"VINCE_ENABLE_ALERTS"},
		},
		&cli.StringFlag{
			Category:    "alerts",
			Name:        "alerts-source",
			Usage:       "path to directory with alerts scripts",
			Destination: &o.Alerts.Source,
			EnvVars:     []string{"VINCE_ALERTS_SOURCE"},
		},

		&cli.StringFlag{
			Category:    "cors",
			Name:        "cors-origin",
			Value:       "*",
			Destination: &o.Cors.Origin,
			EnvVars:     []string{"VINCE_CORS_ORIGIN"},
		},
		&cli.BoolFlag{
			Category:    "cors",
			Name:        "cors-credentials",
			Value:       true,
			Destination: &o.Cors.Credentials,
			EnvVars:     []string{"VINCE_CORS_ORIGIN"},
		},
		&cli.IntFlag{
			Category:    "cors",
			Name:        "cors-max-age",
			Value:       1_728_000,
			Destination: &o.Cors.MaxAge,
			EnvVars:     []string{"VINCE_CORS_MAX_AGE"},
		},
		&cli.StringSliceFlag{
			Category:    "cors",
			Name:        "cors-headers",
			Value:       []string{"Authorization", "Content-Type", "Accept", "Origin", "User-Agent", "DNT", "Cache-Control", "X-Mx-ReqToken", "Keep-Alive", "X-Requested-With", "If-Modified-Since", "X-CSRF-Token"},
			Destination: &o.Cors.Headers,
			EnvVars:     []string{"VINCE_CORS_HEADERS"},
		},
		&cli.StringSliceFlag{
			Category:    "cors",
			Name:        "cors-expose",
			Destination: &o.Cors.Expose,
			EnvVars:     []string{"VINCE_CORS_EXPOSE"},
		},
		&cli.StringSliceFlag{
			Category:    "cors",
			Name:        "cors-methods",
			Value:       []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			Destination: &o.Cors.Methods,
			EnvVars:     []string{"VINCE_CORS_METHODS"},
		},
		&cli.BoolFlag{
			Category:    "cors",
			Name:        "cors-send-preflight-response",
			Value:       true,
			Destination: &o.Cors.SendPreflightResponse,
			EnvVars:     []string{"VINCE_CORS_SEND_PREFLIGHT_RESPONSE"},
		},
		&cli.Uint64SliceFlag{
			Category:    "core",
			Name:        "super-users",
			Usage:       "a list of user ID with super privilege",
			Destination: &o.SuperUserId,
			EnvVars:     []string{"VINCE_SUPER_USERS"},
		},
		&cli.BoolFlag{
			Category:    "firewall",
			Name:        "enable-firewall",
			Usage:       "allow blocking ip address",
			Destination: &o.Firewall.Enabled,
			EnvVars:     []string{"VINCE_ENABLE_FIREWALL"},
		},
		&cli.StringSliceFlag{
			Category:    "firewall",
			Name:        "firewall-block-list",
			Usage:       "block  ip address from this list",
			Destination: &o.Firewall.BlockIP,
			EnvVars:     []string{"VINCE_FIREWALL_BLOCK_LIST"},
		},
		&cli.StringSliceFlag{
			Category:    "firewall",
			Name:        "firewall-allow-list",
			Usage:       "allow  ip address from this list",
			Destination: &o.Firewall.AllowIP,
			EnvVars:     []string{"VINCE_FIREWALL_ALLOW_LIST"},
		},
	}
}

func (o *Options) IsSuperUser(uid uint64) bool {
	for _, v := range o.SuperUserId {
		if v == uid {
			return true
		}
	}
	return false
}

// Test returns Options with initialized values. This rely only on env Variables
// for flags.
func Test(fn ...func(*Options)) *Options {
	o := &Options{}
	o.NoSignal = true
	set := flag.NewFlagSet("vince", flag.ContinueOnError)
	for _, f := range o.Flags() {
		err := f.Apply(set)
		if err != nil {
			log.Get().Fatal().Err(err).Msg("failed to apply flag")
		}
	}
	// setup http and https listeners
	o.Listen = randomAddress()
	o.TLS.Address = randomAddress()
	o.Mailer.SMTP.Address = randomAddress()

	// setup secrets
	o.Bootstrap.Key = base64.StdEncoding.EncodeToString(secrets.APIKey())
	o.Secrets.Secret = base64.StdEncoding.EncodeToString(secrets.ED25519())
	o.Secrets.Age = base64.StdEncoding.EncodeToString(secrets.AGE())

	// setup default user. We don't enable bootstrapping
	o.Bootstrap.Name = TestName
	o.Bootstrap.Email = TestEmail
	o.Bootstrap.Password = TestPassword

	for _, f := range fn {
		f(o)
	}
	return o
}

func randomAddress() string {
	ls, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Get().Fatal().Err(err).Msg("failed creating random listener")
	}
	a := ls.Addr().String()
	ls.Close()
	return a
}

func (o *Options) Magic() *certmagic.Config {
	magic := certmagic.NewDefault()
	path := o.Acme.CertsPath
	if path == "" {
		// Default to the data directory
		path = filepath.Join(o.DataPath, "certs")
	}
	magic.Storage = &certmagic.FileStorage{Path: path}
	issuer := o.Acme.Issuer
	if o.Acme.ExternalAccount.KeyID != "" {
		issuer.ExternalAccount = &o.Acme.ExternalAccount
	}
	magic.Issuers = append(magic.Issuers,
		certmagic.NewACMEIssuer(magic, issuer),
	)
	return magic
}
