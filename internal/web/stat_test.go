package web

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/alicia"
)

func TestTopfields(t *testing.T) {
	require.Equal(t, []alicia.Field{
		alicia.ID,
		alicia.BOUNCE,
		alicia.SESSION,
		alicia.VIEW,
		alicia.DURATION,
	}, topFields)
}
