package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func UnimplementedStat(db *db.Config, w http.ResponseWriter, r *http.Request) {
}

func Sources(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(db.Get(), r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		// params.Filter(),
		nil,
		metrics, ro2.SourceField)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		// params.Filter(),
		nil,
		metrics, ro2.Utm_mediumField)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		// params.Filter(),
		nil,
		metrics, ro2.Utm_campaignField)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		// params.Filter(),
		nil,
		metrics, ro2.Utm_contentField)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		// params.Filter(),
		nil,
		metrics, ro2.Utm_termField)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		// params.Filter(),
		nil,
		metrics, ro2.Utm_sourceField)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	referrer := r.PathValue("referrer")

	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	_ = referrer
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		// oracle.NewAnd(params.Filter(),
		// 	oracle.NewEq("referrer", referrer)),
		nil,
		metrics, ro2.PageField)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "pageviews", "bounce_rate")
	}
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		// params.Filter(),
		nil,
		metrics, ro2.PageField)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().Breakdown(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
		[]string{"visitors", "visits", "visit_duration"},
		ro2.Entry_pageField,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().BreakdownExitPages(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
	}
	db.JSON(w, o)
}

func Countries(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().Breakdown(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
		[]string{"visitors"},
		ro2.CountryField,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().Breakdown(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
		[]string{"visitors"},
		ro2.Subdivision1_codeField,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		code := m["subdivision1_codeField"].(string)
		delete(m, "subdivision1_codeField")
		reg := location.GetRegion(code)
		m["code"] = code
		m["name"] = reg.Name
		m["country_flag"] = reg.Flag
	}
	db.JSON(w, o)
}

func Cities(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().BreakdownCity(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
	}
	db.JSON(w, o)
}

func Browsers(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
		ro2.BrowserField,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
		ro2.Browser_versionField,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
		ro2.OsField,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
		ro2.Os_versionField,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
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
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(
		params.Start(),
		params.End(),
		site.Domain,
		// params.Filter(),
		nil,
		ro2.DeviceField,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["device"]
		delete(m, "device")
	}
	db.JSON(w, o)
}
