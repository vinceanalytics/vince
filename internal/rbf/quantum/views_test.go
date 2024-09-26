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
	require.Equal(t, []string{"F_200001", "F_20000102", "F_2000010203", "F_200001020304", "F_iso199952"}, got)
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
		require.Equal(t, []string{"F_200001", "F_200002"}, got)
	})
	t.Run("weeks", func(t *testing.T) {
		var got []string
		f.Week("F", ts, addMonth(ts), func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_iso199952", "F_iso200001", "F_iso200002", "F_iso200003", "F_iso200004"}, got)
	})
	t.Run("days", func(t *testing.T) {
		var got []string
		f.Day("F", ts, ts.AddDate(0, 0, 3), func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_20000102", "F_20000103", "F_20000104"}, got)
	})
	t.Run("hours", func(t *testing.T) {
		var got []string
		f.Hour("F", ts, ts.Add(3*time.Hour), func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_2000010203", "F_2000010204", "F_2000010205"}, got)
	})
	t.Run("minute", func(t *testing.T) {
		var got []string
		f.Minute("F", ts, ts.Add(3*time.Minute), func(b []byte) error {
			got = append(got, string(b))
			return nil
		})
		require.Equal(t, []string{"F_200001020304", "F_200001020305", "F_200001020306"}, got)
	})
}
func TestParse(t *testing.T) {
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
			got = append(got, Parse(b[2:]))
			return nil
		})
		require.Equal(t, []string{"2000-01-01", "2000-02-01"}, got)
	})
	t.Run("weeks", func(t *testing.T) {
		var got []string
		f.Week("F", ts, addMonth(ts), func(b []byte) error {
			got = append(got, Parse(b[2:]))
			return nil
		})
		require.Equal(t, []string{"1999-12-29", "2000-01-07", "2000-01-14", "2000-01-21", "2000-01-28"}, got)
	})
	t.Run("days", func(t *testing.T) {
		var got []string
		f.Day("F", ts, ts.AddDate(0, 0, 3), func(b []byte) error {
			got = append(got, Parse(b[2:]))
			return nil
		})
		require.Equal(t, []string{"2000-01-02", "2000-01-03", "2000-01-04"}, got)
	})
	t.Run("hours", func(t *testing.T) {
		var got []string
		f.Hour("F", ts, ts.Add(3*time.Hour), func(b []byte) error {
			got = append(got, Parse(b[2:]))
			return nil
		})
		require.Equal(t, []string{"2000-01-02 03:00:00", "2000-01-02 04:00:00", "2000-01-02 05:00:00"}, got)
	})
	t.Run("minute", func(t *testing.T) {
		var got []string
		f.Minute("F", ts, ts.Add(3*time.Minute), func(b []byte) error {
			got = append(got, Parse(b[2:]))
			return nil
		})
		require.Equal(t, []string{"2000-01-02 03:04:00", "2000-01-02 03:05:00", "2000-01-02 03:06:00"}, got)
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
