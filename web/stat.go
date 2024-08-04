package web

import (
	"net/http"

	"github.com/gernest/len64/internal/location"
	"github.com/gernest/len64/internal/oracle"
	"github.com/gernest/len64/web/db"
	"github.com/gernest/len64/web/query"
)

func UnimplementedStat(db *db.Config, w http.ResponseWriter, r *http.Request) {
}

func Countries(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().Breakdown(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
		[]string{"visitors"},
		"country",
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		code := m["country"].(string)
		c := location.GetCountry(code)
		m["code"] = code
		m["name"] = c.Name
		m["country_flag"] = c.Flag
	}
	db.JSON(w, o)
}

func Regions(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().Breakdown(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
		[]string{"visitors"},
		"region",
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		code := m["region"].(string)
		reg := location.GetRegion(code)
		m["code"] = code
		m["name"] = reg.Name
		m["country_flag"] = reg.Flag
	}
	db.JSON(w, o)
}

func Cities(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().BreakdownCity(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	db.JSON(w, o)
}

func Browsers(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
		"browser",
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["browser"]
		delete(m, "browser")
	}
	db.JSON(w, o)
}

func BrowserVersions(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
		"browser_version",
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["browser_version"]
		delete(m, "browser_version")
	}
	db.JSON(w, o)
}

func Os(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
		"os",
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["os"]
		delete(m, "os")
	}
	db.JSON(w, o)
}

func OsVersion(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
		"os_version",
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["os_version"]
		delete(m, "os_version")
	}
	db.JSON(w, o)
}

func ScreenSize(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
		"device",
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["device"]
		delete(m, "device")
	}
	db.JSON(w, o)
}
