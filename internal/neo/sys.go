package neo

import (
	"time"
)

type Sys struct {
	Timestamp time.Time
	Labels    map[string]string
	Name      string
	Value     float64
}
