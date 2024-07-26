package web

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFocus(t *testing.T) {
	var o bytes.Buffer
	err := home.Execute(&o, map[string]any{})
	require.NoError(t, err)
	fmt.Println(o.String())
	t.Error()
}
