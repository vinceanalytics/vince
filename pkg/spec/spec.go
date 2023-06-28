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

type Site struct {
	Domain      string    `json:"domain"`
	Public      bool      `json:"public,omitempty"`
	Description string    `json:"desc,omitempty"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
