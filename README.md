# vince

The Cloud Native Web Analytics Platform.


# Features

- Alert notification
- Time on site tracking
- Conversion tracking 
- Multiple site management
- User Interaction Management 
- Campaign Management 
- Report Generation
- Goal Tracking 
- Event Tracking 

# Origins

This started as a go port of [Plausible](https://github.com/plausible/analytics), with 
the intention to remove clickhouse and postgresql dependency aiming for a self hosted solution
used by all team sizes.

We use 
- sqlite for operational data (users,sites etc)
- badger for persistance of aggregates
- apache parquet/arrow for timeseries data (system stats)

# Downloads

Files are signed with [minisign](https://jedisct1.github.io/minisign/) using this public key:
```
RWSA5ztaWA/0ny2xL3U6ZQBgmfbECNm7qCPZA1VEWeGCE51WuWkj9Tt4
```


**v8s**
|                                                                                                                                    filename|                                                                                                               signature|size|
|                                                                                                                                        ----|                                                                                                                    ----|----|
|                             [v8s_windows_arm64.zip](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_windows_arm64.zip)|               [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_windows_arm64.zip.minisig)|`9mb`|
|                           [v8s_windows_x86_64.zip](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_windows_x86_64.zip)|              [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_windows_x86_64.zip.minisig)|`10mb`|
|                       [v8s_darwin_x86_64.tar.gz](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_darwin_x86_64.tar.gz)|            [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_darwin_x86_64.tar.gz.minisig)|`11mb`|
|                         [v8s_darwin_arm64.tar.gz](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_darwin_arm64.tar.gz)|             [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_darwin_arm64.tar.gz.minisig)|`10mb`|
|                           [v8s_linux_arm64.tar.gz](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_linux_arm64.tar.gz)|              [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_linux_arm64.tar.gz.minisig)|`9mb`|
|                         [v8s_linux_x86_64.tar.gz](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_linux_x86_64.tar.gz)|             [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_linux_x86_64.tar.gz.minisig)|`10mb`|
|                 [v8s_v0.0.0_linux_x86_64.deb](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_x86_64.deb)|         [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_x86_64.deb.minisig)|`11mb`|
|                   [v8s_v0.0.0_linux_arm64.deb](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_arm64.deb)|          [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_arm64.deb.minisig)|`10mb`|
|                   [v8s_v0.0.0_linux_arm64.apk](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_arm64.apk)|          [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_arm64.apk.minisig)|`10mb`|
|                 [v8s_v0.0.0_linux_x86_64.apk](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_x86_64.apk)|         [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_x86_64.apk.minisig)|`11mb`|
|   [v8s_v0.0.0_linux_arm64.pkg.tar.zst](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_arm64.pkg.tar.zst)|  [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_arm64.pkg.tar.zst.minisig)|`8mb`|
|                 [v8s_v0.0.0_linux_x86_64.rpm](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_x86_64.rpm)|         [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_x86_64.rpm.minisig)|`11mb`|
|                   [v8s_v0.0.0_linux_arm64.rpm](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_arm64.rpm)|          [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_arm64.rpm.minisig)|`10mb`|
| [v8s_v0.0.0_linux_x86_64.pkg.tar.zst](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_x86_64.pkg.tar.zst)| [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/v8s_v0.0.0_linux_x86_64.pkg.tar.zst.minisig)|`9mb`|


**vince**
|                                                                                                                                        filename|                                                                                                                 signature|size|
|                                                                                                                                            ----|                                                                                                                      ----|----|
|                             [vince_windows_arm64.zip](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_windows_arm64.zip)|               [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_windows_arm64.zip.minisig)|`60mb`|
|                           [vince_windows_x86_64.zip](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_windows_x86_64.zip)|              [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_windows_x86_64.zip.minisig)|`61mb`|
|                           [vince_linux_arm64.tar.gz](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_linux_arm64.tar.gz)|              [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_linux_arm64.tar.gz.minisig)|`60mb`|
|                         [vince_darwin_arm64.tar.gz](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_darwin_arm64.tar.gz)|             [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_darwin_arm64.tar.gz.minisig)|`61mb`|
|                         [vince_linux_x86_64.tar.gz](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_linux_x86_64.tar.gz)|             [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_linux_x86_64.tar.gz.minisig)|`60mb`|
|                       [vince_darwin_x86_64.tar.gz](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_darwin_x86_64.tar.gz)|            [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_darwin_x86_64.tar.gz.minisig)|`61mb`|
|                   [vince_v0.0.0_linux_arm64.deb](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_arm64.deb)|          [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_arm64.deb.minisig)|`61mb`|
|                 [vince_v0.0.0_linux_x86_64.deb](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_x86_64.deb)|         [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_x86_64.deb.minisig)|`61mb`|
|                   [vince_v0.0.0_linux_arm64.apk](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_arm64.apk)|          [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_arm64.apk.minisig)|`61mb`|
|                 [vince_v0.0.0_linux_x86_64.apk](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_x86_64.apk)|         [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_x86_64.apk.minisig)|`61mb`|
| [vince_v0.0.0_linux_x86_64.pkg.tar.zst](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_x86_64.pkg.tar.zst)| [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_x86_64.pkg.tar.zst.minisig)|`61mb`|
|   [vince_v0.0.0_linux_arm64.pkg.tar.zst](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_arm64.pkg.tar.zst)|  [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_arm64.pkg.tar.zst.minisig)|`60mb`|
|                   [vince_v0.0.0_linux_arm64.rpm](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_arm64.rpm)|          [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_arm64.rpm.minisig)|`61mb`|
|                 [vince_v0.0.0_linux_x86_64.rpm](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_x86_64.rpm)|         [minisig](https://github.com/vinceanalytics/vince/releases/download/v0.0.0/vince_v0.0.0_linux_x86_64.rpm.minisig)|`61mb`|



## Container Image
**v8s**
```
ghcr.io/vinceanalytics/v8s:v0.0.0
```

**vince**
```
ghcr.io/vinceanalytics/vince:v0.0.0
```


## Brew


**v8s**
```
brew install vinceanalytics/tap/v8s
```

**vince**
```
brew install vinceanalytics/tap/vince
```




# The name vince 

Vince is named after [Vince Staples](https://en.wikipedia.org/wiki/Vince_Staples) , 
my favorite hip hop artist.

It was late night, 1 year after I quit my job and took a sabbatical, I was listening
to one of his album [Big Fish Theory](https://en.wikipedia.org/wiki/Big_Fish_Theory)
. The song Big Fish inspired me to program again.

The lirics that got me going.
```
I was up late night ballin'
Countin' up hundreds by the thousand
```

So, enjoy the outcome of my late nights, hoping to be counting hundreds by the thousand
soon, or in my next life.