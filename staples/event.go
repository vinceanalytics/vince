package staples

import (
	"context"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/geo"
	"github.com/vinceanalytics/staples/staples/logger"
	"github.com/vinceanalytics/staples/staples/ref"
	"github.com/vinceanalytics/staples/staples/ua"
)

type Event struct {
	Timestamp int64
	ID        int64
	// When a new session is established for the first time we set Bounce to 1, if
	// a user visits another page within the same session for the first time Bounce
	// is set to -1, any subsequent visits within the session sets Bounce to 0.
	//
	// This allows effective calculation of bounce rate by just summing the Bounce
	// column with faster math.Int64.Sum.
	//
	// NOTE: Bounce is calculated per session. We simply want to know if a user
	// stayed and browsed the website.
	Bounce         int64
	Session        int64
	Duration       float64
	Browser        string
	BrowserVersion string
	City           string
	Country        string
	Domain         string
	EntryPage      string
	ExitPage       string
	Host           string
	Event          string
	Os             string
	OsVersion      string
	Path           string
	Referrer       string
	ReferrerSource string
	Region         string
	Screen         string
	UtmCampaign    string
	UtmContent     string
	UtmMedium      string
	UtmSource      string
	UtmTerm        string
}

var eventsPool = &sync.Pool{New: func() any { return new(Event) }}

func NewEvent() *Event {
	return eventsPool.Get().(*Event)
}

func (e *Event) Release() {
	*e = Event{}
	eventsPool.Put(e)
}

func (e *Event) TS() int64 { return e.Timestamp }

func (e *Event) Hit() {
	e.EntryPage = e.Path
	e.Bounce = 1
	e.Session = 1
}

func (s *Event) Update(e *Event) {
	if s.Bounce == 1 {
		s.Bounce, e.Bounce = -1, -1
	} else {
		s.Bounce, e.Bounce = 0, 0
	}
	e.Session = 0
	e.ExitPage = e.Path
	// Track duration since last visit.
	e.Duration = time.UnixMilli(e.Timestamp).Sub(time.UnixMilli(s.Timestamp)).Seconds()
	s.Timestamp = e.Timestamp
}

func Parse(ctx context.Context, req *v1.Event) *Event {
	log := logger.Get(ctx)
	if req.Url == "" || req.N == "" || req.D == "" {
		log.Error("invalid request")
		return nil
	}
	uri, err := url.Parse(req.Url)
	if err != nil {
		log.Error("failed parsing event url", "err", err)
		return nil
	}
	if uri.Scheme == "data" {
		log.Error("Data url scheme is not supported")
		return nil
	}
	host := sanitizeHost(uri.Host)
	query := uri.Query()

	ref, src, err := refSource(query, req.R)
	if err != nil {
		log.Error("failed parsing referer", "err", err)
		return nil
	}
	path := uri.Path
	if len(path) > 2000 {
		log.Error("Path too long", "path", path)
		return nil
	}
	if req.H && uri.Fragment != "" {
		path += "#" + uri.Fragment
	}
	domain := req.D
	agent := ua.Get(req.Ua)
	var city geo.Info
	if req.Ip != "" {
		ip := net.ParseIP(req.Ip)
		city = geo.Get(ctx).Get(ip)
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
	userID := Fingerprint(req.Ip, req.Ua, domain, host)
	e := NewEvent()
	e.ID = int64(userID)
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
	e.Timestamp = req.Timestamp.AsTime().UnixMilli()
	return e
}

func sanitizeHost(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(s), "www.")
}

func refSource(q url.Values, u string) (xref, source string, err error) {
	r, err := url.Parse(u)
	if err != nil {
		return "", "", err
	}
	r.Host = sanitizeHost(r.Host)
	r.Path = strings.TrimSuffix(r.Path, "/")
	xref = r.String()
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
	source = ref.Search(r.Host)
	return
}
func Fingerprint(remoteIP, userAgent, domain, host string) (sum uint64) {
	var h xxhash.Digest
	h.WriteString(remoteIP)
	h.WriteString(userAgent)
	h.WriteString(domain)
	h.WriteString(host)
	sum = h.Sum64()
	return
}
