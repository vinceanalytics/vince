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

By default `256MB` is used. You need to adjust this depending on expected traffick.

> Unless otherwise `256MB` is a very conservative balance. This option is for power users.
> You can have production deployments sticking with the defaults.

*env*
: `VINCE_GRANULE_SIZE` example `VINCE_GRANULE_SIZE=268435456`

*flag*
: `--granuleSize` example `--granuleSize=268435456`

*file*
: `granuleSize` 

```json
{"granuleSize":"268435456"}
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