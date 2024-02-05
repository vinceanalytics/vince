package ua

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBot(t *testing.T) {
	m := Get("monitoring360bot/1.1")
	require.True(t, m.IsBot)
}
