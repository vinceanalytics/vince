package license

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidMessage(t *testing.T) {
	msg := `-----BEGIN LICENSE KEY-----

xA0DAAgT7XgNsLV7NPIBy+ViAAAAAAAIkE4QChgDIPqpp6eLMyoQdGVzdEBleGFt
cOJsZS5j4W9tAMKiBAATCAAQBQJmwR54CRDteA2wtXs08gAAMYYCCQHrHH+2nN6Q
IBPeQu6l1leabWD4hZImnHR3dwMWnyQmwsyrX35a4q1Yko80o3ucTRpTMCvwL+Uo
abG5jqHJZZC0/AIJAbgeWCLV+1w7VhT25MfloaDI16ERnK4RcgJrmStZE+XUjqoa
9icD8rkVem3BnaUEPphX43scxe4zYVLUhy/ny4cr
=Qnfu
-----END LICENSE KEY-----`

	ls, err := Verify([]byte(msg))
	require.NoError(t, err)
	require.Equal(t, "test@example.com", ls.Email)
}
