package web

import (
	"math"
	"net/http"
	"slices"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func UnimplementedStat(db *db.Config, w http.ResponseWriter, r *http.Request) {
}

func MainGraph(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metric := params.Metric()

	result, err := db.Get().Timeseries(site.Domain, params, []string{metric})
	if err != nil {
		db.Logger().Error("reading main graph", "err", err)
		db.JSON(w, map[string]any{
			"labels":   []string{},
			"plot":     []float64{},
			"metric":   metric,
			"interval": params.Interval().String(),
		})
		return
	}
	size := len(result)
	labels := make([]string, 0, size)
	for k := range result {
		labels = append(labels, k)
	}
	slices.Sort(labels)
	plot := make([]float64, 0, size)
	var reduce func(r *ro2.Stats) float64
	switch metric {
	case "visitors":
		reduce = func(r *ro2.Stats) float64 { return r.Visitors }
	case "visits":
		reduce = func(r *ro2.Stats) float64 { return r.Visits }
	case "pageview":
		reduce = func(r *ro2.Stats) float64 { return r.PageViews }
	case "views_perVisit":
		reduce = func(r *ro2.Stats) float64 { return r.ViewsPerVisits }
	case "bounce_rate":
		reduce = func(r *ro2.Stats) float64 { return r.BounceRate }
	case "visit_duration":
		reduce = func(r *ro2.Stats) float64 { return r.VisitDuration }
	default:
		reduce = func(_ *ro2.Stats) float64 { return 0 }
	}
	for i := range labels {
		stat := result[labels[i]]
		stat.Compute()
		plot = append(plot, reduce(stat))
	}
	db.JSON(w, map[string]any{
		"labels":   labels,
		"plot":     plot,
		"metric":   metric,
		"interval": params.Interval().String(),
	})
}

var topFields = ro2.MetricsToProject(
	[]string{"visitors", "visits", "pageviews", "views_per_visit", "bounce_rate", "visit_duration"},
)

func TopStats(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())

	metrics := []string{"visitors", "visits", "pageviews", "views_per_visit", "bounce_rate", "visit_duration"}
	stats, err := db.Get().Stats(site.Domain, params.Start(), params.End(), params.Interval(), params.Filter(), metrics)
	if err != nil {
		db.Logger().Error("reading top stats", "err", err)
	}
	stats.Compute()
	cmp := new(ro2.Stats)

	if x := params.Compare(); x != nil {
		cmp, err = db.Get().Stats(site.Domain, x.Start, x.End, params.Interval(), params.Filter(), metrics)
		if err != nil {
			db.Logger().Error("reading top stats comparison", "err", err)
		}
	}
	cmp.Compute()
	db.JSON(w, map[string]any{
		"from":     params.From(),
		"to":       params.To(),
		"interval": params.Interval().String(),
		"top_stats": []any{
			entry(stats.Visitors, cmp.Visitors, "Unique visitors", "visitors"),
			entry(stats.Visits, cmp.Visits, "Total visits", "visits"),
			entry(stats.PageViews, cmp.PageViews, "Total pageviews", "pageviews"),
			entry(stats.ViewsPerVisits, cmp.ViewsPerVisits, "Views per visit", "views_per_visit"),
			entry(stats.BounceRate, cmp.BounceRate, "Bounce rate", "bounce_rate"),
			entry(stats.VisitDuration, cmp.VisitDuration, "Visit duration", "visit_duration"),
		},
	})
}

func entry(curr, prev float64, name, key string) map[string]any {
	m := map[string]any{
		"name":         name,
		"value":        curr,
		"graph_metric": key,
	}
	var change float64
	if key == "bounce_rate" {
		change = curr - prev
	} else {
		switch {
		case prev == 0 && curr > 0:
			change = 100
		case prev == 0 && curr == 0:
		default:
			change = math.Round((curr - prev) / prev * 100)
		}
	}
	m["comparison_value"] = prev
	m["change"] = change
	return m
}

func CurrentVisitors(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	visitors, err := db.Get().CurrentVisitors(site.Domain)
	if err != nil {
		db.Logger().Error("computing current visitors", "err", err)
	}
	db.JSON(w, visitors)
}

func Sources(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(site.Domain, params, metrics, alicia.SOURCE)
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
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(site.Domain, params, metrics, alicia.UTM_MEDIUM)
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
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(site.Domain, params, metrics, alicia.UTM_CAMPAIGN)
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
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(site.Domain, params, metrics, alicia.UTM_CONTENT)
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
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(site.Domain, params, metrics, alicia.UTM_TERM)
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
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(site.Domain, params, metrics, alicia.UTM_SOURCE)
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
	params := query.New(r.URL.Query())
	referrer := r.PathValue("referrer")

	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	_ = referrer
	o, err := db.Get().Breakdown(site.Domain, params, metrics, alicia.REFERRER)
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
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "pageviews", "bounce_rate")
	}
	o, err := db.Get().Breakdown(site.Domain, params, metrics, alicia.PAGE)
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
	params := query.New(r.URL.Query())
	o, err := db.Get().Breakdown(
		site.Domain,
		params,
		[]string{"visitors", "visits", "visit_duration"},
		alicia.ENTRY_PAGE,
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
	params := query.New(r.URL.Query())
	o, err := db.Get().BreakdownExitPages(site.Domain, params)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
	}
	db.JSON(w, o)
}

func Countries(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Get().Breakdown(site.Domain, params, []string{"visitors"}, alicia.COUNTRY)
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
		m["flag"] = c.Flag
	}
	db.JSON(w, o)
}

func Regions(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Get().Breakdown(site.Domain, params, []string{"visitors"}, alicia.SUB1_CODE)
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
		m["flag"] = reg.Flag
	}
	db.JSON(w, o)
}

func Cities(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Get().BreakdownCity(site.Domain, params)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &ro2.Result{}
	}
	db.JSON(w, o)
}

func Browsers(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(site.Domain, params, alicia.BROWSER)
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
	params := query.New(r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(site.Domain, params, alicia.BROWSER_VESRION)
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
	params := query.New(r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(site.Domain, params, alicia.OS)
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
	params := query.New(r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(site.Domain, params, alicia.OS_VERSION)
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
	params := query.New(r.URL.Query())
	o, err := db.Get().BreakdownVisitorsWithPercentage(site.Domain, params, alicia.DEVICE)
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
