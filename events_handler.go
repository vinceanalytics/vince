package vince

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gernest/vince/geoip"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (v *Vince) EventsEndpoint(w http.ResponseWriter, r *http.Request) {
	if !v.processEvent(r) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (v *Vince) processEvent(r *http.Request) bool {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		xlg.Err(err).Msg("Failed decoding json")
		return false
	}
	remoteIp := GetRemoteIP(r)
	if req.URI == "" {
		xlg.Debug().Msg("url is required")
		return false
	}
	uri, err := url.Parse(req.URI)
	if err != nil {
		xlg.Err(err).Msg("Failed parsing url")
		return false
	}
	if uri.Scheme == "data" {
		xlg.Debug().Msg("url scheme is not allowed")
		return false
	}

	host := sanitizeHost(uri.Host)
	userAgent := r.UserAgent()

	referrer := req.Referrer
	refUrl, err := url.Parse(referrer)
	if err != nil {
		xlg.Err(err).Msg("Failed parsing referrer url")
		return false
	}
	path := uri.Path
	if req.HashMode && uri.Fragment != "" {
		path += "#" + uri.Fragment
	}
	if len(path) > 2000 {
		xlg.Debug().Msg("pathname too long")
		return false
	}
	if req.EventName == "" {
		xlg.Debug().Msg("event_name is required")
		return false
	}
	if req.Domain == "" {
		xlg.Debug().Msg("domain is required")
		return false
	}
	var domains []string
	for _, d := range strings.Split(req.Domain, ",") {
		domains = append(domains, sanitizeHost(d))
	}

	query := r.URL.Query()
	now := time.Now()
	var operatingSystem string
	var operatingSystemVersion string
	var browser string
	var browserVersion string
	if ua := parseUA(userAgent); ua != nil {
		if ua.os != nil {
			operatingSystem = ua.os.name
			operatingSystemVersion = ua.os.version
		}
		if ua.client != nil {
			browser = ua.client.name
			browserVersion = ua.client.version
		}
	}
	// handle referrer
	ref := ParseReferrer(req.Referrer)
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
	referrer = cleanReferrer(referrer)

	var countryCode string
	var cityGeonameId uint32
	if remoteIp != "" {
		ip := net.ParseIP(remoteIp)
		city, err := geoip.Lookup(ip)
		if err == nil {
			countryCode = city.Country.IsoCode
			cityGeonameId = uint32(city.Country.GeoNameID)
		}
	}
	var screenSize string
	switch {
	case req.ScreenWidth < 576:
		screenSize = "Mobile"
	case req.ScreenWidth < 992:
		screenSize = "Tablet"
	case req.ScreenWidth < 1440:
		screenSize = "Laptop"
	case req.ScreenWidth >= 1440:
		screenSize = "Desktop"
	}
	var labels []*Label
	for k, v := range req.Meta {
		if k == "" || v == "" {
			continue
		}
		labels = append(labels, &Label{
			Name: k, Value: v,
		})
	}

	for _, domain := range domains {
		userID := seedID.Gen(remoteIp, userAgent, domain, host)
		e := GetEvent()
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
		e.OperatingSystem = operatingSystem
		e.OperatingSystemVersion = operatingSystemVersion
		e.Browser = browser
		e.BrowserVersion = browserVersion
		e.ReferrerSource = source
		e.Referrer = referrer
		e.CountryCode = countryCode
		e.CityGeonameId = cityGeonameId
		e.ScreenSize = screenSize
		e.Labels = labels
		e.Timestamp = timestamppb.New(now)
		previousUUserID := seedID.GenPrevious(remoteIp, userAgent, domain, host)
		e.SessionId = v.session.RegisterSession(e, previousUUserID)
		v.events <- e
	}
	return true
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
