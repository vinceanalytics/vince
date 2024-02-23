package db

import (
	"bytes"
	"os"
	"testing"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/buffers"
	"github.com/vinceanalytics/vince/internal/closter/events"
)

func TestArrowToParquet(t *testing.T) {
	r := events.SampleRecord()
	defer r.Release()
	b := buffers.Bytes()
	defer b.Release()

	err := ArrowToParquet(b, memory.DefaultAllocator, r)
	require.NoError(t, err)
	// os.WriteFile("testdata/record.parquet", b.Bytes(), 0600)
	data, err := os.ReadFile("testdata/record.parquet")
	require.NoError(t, err)
	require.True(t, bytes.Equal(b.Bytes(), data))
}
