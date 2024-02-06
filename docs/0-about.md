# What is vince ?
Vince is a modern server for collecting and analyzing website analytics. `Vince` focuses on modern web application development by emphasizing easy of use for both deployment, maintenance and integration with existing infrastructure.

It ships with a standalone binary with zero dependency called `vince`


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