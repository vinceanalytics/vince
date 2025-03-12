package models

import v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"

const (
	MutexFieldSize       = v1.Field_session + 1
	TranslatedFieldsSize = v1.Field_subdivision2_code + 1
	SearchFieldSize      = v1.Field_city + 1
	BSIFieldsSize        = v1.Field_duration - v1.Field_timestamp + 1
	AllFields            = v1.Field_duration + 1
)

func AsMutex(f Field) byte {
	return byte(f)
}

func AsBSI(f Field) byte {
	return byte(f - v1.Field_timestamp)
}

func BSI(i int) Field {
	return Field(i) + v1.Field_timestamp
}

func Mutex(i int) Field {
	return Field(i)
}

func DataForMetrics(m ...string) BitSet {
	var s BitSet
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
		case "events":
			s.Set(v1.Field_event)
		}
	}
	return s
}
