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
		f.Month("F", ts, end, func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_200001", "F_200002", "F_200003"}, got)
	})
	t.Run("weeks", func(t *testing.T) {
		var got []string
		f.Week("F", ts, addMonth(ts), func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_20000152", "F_20000101", "F_20000102", "F_20000103", "F_20000104"}, got)
	})
	t.Run("days", func(t *testing.T) {
		var got []string
		f.Day("F", ts, ts.AddDate(0, 0, 3), func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_2000015202", "F_2000010103", "F_2000010104", "F_2000010105"}, got)
	})
	t.Run("hours", func(t *testing.T) {
		var got []string
		f.Hour("F", ts, ts.Add(3*time.Hour), func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_200001520203", "F_200001520204", "F_200001520205", "F_200001520206"}, got)
	})
	t.Run("minute", func(t *testing.T) {
		var got []string
		f.Minute("F", ts, ts.Add(3*time.Minute), func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_20000152020304", "F_20000152020305", "F_20000152020306", "F_20000152020307"}, got)
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
