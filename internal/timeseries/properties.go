package timeseries

import "github.com/vinceanalytics/vince/pkg/property"

type Property = property.Property

const (
	Base           = property.Base
	Event          = property.Event
	Page           = property.Page
	EntryPage      = property.EntryPage
	ExitPage       = property.ExitPage
	Referrer       = property.Referrer
	UtmMedium      = property.UtmMedium
	UtmSource      = property.UtmSource
	UtmCampaign    = property.UtmCampaign
	UtmContent     = property.UtmContent
	UtmTerm        = property.UtmTerm
	UtmDevice      = property.UtmDevice
	UtmBrowser     = property.UtmBrowser
	BrowserVersion = property.BrowserVersion
	Os             = property.Os
	OsVersion      = property.OsVersion
	Country        = property.Country
	Region         = property.Region
	City           = property.City
)

const BaseKey = property.BaseKey

type Metric = property.Metric

const (
	Visitors       = property.Visitors
	Views          = property.Views
	Events         = property.Events
	Visits         = property.Visits
	BounceRates    = property.BounceRates
	VisitDurations = property.VisitDurations
)
