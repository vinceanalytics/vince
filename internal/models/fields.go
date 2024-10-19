package models

const (
	TranslatedFieldsSize = Field_subdivision2_code - Field_view + 1
	BSIFieldsSize        = Field_city - Field_timestamp + 1
)

func (f Field) Mutex() byte {
	return byte(f - Field_view)
}

func (f Field) BSI() byte {
	return byte(f - Field_timestamp)
}

func BSI(i int) Field {
	return Field(i) + Field_timestamp
}

func Mutex(i int) Field {
	return Field(i) + Field_view
}
