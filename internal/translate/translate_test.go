package translate

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/models"
)

func TestTranslate(t *testing.T) {

	t.Run("panics on non mutex fields", func(t *testing.T) {
		ts := New()
		for f := v1.Field_view; f <= v1.Field_duration; f++ {
			require.Panics(t, func() { ts.Get(f, nil) })
		}
	})
	t.Run("works for mutex fields", func(t *testing.T) {
		ts := New()
		for f := v1.Field_domain; f <= v1.Field_subdivision2_code; f++ {
			id, ok := ts.Get(f, []byte("hello"))
			require.Equal(t, uint64(1), id)
			require.False(t, ok)
		}
	})

	t.Run("safe bounds", func(t *testing.T) {
		ts := New()
		require.Equal(t, uint8(0), models.AsMutex(v1.Field_domain))
		require.Equal(t, uint8(len(ts.mapping)-1), models.AsMutex(v1.Field_subdivision2_code))
		_ = ts
	})
}
