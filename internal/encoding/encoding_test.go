package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/models"
)

func TestBitmap(t *testing.T) {
	rs := Bitmap(0, 0, models.Field_domain)
	require.Equal(t, []byte{0, 0, 0, 12}, rs)
	require.Equal(t, 4, cap(rs))
}

func TestShard(t *testing.T) {
	rs := Shard(0)
	require.Equal(t, []byte{3, 0}, rs)
	require.Equal(t, 2, cap(rs))
}
