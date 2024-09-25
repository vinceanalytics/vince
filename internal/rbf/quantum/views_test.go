package quantum

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFieldViews(t *testing.T) {
	f := NewField()
	defer f.Release()
	ts := time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC)
	f.ViewsByTimeInto(ts)
	var got []string
	f.Views("F", func(view string) {
		got = append(got, view)
	})
	require.Equal(t, []string{"F_200001", "F_20000152", "F_2000015202", "F_200001520203", "F_20000152020304"}, got)
}

func TestFieldTimeRange(t *testing.T) {
	f := NewField()
	defer f.Release()
	ts := time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC)

	t.Run("months", func(t *testing.T) {
		end := ts
		for range 2 {
			end = addMonth(end)
		}
		var got []string
		f.Month("F", ts, end, func(b []byte) {
			got = append(got, string(b))
		})
		require.Equal(t, []string{"F_200001", "F_200002", "F_200003"}, got)
	})
}

func BenchmarkField_ViewsByTimeInto(b *testing.B) {
	f := NewField()
	defer f.Release()
	ts := time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC)

	for range b.N {
		f.ViewsByTimeInto(ts)
	}
}
func BenchmarkField_Views(b *testing.B) {
	f := NewField()
	defer f.Release()
	ts := time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC)
	f.ViewsByTimeInto(ts)

	for range b.N {
		f.Views("F", func(view string) {})
	}
}
