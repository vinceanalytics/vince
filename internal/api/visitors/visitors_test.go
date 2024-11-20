package visitors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

func TestCurrentRange(t *testing.T) {
	end := xtime.Test()
	start := end.Add(-5 * time.Minute)
	seq := compute.Range(encoding.Minute, start, end)
	var from, to []int64

	for k, v := range seq {
		from = append(from, k)
		to = append(to, v)
	}

	t.Run("ensure last time range is included", func(t *testing.T) {
		require.Equal(t, end.Truncate(time.Minute).UnixMilli(), from[len(from)-1])
	})

	require.Equal(t, 5, len(from))
}
