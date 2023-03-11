module github.com/gernest/vince

go 1.20

require (
	github.com/apache/arrow/go/v12 v12.0.0-20230310195443-5b2fbade23ed
	github.com/aws/aws-sdk-go-v2 v1.17.5
	github.com/aws/aws-sdk-go-v2/config v1.18.15
	github.com/aws/aws-sdk-go-v2/service/s3 v1.30.5
	github.com/aws/smithy-go v1.13.5
	github.com/beeker1121/goque v2.1.0+incompatible
	github.com/belak/octicon v0.0.0-20220531043918-8d81fefac757
	github.com/dchest/captcha v1.0.0
	github.com/dchest/siphash v1.2.3
	github.com/dgraph-io/badger/v3 v3.2103.5
	github.com/dgraph-io/ristretto v0.1.1
	github.com/dlclark/regexp2 v1.7.0
	github.com/docker/go-units v0.5.0
	github.com/eggsampler/acme/v3 v3.3.0
	github.com/emersion/go-message v0.16.0
	github.com/emersion/go-sasl v0.0.0-20200509203442-7bfe0ed36a21
	github.com/emersion/go-smtp v0.16.0
	github.com/gammazero/deque v0.2.1
	github.com/glebarez/sqlite v1.7.0
	github.com/go-gorp/gorp/v3 v3.1.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.7.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/golang/protobuf v1.5.2
	github.com/google/certificate-transparency-go v1.1.4
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hpcloud/tail v1.0.0
	github.com/jmhodges/clock v1.2.0
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/compress v1.16.0
	github.com/lestrrat-go/dataurl v0.0.0-20220721131304-b60017625013
	github.com/letsencrypt/challtestsrv v1.2.1
	github.com/letsencrypt/pkcs11key/v4 v4.0.0
	github.com/miekg/dns v1.1.51
	github.com/miekg/pkcs11 v1.1.1
	github.com/oklog/ulid/v2 v2.1.0
	github.com/oschwald/geoip2-golang v1.8.0
	github.com/oschwald/maxminddb-golang v1.10.0
	github.com/prometheus/client_golang v1.13.0
	github.com/prometheus/client_model v0.2.0
	github.com/rs/zerolog v1.21.0
	github.com/segmentio/parquet-go v0.0.0-20230106170957-952b1613a191
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399
	github.com/urfave/cli/v2 v2.23.7
	github.com/weppos/publicsuffix-go v0.30.0
	github.com/zmap/zcrypto v0.0.0-20230205235340-d51ce4775101
	github.com/zmap/zlint/v3 v3.4.1
	golang.org/x/crypto v0.1.0
	golang.org/x/exp v0.0.0-20230118134722-a68e582fa157
	golang.org/x/net v0.7.0
	golang.org/x/sync v0.1.0
	golang.org/x/term v0.5.0
	golang.org/x/text v0.7.0
	golang.org/x/time v0.3.0
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/go-jose/go-jose.v2 v2.6.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/gorm v1.24.5
)

require (
	github.com/JohnCGriffin/overflow v0.0.0-20211019200055-46fa312c352c // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/apache/thrift v0.16.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.15 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emersion/go-textwrapper v0.0.0-20200911093747-65d896831594 // indirect
	github.com/glebarez/go-sqlite v1.20.3 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/lestrrat-go/option v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/segmentio/encoding v0.3.5 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/genproto v0.0.0-20220706185917-7780775163c4 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	modernc.org/libc v1.22.2 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.20.3 // indirect
)
