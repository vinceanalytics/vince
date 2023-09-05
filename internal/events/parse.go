package events

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/geoip"
	"github.com/vinceanalytics/vince/internal/referrer"
	"github.com/vinceanalytics/vince/internal/ua"
	"github.com/vinceanalytics/vince/internal/userid"
)

var ErrInvalid = errors.New("missing uri")
var ErrDataScheme = errors.New("data scheme not supported")

func Parse(req *entry.Request, ts time.Time) (*entry.Entry, error) {
	if req.Url == "" || req.N == "" || req.D == "" {
		return nil, ErrInvalid
	}

	uri, err := url.Parse(req.Url)
	if err != nil {
		return nil, err
	}
	if uri.Scheme == "data" {
		return nil, ErrDataScheme
	}
	host := sanitizeHost(uri.Host)
	query := uri.Query()

	ref, src, err := refSource(query, req.R)
	if err != nil {
		return nil, err
	}
	path := uri.Path
	if len(path) > 2000 {
		return nil, ErrInvalid
	}
	if req.H && uri.Fragment != "" {
		path += "#" + uri.Fragment
	}
	domain := req.D
	agent := ua.Parse(req.Ua)
	var city geoip.Info
	if req.Ip != "" {
		ip := net.ParseIP(req.Ip)
		city = geoip.Lookup(ip)
	}
	var screenSize string
	switch {
	case req.W < 576:
		screenSize = "mobile"
	case req.W < 992:
		screenSize = "tablet"
	case req.W < 1440:
		screenSize = "laptop"
	case req.W >= 1440:
		screenSize = "desktop"
	}
	userID := userid.Hash(req.Ip, req.Ua, domain, host)
	e := &entry.Entry{}
	e.ID = userID
	e.Event = req.N
	e.Host = host
	e.Path = path
	e.Domain = domain
	e.UtmMedium = query.Get("utm_medium")
	e.UtmSource = query.Get("utm_source")
	e.UtmCampaign = query.Get("utm_campaign")
	e.UtmContent = query.Get("utm_content")
	e.UtmTerm = query.Get("utm_term")
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
	return e, nil
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
