package db

import (
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/vinceanalytics/vince/internal/location"
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

var (
	ErrDrop = errors.New("event dropped")
)

func (db *Config) parse(r *http.Request) (*models.Model, error) {
	req := newRequest()
	defer req.Release()

	err := req.Parse(r)
	if err != nil {
		return nil, err
	}
	domain := req.domains[0]
	if !db.ops.HasSite(domain) {
		return nil, ErrDrop
	}

	host := req.hostname
	query := req.uri.Query()

	ref, src, err := refSource(db.lo, query, req.referrer)
	if err != nil {
		return nil, fmt.Errorf("parsing referer %w", err)
	}
	path := req.pathname

	userID := uniqueID(req.remoteIp, req.userAgent, domain, host)
	e := models.Get()
	if req.remoteIp != "" {
		ip := net.ParseIP(req.remoteIp)
		err := db.geo.UpdateCity(ip, e)
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
	if ua2.Parse(req.userAgent, e) {
		// It is a bot, drop this event
		return nil, ErrDrop
	}
	e.Source = src
	e.Referrer = []byte(ref)
	e.Timestamp = req.ts.UnixMilli()
	return e, nil
}

func sanitizeHost(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(s), "www.")
}

func refSource(lo *location.Location, q url.Values, u string) (xref string, source []byte, err error) {
	if !strings.HasPrefix(u, "http") {
		u = "https://" + u
	}
	r, err := url.Parse(u)
	if err != nil {
		return "", nil, err
	}
	r.Host = sanitizeHost(r.Host)
	r.Path = strings.TrimSuffix(r.Path, "/")
	xref = r.String()
	source = []byte(q.Get("utm_source"))
	if len(source) == 0 {
		source = []byte(q.Get("source"))
	}
	if len(source) == 0 {
		source = []byte(q.Get("ref"))
	}
	if len(source) != 0 {
		return
	}
	source, err = ref.Search(lo, r.Host)
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
