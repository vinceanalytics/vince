package events

import (
	"crypto/sha512"
	"encoding/binary"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	v2 "github.com/vinceanalytics/vince/gen/go/vince/v2"
	"github.com/vinceanalytics/vince/internal/geo"
	"github.com/vinceanalytics/vince/internal/ref"
	"github.com/vinceanalytics/vince/ua"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const pageView = "pageview"

func Convert(a *v1.Data) *v2.Data {
	o := new(v2.Data)
	o.Timestamp = a.Timestamp
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(a.Id))
	h := sha512.Sum512_224(b[:])
	o.Id = h[:]
	o.Bounce = a.Bounce
	o.Session = a.Session
	o.View = a.View
	o.Duration = time.Duration(a.Duration * float64(time.Second)).Milliseconds()

	o.Page = a.Page
	o.Host = a.Host
	o.Domain = a.Domain
	o.UtmMedium = a.UtmMedium
	o.UtmSource = a.UtmSource
	o.UtmCampaign = a.UtmCampaign
	o.UtmContent = a.UtmContent
	o.UtmTerm = a.UtmTerm
	o.Os = a.Os
	o.OsVersion = a.OsVersion
	o.Browser = a.Browser
	o.BrowserVersion = a.BrowserVersion
	o.Source = a.Source
	o.Referrer = a.Referrer
	o.Country = a.Country
	o.Region = a.Region
	o.City = a.City
	o.Device = a.Device
	return o
}

func Hit(e *v2.Data) {
	e.EntryPage = e.Page
	e.Bounce = True
	e.Session = true
	if e.Event == pageView {
		e.Event = ""
		e.View = true
	}
}

func Update(fromSession *v2.Data, event *v2.Data) {
	if fromSession.Bounce == True {
		fromSession.Bounce, event.Bounce = nil, nil
	} else {
		fromSession.Bounce, event.Bounce = False, False
	}
	event.Session = false
	event.ExitPage = event.Page
	// Track duration since last visit.
	event.Duration = time.UnixMilli(event.Timestamp).Sub(time.UnixMilli(fromSession.Timestamp)).Milliseconds()
	fromSession.Timestamp = event.Timestamp
}

var True = ptr(true)

var False = ptr(false)

func Parse(log *slog.Logger, g *geo.Geo, req *v1.Event) *v2.Data {
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
		city = g.Get(ip)
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
	userID := uniqueID(req.Ip, req.Ua, domain, host)
	e := new(v2.Data)
	e.Id = userID
	e.Event = req.N
	e.Page = path
	e.Host = host
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
	e.Source = src
	e.Referrer = ref
	e.Country = city.Country
	e.Region = city.Region
	e.City = city.City
	e.Device = screenSize
	e.Timestamp = req.Timestamp.AsTime().UnixMilli()
	return e
}

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

func uniqueID(remoteIP, userAgent, domain, host string) []byte {
	h := sha512.New512_224()
	h.Write([]byte(remoteIP))
	h.Write([]byte(userAgent))
	h.Write([]byte(domain))
	h.Write([]byte(host))
	return h.Sum(nil)
}
