# NAME

vince - The Cloud Native Web Analytics Platform.

# SYNOPSIS

vince

```
[--acme-domain]=[value]
[--acme-email]=[value]
[--alerts-source]=[value]
[--backup-dir]=[value]
[--bootstrap-email]=[value]
[--bootstrap-key]=[value]
[--bootstrap-name]=[value]
[--bootstrap-password]=[value]
[--cache-refresh-interval]=[value]
[--cors-credentials]
[--cors-expose]=[value]
[--cors-headers]=[value]
[--cors-max-age]=[value]
[--cors-methods]=[value]
[--cors-origin]=[value]
[--cors-send-preflight-response]
[--data]=[value]
[--enable-alerts]
[--enable-auto-tls]
[--enable-backup]
[--enable-bootstrap]
[--enable-email]
[--enable-firewall]
[--enable-profile]
[--enable-tls]
[--env]=[value]
[--firewall-allow-list]=[value]
[--firewall-block-list]=[value]
[--listen]=[value]
[--log-level]=[value]
[--mailer-address-name]=[value]
[--mailer-address]=[value]
[--mailer-smtp-address]=[value]
[--mailer-smtp-anonymous-enable]
[--mailer-smtp-anonymous-trace]=[value]
[--mailer-smtp-enable-mailhog]
[--mailer-smtp-oauth-host]=[value]
[--mailer-smtp-oauth-port]=[value]
[--mailer-smtp-oauth-token]=[value]
[--mailer-smtp-oauth-username]
[--mailer-smtp-plain-enabled]
[--mailer-smtp-plain-identity]=[value]
[--mailer-smtp-plain-password]=[value]
[--mailer-smtp-plain-username]=[value]
[--secret-age]=[value]
[--secret]=[value]
[--super-users]=[value]
[--tls-address]=[value]
[--tls-cert]=[value]
[--tls-key]=[value]
[--ts-buffer-sync-interval]=[value]
[--url]=[value]
```

**Usage**:

```
vince [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# GLOBAL OPTIONS

**--acme-domain**="": Domain to use with letsencrypt.

**--acme-email**="": Email address to use with letsencrypt.

**--alerts-source**="": path to directory with alerts scripts

**--backup-dir**="": directory where backups are stored

**--bootstrap-email**="": Email address of the user to bootstrap.

**--bootstrap-key**="": API Key of the user to bootstrap.

**--bootstrap-name**="": Full name of the user to bootstrap.

**--bootstrap-password**="": Password of the user to bootstrap.

**--cache-refresh-interval**="": window for refreshing sites cache (default: 15m0s)

**--cors-credentials**: 

**--cors-expose**="":  (default: [])

**--cors-headers**="":  (default: [Authorization Content-Type Accept Origin User-Agent DNT Cache-Control X-Mx-ReqToken Keep-Alive X-Requested-With If-Modified-Since X-CSRF-Token])

**--cors-max-age**="":  (default: 1728000)

**--cors-methods**="":  (default: [GET POST PUT PATCH DELETE OPTIONS])

**--cors-origin**="":  (default: *)

**--cors-send-preflight-response**: 

**--data**="": path to data directory (default: .vince)

**--enable-alerts**: allows loading and executing alerts

**--enable-auto-tls**: Enables using acme for automatic https.

**--enable-backup**: Allows backing up and restoring

**--enable-bootstrap**: allows creating a user and api key on startup.

**--enable-email**: allows sending emails

**--enable-firewall**: allow blocking ip address

**--enable-profile**: Expose /debug/pprof endpoint

**--enable-tls**: Enables serving https traffic.

**--env**="": environment on which vince is run (dev,staging,production) (default: dev)

**--firewall-allow-list**="": allow  ip address from this list (default: [])

**--firewall-block-list**="": block  ip address from this list (default: [])

**--listen**="": http address to listen to (default: :8080)

**--log-level**="": log level, values are (trace,debug,info,warn,error,fatal,panic) (default: debug)

**--mailer-address**="": email address used for the sender of outgoing emails  (default: vince@mailhog.example)

**--mailer-address-name**="": email address name  used for the sender of outgoing emails  (default: gernest from vince analytics)

**--mailer-smtp-address**="": host:port address of the smtp server used for outgoing emails (default: localhost:1025)

**--mailer-smtp-anonymous-enable**: enables anonymous authenticating smtp client

**--mailer-smtp-anonymous-trace**="": trace value for anonymous smtp auth

**--mailer-smtp-enable-mailhog**: port address of the smtp server used for outgoing emails

**--mailer-smtp-oauth-host**="": host value for oauth bearer smtp auth

**--mailer-smtp-oauth-port**="": port value for oauth bearer smtp auth (default: 0)

**--mailer-smtp-oauth-token**="": token value for oauth bearer smtp auth

**--mailer-smtp-oauth-username**: allows oauth authentication on smtp client

**--mailer-smtp-plain-enabled**: enables PLAIN authentication of smtp client

**--mailer-smtp-plain-identity**="": identity value for plain smtp auth

**--mailer-smtp-plain-password**="": password value for plain smtp auth

**--mailer-smtp-plain-username**="": username value for plain smtp auth

**--secret**="": path to a file with  ed25519 private key

**--secret-age**="": path to file with age.X25519Identity

**--super-users**="": a list of user ID with super privilege (default: [])

**--tls-address**="": https address to listen to. You must provide tls-key and tls-cert or configure auto-tls (default: :8443)

**--tls-cert**="": Path to certificate file used for https

**--tls-key**="": Path to key file used for https

**--ts-buffer-sync-interval**="": window for buffering timeseries in memory before savin them (default: 1m0s)

**--url**="": url for the server on which vince is hosted(it shows up on emails) (default: http://localhost:8080)


# COMMANDS

## config

generates configurations for vince

## version

prints version information
