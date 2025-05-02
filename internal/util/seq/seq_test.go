package seq

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeq(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seq")
	{
		s, err := New(path)
		require.NoError(t, err)
		require.Equal(t, uint64(0), s.Load())

		for range 5 {
			s.Next()
		}
		require.Equal(t, uint64(5), s.Load())
		require.NoError(t, s.Close())
	}
	s, err := New(path)
	require.NoError(t, err)
	require.Equal(t, uint64(5), s.Load())
}
