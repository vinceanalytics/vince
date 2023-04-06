package system

import "sync"

var set = &sync.Map{}

func Get(name string) *Histogram {
	v, _ := set.LoadOrStore(name, new(Histogram))
	return v.(*Histogram)
}

type Reader interface {
	Read(*History)
}

func All(f func(Reader)) {
	set.Range(func(key, value any) bool {
		f(value.(*Histogram))
		return true
	})
}
