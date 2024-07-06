package events

import (
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/geo"
	"github.com/vinceanalytics/vince/internal/ref"
	"github.com/vinceanalytics/vince/ua"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const pageView = "pageview"

func Hit(e *v1.Data) {
	e.EntryPage = e.Page
	e.Bounce = True
	e.Session = true
	if e.Event == pageView {
		e.Event = ""
		e.View = true
	}
}

func Update(fromSession *v1.Data, event *v1.Data) {
	if fromSession.Bounce == True {
		fromSession.Bounce, event.Bounce = nil, nil
	} else {
		fromSession.Bounce, event.Bounce = False, False
	}
	event.Session = false
	event.ExitPage = event.Page
	// Track duration since last visit.
	event.Duration = time.UnixMilli(event.Timestamp).Sub(time.UnixMilli(fromSession.Timestamp)).Seconds()
	fromSession.Timestamp = event.Timestamp
}

var True = ptr(true)

var False = ptr(false)

func Parse(log *slog.Logger, g *geo.Geo, req *v1.Event) *v1.Data {
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
	e := One()
	e.Id = int64(userID)
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

func uniqueID(remoteIP, userAgent, domain, host string) (sum uint64) {
	var h xxhash.Digest
	h.WriteString(remoteIP)
	h.WriteString(userAgent)
	h.WriteString(domain)
	h.WriteString(host)
	sum = h.Sum64()
	return
}

func Clone(e *v1.Data) *v1.Data {
	o := One()
	proto.Merge(o, e)
	return o
}
