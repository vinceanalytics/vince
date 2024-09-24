package quantum

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestViews(t *testing.T) {
	f := NewField()
	defer f.Release()
	ts := time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC)
	f.ViewsByTimeInto(ts)
	var got []string
	require.NoError(t, f.Views("F", func(view string) error {
		got = append(got, view)
		return nil
	}))
	require.Equal(t, []string{"F_2000", "F_200001", "F_20000102", "F_2000010203", "F_200001020304"}, got)
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
		f.Views("F", func(view string) error {
			return nil
		})
	}
}
