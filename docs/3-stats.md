

# Stats API

> **CC-BY-S-4.0** *This section was initially copied from Plausible Analytics docs*

The`vince` API offers a way to retrieve your stats programmatically. It's a read-only interface to present historical and real-time stats only. Take a look at our [events API](#events-api) if you want to send pageviews or custom events.

The API accepts GET requests with query parameters and returns standard HTTP responses along with a JSON-encoded body. All API requests must be made over HTTPS. Calls made over plain HTTP will fail. API requests without authentication will also fail.

Each request must be authenticated with an [authToken](#authorization) using the Bearer Token method.


## Concepts

Querying the `vince` API will feel familiar if you have used time-series databases before. You can't query individual records from
our stats database. You can only request aggregated metrics over a certain time period.

Each request requires a `site_id` parameter which is the domain of your site as configured in [domains](#domains). 

### Metrics

You can specify a `metrics` option in the query, to choose the metrics for each instance returned. See here for a full overview of [metrics and their definitions](#all-metrics). The metrics currently supported in Stats API are:

| Metric            | Description                                                                                                                                               |
|-------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| `visitors`        | The number of unique visitors.                                                                                                                            |
| `visits`          | The number of visits/sessions                                                                                                                             |
| `pageviews`       | The number of pageview events                                                                                                                             |
| `views_per_visit` | The number of pageviews divided by the number of visits. Returns a floating point number. currently only supported in Aggregate and Timeseries endpoints. |
| `bounce_rate`     | Bounce rate percentage                                                                                                                                    |
| `visit_duration`  | Visit duration in seconds                                                                                                                                 |
| `events`          | The number of events (pageviews + custom events)                                                                                                          |

## Time periods

The options are identical for each endpoint that supports configurable time periods. Each period is relative to a `date` parameter. The date should follow the standard ISO-8601 format. When not specified, the `date` field defaults to `today(site.timezone)`.
All time calculations on our backend are done in the time zone that the site is configured in.

* `12mo,6mo` - Last n calendar months relative to `date`.
* `month` - The calendar month that `date` falls into.
* `30d,7d` - Last n days relative to `date`.
* `day` - Stats for the full day specified in `date`.
* `custom` - Provide a custom range in the `date` parameter.

When using a custom range, the `date` parameter expects two ISO-8601 formatted dates joined with a comma as follows `?period=custom&date=2021-01-01,2021-01-31`.
Stats will be returned for the whole date range inclusive of the start and end dates.

## Properties

Each pageview and custom event in our database has some predefined _properties_ associated with it. In other analytics tools, these
are often referred to as _dimensions_ as well. Properties can be used for filtering and breaking down your stats to drill into
more depth. Here's the full list of properties we collect automatically:

| Property              | Example                       | Description                                                                                                                             |
| --------------------- | ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `page`            | /blog/remove-google-analytics | Pathname of the page where the event is triggered. You can also use an asterisk to group multiple pages (`/blog*`)  |
| `entry_page`      | /home                         | Page on which the visit session started (landing page).                                                                                 |
| `exit_page`       | /home                         | Page on which the visit session ended (last page viewed).                                                                               |
| `source`          | Twitter                       | Visit source, populated from an url query parameter tag (`utm_source`, `source` or `ref`) or the `Referer` HTTP header.                 |
| `referrer`        | t.co/fzWTE9OTPt               | Raw `Referer` header without `http://`, `http://` or `www.`.                                                                            |
| `utm_medium`      | social                        | Raw value of the `utm_medium` query param on the entry page.                                                                            |
| `utm_source`      | twitter                       | Raw value of the `utm_source` query param on the entry page.                                                                            |
| `utm_campaign`    | profile                       | Raw value of the `utm_campaign` query param on the entry page.                                                                          |
| `utm_content`     | banner                        | Raw value of the `utm_content` query param on the entry page.                                                                           |
| `utm_term`        | keyword                       | Raw value of the `utm_term` query param on the entry page.                                                                              |
| `device`          | Desktop                       | Device type. Possible values are `Desktop`, `Laptop`, `Tablet` and `Mobile`.                                                            |
| `browser`         | Chrome                        | Name of the browser vendor. Most popular ones are `Chrome`, `Safari` and `Firefox`.                                                     |
| `browser_version` | 88.0.4324.146                 | Version number of the browser used by the visitor.                                                                                      |
| os              | Mac                           | Name of the operating system. Most popular ones are `Mac`, `Windows`, `iOS` and `Android`. Linux distributions are reported separately. |
| `os_version`      | 10.6                          | Version number of the operating system used by the visitor.                                                                             |
| `country`         | United Kingdom                 | Country of the visitor country.                                                                                         |
| `region`          | England                         | Region the visitor region.                                                                                                  |
| `city`            | London                      | City of the visitor city.                                                                            |

### Custom properties

TBD

## Filtering

Most endpoints support a `filters` query parameter to drill down into your data. You can filter by all properties described in the [Properties table](#properties), using the following operators:

| Operator        | Usage example                              | Explanation                                                               |
|-----------------|--------------------------------------------|---------------------------------------------------------------------------|
| `==`            | `name==Signup`                       | Simple equality - custom event "Signup"                                 |
| `!=`            | `country!=Tanzania`                        | Simple inequality - country is not Tanzania                                 |
| `~=`             | `page~=^/blog/.*?`                      | Regex - matches a regular expression
| `!~`             | `page!~^/blog/.*?`                      | Regex - matches not  a regular expression



## Endpoints

### GET /api/v1/stats/realtime/visitors

This endpoint returns the number of current visitors on your site. A current visitor is defined as a visitor who triggered a pageview on your site
in the last 5 minutes.

```bash
+ curl -X GET 'http://localhost:8080/api/v1/stats/realtime/visitors?site_id=vinceanalytics.com'
6
```



#### Parameters
<hr/>

**site_id** <span class="required">REQUIRED<span/>

Domain of your site on `vince`.
<hr / >

### GET /api/v1/stats/aggregate

This endpoint aggregates metrics over a certain time period.It include `Unique Visitors`` Pageviews`, `Bounce rate` and `Visit duration`. You can retrieve any number and combination of these metrics in one request.


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


#### Parameters
<hr / >

**site_id** <span class="required">REQUIRED<span/>

Domain of your site on `vince`.

<hr / >

**period** <span class="optional">optional<span/>

See [time periods](#time-periods). If not specified, it will default to `30d`.

<hr / >

**metrics** <span class="optional">optional<span/>

List of metrics to aggregate. Valid options are `visitors`, `visits`, `pageviews`, `views_per_visit`, `bounce_rate`, `visit_duration` and `events`. If not specified, it will default to `visitors`.

<hr / >

**filters** <span class="optional">optional<span/>

See [filtering](#filtering)
<hr / >

### GET /api/v1/stats/timeseries

This endpoint provides timeseries data over a certain time period.

```bash
+ curl -X GET 'http://localhost:8080/api/v1/stats/timeseries?period=6mo&site_id=vinceanalytics.com'
{
  "results": [
    {
      "timestamp": "2024-02-19T00:00:00Z",
      "values": [
        {
          "metric": "visitors",
          "value": 8
        }
      ]
    }
  ]
}
```


#### Parameters
<hr / >

**site_id** <span class="required">REQUIRED<span/>

Domain of your site on `vince`.

<hr / >

**period** <span class="optional">optional<span/>

See [time periods](#time-periods). If not specified, it will default to `30d`.

<hr / >

**filters** <span class="optional">optional<span/>

See [filtering](#filtering)

<hr / >

**metrics** <span class="optional">optional<span/>

Comma-separated list of metrics to show for each time bucket. Valid options are `visitors`, `visits`, `pageviews`, `views_per_visit`, `bounce_rate`, `visit_duration` and `events`. If not
specified, it will default to `visitors`.

<hr / >

**interval** <span class="optional">optional<span/>

Choose your reporting interval. Valid options are `date` (always) and `month` (when specified period is longer than one calendar month). Defaults to
`month` for `6mo` and `12mo`, otherwise falls back to `date`.

### GET /api/v1/stats/breakdown

This endpoint allows you to break down your stats by some property. If you are familiar with SQL family databases, this endpoint corresponds to
running `GROUP BY` on a certain property in your stats, then ordering by the count.

Check out the [properties](#properties) section for a reference of all the properties you can use in this query.

This endpoint can be used to fetch data for `Top sources`, `Top pages`, `Top countries` and similar reports.



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


#### Parameters
<hr / >

**site_id** <span class="required">REQUIRED<span/>

Domain of your site on `vince`.

<hr / >

**property** <span class="required">REQUIRED<span/>

Which property to break down the stats by. Valid options are listed in the [properties](#properties) section above.

<hr / >

**period** <span class="optional">optional<span/>

See [time periods](#time-periods). If not specified, it will default to `30d`.

<hr / >

**metrics** <span class="optional">optional<span/>

Comma-separated list of metrics to show for each item in breakdown. Valid options are `visitors`, `pageviews`, `bounce_rate`, `visit_duration`, `visits` and `events`. If not
specified, it will default to `visitors`.

<hr / >

**limit** <span class="optional">optional<span/>

Limit the number of results. Maximum value is 1000. Defaults to 100. If you want to get more than 1000 results, you can make multiple requests and paginate the results by specifying the `page` parameter (e.g. make the same request with `page=1`, then `page=2`, etc)

<hr / >
**page** <span class="optional">optional<span/>

Number of the page, used to paginate results. Importantly, the page numbers start from 1 not 0.

<hr / >
**filters** <span class="optional">optional<span/>

See [filtering](#filtering)

