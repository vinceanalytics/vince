package query

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	query := "period=30d&date=2024-07-24&filters=%5B%5B%22is%22%2C%22event%3Apage%22%2C%5B%22%2F%3Adashboard%22%2C%22%2Fsites%22%5D%5D%2C%5B%22is%22%2C%22visit%3Autm_source%22%2C%5B%22Twitter%22%5D%5D%5D&with_imported=true&limit=300"
	params, err := url.ParseQuery(query)
	require.NoError(t, err)
	var a []Filter
	err = json.Unmarshal([]byte(params.Get("filters")), &a)
	require.NoError(t, err)
	want := []Filter{
		{Op: "is", Key: "event:page", Value: []string{"/:dashboard", "/sites"}},
		{Op: "is", Key: "visit:utm_source", Value: []string{"Twitter"}},
	}
	require.Equal(t, want, a)
}

func TestParseDate(t *testing.T) {
	t.Error(time.Parse(time.RFC3339, "2021-01-01"))
}
