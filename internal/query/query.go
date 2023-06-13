package query

import (
	"path"
	"regexp"
	"time"

	"github.com/dop251/goja"
)

type Query struct {
	Offset *Duration `json:"offset,omitempty"`
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

type Metrics struct {
	Visitors       *Select `json:"visitors,omitempty"`
	Views          *Select `json:"views,omitempty"`
	Events         *Select `json:"events,omitempty"`
	Visits         *Select `json:"visits,omitempty"`
	BounceRates    *Select `json:"bounceRates,omitempty"`
	VisitDurations *Select `json:"visitDurations,omitempty"`
}

type Select struct {
	Exact *Value `json:"exact,omitempty"`
	Re    *Value `json:"re,omitempty"`
	Glob  *Value `json:"glob,omitempty"`
	re    *regexp.Regexp
}

func (s *Select) Match(txt []byte) bool {
	if s.Exact != nil {
		return s.Exact.Value == string(txt)
	}
	if s.Glob != nil {
		ok, _ := path.Match(s.Glob.Value, string(txt))
		return ok
	}
	if s.Re != nil {
		if s.re == nil {
			s.re = regexp.MustCompile(s.Re.Value)
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
