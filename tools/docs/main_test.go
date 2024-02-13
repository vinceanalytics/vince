package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	var b bytes.Buffer
	err := buildDocs(&b, "testdata/")
	require.NoError(t, err)
	os.WriteFile("testdata/index.html", b.Bytes(), 0600)
}
