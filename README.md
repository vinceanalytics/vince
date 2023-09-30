
<p align="center">
    <img src="./assets/ui/logo.svg" alt="Vince Logo" />
</p>

> :warning: **This is still under early development its not in a usable state yet**. 

# vince

The Cloud Native Web Analytics Platform. Built on Apache Arrow and Apache Parquet.

# Features

- [x] SQL for querying stats (All MySQL compatible clients are supported)
- [x] Time on site tracking
- [x] Conversion tracking 
- [x] Multiple site management
- [ ] User Interaction Management 
- [x] Campaign Management 
- [ ] Report Generation
- [x] Goal Tracking 
- [x] Event Tracking 
- [ ] Cloud Native (seamless k8s integration)
- [x] API for sites management
- [x] No runtime dependency (Static binary with everything you need)

## Usage

<details markdown="1">
<summary>Install</summary>

```bash
curl -fsSL https://github.com/vinceanalytics/vince/releases/latest/download/install.sh | bash
```

```bash
brew install vinceanalytics/tap/vince
```

```bash
docker pull ghcr.io/vinceanalytics/vince
```
</details>

<details markdown="1">
<summary>Initialize a project</summary>

```bash
VINCE_ROOT_PASSWORD=xxxxx vince init example
```

</details>

<details markdown="1">
<summary>Serve analytics collection api and web console</summary>

__Start server__
```bash
vince serve example
```

The script for embedding will be served under `localhost:8080/js/vince.js`.
Web analytics events are collected on `localhost:8080/api/events` endpoint.

</details>

<details markdown="1">
<summary>Connect with mysql</summary>

```bash
LIBMYSQL_ENABLE_CLEARTEXT_PLUGIN=y mysql --host 127.0.0.1 --port 3306 -uroot -p$VINCE_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 2
Server version: 5.7.9-Vitess Dolt

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> 
```

</details>

## Contributing

<details markdown="1">
<summary>UI</summary>

Console ui lives in `./ui/` directory. We are using react with `@primer/react` 
components. For the code editor we use `Monaco`

__requirement__:
- Latest node version
- Latest go version `go1.21+`

You need go version because we embed the generated app, so only way to test it/develop is to run embedding step then access your work through `vince serve`

When you are done making changes

```bash
go generate 
```

Will take care of building/embedding and start the server for you. Note that
for this all to work., you must create development project in `.vince` directory.

Basically steps to getting started

- Clone and cd into vince root
- Install and setup  `go1.21+`
- Install latest node version
- `go install`
- `VINCE_ROOT_PASSWORD=xxxxx vince init .vince`

Then you can work on files in `./ui/` when done.

```bash
go generate
```

You can now access the ui with your changes on `localhost:8080`

</details>

<details markdown="1">
<summary>Backend</summary>

You only need the latest Go version `go1.21+`

We recommend using `go install` when developing.

</details>