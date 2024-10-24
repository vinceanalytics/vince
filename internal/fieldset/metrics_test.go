package fieldset

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/models"
)

func TestBS(t *testing.T) {
	bs := From("visitors", "visitors",
		"pageviews", "views_per_visit", "bounce_rate", "visit_duration", "events")
	want := []models.Field{
		models.Field_id,
		models.Field_bounce,
		models.Field_duration,
		models.Field_view,
		models.Field_session,
		models.Field_event,
	}
	var got []models.Field
	bs.Each(func(field models.Field) error {
		got = append(got, field)
		return nil
	})
	require.Equal(t, want, got)
}
