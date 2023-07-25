package api

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/gate"
	"github.com/vinceanalytics/vince/internal/geoip"
	"github.com/vinceanalytics/vince/internal/referrer"
	"github.com/vinceanalytics/vince/internal/remoteip"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/ua"
	"github.com/vinceanalytics/vince/internal/userid"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/log"
)

// Events accepts events payloads from vince client script.
func Events(w http.ResponseWriter, r *http.Request) {
	system.DataPointReceived.Inc()

	w.Header().Set("X-Content-Type-Options", "nosniff")
	xlg := log.Get()
	req := entry.NewRequest()
	defer req.Release()

	err := req.Parse(r.Body)
	if err != nil {
		system.DataPointRejected.Inc()
		xlg.Err(err).Msg("Failed decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteIp := remoteip.Get(r)
	if req.URI == "" {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uri, err := url.Parse(req.URI)
	if err != nil {
		system.DataPointRejected.Inc()
		xlg.Err(err).Msg("Failed parsing url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if uri.Scheme == "data" {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	host := sanitizeHost(uri.Host)
	userAgent := r.UserAgent()

	ref, src, err := refSource(r.URL.Query(), req.Referrer)
	if err != nil {
		system.DataPointRejected.Inc()
		xlg.Err(err).Msg("Failed parsing referrer url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := uri.Path
	if req.HashMode && uri.Fragment != "" {
		path += "#" + uri.Fragment
	}
	if len(path) > 2000 {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.EventName == "" {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Domain == "" {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	domain := req.Domain

	query := uri.Query()
	agent := ua.Parse(userAgent)

	var city geoip.Info
	if remoteIp != "" {
		ip := net.ParseIP(remoteIp)
		city = geoip.Lookup(ip)
	}
	var screenSize string
	switch {
	case req.ScreenWidth < 576:
		screenSize = "mobile"
	case req.ScreenWidth < 992:
		screenSize = "tablet"
	case req.ScreenWidth < 1440:
		screenSize = "laptop"
	case req.ScreenWidth >= 1440:
		screenSize = "desktop"
	}
	ctx := r.Context()
	ts := core.Now(ctx)
	pass := gate.Check(r.Context(), domain)
	if !pass {
		system.DataPointDropped.Inc()
		w.Header().Set("x-vince-dropped", strconv.Itoa(dropped))
		w.WriteHeader(http.StatusAccepted)
		return
	}
	userID := userid.Hash(remoteIp, userAgent, domain, host)
	e := entry.NewEntry()
	e.ID = userID
	e.Name = req.EventName
	e.Host = host
	e.Path = path
	e.Domain = domain
	e.UtmMedium = query.Get("utm_medium")
	e.UtmSource = query.Get("utm_source")
	e.UtmCampaign = query.Get("utm_campaign")
	e.UtmContent = query.Get("utm_content")
	e.UtmTerm = query.Get("utm_content")
	e.Os = agent.Os
	e.OsVersion = agent.OsVersion
	e.Browser = agent.Browser
	e.BrowserVersion = agent.BrowserVersion
	e.ReferrerSource = src
	e.Referrer = ref
	e.Country = city.Country
	e.Region = city.Region
	e.City = city.City
	e.Screen = screenSize
	e.Timestamp = ts
	timeseries.Register(ctx, e)

	system.DataPointAccepted.Inc()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func sanitizeHost(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(s), "www.")
}

func refSource(q url.Values, u string) (ref, source string, err error) {
	r, err := url.Parse(u)
	if err != nil {
		return "", "", err
	}
	r.Host = sanitizeHost(r.Host)
	r.Path = strings.TrimSuffix(r.Path, "/")
	ref = r.String()
	source = q.Get("utm_source")
	if source == "" {
		source = q.Get("source")
	}
	if source == "" {
		source = q.Get("ref")
	}
	if source != "" {
		return
	}
	source = referrer.Parse(r.Host)
	return
}
