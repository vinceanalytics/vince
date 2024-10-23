
## vince

**Vince** is a privacy friendly web analytics server focused on painless self hosting.

![Vince Analytics](desktop.png)


# Features

- [**Automatic TLS**](https://www.vinceanalytics.com/guides/config/auto-tls/) native support for let's encrypt.
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

### Download 

[see release page](https://github.com/vinceanalytics/vince/releases)


## Checking installation

```
vince --version
```

## Start vince

*create admin*
```
❯ vince admin --name acme --password 1234
```

*start server*
```
❯ vince serve                            
2024/10/23 15:32:08 [JOB 1] WAL file vince-data/pebble/000002.log with log number 000002 stopped reading at offset: 124; replayed 1 keys in 1 batches
2024/10/23 15:32:08 INFO starting event processing loop
2024/10/23 15:32:08 INFO starting server addr=:8080
```

# Credit

[Plausible Analytics](https://github.com/plausible/analytics) : `vince` started as a Go port of plausible with a focus on self hosting.