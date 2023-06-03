package alerts

import "github.com/dop251/goja"

type Script struct {
	runtime *goja.Runtime
	m       map[string][]goja.Value
}

func (s *Script) Register(domain string, o goja.Value) {
	s.m[domain] = append(s.m[domain], o)
}
