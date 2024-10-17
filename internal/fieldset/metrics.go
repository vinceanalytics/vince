package fieldset

import "github.com/vinceanalytics/vince/internal/models"

type Set uint8

func From(m ...string) Set {
	var s Set
	for _, v := range m {
		switch v {
		case "visitors":
			s.Set(models.Field_id)
		case "visits":
			s.Set(models.Field_session)
		case "pageviews":
			s.Set(models.Field_view)
		case "views_per_visit":
			s.Set(models.Field_view)
			s.Set(models.Field_session)
		case "bounce_rate":
			s.Set(models.Field_bounce)
			s.Set(models.Field_session)
		case "visit_duration":
			s.Set(models.Field_duration)
			s.Set(models.Field_session)
		case "events":
			s.Set(models.Field_event)
		}
	}
	return s
}

func (s Set) Each(fn func(field models.Field) error) error {
	for f := models.Field_id; f <= models.Field_duration; f++ {
		if s.Has(f) {
			err := fn(f)
			if err != nil {
				return err
			}
		}
	}
	if s.Has(models.Field_event) {
		return fn(models.Field_event)
	}
	return nil
}

func (s *Set) Set(f models.Field) {
	if !s.Has(f) {
		*s |= Set(f)
	}
}

func (s *Set) Has(f models.Field) bool {
	return *s&Set(f) != 0
}
