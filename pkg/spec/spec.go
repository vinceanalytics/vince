package spec

import "time"

type CreateSite struct {
	Domain      string  `json:"domain"`
	Public      *bool   `json:"public,omitempty"`
	Description *string `json:"desc,omitempty"`
}

type UpdateSite struct {
	Public      *bool   `json:"public,omitempty"`
	Description *string `json:"desc,omitempty"`
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

type GlobalList List[Metrics]

type Metrics struct {
	Visitors uint64 `json:"visitors,omitempty"`
	Views    uint64 `json:"views,omitempty"`
	Events   uint64 `json:"events,omitempty"`
	Visits   uint64 `json:"visits,omitempty"`
}

type One[T any] struct {
	Elapsed time.Duration `json:"elapsed"`
	Item    T             `json:"item"`
}

type List[T any] struct {
	Elapsed time.Duration `json:"elapsed"`
	Items   []T           `json:"items"`
}
