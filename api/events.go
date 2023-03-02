package api

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gernest/vince/country"
	"github.com/gernest/vince/gate"
	"github.com/gernest/vince/geoip"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/referrer"
	"github.com/gernest/vince/remoteip"
	"github.com/gernest/vince/timeseries"
	"github.com/gernest/vince/ua"
)

type Request struct {
	EventName   string            `json:"n"`
	URI         string            `json:"url"`
	Referrer    string            `json:"r"`
	Domain      string            `json:"d"`
	ScreenWidth int               `json:"w"`
	HashMode    bool              `json:"h"`
	Meta        map[string]string `json:"m"`
}

func Events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	xlg := log.Get(r.Context())
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		xlg.Err(err).Msg("Failed decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	remoteIp := remoteip.Get(r)
	if req.URI == "" {
		xlg.Debug().Msg("url is required")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uri, err := url.Parse(req.URI)
	if err != nil {
		xlg.Err(err).Msg("Failed parsing url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if uri.Scheme == "data" {
		xlg.Debug().Msg("url scheme is not allowed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	host := sanitizeHost(uri.Host)
	userAgent := r.UserAgent()

	reqReferrer := req.Referrer
	refUrl, err := url.Parse(reqReferrer)
	if err != nil {
		xlg.Err(err).Msg("Failed parsing referrer url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := uri.Path
	if req.HashMode && uri.Fragment != "" {
		path += "#" + uri.Fragment
	}
	if len(path) > 2000 {
		xlg.Debug().Msg("pathname too long")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.EventName == "" {
		xlg.Debug().Msg("event_name is required")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Domain == "" {
		xlg.Debug().Msg("domain is required")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var domains []string
	for _, d := range strings.Split(req.Domain, ",") {
		domains = append(domains, sanitizeHost(d))
	}

	query := r.URL.Query()
	now := time.Now()
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
		if ref != nil {
			if ref.Type == "unknown" {
				source = sanitizeHost(refUrl.Host)
			} else {
				source = ref.Type
			}
		}
	}
	reqReferrer = cleanReferrer(reqReferrer)

	var countryCode country.Code
	var cityGeonameId uint32
	if remoteIp != "" {
		ip := net.ParseIP(remoteIp)
		city, err := geoip.Lookup(ip)
		if err == nil {
			countryCode = country.Lookup(city.Country.IsoCode)
			cityGeonameId = uint32(city.Country.GeoNameID)
		}
	}
	var screenSize timeseries.ScreenSize
	switch {
	case req.ScreenWidth < 576:
		screenSize = timeseries.Mobile
	case req.ScreenWidth < 992:
		screenSize = timeseries.Tablet
	case req.ScreenWidth < 1440:
		screenSize = timeseries.Laptop
	case req.ScreenWidth >= 1440:
		screenSize = timeseries.Desktop
	}
	var dropped int
	for _, domain := range domains {
		b, pass := gate.Check(r.Context(), domain)
		if !pass {
			dropped += 1
			continue
		}
		userID := int64(seedID.Gen(remoteIp, userAgent, domain, host))
		e := new(timeseries.Event)
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
		e.CountryCode = countryCode
		e.CityGeoNameID = cityGeonameId
		e.ScreenSize = screenSize
		e.Timestamp = now
		previousUUserID := int64(seedID.GenPrevious(remoteIp, userAgent, domain, host))
		e.SessionId = b.Register(r.Context(), e, previousUUserID)
	}
	if dropped > 0 {
		w.Header().Set("x-vince-dropped", strconv.Itoa(dropped))
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
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
