package db

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/domains"
	"github.com/vinceanalytics/vince/internal/geo"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ref"
	"github.com/vinceanalytics/vince/internal/ua2"
)

var pageView = []byte("pageview")

func (db *Config) ProcessEvent(r *http.Request) error {
	m, err := db.parse(r)
	if err != nil {
		return err
	}
	db.buffer <- m
	return nil
}

func hit(e *models.Model) {
	e.Bounce = 1
	e.Session = true
	if bytes.Equal(e.Event, pageView) {
		e.Event = nil
		e.View = true
	}
}

func newSessionEvent(e *models.Model) {
	if e.View {
		e.EntryPage = e.Page
		e.ExitPage = e.Page
	} else {
		e.Host = nil
	}
}

func update(session *models.Model, event *models.Model) {
	if session.Bounce == 1 {
		session.Bounce, event.Bounce = -1, -1
	} else {
		session.Bounce, event.Bounce = 0, 0
	}
	event.Session = false
	if len(session.EntryPage) == 0 && event.View {
		event.EntryPage = event.Page
	} else {
		event.EntryPage = session.EntryPage
	}
	if event.View && len(session.Host) == 0 {
	} else {
		event.Host = session.Host
	}
	if event.View {
		event.ExitPage = event.Page
	} else {
		event.ExitPage = session.ExitPage
	}
	event.Duration = int64(time.UnixMilli(event.Timestamp).Sub(time.UnixMilli(session.Timestamp)))
	session.Timestamp = event.Timestamp
}

var (
	True = ptr(true)

	False = ptr(false)

	ErrDrop = errors.New("event dropped")
)

func newEevent() *models.Model {
	return new(models.Model)
}

func releaseEvent(e *models.Model) {
	*e = models.Model{}
}

func (db *Config) parse(r *http.Request) (*models.Model, error) {
	req := newRequest()
	defer req.Release()

	err := req.Parse(r)
	if err != nil {
		return nil, err
	}
	domain := req.domains[0]
	if !domains.Allow(domain) {
		return nil, ErrDrop
	}

	host := req.hostname
	query := req.uri.Query()

	ref, src, err := refSource(query, req.referrer)
	if err != nil {
		return nil, fmt.Errorf("parsing referer %w", err)
	}
	path := req.pathname

	userID := uniqueID(req.remoteIp, req.userAgent, domain, host)
	e := newEevent()
	if req.remoteIp != "" {
		ip := net.ParseIP(req.remoteIp)
		err := geo.UpdateCity(ip, e)
		if err != nil {
			db.logger.Error("updating geo data", "err", err)
		}
	}
	e.Id = userID
	e.Event = []byte(req.eventName)
	e.Page = []byte(path)
	e.Host = []byte(host)
	e.Domain = []byte(domain)
	e.UtmMedium = []byte(query.Get("utm_medium"))
	e.UtmSource = []byte(query.Get("utm_source"))
	e.UtmCampaign = []byte(query.Get("utm_campaign"))
	e.UtmContent = []byte(query.Get("utm_content"))
	e.UtmTerm = []byte(query.Get("utm_term"))
	ua2.Parse(req.userAgent, e)
	e.Source = []byte(src)
	e.Referrer = []byte(ref)
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
	if !strings.HasPrefix(u, "http") {
		u = "https://" + u
	}
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
