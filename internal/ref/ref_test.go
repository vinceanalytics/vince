package ref

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRe(t *testing.T) {
	got, err := Search("mail.126.com")
	require.NoError(t, err)
	require.Equal(t, "126 Mail", string(got))
}

func BenchmarkNew(b *testing.B) {
	for range b.N {
		New()
	}
}
