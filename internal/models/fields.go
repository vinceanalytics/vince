package models

const (
	MutexFieldSize       = Field_session + 1
	TranslatedFieldsSize = Field_subdivision2_code + 1
	BSIFieldsSize        = Field_duration - Field_timestamp + 1
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
