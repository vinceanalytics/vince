---
title: Configuration
---

# Configuration

**Vince** acn be configured with commandline flags or environment variables.

::: tip
We recommend the use of environment variables. They are safer (passing secrets ),
and it allows simpler deployments, where you set the environment and just call 
**vince** command.
:::


# Alerts

## alerts-source
comma separated list of alert files of the form file[name,interval] eg foo.ts[spike,15m]
::: code-group
```shell [flag]
--alerts-source=""
```
```shell [env]
VINCE_ALERTS_SOURCE=""
```
:::
## enable-alerts
allows loading and executing alerts
::: code-group
```shell [flag]
--enable-alerts="false"
```
```shell [env]
VINCE_ENABLE_ALERTS="false"
```
:::
# Backup

## backup-dir
directory where backups are stored
::: code-group
```shell [flag]
--backup-dir="..."
```
```shell [env]
VINCE_BACKUP_DIR="..."
```
:::
## enable-backup
Allows backing up and restoring
::: code-group
```shell [flag]
--enable-backup="false"
```
```shell [env]
VINCE_BACKUP_ENABLED="false"
```
:::
# Bootstrap

## bootstrap-email
Email address of the user to bootstrap.
::: code-group
```shell [flag]
--bootstrap-email="..."
```
```shell [env]
VINCE_BOOTSTRAP_EMAIL="..."
```
:::
## bootstrap-full-name
Full name of the user to bootstrap.
::: code-group
```shell [flag]
--bootstrap-full-name="..."
```
```shell [env]
VINCE_BOOTSTRAP_FULL_NAME="..."
```
:::
## bootstrap-key
API Key of the user to bootstrap.
::: code-group
```shell [flag]
--bootstrap-key="..."
```
```shell [env]
VINCE_BOOTSTRAP_KEY="..."
```
:::
## bootstrap-name
User name of the user to bootstrap.
::: code-group
```shell [flag]
--bootstrap-name="..."
```
```shell [env]
VINCE_BOOTSTRAP_NAME="..."
```
:::
## bootstrap-password
Password of the user to bootstrap.
::: code-group
```shell [flag]
--bootstrap-password="..."
```
```shell [env]
VINCE_BOOTSTRAP_PASSWORD="..."
```
:::
## enable-bootstrap
allows creating a user and api key on startup.
::: code-group
```shell [flag]
--enable-bootstrap="false"
```
```shell [env]
VINCE_ENABLE_BOOTSTRAP="false"
```
:::
# Core

## data
path to data directory
::: code-group
```shell [flag]
--data=".vince"
```
```shell [env]
VINCE_DATA=".vince"
```
:::
## enable-profile
Expose /debug/pprof endpoint
::: code-group
```shell [flag]
--enable-profile="false"
```
```shell [env]
VINCE_ENABLE_PROFILE="false"
```
:::
## listen
http address to listen to
::: code-group
```shell [flag]
--listen=":8080"
```
```shell [env]
VINCE_LISTEN=":8080"
```
:::
## log-level
log level, values are (trace,debug,info,warn,error,fatal,panic)
::: code-group
```shell [flag]
--log-level="debug"
```
```shell [env]
VINCE_LOG_LEVEL="debug"
```
:::
## super-users
a list of user names with super privilege
::: code-group
```shell [flag]
--super-users=""
```
```shell [env]
VINCE_SUPER_USERS=""
```
:::
## uploads-dir
Path to store uploaded assets
::: code-group
```shell [flag]
--uploads-dir="..."
```
```shell [env]
VINCE_UPLOAD_DIR="..."
```
:::
## url
url for the server on which vince is hosted(it shows up on emails)
::: code-group
```shell [flag]
--url="http://localhost:8080"
```
```shell [env]
VINCE_URL="http://localhost:8080"
```
:::
# Cors

## cors-credentials

::: code-group
```shell [flag]
--cors-credentials="true"
```
```shell [env]
VINCE_CORS_ORIGIN="true"
```
:::
## cors-expose

::: code-group
```shell [flag]
--cors-expose=""
```
```shell [env]
VINCE_CORS_EXPOSE=""
```
:::
## cors-headers

::: code-group
```shell [flag]
--cors-headers="Authorization,Content-Type,Accept,Origin,User-Agent,DNT,Cache-Control,X-Mx-ReqToken,Keep-Alive,X-Requested-With,If-Modified-Since,X-CSRF-Token"
```
```shell [env]
VINCE_CORS_HEADERS="Authorization,Content-Type,Accept,Origin,User-Agent,DNT,Cache-Control,X-Mx-ReqToken,Keep-Alive,X-Requested-With,If-Modified-Since,X-CSRF-Token"
```
:::
## cors-max-age

::: code-group
```shell [flag]
--cors-max-age="1728000"
```
```shell [env]
VINCE_CORS_MAX_AGE="1728000"
```
:::
## cors-methods

::: code-group
```shell [flag]
--cors-methods="GET,POST,PUT,PATCH,DELETE,OPTIONS"
```
```shell [env]
VINCE_CORS_METHODS="GET,POST,PUT,PATCH,DELETE,OPTIONS"
```
:::
## cors-origin

