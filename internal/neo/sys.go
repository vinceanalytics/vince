package neo

import (
	"time"

	"github.com/segmentio/parquet-go"
)

type Sys struct {
	Timestamp time.Time         `parquet:"timestamp,timestamp,zstd"`
	Labels    map[string]string `parquet:"labels" parquet-key:",dict,zstd" parquet-value:",dict,zstd"`
	Name      string            `parquet:"name,zstd"`
	Value     float64           `parquet:"value,zstd"`
}

var sysSchema = parquet.SchemaOf(Sys{})
