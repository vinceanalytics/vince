# Events API



The Vince Analytics  Events API can be used to record pageviews and custom events. This is useful when tracking Android or iOS mobile apps, or for server side tracking.

In most cases we recommend installing Vince through provided  [script](#addscript-to-your-website) 

## Unique visitor tracking

> Special care should be taken with two key headers which are used for unique visitor counting
>
>1. The _User-Agent_ header
>2. The _X-Forwarded-For_ header

If these headers are not sent exactly as required, unique visitor counting will not work as intended. Please refer to the [Request headers](#request-headers) section below for more in-depth documentation on each header separately.


## Endpoints
### POST /api/event

Records a pageview or custom event. When using this endpoint, it's crucial to send the HTTP headers correctly, since these are used for unique user counting.

```bash 
curl -i -X POST http://localhost:8080/api/event \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36 OPR/71.0.3770.284' \
  -H 'X-Forwarded-For: 127.0.0.1' \
  -H 'Content-Type: application/json' \
  --data '{"name":"pageview","url":"http://example.com","domain":"example.com"}'
```



### Parameters

<hr / >
**domain** <span class="required">REQUIRED<span/>

Domain name of the site in Vince

<hr / >

**name** <span class="required">REQUIRED<span/>

Name of the event. Can specify `pageview` which is a special type of event in Vince. All other names will be treated as custom events.

<hr / >

**url** <span class="required">REQUIRED<span/>

URL of the page where the event was triggered. If the URL contains UTM parameters, they will be extracted and stored. When using the script, this is set to `window.location.href`.

The maximum size of the URL, excluding the domain and the query string, is 2,000 characters. Additionally, URLs using the [data URI scheme](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URLs) are not supported by the API.

<hr / >

**referrer** <span class="optional">OPTIONAL<span/>

Referrer for this event. When using the standard tracker script, this is set to `document.referrer`

Referrer values are processed heavily for better usability. Consider referrer
URLS like `m.facebook.com/some-path` and `facebook.com/some-other-path`. It's intuitive to think of both of these as coming from a single source: Facebook. In the first example the `referrer` value would be split into `source == Facebook` and `referrer == m.facebook.com/some-path`.

Vince uses the open source [referer-parser](https://github.com/snowplow-referer-parser/referer-parser) database to parse referrers and assign these source categories.

## Request headers


**User-Agent** <span class="required">REQUIRED<span/>

The raw value of User-Agent is used to calculate the *user_id* which identifies a unique visitor
in Vince.

User-Agent is also used to populate the _Devices_ properties in vince. The device data is derived from the open source database [device-detector](https://github.com/matomo-org/device-detector). If your User-Agent is not showing up in your dashboard, it's probably because it is not recognized as one in the _device-detector_ database.

The header is required but bear in mind that browsers and some HTTP libraries automatically add a default User-Agent header to HTTP requests. In case of browsers, we would not recommend overriding the header manually unless you have a specific reason to.

<hr/>
**X-Forwarded-For** <span class="optional">optional<span/>


Used to explicitly set the IP address of the client. If not set, the remote IP of the sender will automatically be used. Depending on your use-case:
1. If sending the event from your visitors' device, this header does not need to be set
2. If sending the event from a backend server or proxy, make sure to override this header with the correct IP address of the client.

The raw value of the IP address is not stored in our database. The IP address is used to calculate the *user_id* which identifies a unique visitor in Vince. It is also used to fill the Location properties with country, region and city data of the visitor.

If the header contains a comma-separated list (as it should if the request is sent through a chain of proxies), then the first valid IP address from the list is used. Both IPv4 and IPv6 addresses are supported. More information about the header format can be found on [MDN docs](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For).

<hr/>

**Content-Type** <span class="required">REQUIRED<span/>

Must be either *application/json* or *text/plain*. In case of *text/plain*, the request body is still interpreted as JSON.