::: code-group
```shell [flag]
--cors-origin="*"
```
```shell [env]
VINCE_CORS_ORIGIN="*"
```
:::
## cors-send-preflight-response

::: code-group
```shell [flag]
--cors-send-preflight-response="true"
```
```shell [env]
VINCE_CORS_SEND_PREFLIGHT_RESPONSE="true"
```
:::
# Firewall

## enable-firewall
allow blocking ip address
::: code-group
```shell [flag]
--enable-firewall="false"
```
```shell [env]
VINCE_ENABLE_FIREWALL="false"
```
:::
## firewall-allow-list
allow  ip address from this list
::: code-group
```shell [flag]
--firewall-allow-list=""
```
```shell [env]
VINCE_FIREWALL_ALLOW_LIST=""
```
:::
## firewall-block-list
block  ip address from this list
::: code-group
```shell [flag]
--firewall-block-list=""
```
```shell [env]
VINCE_FIREWALL_BLOCK_LIST=""
```
:::
# Intervals

## cache-refresh-interval
window for refreshing sites cache
::: code-group
```shell [flag]
--cache-refresh-interval="15m0s"
```
```shell [env]
VINCE_SITE_CACHE_REFRESH_INTERVAL="15m0s"
```
:::
## system-interval
Interval for collecting system metrics
::: code-group
```shell [flag]
--system-interval="1m0s"
```
```shell [env]
VINCE_SYSTEM_INTERVAL="1m0s"
```
:::
## ts-buffer-sync-interval
window for buffering timeseries in memory before savin them
::: code-group
```shell [flag]
--ts-buffer-sync-interval="1m0s"
```
```shell [env]
VINCE_TS_BUFFER_INTERVAL="1m0s"
```
:::
# Mailer

## enable-email
allows sending emails
::: code-group
```shell [flag]
--enable-email="false"
```
```shell [env]
VINCE_ENABLE_EMAIL="false"
```
:::
## mailer-address
email address used for the sender of outgoing emails 
::: code-group
```shell [flag]
--mailer-address="vince@mailhog.example"
```
```shell [env]
VINCE_MAILER_ADDRESS="vince@mailhog.example"
```
:::
## mailer-address-name
email address name  used for the sender of outgoing emails 
::: code-group
```shell [flag]
--mailer-address-name="gernest from vince analytics"
```
```shell [env]
VINCE_MAILER_ADDRESS_NAME="gernest from vince analytics"
```
:::
## mailer-smtp-address
host:port address of the smtp server used for outgoing emails
::: code-group
```shell [flag]
--mailer-smtp-address="localhost:1025"
```
```shell [env]
VINCE_MAILER_SMTP_ADDRESS="localhost:1025"
```
:::
## mailer-smtp-anonymous-enable
enables anonymous authenticating smtp client
::: code-group
```shell [flag]
--mailer-smtp-anonymous-enable="false"
```
```shell [env]
VINCE_MAILER_SMTP_ANONYMOUS_ENABLED="false"
```
:::
## mailer-smtp-anonymous-trace
trace value for anonymous smtp auth
::: code-group
```shell [flag]
--mailer-smtp-anonymous-trace="..."
```
```shell [env]
VINCE_MAILER_SMTP_ANONYMOUS_TRACE="..."
```
:::
## mailer-smtp-enable-mailhog
port address of the smtp server used for outgoing emails
::: code-group
```shell [flag]
--mailer-smtp-enable-mailhog="false"
```
```shell [env]
VINCE_MAILER_SMTP_ENABLE_MAILHOG="false"
```
:::
## mailer-smtp-oauth-host
host value for oauth bearer smtp auth
::: code-group
```shell [flag]
--mailer-smtp-oauth-host="..."
```
```shell [env]
VINCE_MAILER_SMTP_OAUTH_HOST="..."
```
:::
## mailer-smtp-oauth-port
port value for oauth bearer smtp auth
::: code-group
```shell [flag]
--mailer-smtp-oauth-port="0"
```
```shell [env]
VINCE_MAILER_SMTP_OAUTH_PORT="0"
```
:::
## mailer-smtp-oauth-token
token value for oauth bearer smtp auth
::: code-group
```shell [flag]
--mailer-smtp-oauth-token="..."
```
```shell [env]
VINCE_MAILER_SMTP_OAUTH_TOKEN="..."
```
:::
## mailer-smtp-oauth-username
allows oauth authentication on smtp client
::: code-group
```shell [flag]
--mailer-smtp-oauth-username="false"
```
```shell [env]
VINCE_MAILER_SMTP_OAUTH_USERNAME="false"
```
:::
## mailer-smtp-plain-enabled
enables PLAIN authentication of smtp client
::: code-group
```shell [flag]
--mailer-smtp-plain-enabled="false"
```
```shell [env]
VINCE_MAILER_SMTP_PLAIN_ENABLED="false"
```
:::
## mailer-smtp-plain-identity
identity value for plain smtp auth
::: code-group
```shell [flag]
--mailer-smtp-plain-identity="..."
```
```shell [env]
VINCE_MAILER_SMTP_PLAIN_IDENTITY="..."
```
:::
## mailer-smtp-plain-password
password value for plain smtp auth
::: code-group
```shell [flag]
--mailer-smtp-plain-password="..."
```
```shell [env]
VINCE_MAILER_SMTP_PLAIN_PASSWORD="..."
```
:::
## mailer-smtp-plain-username
username value for plain smtp auth
::: code-group
```shell [flag]
--mailer-smtp-plain-username="..."
```
```shell [env]
VINCE_MAILER_SMTP_PLAIN_USERNAME="..."
```
:::
# Secrets

