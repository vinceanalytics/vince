package ref

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRe(t *testing.T) {
	re, err := New()
	require.NoError(t, err)

	got, err := re.Search("mail.126.com")
	require.NoError(t, err)
	require.Equal(t, "126 Mail", got)
}
