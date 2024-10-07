package license

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidMessage(t *testing.T) {
	msg := `-----BEGIN LICENSE KEY-----

xA0DAAgTTlEgdN2He/kBy+ZiAAAAAAB7InZpZXdzIjoxMDAwMCwic2l0ZXMiOjEw
LCJ1c2VycyI6MywiZXhwaXJ5IjoxNzU1ODg5NTk3MjI35CwiZW1haWwiOiJ0ZXN0
QGXjeGFtcGxlLmPib20ifQDCoQQAEwgAEAUCZsjdvQkQTlEgdN2He/kAAKh0AgkB
BdxQbTB0TcI0Kw8+5nbQhFqryBtfw4RcZpcl1tcwSIpJl4LqyPJWyg9FLP8mkFeJ
oq1kGKEd2sv/kN+7TQHrXYgCB1hMXOMaTPzhuI0xSeQwlvesl6nPFE4/IYXjQVsI
VUWBwvJXfkQk94f2jaTTSpArc1h2L3GKQQogfS3lwrS5en9u
=htPY
-----END LICENSE KEY-----`

	ls, err := Parse([]byte(msg))
	require.NoError(t, err)
	require.Equal(t, "test@example.com", ls.Email)
}
