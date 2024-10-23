package web

import (
	"math"
	"net/http"
	"slices"

	"github.com/vinceanalytics/vince/internal/api/aggregates"
	"github.com/vinceanalytics/vince/internal/api/breakdown"
	"github.com/vinceanalytics/vince/internal/api/timeseries"
	"github.com/vinceanalytics/vince/internal/api/visitors"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func UnimplementedStat(db *db.Config, w http.ResponseWriter, r *http.Request) {
}

func MainGraph(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metric := params.Metric()

	result, err := timeseries.Timeseries(ctx, db.TimeSeries(), site.Domain, params, []string{metric})
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
	reduce := aggregates.StatToValue(metric)
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

func TopStats(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())

	metrics := []string{"visitors", "visits", "pageviews", "views_per_visit", "bounce_rate", "visit_duration"}
	stats, err := aggregates.Aggregates(
		ctx, db.TimeSeries(),
		site.Domain, params.Start(), params.End(), params.Interval(), params.Filter(), metrics)
	if err != nil {
		db.Logger().Error("reading top stats", "err", err)
		stats = &aggregates.Stats{}
	}
	stats.Compute()
	cmp := new(aggregates.Stats)

	if x := params.Compare(); x != nil && !params.Realtime() {
		cmp, err = aggregates.Aggregates(
			ctx, db.TimeSeries(),
			site.Domain, x.Start, x.End, params.Interval(), params.Filter(), metrics)
		if err != nil {
			db.Logger().Error("reading top stats comparison", "err", err)
			cmp = &aggregates.Stats{}
		}
	}
	cmp.Compute()
	realtime := params.Realtime()
	db.JSON(w, map[string]any{
		"from":     params.From(),
		"to":       params.To(),
		"interval": params.Interval().String(),
		"top_stats": []any{
			entry(realtime, stats.Visitors, cmp.Visitors, "Unique visitors", "visitors"),
			entry(realtime, stats.Visits, cmp.Visits, "Total visits", "visits"),
			entry(realtime, stats.PageViews, cmp.PageViews, "Total pageviews", "pageviews"),
			entry(realtime, stats.ViewsPerVisits, cmp.ViewsPerVisits, "Views per visit", "views_per_visit"),
			entry(realtime, stats.BounceRate, cmp.BounceRate, "Bounce rate", "bounce_rate"),
			entry(realtime, stats.VisitDuration, cmp.VisitDuration, "Visit duration", "visit_duration"),
		},
	})
}

func entry(realtime bool, curr, prev float64, name, key string) map[string]any {
	m := map[string]any{
		"name":         name,
		"value":        curr,
		"graph_metric": key,
	}
	if realtime {
		return m
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
	visitors, err := visitors.Current(r.Context(), db.TimeSeries(), site.Domain)
	if err != nil {
		db.Logger().Error("computing current visitors", "err", err)
	}
	db.JSON(w, visitors)
}

func Sources(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, metrics, models.Field_source)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["source"]
		delete(m, "source")
	}
	db.JSON(w, o)
}

func UtmMediums(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, metrics, models.Field_utm_medium)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_medium"]
		delete(m, "utm_medium")
	}
	db.JSON(w, o)
}

func UtmCampaigns(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, metrics, models.Field_utm_campaign)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_campaign"]
		delete(m, "utm_campaign")
	}
	db.JSON(w, o)
}

func UtmContents(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, metrics, models.Field_utm_content)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_content"]
		delete(m, "utm_content")
	}
	db.JSON(w, o)
}

func UtmTerms(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, metrics, models.Field_utm_term)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_term"]
		delete(m, "utm_term")
	}
	db.JSON(w, o)
}

func UtmSources(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, metrics, models.Field_utm_source)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["utm_source"]
		delete(m, "utm_source")
	}
	db.JSON(w, o)
}

func Referrer(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	referrer := r.PathValue("referrer")
	if referrer == "Google" {
		db.JSONCode(http.StatusUnprocessableEntity, w, map[string]any{
			"error":    "The site is not connected to Google Search Keywords",
			"is_admin": db.CurrentUser() != "",
		})
		return
	}

	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params.With(&query.Filter{
		Op:    "is",
		Key:   models.Field_source.String(),
		Value: []string{referrer},
	}), metrics, models.Field_referrer)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["referrer"]
		delete(m, "referrer")
	}
	db.JSON(w, o)
}

func Pages(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "pageviews", "bounce_rate")
	}
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, metrics, models.Field_page)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["page"]
		delete(m, "page")
	}
	db.JSON(w, o)
}

func EntryPages(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(),
		site.Domain,
		params,
		[]string{"visitors", "visits", "visit_duration"},
		models.Field_entry_page,
	)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["entry_page"]
		delete(m, "entry_page")
	}
	db.JSON(w, o)
}

func ExitPages(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.BreakdownExitPages(ctx, db.TimeSeries(), site.Domain, params)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	db.JSON(w, o)
}

func Countries(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, []string{"visitors"}, models.Field_country)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		code := m[models.Field_country.String()].(string)
		delete(m, models.Field_country.String())
		c := location.GetCountry(code)
		m["code"] = code
		m["alpha_3"] = c.Alpha
		m["name"] = c.Name
		m["flag"] = c.Flag
	}
	db.JSON(w, o)
}

func Regions(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.Breakdown(ctx, db.TimeSeries(), site.Domain, params, []string{"visitors"}, models.Field_subdivision1_code)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		code := m[models.Field_subdivision1_code.String()].(string)
		delete(m, models.Field_subdivision1_code.String())
		reg := location.GetRegion([]byte(code))
		m["code"] = code
		m["name"] = reg.Name
		m["flag"] = reg.Flag
	}
	db.JSON(w, o)
}

func Cities(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.BreakdownCity(ctx, db.TimeSeries(), site.Domain, params, []string{"visitors"})
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	db.JSON(w, o)
}

func Browsers(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.BreakdownVisitorsWithPercentage(ctx, db.TimeSeries(), site.Domain, params, models.Field_browser)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["browser"]
		delete(m, "browser")
	}
	db.JSON(w, o)
}

func BrowserVersions(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.BreakdownVisitorsWithPercentage(ctx, db.TimeSeries(), site.Domain, params, models.Field_browser_version)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["browser_version"]
		delete(m, "browser_version")
	}
	db.JSON(w, o)
}

func Os(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.BreakdownVisitorsWithPercentage(ctx, db.TimeSeries(), site.Domain, params, models.Field_os)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["os"]
		delete(m, "os")
	}
	db.JSON(w, o)
}

func OsVersion(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.BreakdownVisitorsWithPercentage(ctx, db.TimeSeries(), site.Domain, params, models.Field_os_version)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["os_version"]
		delete(m, "os_version")
	}
	db.JSON(w, o)
}

func ScreenSize(db *db.Config, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := db.CurrentSite()
	params := query.New(r.URL.Query())
	o, err := breakdown.BreakdownVisitorsWithPercentage(ctx, db.TimeSeries(), site.Domain, params, models.Field_device)
	if err != nil {
		db.Logger().Error("breaking down", "err", err)
		o = &breakdown.Result{}
	}
	for i := range o.Results {
		m := o.Results[i]
		m["name"] = m["device"]
		delete(m, "device")
	}
	db.JSON(w, o)
}
