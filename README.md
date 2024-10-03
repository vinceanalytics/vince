
<p align="center">
    <img src="./logo.svg" alt="Vince Logo" />
    <br>
    <a href="https://vinceanalytics.com/">Website</a> |
    <a href="https://www.vinceanalytics.com/guides/deployment/local/">Install</a>
</p>


## vince

**Vince** is a privacy friendly web analytics server focused on painless self hosting.

![Vince Analytics](desktop.png)



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

### From source

```
go install github.com/vinceanalytics/vince@latest
```

### Download 

[see release page](https://github.com/vinceanalytics/vince/releases)


## Checking installation

```
vince --version
```

## Start vince

```shell
vince serve --adminName=acme \
  --adminPassword=1234\
  --adminEmail=trial@vinceanalytics.com \
  --domains=vinceanalytics.com \
  --license=path/to/license_key
```

This command will start vince on `localhost:8080`.

# FAQ

## How do I bypass license key?

use `vince crack` command to patch license key, you can choose how long you want 
the cracked key to be valid with `--expires`  flag.

```
NAME:
   vince crack - Cracks vince license key

USAGE:
   vince crack [command [command options]] 

DESCRIPTION:
   Allows users to use vince without a valid license key.
       # vince crack /path/to/vince/data

OPTIONS:
   --expires value  Duration of the patched license (default: 24h0m0s)
   --help, -h       show help (default: false)
```

**example:**
Assuming your data directory is `vince-data`

```
❯ vince crack vince-data
VINCE_ADMIN_EMAIL        Expires                              
crack@vinceanalytics.com 2024-10-04 06:52:01.123432 +0000 UTC 
```
After running the command, start vince with f `--adminEmail=crack@vinceanalytics.com`.
You can omit the `--license` flag.

*This only works when running vince with persistence storage*

# Credit

[Plausible Analytics](https://github.com/plausible/analytics) : `vince` started as a Go port of plausible with a focus on self hosting.