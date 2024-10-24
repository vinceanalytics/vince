package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/models"
)

func TestY(t *testing.T) {
	rs := Bitmap(0, 0, models.Field_domain)
	require.Equal(t, []byte{0, 0, 0, 12}, rs)
	require.Equal(t, 4, cap(rs))
}