## secret
path to a file with  ed25519 private key
::: code-group
```shell [flag]
--secret="..."
```
```shell [env]
VINCE_SECRET="..."
```
:::
## secret-age
path to file with age.X25519Identity
::: code-group
```shell [flag]
--secret-age="..."
```
```shell [env]
VINCE_SECRET_AGE="..."
```
:::
# Tls

## acme-certs-path
Patch where issued certs will be stored
::: code-group
```shell [flag]
--acme-certs-path="..."
```
```shell [env]
VINCE_ACME_CERTS_PATH="..."
```
:::
## acme-domain
Domain to use with letsencrypt.
::: code-group
```shell [flag]
--acme-domain="..."
```
```shell [env]
VINCE_ACME_DOMAIN="..."
```
:::
## acme-issuer-account-key-pem
The PEM-encoded private key of the ACME account to use
::: code-group
```shell [flag]
--acme-issuer-account-key-pem="..."
```
```shell [env]
VINCE_ACME_ISSUER_ACCOUNT_KEY_PEM="..."
```
:::
## acme-issuer-agreed
Agree to CA's subscriber agreement
::: code-group
```shell [flag]
--acme-issuer-agreed="true"
```
```shell [env]
VINCE_ACME_ISSUER_AGREED="true"
```
:::
## acme-issuer-ca
The endpoint of the directory for the ACME  CA
::: code-group
```shell [flag]
--acme-issuer-ca="https://acme-v02.api.letsencrypt.org/directory"
```
```shell [env]
VINCE_ACME_ISSUER_CA="https://acme-v02.api.letsencrypt.org/directory"
```
:::
## acme-issuer-disable-http-challenge

::: code-group
```shell [flag]
--acme-issuer-disable-http-challenge="false"
```
```shell [env]
VINCE_ACME_ISSUER_DISABLE_HTTP_CHALLENGE="false"
```
:::
## acme-issuer-disable-tls-alpn-challenge

::: code-group
```shell [flag]
--acme-issuer-disable-tls-alpn-challenge="false"
```
```shell [env]
VINCE_ACME_ISSUER_DISABLE_TLS_ALPN_CHALLENGE="false"
```
:::
## acme-issuer-email
The email address to use when creating or selecting an existing ACME server account
::: code-group
```shell [flag]
--acme-issuer-email="..."
```
```shell [env]
VINCE_ACME_ISSUER_EMAIL="..."
```
:::
## acme-issuer-external-account-key-id

::: code-group
```shell [flag]
--acme-issuer-external-account-key-id="..."
```
```shell [env]
VINCE_ACME_ISSUER_EXTERNAL_ACCOUNT_KEY_ID="..."
```
:::
## acme-issuer-external-account-mac-key

::: code-group
```shell [flag]
--acme-issuer-external-account-mac-key="..."
```
```shell [env]
VINCE_ACME_ISSUER_EXTERNAL_ACCOUNT_MAC_KEY="..."
```
:::
## acme-issuer-test-ca
The endpoint of the directory for the ACME  CA to use to test domain validation
::: code-group
```shell [flag]
--acme-issuer-test-ca="https://acme-staging-v02.api.letsencrypt.org/directory"
```
```shell [env]
VINCE_ACME_ISSUER_TEST_CA="https://acme-staging-v02.api.letsencrypt.org/directory"
```
:::
## enable-auto-tls
Enables using acme for automatic https.
::: code-group
```shell [flag]
--enable-auto-tls="false"
```
```shell [env]
VINCE_AUTO_TLS="false"
```
:::
## enable-tls
Enables serving https traffic.
::: code-group
```shell [flag]
--enable-tls="false"
```
```shell [env]
VINCE_ENABLE_TLS="false"
```
:::
## tls-address
https address to listen to. You must provide tls-key and tls-cert or configure auto-tls
::: code-group
```shell [flag]
--tls-address=":8443"
```
```shell [env]
VINCE_TLS_LISTEN=":8443"
```
:::
## tls-cert
Path to certificate file used for https
::: code-group
```shell [flag]
--tls-cert="..."
```
```shell [env]
VINCE_TLS_CERT="..."
```
:::
## tls-key
Path to key file used for https
::: code-group
```shell [flag]
--tls-key="..."
```
```shell [env]
VINCE_TLS_KEY="..."
```
:::
