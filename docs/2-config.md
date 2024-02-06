# Configuration

`vince` has three way of passing configuration, commandline flags, environment
variables and configuration file.

All three ways can be combined to form a secure  deployments. The level of precedence follows 
 cli `->` env `->` file. So if lets say `listen` is provided by all ways, then
the value set in `file` will be used.

> We recommend using *commandline flags and environment variables*.
> Anything that can be expressed in configuration file can also be expressed 
> with commandline flags and environment variables.


## Data
Path to directory where `vince` will store persisting data. This option is required for `vince` to be operational.

*env*
: `VINCE_DATA` example `VINCE_DATA=path/to/storage`

*flag*
: `--data` example `--data=path/to/storage`

*file*
: `data` 

```json
{"data":"path/to/storage"}
```