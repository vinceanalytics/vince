package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	require.False(t, Build().IsZero())
}
