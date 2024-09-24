package dsl

func (s *Store[T]) Append(data []T) {
	for i := range data {
		s.schema.Write(data[i])
	}
}

func (s *Store[T]) Flush() error {
	if len(s.schema.ids) == 0 {
		return nil
	}
	return s.schema.process(s)
}
