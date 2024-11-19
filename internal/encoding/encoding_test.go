package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/models"
)

func TestKey(t *testing.T) {

	var k Key
	k.Write(models.Field_domain, Minute, 1, 1)

	f, re, v, co := k.Component()
	require.Equal(t, models.Field_domain, f)
	require.Equal(t, Minute, re)
	require.Equal(t, uint64(1), v)
	require.Equal(t, uint64(1), co)

	b := From(k.Bytes())

	f, re, v, co = b.Component()
	require.Equal(t, models.Field_domain, f)
	require.Equal(t, Minute, re)
	require.Equal(t, uint64(1), v)
	require.Equal(t, uint64(1), co)
}
