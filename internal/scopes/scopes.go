// Package scopes defines all resources and resource operations exposed by
// vince. Resources are mainly grpc services and the resource operations are the
// grpc full methods.
package scopes

import (
	"encoding/json"
	"fmt"

	"github.com/bits-and-blooms/bitset"
)

const (
	ResourceBaseURL = "https://vinceanalytics.com"
)

type Scope uint

const (
	All Scope = 1 + iota
	Query

	CreateSite
	GetSite
	ListSites
	DeleteSite

	CreateSnippet
	UpdateSnippet
	ListSnippets
	DeleteSnippet

	CreateGoal
	GetGoal
	ListGoals
	DeleteGoal

	Version
)

var _ json.Marshaler = (*Scope)(nil)

var _ json.Unmarshaler = (*Scope)(nil)

func (s Scope) String() string {
	return nameToValue[s]
}

func (s Scope) FullMethod() string {
	return nameToValue[s]
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Scope) UnmarshalJSON(b []byte) error {
	var o string
	err := json.Unmarshal(b, &o)
	if err != nil {
		return err
	}
	return s.Parse(o)
}

func (s *Scope) Parse(o string) error {
	scope := valueToName[o]
	if scope == 0 {
		return fmt.Errorf("invalid scope :%q", o)
	}
	*s = scope
	return nil
}

var nameToValue = map[Scope]string{
	All:           "*",
	Query:         "/v1.Query/Query",
	CreateSite:    "/v1.Sites/CreateSite",
	GetSite:       "/v1.Sites/GetSite",
	ListSites:     "/v1.Sites/ListSites",
	DeleteSite:    "/v1.Sites/DeleteSite",
	CreateSnippet: "/v1.Snippets/CreateSnippet",
	UpdateSnippet: "/v1.Snippets/UpdateSnippet",
	ListSnippets:  "/v1.Snippets/ListSnippets",
	DeleteSnippet: "/v1.Snippets/DeleteSnippet",
	CreateGoal:    "/v1.Goals/CreateGoal",
	GetGoal:       "/v1.Goals/GetGoal",
	ListGoals:     "/v1.Goals/ListGoals",
	DeleteGoal:    "/v1.Goals/DeleteGoal",
	Version:       "/v1.Vince/Version",
}

var valueToName = map[string]Scope{
	"*":                          All,
	"/v1.Query/Query":            Query,
	"/v1.Sites/CreateSite":       CreateSite,
	"/v1.Sites/GetSite":          GetSite,
	"/v1.Sites/ListSites":        ListSites,
	"/v1.Sites/DeleteSite":       DeleteSite,
	"/v1.Snippets/CreateSnippet": CreateSnippet,
	"/v1.Snippets/UpdateSnippet": UpdateSnippet,
	"/v1.Snippets/ListSnippets":  ListSnippets,
	"/v1.Snippets/DeleteSnippet": DeleteSnippet,
	"/v1.Goals/CreateGoal":       CreateGoal,
	"/v1.Goals/GetGoal":          GetGoal,
	"/v1.Goals/ListGoals":        ListGoals,
	"/v1.Goals/DeleteGoal":       DeleteGoal,
	"/v1.Vince/Version":          Version,
}

type List = bitset.BitSet
