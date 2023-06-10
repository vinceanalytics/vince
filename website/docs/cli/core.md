---
title: cli core
---

These are options that are essential for vince operations. They must be set
before starting vince instance. They have sensible defaults to get you up to
speed.


# data  
- Default: `.vince`

Path to a directory where all files needed to operate vince are stored. Can be
configured by commandline flag or environment variable.
::: code-group
```shell [flag]
vince --data='.vince'
```


```shell [env]
export VINCE_DATA=".vince"
```
:::


::: warning
The structure and contents of this directory is not stable and subject to
changes between versions without notice. Don't touch this directory, treat it
like a big black box.
:::

Currently there are  two kind of data  managed by vince
 - *Operational data*: stored in sqlite database named `vince.db` in the root
 of data directory.

::: tip
Don't manually copy this file for backup, read [Backup and
Restore](../guide/backup.md) on how to backup your data and restoring it.
:::
- *Time Series data*: Everything under `ts` subdirectory. 

::: info
There are many directories under `ts` sub directory serving different purposes
 mostly are badger databases). They are explained in detail in
 [Timeseries](../guide/timeseries.md) page.
:::


# listen
- Default: `:8080`

Host:Port address on which to listen for http.
::: code-group
```shell [flag]
vince --listen=':8080'
```

```shell [env]
export VINCE_LISTEN=":8080"
```
:::

::: tip
If you want to serve https, see [TLS](./tls.md)
:::

# log-level
- Default: `debug`

Controls how much information is logged by vince.
::: code-group
```shell [flag]
vince --log-level='trace | debug | info | warn | error | fatal | panic'
```

```shell [env]
export VINCE_LOG_LEVEL="trace | debug | info | warn | error | fatal | panic"
```
:::

# enable-profile
- Default: `false

Expose profile data on `/debug/pprof endpoint` endpoint. Useful when debugging
performance issues.
::: code-group
```shell [flag]
vince --enable-profile
```

```shell [env]
export VINCE_ENABLE_PROFILE="true"
```
:::

# url
- Default: `http://localhost:8080`

::: code-group
```shell [flag]
vince --url='https://vince.example.com'
```

```shell [env]
export VINCE_URL="https://vince.example.com"
```
:::

Resolvable domain name that vince is hosted on. This is used on emails and other features like script links generation, shared link generations.

::: danger
You must update this before production deployments.
:::