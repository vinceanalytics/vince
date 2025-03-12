package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranslate(t *testing.T) {
	for n := range int(MutexFieldSize) {
		require.Equal(t, n, int(AsMutex(Mutex(n))))
	}
	for n := range int(BSIFieldsSize) {
		require.Equal(t, n, int(AsBSI(BSI(n))))
	}
}
