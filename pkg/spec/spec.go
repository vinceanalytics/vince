package spec

import (
	"time"

	"github.com/vinceanalytics/vince/pkg/property"
)

type CreateSite struct {
	Domain      string  `json:"domain"`
	Public      *bool   `json:"public,omitempty"`
	Description *string `json:"description,omitempty"`
}

type UpdateSite struct {
	Public      *bool   `json:"public,omitempty"`
	Description *string `json:"description,omitempty"`
}

type Site One[Site_]

type Site_ struct {
	Domain      string    `json:"domain"`
	Public      bool      `json:"public,omitempty"`
	Description string    `json:"desc,omitempty"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SiteList List[Site_]

type QueryOptions struct {
	Window time.Duration   `json:"window,omitempty"`
	Offset time.Duration   `json:"offset,omitempty"`
	Metric property.Metric `json:"metric,omitempty"`
}

type QueryPropertyOptions struct {
	Window   time.Duration     `json:"window,omitempty"`
	Offset   time.Duration     `json:"offset,omitempty"`
	Metric   property.Metric   `json:"metric,omitempty"`
	Property property.Property `json:"property,omitempty"`
	Selector Select            `json:"selector,omitempty"`
}

type PropertyResult[T uint64 | []uint64] struct {
	Timestamps []int64       `json:"timestamps"`
	Elapsed    time.Duration `json:"elapsed"`
	Result     map[string]T  `json:"result"`
}

type Series struct {
	Timestamps []int64       `json:"timestamps"`
	Elapsed    time.Duration `json:"elapsed"`
	Result     []uint64      `json:"result"`
}

type Metrics struct {
	Visitors uint64 `json:"visitors,omitempty"`
	Views    uint64 `json:"views,omitempty"`
	Events   uint64 `json:"events,omitempty"`
	Visits   uint64 `json:"visits,omitempty"`
}

type ResultSet[T any] struct {
	Timestamps []int64       `json:"timestamps,omitempty"`
	Elapsed    time.Duration `json:"elapsed"`
	Result     T             `json:"result"`
}

type Stat One[uint64]

type Stats One[Metrics]

type Global[T uint64 | Metrics] struct {
	Elapsed time.Duration `json:"elapsed"`
	Result  T             `json:"result"`
}

type One[T any] struct {
	Elapsed time.Duration `json:"elapsed"`
	Item    T             `json:"item"`
}

type List[T any] struct {
	Elapsed time.Duration `json:"elapsed"`
	Items   []T           `json:"items"`
}

type Select struct {
	Exact *string `json:"exact,omitempty"`
	Re    *string `json:"re,omitempty"`
	Glob  *string `json:"glob,omitempty"`
}
