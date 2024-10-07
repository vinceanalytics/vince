package ua2

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func TestParse(t *testing.T) {
	userAgent := `Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1`
	require.Equal(t, &v1.Agent{
		Os: "iOS", OsVersion: "11_0", Browser: "Mobile Safari", BrowserVersion: "11.0", Device: "Mobile",
	}, Parse(userAgent))
}
