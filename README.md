
<p align="center">
    <img src="./logo.svg" alt="Vince Logo" />
    <br>
    <a href="https://vinceanalytics.com/">Website</a> |
    <a href="https://vinceanalytics.com/#getting-started">Getting started</a> |
    <a href="https://vinceanalytics.com/#stats-api">API</a>
</p>


## What ?

`vince` is a highly optimized ,privacy friendly modern server for collecting and analyzing website analytics.

## Why ?

- You want to self host(own your data) painlessly
- You want to save money (use very little resources) 
- You only need to consume the API to access the stats
- You manage large number of sites

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


### Getting current visitors


```bash
+ curl -X GET 'http://localhost:8080/api/v1/stats/realtime/visitors?site_id=vinceanalytics.com'
6
```

### Getting aggregate metrics


```bash
+ curl -X GET 'http://localhost:8080/api/v1/stats/aggregate?metrics=visitors%2Cvisits%2Cpageviews%2Cviews_per_visit%2Cbounce_rate%2Cvisit_duration%2Cevents&site_id=vinceanalytics.com'
{
  "results": {
    "bounce_rate": {
      "value": 0.8888888888888888
    },
    "events": {
      "value": 10
    },
    "pageviews": {
      "value": 10
    },
    "views_per_visit": {
      "value": 1.1111111111111112
    },
    "visit_duration": {
      "value": 0.0013333333333333333
    },
    "visitors": {
      "value": 8
    },
    "visits": {
      "value": 9
    }
  }
}
```


### Breaking down stats by property

```bash
+ curl -X GET 'http://localhost:8080/api/v1/stats/breakdown?metrics=visitors%2Cbounce_rate&property=browser&site_id=vinceanalytics.com'
{
  "results": {
    "browser": {
      "Chrome Mobile": {
        "bounce_rate": 0.6666666666666666,
        "visitors": 6
      },
      "Chrome Webview": {
        "bounce_rate": 1,
        "visitors": 1
      }
    }
  }
}
```

Check out the [getting started](https://vinceanalytics.com/#getting-started) instructions if you want to give `vince` a try.