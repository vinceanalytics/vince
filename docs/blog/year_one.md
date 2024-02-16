# Year One

###### createdAt 2024-02-16
###### updatedAt 2024-02-16
###### author Geofrey Ernest
###### section Roadmap


This marks the beginning of yearly tradition to reflect on what happened and put
down what I anticipate for the upcoming year.


Until now, I have had many iterations on `vince`. The goal has always been to build
the cheapest, most efficient and feature packed web analytics server.

It took me three years, across three major stack changes to settle down on what we have
now.

- envoy + wasm : I started with web events collection only using envoy proxy with 
a `wasm` plugin written in `zig`. The idea fell apart after I discovered there is no api to clear counters using the proxy wasm. I learned a lot about web events collection and how to hand PI data.

- go + clickhouse : This was a safe bet, inspired by Plausible. It was a direct Go port of Plausible.
It was very promising but I was not happy that I had to mandate users to install `clickhouse`.
Deployments were not painless so I was on the lookout for embedded options.

- go + `duckdb` :Compile time were massive, development slowed down. I was back on the market after a month porting to `duckdb`

- go + `frostdb`: This was pure Go solution. Worked very well,  as I was adding more features I realized I was removing dependency on it. By now I have amassed a trove of knowledge about the shape and nature of data I was collecting and access pattern. `frostdb` became a heavy dependency that I really didn't need anymore.

- go +apache arrow + apache parquet: This is what I have settled on. `vince` is insanely fast and very lobust in terms of resource consumption and throughput.

I will write a lot about how we are going to disrupt this space with a better solution especially for modern deployments on this blog.


### What works ?

- **events ingestion ,storage and querying**. We keep hot data in memory as `arrow.Record` and move it to data based on configurable threshold. On disk we store the record in `parquet`, the record parts are compacted and compressed before storing in cold key/value store saving a lot memory.
Querying is only supported for hot data. Cold data query is on the pipeline

- Plausible like API
 - current visitors `/api/v1/stats/realtime/visitors`
 - aggregates `/api/v1/stats/aggregate`
 - breakdown `/api/v1/stats/breakdown`

### What is not working ?

- Plausible like API
 - timeseries `/api/v1/stats/timeseries` : I am currently working on this so it will be done soon.


### Plan for 2024

- Support cold data querying
- Write more blog posts about `vince` internals
- Add `Consulting` page
- Ship the first public release of `vince`



Thanks for reading this, until next time, have a good day.


