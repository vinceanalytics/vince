package events

import (
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

var onePool = &sync.Pool{New: func() any { return new(v1.Data) }}

var listPool = &sync.Pool{New: func() any { return new(v1.Data_List) }}

func One() *v1.Data {
	return onePool.Get().(*v1.Data)
}

func PutOne(d *v1.Data) {
	d.Reset()
	onePool.Put(d)
}

func List() *v1.Data_List {
	return listPool.Get().(*v1.Data_List)
}

func Put(ls *v1.Data_List) {
	for i, e := range ls.Items {
		e.Reset()
		onePool.Put(e)
		ls.Items[i] = nil
	}
	ls.Reset()
	listPool.Put(ls)
}
