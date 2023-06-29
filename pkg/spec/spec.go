package spec

import "time"

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

type Global One[Metrics]

type GlobalQuery_ struct {
	Visitors []uint64 `json:"visitors,omitempty"`
	Views    []uint64 `json:"views,omitempty"`
	Events   []uint64 `json:"events,omitempty"`
	Visits   []uint64 `json:"visits,omitempty"`
}

type QueryOptions struct {
	Window time.Duration `json:"window,omitempty"`
	Offset time.Duration `json:"offset,omitempty"`
}

// QueryGlobal result of querying global stats for all metrics.
type QueryGlobal Series[GlobalQuery_]

// QueryGlobalMetric result for querying global stats for a single metric.
type QueryGlobalMetric Series[[]uint64]

type Series[T any] struct {
	Timestamps []int64       `json:"timestamps"`
	Elapsed    time.Duration `json:"elapsed"`
	Result     T             `json:"result"`
}

type Metrics struct {
	Visitors uint64 `json:"visitors,omitempty"`
	Views    uint64 `json:"views,omitempty"`
	Events   uint64 `json:"events,omitempty"`
	Visits   uint64 `json:"visits,omitempty"`
}

type Metric One[uint64]

type One[T any] struct {
	Elapsed time.Duration `json:"elapsed"`
	Item    T             `json:"item"`
}

type List[T any] struct {
	Elapsed time.Duration `json:"elapsed"`
	Items   []T           `json:"items"`
}
