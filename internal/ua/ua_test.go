package ua

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	a := Get("monitoring360bot/1.1")
	require.True(t, a.Bot)
}
