package plot

import (
	"encoding/json"
	"io"
	"math"
)

type Property uint

const (
	Event Property = iota
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
	UtmBrowserVersion
	Os
	OsVersion
	Country
	Region
	City
)

type Aggregate uint

const (
	Visitors Aggregate = iota
	Views
	Events
	Visits
	BounceRate
	VisitDuration
	ViewsPerVisit
)

type Data struct {
	All               *Aggr     `json:"all,omitempty"`
	Event             *EntryMap `json:"event,omitempty"`
	Page              *EntryMap `json:"page,omitempty"`
	EntryPage         *EntryMap `json:"entryPage,omitempty"`
	ExitPage          *EntryMap `json:"exitPage,omitempty"`
	Referrer          *EntryMap `json:"referrer,omitempty"`
	UtmMedium         *EntryMap `json:"utmMedium,omitempty"`
	UtmSource         *EntryMap `json:"utmSource,omitempty"`
	UtmCampaign       *EntryMap `json:"utmCampaign,omitempty"`
	UtmContent        *EntryMap `json:"utmContent,omitempty"`
	UtmTerm           *EntryMap `json:"utmTerm,omitempty"`
	UtmDevice         *EntryMap `json:"utmDevice,omitempty"`
	UtmBrowser        *EntryMap `json:"utmBrowser,omitempty"`
	UtmBrowserVersion *EntryMap `json:"utmBrowserVersion,omitempty"`
	Os                *EntryMap `json:"os,omitempty"`
	OsVersion         *EntryMap `json:"osVersion,omitempty"`
	Country           *EntryMap `json:"country,omitempty"`
	Region            *EntryMap `json:"region,omitempty"`
	City              *EntryMap `json:"city,omitempty"`
}

type AggregateEntry struct {
	Options   AggregateOptions `json:"-"`
	Prop      Property         `json:"-"`
	Aggregate Aggr             `json:"aggregate"`
}

type AggregateOptions struct {
	NoSum     bool
	NoPercent bool
}

type Entry struct {
	Sum    float64   `json:"sum"`
	Values []float64 `json:"values"`
}

func (e *Entry) Build() {
	for i := range e.Values {
		e.Sum += e.Values[i]
	}
}

type EntryMap struct {
	Entries map[string]*AggregateEntry `json:"entries,omitempty"`
	Sum     *Summary                   `json:"sum,omitempty"`
	Percent *Summary                   `json:"percent,omitempty"`
}

type Summary struct {
	Visitors      []*Item `json:"visitors,omitempty"`
	Views         []*Item `json:"views,omitempty"`
	Events        []*Item `json:"events,omitempty"`
	Visits        []*Item `json:"visits,omitempty"`
	BounceRate    []*Item `json:"bounce_rate,omitempty"`
	VisitDuration []*Item `json:"visitDuration,omitempty"`
	ViewsPerVisit []*Item `json:"viewsPerVisit,omitempty"`
}

type Aggr struct {
	Visitors      *Entry `json:"visitors,omitempty"`
	Views         *Entry `json:"views,omitempty"`
	Events        *Entry `json:"events,omitempty"`
	Visits        *Entry `json:"visits,omitempty"`
	BounceRate    *Entry `json:"bounce_rate,omitempty"`
	VisitDuration *Entry `json:"visitDuration,omitempty"`
	ViewsPerVisit *Entry `json:"viewsPerVisit,omitempty"`
}

func build(b ...builder) {
	for _, v := range b {
		v.Build()
	}
}

func (a *Aggr) Build() {
	if a == nil {
		return
	}
	build(
		a.Visitors, a.Views, a.Events, a.Visits, a.BounceRate, a.VisitDuration, a.ViewsPerVisit,
	)
}

type Item struct {
	Key   string  `json:"skey"`
	Value float64 `json:"value"`
}

type AggregateValues struct {
	Visitors      []float64 `json:"visitors,omitempty"`
	Views         []float64 `json:"views,omitempty"`
	Events        []float64 `json:"events,omitempty"`
	Visits        []float64 `json:"visits,omitempty"`
	BounceRate    []float64 `json:"bounce_rate,omitempty"`
	VisitDuration []float64 `json:"visitDuration,omitempty"`
	ViewsPerVisit []float64 `json:"viewsPerVisit,omitempty"`
}

