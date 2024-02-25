package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// QueryParams represents the query parameters passed in an HTTP request.
// Query parameter keys are case-sensitive, as per the HTTP spec.
type QueryParams map[string]string

// NewQueryParams returns a new QueryParams from the given HTTP request.
func NewQueryParams(r *http.Request) (QueryParams, error) {
	qp := make(QueryParams)
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, err
	}
	for k, v := range values {
		qp[k] = v[0]
	}

	if _, ok := qp["freshness_strict"]; ok {
		if _, ok := qp["freshness"]; !ok {
			return nil, fmt.Errorf("freshness_strict requires freshness")
		}
	}
	for _, k := range []string{"timeout", "freshness", "db_timeout"} {
		t, ok := qp[k]
		if ok {
			_, err := time.ParseDuration(t)
			if err != nil {
				return nil, fmt.Errorf("%s is not a valid duration", k)
			}
		}
	}
	for _, k := range []string{"retries"} {
		r, ok := qp[k]
		if ok {
			_, err := strconv.Atoi(r)
			if err != nil {
				return nil, fmt.Errorf("%s is not a valid integer", k)
			}
		}
	}
	q, ok := qp["q"]
	if ok {
		if q == "" {
			return nil, fmt.Errorf("query parameter not set")
		}
	}
	return qp, nil
}

func (qp QueryParams) SiteID() string {
	return qp["site_id"]
}

func (qp QueryParams) TenantID() string {
	return qp["tenant_id"]
}

// Timings returns true if the query parameters indicate timings should be returned.
func (qp QueryParams) Timings() bool {
	return qp.HasKey("timings")
}

// Tx returns true if the query parameters indicate the query should be executed in a transaction.
func (qp QueryParams) Tx() bool {
	return qp.HasKey("transaction")
}

// Query returns true if the query parameters request queued operation
func (qp QueryParams) Queue() bool {
	return qp.HasKey("queue")
}

// Pretty returns true if the query parameters indicate pretty-printing should be returned.
func (qp QueryParams) Pretty() bool {
	return qp.HasKey("pretty")
}

// Bypass returns true if the query parameters indicate bypass mode.
func (qp QueryParams) Bypass() bool {
	return qp.HasKey("bypass")
}

// Wait returns true if the query parameters indicate the query should wait.
func (qp QueryParams) Wait() bool {
	return qp.HasKey("wait")
}

// Associative returns true if the query parameters request associative results.
func (qp QueryParams) Associative() bool {
	return qp.HasKey("associative")
}

// NoRewrite returns true if the query parameters request no rewriting of queries.
func (qp QueryParams) NoRewriteRandom() bool {
	return qp.HasKey("norwrandom")
}

// NonVoters returns true if the query parameters request non-voters to be included in results.
func (qp QueryParams) NonVoters() bool {
	return qp.HasKey("nonvoters")
}

// NoLeader returns true if the query parameters request no leader mode.
func (qp QueryParams) NoLeader() bool {
	return qp.HasKey("noleader")
}

// Redirect returns true if the query parameters request redirect mode.
func (qp QueryParams) Redirect() bool {
	return qp.HasKey("redirect")
}

// Vacuum returns true if the query parameters request vacuum mode.
func (qp QueryParams) Vacuum() bool {
	return qp.HasKey("vacuum")
}

// Compress returns true if the query parameters request compression.
func (qp QueryParams) Compress() bool {
	return qp.HasKey("compress")
}

// Key returns the value of the key named "key".
func (qp QueryParams) Key() string {
	return qp["key"]
}

// DBTimeout returns the value of the key named "db_timeout".
func (qp QueryParams) DBTimeout(def time.Duration) time.Duration {
	t, ok := qp["db_timeout"]
	if !ok {
		return def
	}
	d, _ := time.ParseDuration(t)
	return d
}

// Query returns the requested query.
func (qp QueryParams) Query() string {
	return qp["q"]
}

// Freshness returns the requested freshness duration.
func (qp QueryParams) Freshness() time.Duration {
	f := qp["freshness"]
	d, _ := time.ParseDuration(f)
	return d
}

// FreshnessStrict returns true if the query parameters indicate strict freshness.
func (qp QueryParams) FreshnessStrict() bool {
	return qp.HasKey("freshness_strict")
}

// Sync returns whether the sync flag is set.
func (qp QueryParams) Sync() bool {
	return qp.HasKey("sync")
}

// Timeout returns the requested timeout duration.
func (qp QueryParams) Timeout(def time.Duration) time.Duration {
	t, ok := qp["timeout"]
	if !ok {
		return def
	}
	d, _ := time.ParseDuration(t)
	return d
}

// Retries returns the requested number of retries.
func (qp QueryParams) Retries(def int) int {
	i, ok := qp["retries"]
	if !ok {
		return def
	}
	r, _ := strconv.Atoi(i)
	return r
}

// Version returns the requested version.
func (qp QueryParams) Version() string {
	return qp["ver"]
}

// HasKey returns true if the given key is present in the query parameters.
func (qp QueryParams) HasKey(k string) bool {
	_, ok := qp[k]
	return ok
}
