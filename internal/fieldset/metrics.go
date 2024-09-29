package fieldset

import v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"

type Set uint8

func From(m ...string) Set {
	var s Set
	for _, v := range m {
		switch v {
		case "visitors":
			s.Set(v1.Field_id)
		case "visits":
			s.Set(v1.Field_session)
		case "pageviews":
			s.Set(v1.Field_view)
		case "views_per_visit":
			s.Set(v1.Field_view)
			s.Set(v1.Field_session)
		case "bounce_rate":
			s.Set(v1.Field_bounce)
			s.Set(v1.Field_session)
		case "visit_duration":
			s.Set(v1.Field_duration)
			s.Set(v1.Field_session)
		}
	}
	return s
}

func (s Set) Each(fn func(field v1.Field) error) error {
	for f := v1.Field_id; f <= v1.Field_duration; f++ {
		if s.Has(f) {
			err := fn(f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Set) Set(f v1.Field) {
	if !s.Has(f) {
		*s |= Set(f)
	}
}

func (s *Set) Has(f v1.Field) bool {
	return *s&Set(f) != 0
}
