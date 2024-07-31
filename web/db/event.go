package db

import (
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"github.com/gernest/len64/geo"
	"github.com/gernest/len64/ref"
	"github.com/gernest/len64/ua"
)

const pageView = "pageview"

func (db *Config) ProcessEvent(r *http.Request) error {
	m, err := db.parse(r)
	if err != nil {
		if errors.Is(err, ErrDrop) {
			return nil
		}
		return err
	}
	db.append(m)
	return nil
}

func hit(e *v1.Model) {
	e.EntryPage = e.Page
	e.Bounce = True
	e.Session = true
	if e.Event == pageView {
		e.Event = ""
		e.View = true
	}
}

func update(fromSession *v1.Model, event *v1.Model) {
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

var (
	True = ptr(true)

	False = ptr(false)

	ErrDrop = errors.New("dop")
)

func (db *Config) parse(r *http.Request) (*v1.Model, error) {
	req := newRequest()
	defer req.Release()

	err := req.Parse(r)
	if err != nil {
		return nil, err
	}
	domain := req.domains[0]
	if !db.domains.Contains(domain) {
		return nil, ErrDrop
	}

	host := req.hostname
	query := req.uri.Query()

	ref, src, err := refSource(query, req.referrer)
	if err != nil {
		return nil, fmt.Errorf("arsing referer%w", err)
	}
	path := req.pathname
	agent, err := ua.Get(req.userAgent)
	if err != nil {
		return nil, err
	}
	var city geo.Info
	if req.remoteIp != "" {
		ip := net.ParseIP(req.remoteIp)
		city, err = geo.Get(ip)
		if err != nil {
			return nil, err
		}
	}
	userID := uniqueID(req.remoteIp, req.userAgent, domain, host)
	e := new(v1.Model)
	e.Id = userID
	e.Event = req.eventName
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
	e.Device = agent.Device
	e.Timestamp = req.ts.UnixMilli()
	return e, nil
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
	source, err = ref.Search(r.Host)
	return
}

func uniqueID(remoteIP, userAgent, domain, host string) uint64 {
	h := crc32.NewIEEE()
	h.Write([]byte(remoteIP))
	h.Write([]byte(userAgent))
	h.Write([]byte(domain))
	h.Write([]byte(host))
	return uint64(h.Sum32())
}
