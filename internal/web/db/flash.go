package db

type Flash map[string]map[string]string

func (f Flash) Success(msg string) {
	s, ok := f["success"]
	if !ok {
		s = map[string]string{}
		f["success"] = s
	}
	s["message"] = msg
}

func (f Flash) Error(msg string) {
	s, ok := f["error"]
	if !ok {
		s = map[string]string{}
		f["error"] = s
	}
	s["message"] = msg
}
