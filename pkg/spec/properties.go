package spec

import (
	"encoding/json"
	"fmt"
)

const BaseKey = "__root__"

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
