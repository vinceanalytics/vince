package neo

import (
	"bytes"
	"os"
	"testing"

	"github.com/vinceanalytics/vince/internal/must"
)

func TestComputedPartition(t *testing.T) {
	t.Run("defaults to only metrics and timestamp", func(t *testing.T) {
		got := []byte(computedPartition().String())
		// os.WriteFile("testdata/computed_partition_schema_default.txt", got, 0600)
		want := must.
			Must(os.ReadFile("testdata/computed_partition_schema_default.txt"))()
		if !bytes.Equal(got, want) {
			t.Error("schema changed")
		}
	})
	t.Run("with partition key", func(t *testing.T) {
		got := []byte(computedPartition("path").String())
		os.WriteFile("testdata/computed_partition_schema_with_keys.txt", got, 0600)
		want := must.
			Must(os.ReadFile("testdata/computed_partition_schema_with_keys.txt"))()
		if !bytes.Equal(got, want) {
			t.Error("schema changed")
		}
	})
}
