package staples

import (
	"os"
	"testing"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/stretchr/testify/require"
)

func TestSchema(t *testing.T) {
	b := NewArrow[Event](memory.NewGoAllocator())
	defer b.Release()

	as := b.build.Schema()
	// os.WriteFile("testdata/schema", []byte(as.String()), 0600)
	want, err := os.ReadFile("testdata/schema")
	require.NoError(t, err)
	require.Equal(t, string(want), as.String())
}
