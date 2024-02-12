# Configuration

`vince` has three way of passing configuration, commandline flags, environment
variables and configuration file.

All three ways can be combined to form a secure  deployments. The level of precedence follows 
 cli `->` env `->` file. So if lets say `listen` is provided by all ways, then
the value set in `file` will be used.

> We recommend using *commandline flags and environment variables*.
> Anything that can be expressed in configuration file can also be expressed 
> with commandline flags and environment variables.


## Data
Path to directory where `vince` will store persisting data. This option is required for `vince` to be operational.

*env*
: `VINCE_DATA` example `VINCE_DATA=path/to/storage`

*flag*
: `--data` example `--data=path/to/storage`

*file*
: `data` 

```json
{"data":"path/to/storage"}
```


## Listen
HTTP `host:port` that vince server will listen for http requests

*env*
: `VINCE_LISTEN` example `VINCE_LISTEN=:8080`

*flag*
: `--listen` example `--listen=:8080`

*file*
: `data` 

```json
{"listen":":8080"}
```

## Rate Limit
A float representing `requests/second`. This is applied on `/api/event` and `/api/v1/event` protect them against excessive incoming requests.

By default there is no limit.

> Limits are applied globally across all sites. We don't support per site
> limits yet, but it is something planned for the future releases

*env*
: `VINCE_RATE_LIMIT` example `VINCE_RATE_LIMIT=1.7976931348623157e+308`

*flag*
: `--rateLimit` example `--rateLimit=1.7976931348623157e+308`

*file*
: `rateLimit` 

```json
{"rateLimit":1.7976931348623157e+308}
```

## Granule Size
Size in bytes for indexes+parts that when reached they are compacted and stored in disk. 

By default `16MB` is used. You need to adjust this depending on your deployment resources.

> Unless otherwise `16MB` is a very conservative balance. This option is for power users.
> You can have production deployments sticking with the defaults.

When in memory index+`arrow.Record` reaches this size, we convert the arrow
record to parquet file with compression enabled, then the resulting file is stored in durable key/value store with configured retention period.

So, While `16MB` is in memory, what actually goes to disk is smaller than this value. We also compress the index  before storing it as well.


*env*
: `VINCE_GRANULE_SIZE` example `VINCE_GRANULE_SIZE=16777216`

*flag*
: `--granuleSize` example `--granuleSize=16777216`

*file*
: `granuleSize` 

```json
{"granuleSize":"16777216"}
```

## Geo IP DB
Path to geo ip database. We use this to  obtain `city`, `country` and `region` information of an event.

This is optional.

*env*
: `VINCE_GEOIP_DB` example `VINCE_GEOIP_DB=path/to/geoip_db`

*flag*
: `--geoipDbPath` example `--geoipDbPath=path/to/geoip_db`

*file*
: `geoipDbPath` 

```json
{"geoipDbPath":"path/to/geoip_db"}
```

## Domains
A list of domains managed by `vince` instance. To send events to `vince` there is no extra configuration on the client side.

Events submitted that have no registered domain are rejected. If you no longer manage the site simply removing it on this list will stop accepting events from it.


> Domain is  the hostname. For example `https://vinceanalytics.com` has domain `vinceanalytics.com`

*env*
: `VINCE_DOMAINS` example `VINCE_DOMAINS=vinceanalytics.com,example.com`

*flag*
: `--domains` example `--domains=vinceanalytics.com,example.com`

*file*
: `domains` 

```json
{"domains":["vinceanalytics.com","example.com"]}
```

## Configuration file path
Path to configuration file. When provided it will be read and merged with the rest of configurations. Values set in this file takes precedence.

*env*
: `VINCE_CONFIG` example `VINCE_CONFIG=path/to/config.json`

*flag*
: `--config` example `--config=path/to/config.json`


## Log level
How much will be logged on `stdout`

Default is `INFO`. Values are `INFO`,`DEBUG`, `WARN` and `ERROR`.

*env*
: `VINCE_LOG_LEVEL` example `VINCE_LOG_LEVEL=INFO`

*flag*
: `--logLevel` example `--logLevel=INFO`



## Retention Period
How long data will stay in permanent storage. Older data will automatically be deleted and the space reclaimed.

Default is `30` days.


*env*
: `VINCE_RETENTION_PERIOD` example `VINCE_RETENTION_PERIOD=720h0m0s`

*flag*
: `--retentionPeriod` example `--retentionPeriod=720h0m0s`

*file*
: `retentionPeriod` 

```json
{"retentionPeriod":"2592000s"}
```

## Automatic TLS
`vince` supports automatic tls using acme client with let's encrypt. 

### Enabling automatic tls

*env*
: `VINCE_AUTO_TLS` example `VINCE_AUTO_TLS=true`

*flag*
: `--autoTls` example `--autoTls=true`

*file*
: `autoTls` 

```json
{"autoTls":"true"}
```

You need to setup account email address and the domain to generate certificate for, using `acmeEmail` and `acmeDomain` options.

### acmeEmail

This is used by CAs, such as Let's Encrypt, to notify about problems
with issued certificates.

*env*
: `VINCE_ACME_EMAIL` example `VINCE_ACME_EMAIL=example@example.org`

*flag*
: `--acmeEmail` example `--acmeEmail=example@example.org`

### acmeDomain
> `acmeDomain` should be the domain name that is used to point to your server. 
> For example we host vince instance on `cloud.vinceanalytics.com` so we use this as `acmeDomain`

*env*
: `VINCE_ACME_DOMAIN` example `VINCE_ACME_DOMAIN=example.org`

*flag*
: `--acmeDomain` example `--acmeDomain=example.org`

*file*

```json
{"acme":{"email":"example@example.org","domain":"example.org"}}
```


## Authorization
`vince` supports bearer token authorization via `authToken` option. All endpoints except `/api/event` will be protected.

### authToken
When  set, clients calls without this bearer token will be rejected. 

> This is sensitive info use env var to set it.

*env*
: `VINCE_AUTH_TOKEN` example `VINCE_AUTH_TOKEN=xxx`

*flag*
: `--authToken` example `--authToken=xxx`

