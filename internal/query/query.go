package query

import (
	"path"
	"regexp"
	"time"

	"github.com/dop251/goja"
	"github.com/vinceanalytics/vince/pkg/log"
)

type Query struct {
	Offset *Duration `json:"offset,omitempty"`
	Sum    bool      `json:"sum,omitempty"`
	Window *Duration `json:"window,omitempty"`
	Props  *Props    `json:"props,omitempty"`
}

type Props struct {
	Base           *Metrics `json:"base,omitempty"`
	Event          *Metrics `json:"event,omitempty"`
	Page           *Metrics `json:"page,omitempty"`
	EntryPage      *Metrics `json:"entryPage,omitempty"`
	ExitPage       *Metrics `json:"exitPage,omitempty"`
	Referrer       *Metrics `json:"referrer,omitempty"`
	UtmMedium      *Metrics `json:"utmMedium,omitempty"`
	UtmSource      *Metrics `json:"utmSource,omitempty"`
	UtmCampaign    *Metrics `json:"utmCampaign,omitempty"`
	UtmContent     *Metrics `json:"utmContent,omitempty"`
	UtmTerm        *Metrics `json:"utmTerm,omitempty"`
	UtmDevice      *Metrics `json:"UtmDevice,omitempty"`
	UtmBrowser     *Metrics `json:"utmBrowser,omitempty"`
	BrowserVersion *Metrics `json:"browserVersion,omitempty"`
	Os             *Metrics `json:"os,omitempty"`
	OsVersion      *Metrics `json:"osVersion,omitempty"`
	Country        *Metrics `json:"country,omitempty"`
	Region         *Metrics `json:"region,omitempty"`
	City           *Metrics `json:"city,omitempty"`
}

func (p *Props) Set(prop string, m *Metrics) {
	switch prop {
	case "base":
		p.Base = m
	case "event":
		p.Event = m
	case "page":
		p.Page = m
	case "entryPage":
		p.EntryPage = m
	case "exitPage":
		p.ExitPage = m
	case "referrer":
		p.Referrer = m
	case "utmMedium":
		p.UtmMedium = m
	case "utmSource":
		p.UtmSource = m
	case "utmCampaign":
		p.UtmCampaign = m
	case "utmContent":
		p.UtmContent = m
	case "utmTerm":
		p.UtmTerm = m
	case "utmDevice":
		p.UtmDevice = m
	case "utmBrowser":
		p.UtmBrowser = m
	case "browserVersion":
		p.BrowserVersion = m
	case "os":
		p.Os = m
	case "osVersion":
		p.OsVersion = m
	case "country":
		p.Country = m
	case "region":
		p.Region = m
	case "city":
		p.City = m
	}
}

func (p *Props) All(f func(string, *Metrics)) {
	p.Base = &Metrics{}
	f("base", p.Page)
	p.Event = &Metrics{}
	f("base", p.Event)
	p.Page = &Metrics{}
	f("base", p.Page)
	p.EntryPage = &Metrics{}
	f("EntryPage", p.EntryPage)
	p.ExitPage = &Metrics{}
	f("ExitPage", p.ExitPage)
	p.Referrer = &Metrics{}
	f("Referrer", p.Referrer)
	p.UtmMedium = &Metrics{}
	f("UtmMedium", p.UtmMedium)
	p.UtmSource = &Metrics{}
	f("UtmSource", p.UtmSource)
	p.UtmCampaign = &Metrics{}
	f("UtmCampaign", p.UtmCampaign)
	p.UtmContent = &Metrics{}
	f("UtmContent", p.UtmContent)
	p.UtmTerm = &Metrics{}
	f("UtmTerm", p.UtmTerm)
	p.UtmDevice = &Metrics{}
	f("UtmDevice", p.UtmDevice)
	p.UtmBrowser = &Metrics{}
	f("UtmBrowser", p.UtmBrowser)
	p.BrowserVersion = &Metrics{}
	f("BrowserVersion", p.BrowserVersion)
	p.Os = &Metrics{}
	f("Os", p.Os)
	p.OsVersion = &Metrics{}
	f("OsVersion", p.OsVersion)
	p.Country = &Metrics{}
	f("Country", p.Country)
	p.Region = &Metrics{}
	f("Region", p.Region)
	p.City = &Metrics{}
	f("City", p.City)
}

type Metrics struct {
	Visitors       *Select `json:"visitors,omitempty"`
	Views          *Select `json:"views,omitempty"`
	Events         *Select `json:"events,omitempty"`
	Visits         *Select `json:"visits,omitempty"`
	BounceRates    *Select `json:"bounceRates,omitempty"`
	VisitDurations *Select `json:"visitDurations,omitempty"`
}

func (m *Metrics) Set(key string, sel *Select) {
	switch key {
	case "visitors":
		m.Visitors = sel
	case "views":
		m.Views = sel
	case "events":
		m.Events = sel
	case "visits":
		m.Visits = sel
	case "bounceRates":
		m.BounceRates = sel
	case "visitDurations":
		m.VisitDurations = sel
	}
}

type Select struct {
	Exact   *Value `json:"exact,omitempty"`
	Re      *Value `json:"re,omitempty"`
	Glob    *Value `json:"glob,omitempty"`
	re      *regexp.Regexp
	invalid bool
}

