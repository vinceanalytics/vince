package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateDates(t *testing.T) {
	ts, _ := time.Parse(time.RFC822, time.RFC822)
	ts = ts.UTC()
	g := generateDates(5, ts)
	got := make([]string, len(g))
	for i := range g {
		got[i] = g[i].Format(time.DateOnly)
	}
	require.True(t, ts.Equal(g[len(g)-1]))
	require.Equal(t, []string{"2005-12-29", "2005-12-30", "2005-12-31", "2006-01-01", "2006-01-02"}, got)
}
