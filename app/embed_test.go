package app

import (
	"fmt"
	"io/fs"
	"testing"
)

func TestY(t *testing.T) {
	fs.WalkDir(Public, ".", func(path string, d fs.DirEntry, err error) error {

		fmt.Println(path, err)
		return nil
	})
	t.Error()
}
