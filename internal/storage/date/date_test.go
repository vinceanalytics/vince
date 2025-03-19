package date

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolution(t *testing.T) {
	ts, _ := time.Parse(time.RFC822, time.RFC822)
	got := Debug(ts)
	// os.WriteFile("testdata/time_resolution", []byte(got), 0600)
	want, err := os.ReadFile("testdata/time_resolution")
	require.NoError(t, err)
	require.Equal(t, string(want), got)
}
