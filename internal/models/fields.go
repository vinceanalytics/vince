package models

const (
	TranslatedFieldsSize = Field_subdivision2_code - Field_browser
)

type Field uint8

const (
	Field_unknown           Field = 0
	Field_timestamp         Field = 1
	Field_id                Field = 2
	Field_bounce            Field = 3
	Field_session           Field = 4
	Field_view              Field = 5
	Field_duration          Field = 6
	Field_city              Field = 7
	Field_browser           Field = 8
	Field_browser_version   Field = 9
	Field_country           Field = 10
	Field_device            Field = 11
	Field_domain            Field = 12
	Field_entry_page        Field = 13
	Field_event             Field = 14
	Field_exit_page         Field = 15
	Field_host              Field = 16
	Field_os                Field = 17
	Field_os_version        Field = 18
	Field_page              Field = 19
	Field_referrer          Field = 20
	Field_source            Field = 21
	Field_utm_campaign      Field = 22
	Field_utm_content       Field = 23
	Field_utm_medium        Field = 24
	Field_utm_source        Field = 25
	Field_utm_term          Field = 26
	Field_subdivision1_code Field = 27
	Field_subdivision2_code Field = 28
)

// Enum value maps for Field.
var (
	Field_name = map[uint8]string{
		0:  "unknown",
		1:  "timestamp",
		2:  "id",
		3:  "bounce",
		4:  "session",
		5:  "view",
		6:  "duration",
		7:  "city",
		8:  "browser",
		9:  "browser_version",
		10: "country",
		11: "device",
		12: "domain",
		13: "entry_page",
		14: "name",
		15: "exit_page",
		16: "host",
		17: "os",
		18: "os_version",
		19: "page",
		20: "referrer",
		21: "source",
		22: "utm_campaign",
		23: "utm_content",
		24: "utm_medium",
		25: "utm_source",
		26: "utm_term",
		27: "subdivision1_code",
		28: "subdivision2_code",
	}
	Field_value = map[string]uint8{
		"unknown":           0,
		"timestamp":         1,
		"id":                2,
		"bounce":            3,
		"session":           4,
		"view":              5,
		"duration":          6,
		"city":              7,
		"browser":           8,
		"browser_version":   9,
		"country":           10,
		"device":            11,
		"domain":            12,
		"entry_page":        13,
		"name":              14,
		"exit_page":         15,
		"host":              16,
		"os":                17,
		"os_version":        18,
		"page":              19,
		"referrer":          20,
		"source":            21,
		"utm_campaign":      22,
		"utm_content":       23,
		"utm_medium":        24,
		"utm_source":        25,
		"utm_term":          26,
		"subdivision1_code": 27,
		"subdivision2_code": 28,
	}
)

func (f Field) TranslateIndex() byte {
	return byte(f - Field_browser)
}

func TranslateIndex(i int) Field {
	return Field(i) + Field_browser
}

func (f Field) String() string {
	return Field_name[uint8(f)]
}
