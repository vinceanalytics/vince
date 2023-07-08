package spec

import (
	"time"
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
	Window time.Duration `json:"window,omitempty"`
	Offset time.Duration `json:"offset,omitempty"`
	Metric Metric        `json:"metric,omitempty"`
}

type QueryPropertyOptions struct {
	Window   time.Duration `json:"window,omitempty"`
	Offset   time.Duration `json:"offset,omitempty"`
	Metric   Metric        `json:"metric,omitempty"`
	Selector Select        `json:"selector,omitempty"`
}

func (q *QueryPropertyOptions) Defaults() {
	q.Window = time.Hour * 24
}

type PropertyResult[T uint64 | []uint64] struct {
	Timestamps []int64       `json:"timestamps"`
	Elapsed    time.Duration `json:"elapsed"`
	Result     map[string]T  `json:"result"`
}

type Metrics struct {
	Visitors uint64 `json:"visitors,omitempty"`
	Views    uint64 `json:"views,omitempty"`
	Events   uint64 `json:"events,omitempty"`
	Visits   uint64 `json:"visits,omitempty"`
}

type System struct {
	Name       string        `json:"system"`
	Timestamps []int64       `json:"timestamps"`
	Elapsed    time.Duration `json:"elapsed"`
	Result     []int64       `json:"result"`
}

type Series[T uint64 | []uint64] struct {
	Timestamps []int64       `json:"timestamps,omitempty"`
	Elapsed    time.Duration `json:"elapsed"`
	Result     T             `json:"result"`
}

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
