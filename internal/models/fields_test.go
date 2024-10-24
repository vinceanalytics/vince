package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranslate(t *testing.T) {
	for n := range int(TranslatedFieldsSize) {
		require.Equal(t, n, int(Mutex(n).Mutex()))
	}
	for n := range int(BSIFieldsSize) {
		require.Equal(t, n, int(BSI(n).BSI()))
	}
}
