package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/oracle"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func UnimplementedStat(db *db.Config, w http.ResponseWriter, r *http.Request) {
}

func Sources(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Oracle().Breakdown(params.Start(), params.End(), site.Domain,
		params.Filter(), metrics, "source")
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["source"]
		delete(m, "source")
	}
	db.JSON(w, o)
}

func UtmMediums(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Oracle().Breakdown(params.Start(), params.End(), site.Domain,
		params.Filter(), metrics, "utm_medium")
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_medium"]
		delete(m, "utm_medium")
	}
	db.JSON(w, o)
}

func UtmCampaigns(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Oracle().Breakdown(params.Start(), params.End(), site.Domain,
		params.Filter(), metrics, "utm_campaign")
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_campaign"]
		delete(m, "utm_campaign")
	}
	db.JSON(w, o)
}

func UtmContents(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Oracle().Breakdown(params.Start(), params.End(), site.Domain,
		params.Filter(), metrics, "utm_content")
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_content"]
		delete(m, "utm_content")
	}
	db.JSON(w, o)
}

func UtmTerms(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Oracle().Breakdown(params.Start(), params.End(), site.Domain,
		params.Filter(), metrics, "utm_term")
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_term"]
		delete(m, "utm_term")
	}
	db.JSON(w, o)
}

func UtmSources(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Oracle().Breakdown(params.Start(), params.End(), site.Domain,
		params.Filter(), metrics, "utm_source")
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_source"]
		delete(m, "utm_source")
	}
	db.JSON(w, o)
}

func Referrer(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	referrer := r.PathValue("referrer")

	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Oracle().Breakdown(params.Start(), params.End(), site.Domain,
		oracle.NewAnd(params.Filter(),
			oracle.NewEq("referrer", referrer)), metrics, "page")
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["referrer"]
		delete(m, "referrer")
	}
	db.JSON(w, o)
}

func Pages(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "pageviews", "bounce_rate")
	}
	o, err := db.Oracle().Breakdown(params.Start(), params.End(), site.Domain, params.Filter(), metrics, "page")
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["page"]
		delete(m, "page")
	}
	db.JSON(w, o)
}

func EntryPages(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().Breakdown(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
		[]string{"visitors", "visits", "visit_duration"},
		"entry_page",
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &oracle.Breakdown{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["entry_page"]
		delete(m, "entry_page")
	}
	db.JSON(w, o)
}

func ExitPages(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Oracle().BreakdownExitPages(
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
