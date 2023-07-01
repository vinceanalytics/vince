package property

import (
	"encoding/json"
	"fmt"
)

type Property uint8

var _ json.Marshaler = (*Property)(nil)
var _ json.Unmarshaler = (*Property)(nil)

const (
	Base Property = iota
	Event
	Page
	EntryPage
	ExitPage
	Referrer
	UtmMedium
	UtmSource
	UtmCampaign
	UtmContent
	UtmTerm
	UtmDevice
	UtmBrowser
	BrowserVersion
	Os
	OsVersion
	Country
	Region
	City
)

const BaseKey = "__root__"

// Enum value maps for Property.
var (
	_prop_name = map[Property]string{
		Event:          "event",
		Page:           "page",
		EntryPage:      "entryPage",
		ExitPage:       "exitPage",
		Referrer:       "referrer",
		UtmMedium:      "utmMedium",
		UtmSource:      "utmSource",
		UtmCampaign:    "utmCampaign",
		UtmContent:     "utmContent",
		UtmTerm:        "utmTerm",
		UtmDevice:      "UtmDevice",
		UtmBrowser:     "utmBrowser",
		BrowserVersion: "browserVersion",
		Os:             "os",
		OsVersion:      "osVersion",
		Country:        "country",
		Region:         "region",
		City:           "city",
	}
	_prop_value = map[string]Property{
		"event":          Event,
		"page":           Page,
		"entryPage":      EntryPage,
		"exitPage":       ExitPage,
		"referrer":       Referrer,
		"utmMedium":      UtmMedium,
		"utmSource":      UtmSource,
		"utmCampaign":    UtmCampaign,
		"utmContent":     UtmContent,
		"utmTerm":        UtmTerm,
		"utmDevice":      UtmDevice,
		"utmBrowser":     UtmBrowser,
		"browserVersion": BrowserVersion,
		"os":             Os,
		"osVersion":      OsVersion,
		"country":        Country,
		"region":         Region,
		"city":           City,
	}
)

func (p Property) String() string {
	return _prop_name[p]
}

func ParseProperty(k string) Property {
	return _prop_value[k]
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
	Visitors Metric = iota
	Views
	Events
	Visits
	BounceRates
	VisitDurations
)

// Enum value maps for Metric.
var (
	_metric_name = map[Metric]string{
		Visitors:       "visitors",
		Views:          "views",
		Events:         "events",
		Visits:         "visits",
		BounceRates:    "bounceRates",
		VisitDurations: "visitDurations",
	}
	_metric_label = map[Metric]string{
		Visitors:       " Visitors",
		Views:          "Views",
		Events:         "Events",
		Visits:         "Sessions",
		BounceRates:    "Bounce",
		VisitDurations: "Duration",
	}
	_metric_value = map[string]Metric{
		"visitors":       Visitors,
		"views":          Views,
		"events":         Events,
		"visits":         Visits,
		"bounceRates":    BounceRates,
		"visitDurations": VisitDurations,
	}
)

func (m Metric) String() string {
	return _metric_name[m]
}

func (m Metric) Label() string {
	return _metric_label[m]
}

func ParsMetric(k string) Metric {
	return Metric(_metric_value[k])
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
