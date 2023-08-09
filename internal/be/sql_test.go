package be

import "testing"

func TestParse(t *testing.T) {
	p := NewParser()
	p.Parse("select * from 'vince.io'")
	t.Error()
}
