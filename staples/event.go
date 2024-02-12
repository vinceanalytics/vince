package staples

import (
	"context"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/cespare/xxhash/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/geo"
	"github.com/vinceanalytics/vince/logger"
	"github.com/vinceanalytics/vince/ref"
	"github.com/vinceanalytics/vince/ua"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	Bounce   *bool
	Session  bool
	Duration float64

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

// Size without strings
var baseSize = unsafe.Sizeof(Event{})

// Size in bytes of e in memory. We use this as a cost to control cache size.
func (e *Event) Size() (n int) {
	n = int(baseSize)
	n += len(e.Browser)
	n += len(e.BrowserVersion)
	n += len(e.City)
	n += len(e.Country)
	n += len(e.Domain)
	n += len(e.EntryPage)
	n += len(e.ExitPage)
	n += len(e.Host)
	n += len(e.Event)
	n += len(e.Os)
	n += len(e.OsVersion)
	n += len(e.Path)
	n += len(e.Referrer)
	n += len(e.ReferrerSource)
	n += len(e.Region)
	n += len(e.Screen)
	n += len(e.UtmCampaign)
	n += len(e.UtmContent)
	n += len(e.UtmMedium)
	n += len(e.UtmSource)
	n += len(e.UtmTerm)
	return
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
	e.Bounce = True
	e.Session = true
}

func (s *Event) Update(e *Event) {
	if s.Bounce == True {
		s.Bounce, e.Bounce = nil, nil
	} else {
		s.Bounce, e.Bounce = False, False
	}
	e.Session = false
	e.ExitPage = e.Path
	// Track duration since last visit.
	e.Duration = time.UnixMilli(e.Timestamp).Sub(time.UnixMilli(s.Timestamp)).Seconds()
	s.Timestamp = e.Timestamp
}

func Parse(ctx context.Context, req *v1.Event) *Event {
	log := logger.Get(ctx)
	if req.U == "" || req.N == "" || req.D == "" {
		log.Error("invalid request")
		return nil
	}
	uri, err := url.Parse(req.U)
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
	if req.Timestamp == nil {
		req.Timestamp = timestamppb.Now()
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

var True = ptr(true)

var False = ptr(false)

func ptr[T any](a T) *T {
	return &a
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
