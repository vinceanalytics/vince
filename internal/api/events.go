package api

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/gate"
	"github.com/vinceanalytics/vince/internal/geoip"
	"github.com/vinceanalytics/vince/internal/referrer"
	"github.com/vinceanalytics/vince/internal/remoteip"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/internal/ua"
	"github.com/vinceanalytics/vince/internal/userid"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/log"
)

// Events accepts events payloads from vince client script.
func Events(w http.ResponseWriter, r *http.Request) {
	system.DataPoint.WithLabelValues("received").Inc()

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

	reqReferrer := req.Referrer
	refUrl, err := url.Parse(reqReferrer)
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
	var domains []string
	for _, d := range strings.Split(req.Domain, ",") {
		domains = append(domains, sanitizeHost(d))
	}

	query := uri.Query()
	agent := ua.Parse(userAgent)
	// handle referrer
	ref := referrer.ParseReferrer(req.Referrer)
	source := query.Get("utm_source")
	if source == "" {
		source = query.Get("source")
	}
	if source == "" {
		source = query.Get("ref")
	}
	if source == "" {
		source = ref
		if source == "" {
			source = sanitizeHost(refUrl.Host)
		}
	}
	reqReferrer = cleanReferrer(reqReferrer)

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
	var dropped int
	ts := time.Now()
	unix := ts.Unix()
	ctx := r.Context()
	uid := userid.Get(ctx)
	for _, domain := range domains {
		b, pass := gate.Check(r.Context(), domain)
		if !pass {
			dropped += 1
			continue
		}
		userID := uid.Hash(remoteIp, userAgent, domain, host)
		e := entry.NewEntry()
		e.UserId = userID
		e.Name = req.EventName
		e.Hostname = host
		e.Domain = domain
		e.Pathname = path
		e.UtmMedium = query.Get("utm_medium")
		e.UtmSource = query.Get("utm_source")
		e.UtmCampaign = query.Get("utm_campaign")
		e.UtmContent = query.Get("utm_content")
		e.UtmTerm = query.Get("utm_content")
		e.OperatingSystem = agent.Os
		e.OperatingSystemVersion = agent.OsVersion
		e.Browser = agent.Browser
		e.BrowserVersion = agent.BrowserVersion
		e.ReferrerSource = source
		e.Referrer = reqReferrer
		e.Country = city.Country
		e.Region = city.Region
		e.City = city.City
		e.ScreenSize = screenSize
		e.Timestamp = unix
		previousUUserID := uid.HashPrevious(remoteIp, userAgent, domain, host)
		b.Register(r.Context(), e, previousUUserID)
	}
	if dropped > 0 {
		system.DataPointDropped.Inc()
		w.Header().Set("x-vince-dropped", strconv.Itoa(dropped))
		w.WriteHeader(http.StatusAccepted)
		return
	}
	system.DataPointAccepted.Inc()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func sanitizeHost(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(s), "www.")
}

func cleanReferrer(s string) string {
	r, _ := url.Parse(s)
	r.Host = sanitizeHost(r.Host)
	r.Path = strings.TrimSuffix(s, "/")
	return r.String()
}
