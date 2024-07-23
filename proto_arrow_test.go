package len64

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	v13 "go.opentelemetry.io/proto/otlp/logs/v1"
	v12 "go.opentelemetry.io/proto/otlp/metrics/v1"
	v1experimental "go.opentelemetry.io/proto/otlp/profiles/v1experimental"
	v14 "go.opentelemetry.io/proto/otlp/trace/v1"

	"google.golang.org/protobuf/proto"
)

func TestSchema(t *testing.T) {
	samples := []struct {
		name string
		msg  proto.Message
	}{
		{"metric", &v12.Metric{}},
		{"logs", &v13.LogRecord{}},
		{"trace", &v14.Span{}},
		{"profile", &v1experimental.ProfileContainer{}},
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
