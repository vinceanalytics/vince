package web

import (
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func UnimplementedStat(db *db.Config, w http.ResponseWriter, r *http.Request) {
}

func MainGraph(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(db.Get(), r.URL.Query())
	interval := params.Interval()
	shards := db.Get().Shards()
	slices.Sort(shards)
	quantim := roaring64.NewDefaultBSI()
	m := ro2.NewData()
	defer m.Release()

	err := db.Get().View(func(tx *ro2.Tx) error {
		dom := ro2.NewEq(uint64(alicia.DOMAIN), site.Domain)
		fields := ro2.MetricsToProject([]string{params.Metric()})
		tsField := interval.Field()
		start := params.Start()
		end := params.End()
		for _, shard := range shards {
			b := dom.Match(tx, shard)
			if b.IsEmpty() {
				continue
			}
			ts := tx.Cmp(uint64(tsField), shard, roaring64.RANGE, start, end)
			if ts.IsEmpty() {
				continue
			}
			match := roaring64.And(b, ts)
			if match.IsEmpty() {
				continue
			}
			m.ReadFields(tx, shard, match, fields...)
			tx.ExtractBSI(shard, uint64(tsField), match, quantim.SetValue)
		}
		return nil
	})
	if err != nil {
		db.Logger().Error("reading main graph", "err", err)
	}
	ts := quantim.Transpose()
	size := ts.GetCardinality()
	labels := make([]string, 0, size)
	plot := make([]float64, 0, size)

	it := ts.Iterator()
	format := interval.Format()
	metric := params.Metric()
	for it.HasNext() {
		tsv := it.Next()
		labels = append(labels,
			time.UnixMilli(int64(tsv)).UTC().Format(format))
		match := quantim.CompareValue(0, roaring64.EQ, int64(tsv), 0, nil)
		plot = append(plot, m.Compute(metric, match))

	}
	db.JSON(w, map[string]any{
		"labels": labels,
		"plot":   plot,
		"metric": metric,
	})
}

var topFields = ro2.MetricsToProject(
	[]string{"visitors", "visits", "pageviews", "views_per_visit", "bounce_rate", "visit_duration"},
)

func TopStats(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	params := query.New(db.Get(), r.URL.Query())
	m := ro2.NewData()
	defer m.Release()
	err := db.Get().Select(
		params.Start(), params.End(), site.Domain, params.Filter(),
		func(tx *ro2.Tx, shard uint64, match *roaring64.Bitmap) error {
			m.ReadFields(tx, shard, match, topFields...)
			return nil
		},
	)
	if err != nil {
		db.Logger().Error("reading top stats", "err", err)
	}

	old := ro2.NewData()
	defer old.Release()

	if cmp := params.Compare(); cmp != nil {
		err := db.Get().Select(
			cmp.Start.UnixMilli(),
			cmp.End.UnixMilli(),
			site.Domain, params.Filter(),
			func(tx *ro2.Tx, shard uint64, match *roaring64.Bitmap) error {
				old.ReadFields(tx, shard, match, topFields...)
				return nil
			},
		)
		if err != nil {
			db.Logger().Error("reading top stats comparison", "err", err)
		}
	}
	visitors := m.Visitors(nil)
	visits := m.Visits(nil)
	views := m.View(nil)
	viewsPerVisit := math.Floor(per(views, visits))
	bounceRate := math.Floor(per(m.Bounce(nil), visits) * 100)
	duration := time.Duration(m.Duration(nil)).Seconds()
	if visits != 0 {
		duration = duration / float64(visits)
	}
	db.JSON(w, map[string]any{
		"from":     params.From(),
		"to":       params.To(),
		"interval": params.Interval().String(),
		"top_stats": []any{
			map[string]any{
				"name":  "Unique visitors",
				"value": visitors,
			},
			map[string]any{
				"name":  "Total visits",
				"value": views,
			},
			map[string]any{
				"name":  "Total pageviews",
				"value": views,
			},
			map[string]any{
				"name":  "Views per visit",
				"value": viewsPerVisit,
			},
			map[string]any{
				"name":  "Bounce rate",
				"value": bounceRate,
			},
			map[string]any{
				"name":  "Visit duration",
				"value": duration,
			},
		},
	})
}

func per(a, b uint64) float64 {
	if b == 0 {
		return float64(a)
	}
	return float64(a) / float64(b)
}

func change(old, new float64) float64 {
	switch {
	case old == 0 && new > 0:
		return 100
	case old == 0 && new == 0:
		return 0
	default:
		return math.Round((new - old) / old * 100)
	}
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
	params := query.New(db.Get(), r.URL.Query())
	metrics := []string{"visitors"}
	if r.URL.Query().Get("detailed") != "" {
		metrics = append(metrics, "bounce_rate", "visit_duration")
	}
	o, err := db.Get().Breakdown(params.Start(), params.End(), site.Domain,
		params.Filter(),
		metrics, alicia.SOURCE)
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
		params.Filter(),
		metrics, alicia.UTM_MEDIUM)
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
		params.Filter(),
		metrics, alicia.UTM_CAMPAIGN)
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
		params.Filter(),
		metrics, alicia.UTM_CONTENT)
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
		params.Filter(),
		metrics, alicia.UTM_TERM)
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
		params.Filter(),
		metrics, alicia.UTM_SOURCE)
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
		append(ro2.List{params.Filter()},
			ro2.NewEq(uint64(alicia.REFERRER), referrer)),
		metrics, alicia.PAGE)
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
		params.Filter(),
		metrics, alicia.PAGE)
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
		params.Filter(),
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
	params := query.New(db.Get(), r.URL.Query())
	o, err := db.Get().BreakdownExitPages(
		params.Start(),
		params.End(),
		site.Domain,
		params.Filter(),
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
		params.Filter(),
		[]string{"visitors"},
		alicia.COUNTRY,
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
		m["flag"] = c.Flag
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
		params.Filter(),
		[]string{"visitors"},
		alicia.SUB1_CODE,
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
		m["flag"] = reg.Flag
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
		params.Filter(),
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
		params.Filter(),
		alicia.BROWSER,
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
		params.Filter(),
		alicia.BROWSER_VESRION,
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
		params.Filter(),
		alicia.OS,
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
		params.Filter(),
		alicia.OS_VERSION,
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
		params.Filter(),
		alicia.DEVICE,
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
