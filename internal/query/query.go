package query

import (
	"time"

	"github.com/dop251/goja"
)

type Query struct {
	Offset *Duration `json:"offset,omitempty"`
	Step   *Duration `json:"step,omitempty"`
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
}

type Value struct {
	Value string `json:"value,omitempty"`
}

type Duration struct {
	Value time.Duration `json:"value"`
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
	vm.Set("__Select__", func(call goja.ConstructorCall) *goja.Object {
		r := &Select{}
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