func (d *Data) calendar(prop Property) *EntryMap {
	switch prop {
	case Event:
		if d.Event == nil {
			d.Event = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.Event
	case Page:
		if d.Page == nil {
			d.Page = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.Page
	case EntryPage:
		if d.EntryPage == nil {
			d.EntryPage = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.EntryPage
	case ExitPage:
		if d.ExitPage == nil {
			d.ExitPage = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.ExitPage
	case Referrer:
		if d.Referrer == nil {
			d.Referrer = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.Referrer
	case UtmMedium:
		if d.UtmMedium == nil {
			d.UtmMedium = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.UtmMedium
	case UtmSource:
		if d.UtmSource == nil {
			d.UtmSource = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.UtmSource
	case UtmCampaign:
		if d.UtmCampaign == nil {
			d.UtmCampaign = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.UtmCampaign
	case UtmContent:
		if d.UtmContent == nil {
			d.UtmContent = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.UtmContent
	case UtmTerm:
		if d.UtmTerm == nil {
			d.UtmTerm = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.UtmTerm
	case UtmDevice:
		if d.UtmDevice == nil {
			d.UtmDevice = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.UtmDevice
	case UtmBrowser:
		if d.UtmBrowser == nil {
			d.UtmBrowser = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.UtmBrowser
	case UtmBrowserVersion:
		if d.UtmBrowserVersion == nil {
			d.UtmBrowserVersion = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.UtmBrowserVersion
	case Os:
		if d.Os == nil {
			d.Os = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.Os
	case OsVersion:
		if d.OsVersion == nil {
			d.OsVersion = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.OsVersion
	case Country:
		if d.Country == nil {
			d.Country = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.Country
	case Region:
		if d.Region == nil {
			d.Region = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.Region
	case City:
		if d.City == nil {
			d.City = &EntryMap{
				Entries: map[string]*AggregateEntry{},
			}
		}
		return d.City
	default:
		return nil
	}
}

func (d *Data) Set(prop Property, key string, agg AggregateValues, o ...AggregateOptions) {
	cal := d.calendar(prop)
	a := &AggregateEntry{Prop: prop}
	if len(o) > 0 {
		a.Options = o[0]
	}
	if len(agg.Visitors) != 0 {
		a.Aggregate.Visitors = buildEntry(agg.Visitors)
	}
	if len(agg.Views) != 0 {
		a.Aggregate.Views = buildEntry(agg.Views)
	}
	if len(agg.Events) != 0 {
		a.Aggregate.Events = buildEntry(agg.Events)
	}
	if len(agg.Visits) != 0 {
		a.Aggregate.Visits = buildEntry(agg.Visits)
	}
	if len(agg.BounceRate) != 0 {
		a.Aggregate.BounceRate = buildEntry(agg.BounceRate)
	}
	if len(agg.VisitDuration) != 0 {
		a.Aggregate.VisitDuration = buildEntry(agg.VisitDuration)
	}
	if len(agg.ViewsPerVisit) != 0 {
		a.Aggregate.ViewsPerVisit = buildEntry(agg.ViewsPerVisit)
	}
	cal.Entries[key] = a
}

func buildEntry(a []float64) *Entry {
	return &Entry{
		Sum:    sum(a...),
		Values: a,
	}
}

func sum(ls ...float64) (s float64) {
	for i := 0; i < len(ls); i += 1 {
		s += ls[i]
	}
	return
}

func (e *EntryMap) Build() {
	if e == nil {
		return
	}
	var totals [ViewsPerVisit + 1]float64

	for k, v := range e.Entries {
		if v.Options.NoSum {
			continue
		}
		a := &v.Aggregate
		if e.Sum == nil {
			e.Sum = &Summary{}
		}
		sumProp(v.Prop, k, a.Visitors, &totals[Visitors], &e.Sum.Visitors)
		sumProp(v.Prop, k, a.Views, &totals[Views], &e.Sum.Views)
		sumProp(v.Prop, k, a.Events, &totals[Events], &e.Sum.Events)
		sumProp(v.Prop, k, a.Visits, &totals[Visits], &e.Sum.Visits)
		sumProp(v.Prop, k, a.BounceRate, &totals[BounceRate], &e.Sum.BounceRate)
		sumProp(v.Prop, k, a.VisitDuration, &totals[VisitDuration], &e.Sum.VisitDuration)
		sumProp(v.Prop, k, a.ViewsPerVisit, &totals[ViewsPerVisit], &e.Sum.ViewsPerVisit)
	}
	for k, v := range e.Entries {
		if v.Options.NoPercent {
			continue
		}
		a := &v.Aggregate
		if e.Percent == nil {
			e.Percent = &Summary{}
		}
		percentProp(v.Prop, k, a.Visitors, totals[Visitors], &e.Percent.Visitors)
		percentProp(v.Prop, k, a.Views, totals[Views], &e.Percent.Views)
		percentProp(v.Prop, k, a.Events, totals[Events], &e.Percent.Events)
		percentProp(v.Prop, k, a.Visits, totals[Visits], &e.Percent.Visits)
		percentProp(v.Prop, k, a.BounceRate, totals[BounceRate], &e.Percent.BounceRate)
		percentProp(v.Prop, k, a.VisitDuration, totals[VisitDuration], &e.Percent.VisitDuration)
		percentProp(v.Prop, k, a.ViewsPerVisit, totals[ViewsPerVisit], &e.Percent.ViewsPerVisit)
	}
}

func sumProp(prop Property, k string, e *Entry, totals *float64, f *[]*Item) {
	if e == nil {
		return
	}
	*totals += e.Sum
	*f = append(*f, &Item{
		Key:   k,
		Value: e.Sum,
	})
}

func percentProp(prop Property, k string, e *Entry, totals float64, f *[]*Item) {
	if e == nil {
		return
	}
	var p float64
	if totals != 0 {
		p = math.Round(e.Sum / totals * 100)
	}
	*f = append(*f, &Item{
		Key:   k,
		Value: p,
	})
}

type builder interface {
	Build()
}

func (d *Data) Build() *Data {
	d.All.Build()
	sum := func(ls ...*EntryMap) {
		for _, e := range ls {
			e.Build()
		}
	}
	sum(
		d.Event, d.Page, d.EntryPage, d.ExitPage,
		d.Referrer, d.UtmMedium, d.UtmSource, d.UtmCampaign,
		d.UtmContent, d.UtmTerm, d.UtmDevice, d.UtmBrowser,
		d.UtmBrowserVersion, d.Os, d.OsVersion, d.Country, d.Region, d.City,
	)
	return d
}

func (d *Data) Render(w io.Writer) error {
	sum := func(ls ...*EntryMap) {
		for _, e := range ls {
			e.Build()
		}
	}
	sum(
		d.Event, d.Page, d.EntryPage, d.ExitPage,
		d.Referrer, d.UtmMedium, d.UtmSource, d.UtmCampaign,
		d.UtmContent, d.UtmTerm, d.UtmDevice, d.UtmBrowser,
		d.UtmBrowserVersion, d.Os, d.OsVersion, d.Country, d.Region, d.City,
	)
	e := json.NewEncoder(w)
	return e.Encode(d)
}
