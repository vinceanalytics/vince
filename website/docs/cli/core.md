---
title: cli core
---

# data  


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


# listen <Badge type="danger" text="required" />

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