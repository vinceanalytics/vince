
<p align="center">
    <img src="./assets/logo.svg" alt="Vince Logo" />
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
<summary>Login</summary>

```bash
VINCE_ROOT_PASSWORD=xxxxx vince login http://localhost:8080
```

</details>

<details markdown="1">
<summary>Connect with mysql</summary>

```bash
LIBMYSQL_ENABLE_CLEARTEXT_PLUGIN=y mysql --host 127.0.0.1 --port 3306 -uroot -p$VINCE_ACCESS_TOKEN
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

You can obtain `VINCE_ACCESS_TOKEN` via vince client

```bash
vince login --token
```

</details>
