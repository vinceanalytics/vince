package curl

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"google.golang.org/protobuf/proto"
)

var API = CMD("http://localhost:8080")

func TestVersion(t *testing.T) {
	check(t, false, "version.sh", "/api/v1/version", http.MethodGet, nil, nil)
}
func TestVisitors(t *testing.T) {
	check(t, false, "visitors.sh", "/api/v1/visitors/example.com", http.MethodGet, nil, nil)
}

func check(t *testing.T, write bool, file string, path, method string, headers http.Header, body proto.Message) {
	t.Helper()
	file = filepath.Join("testdata/", file)
	var b bytes.Buffer
	err := API.Format(&b, path, method, headers, body)
	require.NoError(t, err)
	if write {
		os.WriteFile(file, b.Bytes(), 0600)
	}
	want, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, string(want), b.String())
}

func TestConfig(t *testing.T) {
	data, _ := m.Marshal(&v1.Config{
		Data: "path/to/storage",
	})
	os.WriteFile("testdata/config.json", data, 0600)
}
