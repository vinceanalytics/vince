package fieldset

import (
	"github.com/bits-and-blooms/bitset"
	"github.com/vinceanalytics/vince/internal/models"
)

type Set struct {
	bitset.BitSet
}

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
	var b [8]uint
	_, buf := s.NextSetMany(0, b[:])
	for i := range buf {
		err := fn(models.Field(buf[i]))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Set) Set(f models.Field) {
	s.BitSet.Set(uint(f))
}

func (s *Set) Has(f models.Field) bool {
	return s.Test(uint(f))
}
