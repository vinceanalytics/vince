
<p align="center">
  <picture width="640" height="250">
    <source media="(prefers-color-scheme: dark)" srcset="./app/images/logo-darkmode.svg">
    <source media="(prefers-color-scheme: light)" srcset="./app/images/mark.svg">
    <img alt="esbuild: An extremely fast JavaScript bundler" src="logo.svg">
  </picture>
  <br>
  <a href="https://vinceanalytics.com/">Website</a> |
  <a href="https://vinceanalytics.com/blog/deploy-local/">Getting started</a> |
  <a href="https://vinceanalytics.com/tags/api/">API</a>  |
  <a href="https://demo.vinceanalytics.com/v1/share/vinceanalytics.com?auth=Ls9tV4pzqOn7BJ7-&demo=true">Demo</a> 
</p>

**Vince** is a self hosted alternative to Google Analytics.


# Features

- **Automatic TLS** native support for let's encrypt.
- **Drop in replacement for plausible** you can use existing plausible  scripts and just point them to the vince instance (note that vince is lean and only covers features for a single entity self hosting, so it is not our goal to be feature parity with plausible).
- **Outbounds links tracking**
- **File download tracking**
- **404 page tracking**
- **Custom event tracking**
- **Time period comparison**
- **Public dashboards** allow access to the dashoard to anyone(by default all dashboards are private).
- **Unique shared access** generate unique links to dahboards that can be password protected.
- **Zero Dependency**: Ships a single binary with everything in it. No runtime dependency.
- **Easy to operate**: One line commandline flags with env variables is all you need.
- **Unlimited sites**: There is no limit on how many sites you can manage.
- **Unlimited events**: scale according to availbale resources.
- **Privacy friendly**: No cookies and fully compliant with GDPR, CCPA and PECR.


# Installation

Vince ships a single executable without any dependencies.


## Installing

### MacOS and Linux

```
curl -fsSL https://vinceanalytics.com/install.sh | bash
```

### Docker

```
docker pull ghcr.io/vinceanalytics/vince
```

### Helm
```
❯ helm repo add vince http://vinceanalytics.com/charts
❯ helm install vince vince/vince
```

### Download 

[see release page](https://github.com/vinceanalytics/vince/releases)


## Checking installation

```
vince --version
```

## Start vince

***create admin***
```
❯ vince admin --name acme --password 1234
```

***start server***
```
❯ vince serve                            
2024/10/23 15:32:08 [JOB 1] WAL file vince-data/pebble/000002.log with log number 000002 stopped reading at offset: 124; replayed 1 keys in 1 batches
2024/10/23 15:32:08 INFO starting event processing loop
2024/10/23 15:32:08 INFO starting server addr=:8080
```


# Comparison with Plausible Analytics

| feature |  vince | plausible |
|---------|--------| -----------|
| Enterprise features | :x:    | :white_check_mark:|
| Hosted offering | :x:    | :white_check_mark:|
| Multi tenant | :x:    | :white_check_mark:|
| Funnels | :x:    | :white_check_mark:|
| Goals Conversion |  :white_check_mark:  | :white_check_mark:|
| Unique visitors |  :white_check_mark:  | :white_check_mark:|
| Total visits |  :white_check_mark:  | :white_check_mark:|
| Page views |  :white_check_mark:  | :white_check_mark:|
| Views per visit |  :white_check_mark:  | :white_check_mark:|
| Visit duration |  :white_check_mark:  | :white_check_mark:|
| Breakdown by **Cities**, **Sources**, **Pages** and **Devices**   |  :white_check_mark:  | :white_check_mark:|
| Self Hosted |  :white_check_mark:  | :white_check_mark:|
| <1KB script |  :white_check_mark:  | :white_check_mark:|
| No Cookies(GDPR, PECR compliant) |  :white_check_mark:  | :white_check_mark:|
| 100% data ownershiip |  :white_check_mark:  | :white_check_mark:|
| Unique shared access to stats|  :white_check_mark:  | :white_check_mark:|
| Outbound links tracking |  :white_check_mark:  | :white_check_mark:|
| File download tracking |  :white_check_mark:  | :white_check_mark:|
| 404 page tracking |  :white_check_mark:  | :white_check_mark:|
| Time period comparisons |  :white_check_mark:  | :white_check_mark:|
| Unlimited sites |  :white_check_mark:  | :x:|
| Unlimited events |  :white_check_mark:  | :x:|
| Zero dependency |  :white_check_mark:  | :x: (needs elixir, clickhouse, postgresql ...etc)|
| Automatic TLS |  :white_check_mark:  | :x:|


# Credit

[Plausible Analytics](https://github.com/plausible/analytics) : `vince` started as a Go port of plausible with a focus on self hosting.
