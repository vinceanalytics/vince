package main

import (
	"os"
	"testing"
)

func TestBlog(t *testing.T) {
	b, _ := os.ReadFile("testdata/blog/post.md")
	p := renderPost(b)

	t.Errorf("%#v", p)
}
