
<p align="center">
    <img src="./logo.svg" alt="Vince Logo" />
    <br>
    <a href="https://vinceanalytics.com/">Website</a> |
    <a href="https://www.vinceanalytics.com/guides/deployment/local/">Install</a>
</p>


## vince

**Vince** is a privacy friendly web analytics server focused on painless self hosting.

![Vince Analytics](desktop.png)


# Features

- **Drop in replacement for plausible** you can use existing plausible  scripts and just point them to the vince instance (note that vince is lean and only covers features for a single entity self hosting, so it is not our goal to be feature parity with plausible).
- **Outbounds links tracking**
- **File download tracking**
- **404 page tracking**
- **Custom event tracking**
- **Time period comparison**
- **Public dashboards** allow access to the dashoard to anyone(by default all dashboards are private).
- **unique shared access** generate unique links to dahboards that can be password protected.
- **Extremely fast** we use bit sliced index of roaring bitmaps.We base our implementation on [serialized roaring bitmaps](https://github.com/dgraph-io/sroar), there is no decoding while reading, data is loaded directly.
- **Zero Dependency**: Ships a single binary with everything in it. No runtime dependency.
- **High events ingestion rate** : We buffer and use LSM based underlying [key value store](https://github.com/dgraph-io/badger) that is optimized for writes.
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
Then start vince with flag `--adminEmail=crack@vinceanalytics.com`.
You can omit the `--license` flag.

*This only works when running vince with persistence storage*

## Where is the REST API ?
Initially we were api only, we had to pivot to the dashboard version due to lack
of interests. There is no plan to bring it back anytime soon.

## Why all this license key stuff?
I am s solo dev, my last employment ended 2021. I am struggling finding remote work 
because remote now seems to imply Remote US or Remote EU. I implemented the license
stuff to avoid feeling depressed.

vinceanalytics is not a company and I am not selling license keys. Just send me an email if
you want a lifetime license. 

I am ashamed to admit I fumbled the bag, by quitting my last job without a contigency
plan( I had good reason to quit after a burn out). So my linkedin profile showing
I have been working at vince analytics is half true.

It has been three years without work, it is dreadful seeing my career coming to an end.

So, be gentle. I accept discussions, but I am in no position to fix issues. I welcome
pull requests.


# Credit

[Plausible Analytics](https://github.com/plausible/analytics) : `vince` started as a Go port of plausible with a focus on self hosting.