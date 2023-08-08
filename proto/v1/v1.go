package v1

import "bytes"

func (s *StoreKey) Badger() []byte {
	var b bytes.Buffer
	if s.Namespace == "" {
		b.WriteString("vince")
	} else {
		b.WriteString(s.Namespace)
	}
	b.WriteByte('/')
	b.WriteString(s.Prefix.String())
	b.WriteByte('/')
	b.WriteString(s.Key)
	return b.Bytes()
}
