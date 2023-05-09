---
title: Installing vince analytics
---

Vince is distributed in two flavors `vince` which is the application with all
the features inteded for self hosting on VPS and `v8s` which is the vince controller
for kuberenets.

`v8s` automates deployment and management of `vince` instances on `k8s` using custom
resources.


## Brew

::: code-group

```shell[vince]
brew install vinceanalytics/tap/vince
```

```shell[v8s]
brew install vinceanalytics/tap/v8s
```

:::


## Container Image
::: code-group

```shell[vince]
ghcr.io/vinceanalytics/vince:v0.0.0
```

```shell[v8s]
ghcr.io/vinceanalytics/v8s:v0.0.0
```

:::

## Linux packages

Download the `.deb`, `.rpm` or `.apk` packages from tables below and install
them with the appropriate tools


## Downloads

Files are signed with [minisign](https://jedisct1.github.io/minisign/) using this public key:
```
RWSA5ztaWA/0ny2xL3U6ZQBgmfbECNm7qCPZA1VEWeGCE51WuWkj9Tt4
```


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



