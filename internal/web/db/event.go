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
	"sync"

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

var (
	ErrDrop = errors.New("event dropped")
)
var eventsPool = &sync.Pool{New: func() any { return new(models.Model) }}

func newEevent() *models.Model {
	return eventsPool.Get().(*models.Model)
}

func releaseEvent(e *models.Model) {
	*e = models.Model{}
	eventsPool.Put(e)
}

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
	query := parseQueryAdvertisements(req.uri.Query())

	ref, src, err := refSource(db.lo, query, req.referrer)
	if err != nil {
		return nil, fmt.Errorf("parsing referer %w", err)
	}
	path := req.pathname

	userID := uniqueID(req.remoteIp, req.userAgent, domain, host)
	e := newEevent()
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

func parseQueryAdvertisements(q url.Values) url.Values {
	// First handle affiliates before other advertisement channels
	affiliate := ""
	if q.Get("ref") != "" {
		affiliate = q.Get("ref")
	} else if q.Get("affiliate") != "" {
		affiliate = q.Get("affiliate")
	}
	affiliate = strings.TrimSpace(affiliate)
	if affiliate != "" {
		q.Set("utm_medium", "affiliate")
		q.Set("utm_source", strings.ToLower(affiliate))
		return q
	}

	// if there are utm parameters, do not overwrite them
	if q.Get("utm_medium") != "" || q.Get("utm_source") != "" {
		return q
	}

	// Handle other advertisement channels
	if q.Get("fbclid") != "" || q.Get("FBCLID") != "" {
		q.Set("utm_medium", "cpc")
		q.Set("utm_source", "facebook")
	}

	if q.Get("twclid") != "" || q.Get("TWCLID") != "" {
		q.Set("utm_medium", "cpc")
		q.Set("utm_source", "twitter")
	}

	if q.Get("rdt_cid") != "" || q.Get("RDT_CID") != "" {
		q.Set("utm_medium", "cpc")
		q.Set("utm_source", "reddit")
	}

	if q.Get("li_fat_id") != "" || q.Get("LI_FAT_ID") != "" {
		q.Set("utm_medium", "cpc")
		q.Set("utm_source", "linkedin")
	}

	if q.Get("ttclid") != "" || q.Get("TTCLID") != "" {
		q.Set("utm_medium", "cpc")
		q.Set("utm_source", "tiktok")
	}

	if q.Get("ScCid") != "" || q.Get("sccid") != "" || q.Get("SCCID") != "" {
		q.Set("utm_medium", "cpc")
		q.Set("utm_source", "snapchat")
	}

	// Put bing and google as last, in case someone shares the gclid or msclkid on other socials, that takes precedence
	if q.Get("msclkid") != "" || q.Get("MSCLKID") != "" {
		q.Set("utm_medium", "cpc")
		q.Set("utm_source", "bing")
	}

	if q.Get("gclid") != "" || q.Get("GCLID") != "" {
		q.Set("utm_medium", "cpc")
		q.Set("utm_source", "google")
	}

	return q
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
