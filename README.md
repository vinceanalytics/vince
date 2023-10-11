
<p align="center">
    <img src="./assets/logo.svg" alt="Vince Logo" />
</p>


> **Warning**
> This is still under early development its not in a usable state yet

# vince

The Cloud Native Web Analytics Platform. Built on Apache Arrow and Apache Parquet.

> **note**
> Vince does not support realtime queries. Events are processed daily at configured time of the day.
> There is a possibility you will have to wait 24h to be able to get actionable insight from your site
> when you configure it for the first time.

# Features

- :white_check_mark: SQL for querying stats (All MySQL compatible clients are supported)
- :white_check_mark: Time on site tracking
- :white_check_mark: Conversion tracking 
- :white_check_mark: Multiple site management
- :white_check_mark: Campaign Management 
- :x: Report Generation
- :white_check_mark: Goal Tracking 
- :white_check_mark: Event Tracking 
- :x: Cloud Native (seamless k8s integration)
- :white_check_mark: API for sites management
- :white_check_mark: No runtime dependency (Static binary with everything you need)

## Usage

Throughout this guide we will be using `http://localhost:8080` to refer to the url where you self hosted vince instance. We expect the url to change to the internet accessible url where you self hosted your vince instance.`example.com` is used to represent your website that you wish to track.

<details markdown="1">
<summary>Install</summary>

vince provides a single binary `vince` that provides both the server and client
functionality. For now only `Mac OS` and `Linux` are supported. 
</details>


```bash
curl -fsSL https://github.com/vinceanalytics/vince/releases/latest/download/install.sh | bash
```

```bash
brew install vinceanalytics/tap/vince
```

```bash
docker pull ghcr.io/vinceanalytics/vince
```

<details markdown="1">
<summary>Initialize a project</summary>

`vince init` sets up a directory for serving vince instance. This includes creating directories for databases and generating of configurations. You can later edit generated configuration file to reflect what you need.

```bash
NAME:
   vince init - Initializes a vince project

USAGE:
   vince init [command [command options]] [arguments...]

OPTIONS:
   -i                Shows interactive prompt for username and password (default: false) || --no-i  Shows interactive prompt for username and password (default: false)
   --username value  Name of the root user (default: "root") [$VINCE_ROOT_USER]
   --password value  password of the root user (default: "vince") [$VINCE_ROOT_PASSWORD]
   --help, -h        show help (default: false)
```

Vince instances are password protected. Access to resources is provided via JWT tokens served using the builtin oauth2 server.

</details>

```bash
VINCE_ROOT_PASSWORD=xxxxx vince init example
```

<details markdown="1">
<summary>Start  server</summary>
Vince binds to two ports, one for vince api and another for mysql api.
</details>

```bash
vince serve example
```

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

<details markdown=1>
<summary>Add site</summary>

You can add a website to allow collection of web analytics using mysql api with
the procedure `add_site` which accepts the domain name of the site as the first argument and optionally site description as a second argument.

Site domain, is the part of the website url without `http://` or `http://` or
`wwww`. Example domain for `https://www.vinceanalytics.com` is `vinceanalytics.com`

There is no limit on the number of sites that can be added. Also you can setup and sent events for sites that have not been added (the events will just be dropped).

Please see `Embedding js tracker` section on how to setup tracker script on your website to start collecting and send web analytics to your vince instance.
</details>

```shell
mysql> call add_site('example.com');
+--------+
| status |
+--------+
| ok     |
+--------+
1 row in set (0.00 sec)
```


<details markdown="1">
<summary>Embedding js tracker</summary>
Vince instance hosts and serve the javascript tracker that you can embed in
your website.

Update `html` of your website to include the script in the `head` tag of your
`html`
</details>

```html
<script defer data-domain="example.com" src="http://localhost:8080/js/vince.js"></script>
```
