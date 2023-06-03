package api

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/vinceanalytics/vince/gate"
	"github.com/vinceanalytics/vince/geoip"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/referrer"
	"github.com/vinceanalytics/vince/remoteip"
	"github.com/vinceanalytics/vince/system"
	"github.com/vinceanalytics/vince/ua"
	"github.com/vinceanalytics/vince/userid"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var ErrBadJSON = errors.New("api: invalid json ")

// Request is sent by the vince script embedded in client websites. For faster
// performance we use simd if available
type Request struct {
	EventName   string `json:"n"`
	URI         string `json:"url"`
	Referrer    string `json:"r"`
	Domain      string `json:"d"`
	ScreenWidth int    `json:"w"`
	HashMode    bool   `json:"h"`
}

// Parse opportunistic parses request body to r object. This is crucial method
// any gains here translates to smooth  events ingestion pipeline.
//
// A hard size limitation of 32kb is imposed. This is arbitrary value, any change
// to it must be be supported with statistics.
func (r *Request) Parse(body io.Reader) error {
	b := bufPool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		bufPool.Put(b)
	}()
	// My mom used to say, don't trust the internet. Never stream decode payloads
	// directly from strangers. We copy just enough then we  process.
	b.ReadFrom(io.LimitReader(body, 32<<10))
	return json.Unmarshal(b.Bytes(), r)
}

var bufPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// Events accepts events payloads from vince client script.
func Events(w http.ResponseWriter, r *http.Request) {
	system.DataPoint.WithLabelValues("received").Inc()

	w.Header().Set("X-Content-Type-Options", "nosniff")
	xlg := log.Get()
	var req Request
	err := req.Parse(r.Body)
	if err != nil {
		system.DataPointRejected.Inc()
		xlg.Err(err).Msg("Failed decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteIp := remoteip.Get(r)
	if req.URI == "" {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uri, err := url.Parse(req.URI)
	if err != nil {
		system.DataPointRejected.Inc()
		xlg.Err(err).Msg("Failed parsing url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if uri.Scheme == "data" {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	host := sanitizeHost(uri.Host)
	userAgent := r.UserAgent()

	reqReferrer := req.Referrer
	refUrl, err := url.Parse(reqReferrer)
	if err != nil {
		system.DataPointRejected.Inc()
		xlg.Err(err).Msg("Failed parsing referrer url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := uri.Path
	if req.HashMode && uri.Fragment != "" {
		path += "#" + uri.Fragment
	}
	if len(path) > 2000 {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.EventName == "" {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Domain == "" {
		system.DataPointRejected.Inc()
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var domains []string
	for _, d := range strings.Split(req.Domain, ",") {
		domains = append(domains, sanitizeHost(d))
	}

	query := uri.Query()
	agent := ua.Parse(userAgent)
	// handle referrer
	ref := referrer.ParseReferrer(req.Referrer)
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
	reqReferrer = cleanReferrer(reqReferrer)

	var city geoip.Info
	if remoteIp != "" {
		ip := net.ParseIP(remoteIp)
		city = geoip.Lookup(ip)
	}
	var screenSize string
	switch {
	case req.ScreenWidth < 576:
		screenSize = "mobile"
	case req.ScreenWidth < 992:
		screenSize = "tablet"
	case req.ScreenWidth < 1440:
		screenSize = "laptop"
	case req.ScreenWidth >= 1440:
		screenSize = "desktop"
	}
	var dropped int
	ts := time.Now()
	unix := ts.Unix()
	ctx := r.Context()
	uid := userid.Get(ctx)
	for _, domain := range domains {
		b, pass := gate.Check(r.Context(), domain)
		if !pass {
			dropped += 1
			continue
		}
		userID := uid.Hash(remoteIp, userAgent, domain, host)
		e := entry.NewEntry()
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
		e.OperatingSystem = agent.Os
		e.OperatingSystemVersion = agent.OsVersion
		e.Browser = agent.Browser
		e.BrowserVersion = agent.BrowserVersion
		e.ReferrerSource = source
		e.Referrer = reqReferrer
		e.Country = city.Country
		e.Region = city.Region
		e.City = city.City
		e.ScreenSize = screenSize
		e.Timestamp = unix
		previousUUserID := uid.HashPrevious(remoteIp, userAgent, domain, host)
		b.Register(r.Context(), e, previousUUserID)
	}
	if dropped > 0 {
		system.DataPointDropped.Inc()
		w.Header().Set("x-vince-dropped", strconv.Itoa(dropped))
		w.WriteHeader(http.StatusAccepted)
		return
	}
	system.DataPointAccepted.Inc()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
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
