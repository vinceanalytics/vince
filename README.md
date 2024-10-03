
<p align="center">
    <img src="./logo.svg" alt="Vince Logo" />
    <br>
    <a href="https://vinceanalytics.com/">Website</a> |
    <a href="https://www.vinceanalytics.com/guides/deployment/local/">Install</a>
</p>


## vince

**Vince** is a privacy friendly web analytics server focused on painless self hosting.

![Vince Analytics](desktop.png)


# FAQ

## How do I bypass license key?

use `vince crack` command to patch license key, you can choose how long you want 
the cracked key to be valid with `--expires`  flag.

```
NAME:
   vince crack - Cracks vince license key

USAGE:
   vince crack [command [command options]] 

DESCRIPTION:
   Allows users to use vince without a valid license key.
       # vince crack /path/to/vince/data

OPTIONS:
   --expires value  Duration of the patched license (default: 24h0m0s)
   --help, -h       show help (default: false)
```

**example:**
Assuming your data directory is `vince-data`

```
❯ vince crack vince-data
VINCE_ADMIN_EMAIL        Expires                              
crack@vinceanalytics.com 2024-10-04 06:52:01.123432 +0000 UTC 
```
After running the command, start vince with the output of `VINCE_ADMIN_EMAIL` and 
`--data=vince-data`