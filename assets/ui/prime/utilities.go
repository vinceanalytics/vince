package prime

import "html/template"

type Class interface {
	Class() template.HTMLAttr
}

type Style interface {
	Style() template.HTMLAttr
}
type attr interface {
	attr()
}

func BuildAttr[T attr](args ...T) (classes, style []template.HTMLAttr) {
	for _, v := range args {
		switch e := any(v).(type) {
		case Class:
			classes = append(classes, e.Class())
		case Style:
			style = append(style, e.Style())
		}
	}
	return
}
