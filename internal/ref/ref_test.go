package ref

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/location"
)

func TestRe(t *testing.T) {
	lo := location.New()
	defer lo.Close()
	got, err := Search(lo, "mail.126.com")
	require.NoError(t, err)
	require.Equal(t, "126 Mail", string(got))
}
