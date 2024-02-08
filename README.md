
<p align="center">
    <img src="./logo.svg" alt="Vince Logo" />
    <br>
    <a href="https://vinceanalytics.com/">Website</a> |
    <a href="https://vinceanalytics.com/#getting-started">Getting started</a> |
    <a href="https://vinceanalytics.com/#api">API</a>
</p>


## What ?

`vince` is a modern server for collecting and analyzing website analytics. `vince` focuses on modern web application development by emphasizing easy of use for both deployment, maintenance and integration with existing infrastructure.

## Why ?

All existing solutions are hard to self host and license is `AGPL`. Also they emphasis more on visual and a lot of corners are cut in deriving the numbers.

I need my web analytics analysis served raw and cold. So there is only storing and analysis nothing more. And the license is Apache 2.0 so self host as you please. 

- There is no UI . Use http api to query for stats.
- Extremely fast and the api is composable, you can integrate the results anywhere `http` works.


## Features

- **Extremely fast** relative to competitors. Uses apache `arrow` for fast vectorized in memory computation. It is designed from grounds up, and highly optimized for web analytics use case.

- **Zero Dependency**: Ships a single binary with everything in it. No runtime dependency.

- **High events ingestion rate** : Non blocking ingestion, you can deploy for very popular sites without worrying.

- **Fast query api** : Instant results for active and historical data.

- **Easy to operate**: One line commandline flags with env variables is all you need.

- **Works with any language and tooling**: No need for special sdk, a simple `http` `api` is exposed. Anything that can speak `http` can work with `vince`

- **10X more data storage** : We use columnar storage with extensive compression schemes. Don't worry about running out of disk. Store and query large volume of data.

- **Unlimited sites**: There is no limit on how many sites you can manage.

- **Privacy friendly**: No cookies and fully compliant with GDPR, CCPA and PECR.


### example checking for current visitors

```bash
curl -X 'GET' 'http://localhost:8080/api/v1/visitors/example.com'
{"visitors":"4"}
```


Check out the [getting started](https://vinceanalytics.com/#getting-started) instructions if you want to give `vince` a try.