package index

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadIndexFile(t *testing.T) {
	f, err := os.Open("testdata/01HPA98Z0TVKPC4QC7HQBTF30Q")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	m, err := NewFileIndex(f)
	if err != nil {
		t.Fatal(err)
	}
	col, err := m.get("path")
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "path", col.name)
	require.Equal(t, 1, len(col.bitmaps))
}
