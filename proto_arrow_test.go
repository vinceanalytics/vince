package len64

import (
	"os"
	"path/filepath"
	"testing"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/proto"
)

func TestSchema(t *testing.T) {
	samples := []struct {
		name string
		msg  proto.Message
	}{
		{"model", &v1.Model{}},
	}

	for _, e := range samples {
		msg := build(e.msg.ProtoReflect())
		schema := msg.schema.String()
		file := filepath.Join("testdata", e.name+"_arrow_schema.txt")
		os.WriteFile(file, []byte(schema), 0600)
		want, err := os.ReadFile(file)
		require.NoError(t, err)
		require.Equal(t, string(want), schema)
	}
}