func (s *Select) Equal(o *Select) bool {
	if s.Exact != nil && o.Exact != nil {
		return s.Exact.Value == o.Exact.Value
	}
	if s.Re != nil && o.Re != nil {
		return s.Re.Value == o.Re.Value
	}
	if s.Glob != nil && o.Glob != nil {
		return s.Glob.Value == o.Glob.Value
	}
	return false
}

func (s *Select) Match(txt []byte) bool {
	if s.invalid {
		return false
	}
	if s.Exact != nil {
		return s.Exact.Value == string(txt)
	}
	if s.Glob != nil {
		ok, _ := path.Match(s.Glob.Value, string(txt))
		return ok
	}
	if s.Re != nil {
		if s.re == nil {
			var err error
			s.re, err = regexp.Compile(s.Re.Value)
			if err != nil {
				log.Get().Err(err).
					Str("pattern", s.Re.Value).
					Msg("failed to compile regular expression for selector")
				s.invalid = true
				return false
			}
		}
		return s.re.Match(txt)
	}
	return true
}

type Value struct {
	Value string `json:"value,omitempty"`
}

type Duration struct {
	Value time.Duration `json:"value"`
}

type QueryResult struct {
	Timestamps []int64     `json:"timestamps"`
	Props      PropsResult `json:"props"`
}

type PropsResult struct {
	Base           *MetricsResult `json:"base,omitempty"`
	Event          *MetricsResult `json:"event,omitempty"`
	Page           *MetricsResult `json:"page,omitempty"`
	EntryPage      *MetricsResult `json:"entryPage,omitempty"`
	ExitPage       *MetricsResult `json:"exitPage,omitempty"`
	Referrer       *MetricsResult `json:"referrer,omitempty"`
	UtmMedium      *MetricsResult `json:"utmMedium,omitempty"`
	UtmSource      *MetricsResult `json:"utmSource,omitempty"`
	UtmCampaign    *MetricsResult `json:"utmCampaign,omitempty"`
	UtmContent     *MetricsResult `json:"utmContent,omitempty"`
	UtmTerm        *MetricsResult `json:"utmTerm,omitempty"`
	UtmDevice      *MetricsResult `json:"UtmDevice,omitempty"`
	UtmBrowser     *MetricsResult `json:"utmBrowser,omitempty"`
	BrowserVersion *MetricsResult `json:"browserVersion,omitempty"`
	Os             *MetricsResult `json:"os,omitempty"`
	OsVersion      *MetricsResult `json:"osVersion,omitempty"`
	Country        *MetricsResult `json:"country,omitempty"`
	Region         *MetricsResult `json:"region,omitempty"`
	City           *MetricsResult `json:"city,omitempty"`
}

type MetricsResult struct {
	Visitors       map[string][]uint32 `json:"visitors,omitempty"`
	Views          map[string][]uint32 `json:"views,omitempty"`
	Events         map[string][]uint32 `json:"events,omitempty"`
	Visits         map[string][]uint32 `json:"visits,omitempty"`
	BounceRates    map[string][]uint32 `json:"bounceRates,omitempty"`
	VisitDurations map[string][]uint32 `json:"visitDurations,omitempty"`
}

func Register(vm *goja.Runtime) {
	vm.Set("__Duration__", func(call goja.ConstructorCall) *goja.Object {
		o, err := time.ParseDuration(call.Arguments[0].String())
		if err != nil {
			return vm.NewGoError(err)
		}
		r := &Duration{Value: o}
		v := vm.ToValue(r).(*goja.Object)
		v.SetPrototype(call.This.Prototype())
		return v
	})
	vm.Set("__Query__", func(call goja.ConstructorCall) *goja.Object {
		r := &Query{}
		v := vm.ToValue(r).(*goja.Object)
		v.SetPrototype(call.This.Prototype())
		return v
	})
	vm.Set("__Props__", func(call goja.ConstructorCall) *goja.Object {
		r := &Props{}
		v := vm.ToValue(r).(*goja.Object)
		v.SetPrototype(call.This.Prototype())
		return v
	})
	vm.Set("__Metrics__", func(call goja.ConstructorCall) *goja.Object {
		r := &Metrics{}
		v := vm.ToValue(r).(*goja.Object)
		v.SetPrototype(call.This.Prototype())
		return v
	})
	vm.Set("__SelectExact__", func(call goja.ConstructorCall) *goja.Object {
		r := &Select{
			Exact: &Value{
				Value: call.Arguments[0].String(),
			},
		}
		v := vm.ToValue(r).(*goja.Object)
		v.SetPrototype(call.This.Prototype())
		return v
	})
	vm.Set("__SelectRe__", func(call goja.ConstructorCall) *goja.Object {
		r := &Select{
			Re: &Value{
				Value: call.Arguments[0].String(),
			},
		}
		var err error
		r.re, err = regexp.Compile(r.Re.Value)
		if err != nil {
			return vm.NewGoError(err)
		}
		v := vm.ToValue(r).(*goja.Object)
		v.SetPrototype(call.This.Prototype())
		return v
	})
	vm.Set("__SelectGlob__", func(call goja.ConstructorCall) *goja.Object {
		r := &Select{
			Glob: &Value{
				Value: call.Arguments[0].String(),
			},
		}
		v := vm.ToValue(r).(*goja.Object)
		v.SetPrototype(call.This.Prototype())
		return v
	})
	vm.Set("__Value__", func(call goja.ConstructorCall) *goja.Object {
		r := &Select{}
		v := vm.ToValue(r).(*goja.Object)
		v.SetPrototype(call.This.Prototype())
		return v
	})
}
