# http address to listen to
export  VINCE_LISTEN=":8080"
# log level, values are (trace,debug,info,warn,error,fatal,panic)
export  VINCE_LOG_LEVEL="debug"
# Enables serving https traffic.
export  VINCE_ENABLE_TLS="false"
# https address to listen to. You must provide tls-key and tls-cert or configure auto-tls
export  VINCE_TLS_LISTEN=":8443"
# Path to key file used for https
export  VINCE_TLS_KEY=""
# Path to certificate file used for https
export  VINCE_TLS_CERT=""
# path to data directory
export  VINCE_DATA=".vince"
# environment on which vince is run (dev,staging,production)
export  VINCE_ENV="dev"
# url for the server on which vince is hosted(it shows up on emails)
export  VINCE_URL="http://localhost:8080"
# Allows backing up and restoring
export  VINCE_BACKUP_ENABLED="false"
# directory where backups are stored
export  VINCE_BACKUP_DIR=""
# allows sending emails
export  VINCE_ENABLE_EMAIL="false"
# email address used for the sender of outgoing emails 
export  VINCE_MAILER_ADDRESS="vince@mailhog.example"
# email address name  used for the sender of outgoing emails 
export  VINCE_MAILER_ADDRESS_NAME="gernest from vince analytics"
# host:port address of the smtp server used for outgoing emails
export  VINCE_MAILER_SMTP_ADDRESS="localhost:1025"
# port address of the smtp server used for outgoing emails
export  VINCE_MAILER_SMTP_ENABLE_MAILHOG="true"
# enables anonymous authenticating smtp client
export  VINCE_MAILER_SMTP_ANONYMOUS_ENABLED="false"
# trace value for anonymous smtp auth
export  VINCE_MAILER_SMTP_ANONYMOUS_TRACE=""
# enables PLAIN authentication of smtp client
export  VINCE_MAILER_SMTP_PLAIN_ENABLED="false"
# identity value for plain smtp auth
export  VINCE_MAILER_SMTP_PLAIN_IDENTITY=""
# username value for plain smtp auth
export  VINCE_MAILER_SMTP_PLAIN_USERNAME=""
# password value for plain smtp auth
export  VINCE_MAILER_SMTP_PLAIN_PASSWORD=""
# allows oauth authentication on smtp client
export  VINCE_MAILER_SMTP_OAUTH_USERNAME="false"
# token value for oauth bearer smtp auth
export  VINCE_MAILER_SMTP_OAUTH_TOKEN=""
# host value for oauth bearer smtp auth
export  VINCE_MAILER_SMTP_OAUTH_HOST=""
# port value for oauth bearer smtp auth
export  VINCE_MAILER_SMTP_OAUTH_PORT="0"
# window for refreshing sites cache
export  VINCE_SITE_CACHE_REFRESH_INTERVAL="15m0s"
# window for buffering timeseries in memory before savin them
export  VINCE_TS_BUFFER_INTERVAL="1m0s"
# path to a file with  ed25519 private key
export  VINCE_SECRET="LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUVWaVBkUXl3Y3NxUlJaNmtlVFRiTTF2MkZSYWt0cGd6MDR3emJtTXpsL0sKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo="
# path to file with age.X25519Identity
export  VINCE_SECRET_AGE="QUdFLVNFQ1JFVC1LRVktMTU0V1lNRUNRNEFBRENRVEc1UDhHRVlMU0c1QUw5VlJXTlRYNEY2TldLRkpORFk4WkRSSlFHWjJYNEE="
# Enables using acme for automatic https.
export  VINCE_AUTO_TLS="false"
# Email address to use with letsencrypt.
export  VINCE_ACME_EMAIL=""
# Domain to use with letsencrypt.
export  VINCE_ACME_DOMAIN=""

#region bootstrap
# allows creating a user and api key on startup.
export  VINCE_ENABLE_BOOTSTRAP="false"
# Full name of the user to bootstrap.
export  VINCE_BOOTSTRAP_NAME=""
# Email address of the user to bootstrap.
export  VINCE_BOOTSTRAP_EMAIL=""
# Password of the user to bootstrap.
export  VINCE_BOOTSTRAP_PASSWORD=""
#endregion bootstrap

# API Key of the user to bootstrap.
export  VINCE_BOOTSTRAP_KEY="Y1YQnJjrrGrRcWWzUMeOiQ2eITkKOouW9mNYD7dREG+BHUz2HiX+SFUDwzTsiZ6JHqCCwngmIpHT2yvJUSwMVg=="
# Expose /debug/pprof endpoint
export  VINCE_ENABLE_PROFILE="false"
# allows loading and executing alerts
export  VINCE_ENABLE_ALERTS="false"
# path to directory with alerts scripts
export  VINCE_ALERTS_SOURCE=""
# 
export  VINCE_CORS_ORIGIN="*"
# 
export  VINCE_CORS_ORIGIN="true"
# 
export  VINCE_CORS_MAX_AGE="1728000"
# 
export  VINCE_CORS_HEADERS="Authorization,Content-Type,Accept,Origin,User-Agent,DNT,Cache-Control,X-Mx-ReqToken,Keep-Alive,X-Requested-With,If-Modified-Since,X-CSRF-Token"
# 
export  VINCE_CORS_EXPOSE=""
# 
export  VINCE_CORS_METHODS="GET,POST,PUT,PATCH,DELETE,OPTIONS"
# 
export  VINCE_CORS_SEND_PREFLIGHT_RESPONSE="true"
# a list of user ID with super privilege
export  VINCE_SUPER_USERS=""
# allow blocking ip address
export  VINCE_ENABLE_FIREWALL="false"
# block  ip address from this list
export  VINCE_FIREWALL_BLOCK_LIST=""
# allow  ip address from this list
export  VINCE_FIREWALL_ALLOW_LIST=""
