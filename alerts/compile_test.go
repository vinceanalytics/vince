package alerts

import "testing"

func TestCompile(t *testing.T) {
	Compile("scripts")
	t.Error()
}
