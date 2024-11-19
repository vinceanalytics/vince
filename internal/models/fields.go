package models

const (
	MutexFieldSize       = Field_session + 1
	TranslatedFieldsSize = Field_subdivision2_code + 1
	SearchFieldSize      = Field_city + 1
	BSIFieldsSize        = Field_duration - Field_timestamp + 1
	AllFields            = Field_duration + 1
)

func (f Field) Mutex() byte {
	return byte(f)
}

func (f Field) BSI() byte {
	return byte(f - Field_timestamp)
}

func BSI(i int) Field {
	return Field(i) + Field_timestamp
}

func Mutex(i int) Field {
	return Field(i)
}

func DataForMetrics(m ...string) BitSet {
	var s BitSet
	for _, v := range m {
		switch v {
		case "visitors":
			s.Set(Field_id)
		case "visits":
			s.Set(Field_session)
		case "pageviews":
			s.Set(Field_view)
		case "views_per_visit":
			s.Set(Field_view)
			s.Set(Field_session)
		case "bounce_rate":
			s.Set(Field_bounce)
			s.Set(Field_session)
		case "visit_duration":
			s.Set(Field_duration)
			s.Set(Field_session)
		case "events":
			s.Set(Field_event)
		}
	}
	return s
}
