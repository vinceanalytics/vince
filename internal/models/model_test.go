package models

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestSize(t *testing.T) {
	require.Equal(t, 536, int(unsafe.Sizeof(Model{})))
}
