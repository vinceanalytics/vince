package ua

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	a, err := Get("monitoring360bot/1.1")
	require.NoError(t, err)
	require.True(t, a.Bot)
}
