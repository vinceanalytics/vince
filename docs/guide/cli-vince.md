# NAME

vince - The Cloud Native Web Analytics Platform.

# SYNOPSIS

vince

```
[--acme-domain]=[value]
[--acme-email]=[value]
[--backup-dir]=[value]
[--bootstrap-email]=[value]
[--bootstrap-key]=[value]
[--bootstrap-name]=[value]
[--bootstrap-password]=[value]
[--cache-refresh]=[value]
[--config]=[value]
[--data]=[value]
[--enable-auto-tls]
[--enable-bootstrap]
[--enable-email-verification]
[--enable-system-stats]
[--enable-tls]
[--env]=[value]
[--listen]=[value]
[--log]=[value]
[--mailer-address-name]=[value]
[--mailer-address]=[value]
[--mailer-smtp-anonymous]=[value]
[--mailer-smtp-host]=[value]
[--mailer-smtp-oauth-host]=[value]
[--mailer-smtp-oauth-port]=[value]
[--mailer-smtp-oauth-token]=[value]
[--mailer-smtp-oauth-username]=[value]
[--mailer-smtp-plain-identity]=[value]
[--mailer-smtp-plain-password]=[value]
[--mailer-smtp-plain-username]=[value]
[--mailer-smtp-port]=[value]
[--rotation-check]=[value]
[--scrape-interval]=[value]
[--secret-age-priv]=[value]
[--secret-age-pub]=[value]
[--secret-ed-priv]=[value]
[--secret-ed-pub]=[value]
[--tls-address]=[value]
[--tls-cert]=[value]
[--tls-key]=[value]
[--ts-buffer]=[value]
[--url]=[value]
```

**Usage**:

```
vince [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--acme-domain**="": Domain to use with letsencrypt.

**--acme-email**="": Email address to use with letsencrypt.

**--backup-dir**="": directory where backups are stored

**--bootstrap-email**="": Email address of the user to bootstrap.

**--bootstrap-key**="": API Key of the user to bootstrap.

**--bootstrap-name**="": Full name of the user to bootstrap.

**--bootstrap-password**="": Password of the user to bootstrap.

**--cache-refresh**="": window for refreshing sites cache (default: 15m0s)

**--config**="": configuration file in json format (default: vince.json)

**--data**="": path to data directory (default: .vince)

**--enable-auto-tls**: Enables using acme for automatic https.

**--enable-bootstrap**: allows creating a user and api key on startup.

**--enable-email-verification**: send emails for account verification

**--enable-system-stats**: Collect and visualize system stats

**--enable-tls**: Enables serving https traffick.

**--env**="": environment on which vince is run (dev,staging,production) (default: dev)

**--listen**="": http address to listen to (default: :8080)

**--log**="": level of logging (default: info)

**--mailer-address**="": email address used for the sender of outgoing emails  (default: vince@mailhog.example)

**--mailer-address-name**="": email address name  used for the sender of outgoing emails  (default: gernest from vince analytics)

**--mailer-smtp-anonymous**="": trace value for anonymous smtp auth

**--mailer-smtp-host**="": host address of the smtp server used for outgoing emails (default: localhost)

**--mailer-smtp-oauth-host**="": host value for oauth bearer smtp auth

**--mailer-smtp-oauth-port**="": port value for oauth bearer smtp auth (default: 0)

**--mailer-smtp-oauth-token**="": token value for oauth bearer smtp auth

**--mailer-smtp-oauth-username**="": username value for oauth bearer smtp auth

**--mailer-smtp-plain-identity**="": identity value for plain smtp auth

**--mailer-smtp-plain-password**="": password value for plain smtp auth

**--mailer-smtp-plain-username**="": username value for plain smtp auth

**--mailer-smtp-port**="": port address of the smtp server used for outgoing emails (default: 1025)

**--rotation-check**="": window for checking log rotation (default: 1h0m0s)

**--scrape-interval**="": system wide metrics collection interval (default: 1m0s)

**--secret-age-priv**="": path to a file with  age private key

**--secret-age-pub**="": path to a file with  age public key

**--secret-ed-priv**="": path to a file with  ed25519 private key

**--secret-ed-pub**="": path to a file with  ed25519 public key

**--tls-address**="": https address to listen to. You must provide tls-key and tls-cert or configure auto-tls (default: :8443)

**--tls-cert**="": Path to certificate file used for https

**--tls-key**="": Path to key file used for https

**--ts-buffer**="": window for buffering timeseries in memory before savin them (default: 1m0s)

**--url**="": url for the server on which vince is hosted(it shows up on emails)


# COMMANDS

## config

generates configurations for vince

**--path**="": directory to save configurations (including secrets) (default: .vince)

## version

prints version information
