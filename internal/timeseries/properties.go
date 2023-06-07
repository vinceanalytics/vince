package timeseries

import (
	"encoding/json"
	"fmt"
)

type Property uint8

var _ json.Marshaler = (*Property)(nil)
var _ json.Unmarshaler = (*Property)(nil)

const (
	Base           Property = 0
	Event          Property = 1
	Page           Property = 2
	EntryPage      Property = 3
	ExitPage       Property = 4
	Referrer       Property = 5
	UtmMedium      Property = 6
	UtmSource      Property = 7
	UtmCampaign    Property = 8
	UtmContent     Property = 9
	UtmTerm        Property = 10
	UtmDevice      Property = 11
	UtmBrowser     Property = 12
	BrowserVersion Property = 13
	Os             Property = 14
	OsVersion      Property = 15
	Country        Property = 16
	Region         Property = 17
	City           Property = 18
)

const BaseKey = "__root__"

// Enum value maps for Property.
var (
	_prop_name = map[uint8]string{
		0:  "base",
		1:  "event",
		2:  "page",
		3:  "entryPage",
		4:  "exitPage",
		5:  "referrer",
		6:  "UtmMedium",
		7:  "UtmSource",
		8:  "UtmCampaign",
		9:  "UtmContent",
		10: "UtmTerm",
		11: "UtmDevice",
		12: "UtmBrowser",
		13: "browserVersion",
		14: "os",
		15: "osVersion",
		16: "country",
		17: "region",
		18: "city",
	}
	_prop_value = map[string]uint8{
		"base":           0,
		"event":          1,
		"page":           2,
		"entryPage":      3,
		"exitPage":       4,
		"referrer":       5,
		"UtmMedium":      6,
		"UtmSource":      7,
		"UtmCampaign":    8,
		"UtmContent":     9,
		"UtmTerm":        10,
		"UtmDevice":      11,
		"UtmBrowser":     12,
		"browserVersion": 13,
		"os":             14,
		"osVersion":      15,
		"country":        16,
		"region":         17,
		"city":           18,
	}
)

func (p Property) String() string {
	return _prop_name[uint8(p)]
}

func ParseProperty(k string) Property {
	p, ok := _prop_value[k]
	if !ok {
		return Base
	}
	return Property(p)
}

func (p Property) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Property) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	v, ok := _prop_value[s]
	if !ok {
		return fmt.Errorf("unknown property value %q", s)
	}
	*p = Property(v)
	return nil
}

type Metric uint8

var _ json.Marshaler = (*Metric)(nil)
var _ json.Unmarshaler = (*Metric)(nil)

const (
	Visitors       Metric = 0
	Views          Metric = 1
	Events         Metric = 2
	Visits         Metric = 3
	BounceRates    Metric = 4
	VisitDurations Metric = 5
)

// Enum value maps for Metric.
var (
	_metric_name = map[uint8]string{
		0: "visitors",
		1: "views",
		2: "events",
		3: "visits",
		4: "bounceRates",
		5: "visitDurations",
	}
	_metric_value = map[string]uint8{
		"visitors":       0,
		"views":          1,
		"events":         2,
		"visits":         3,
		"bounceRates":    4,
		"visitDurations": 5,
	}
)

func (m Metric) String() string {
	return _metric_name[uint8(m)]
}

func (m Metric) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *Metric) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	v, ok := _metric_value[s]
	if !ok {
		return fmt.Errorf("unknown metric value %q", s)
	}
	*m = Metric(v)
	return nil
}
