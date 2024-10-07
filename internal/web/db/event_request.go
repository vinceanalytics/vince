package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var requestPool = &sync.Pool{New: func() any { return new(Request) }}

type Request struct {
	remoteIp  string
	uri       *url.URL
	userAgent string
	hostname  string
	referrer  string
	domains   []string
	eventName string
	hashMode  int64
	pathname  string
	props     map[string]string
	ts        time.Time
}

func newRequest() *Request {
	return requestPool.Get().(*Request)
}

func (r *Request) Release() {
	*r = Request{}
	requestPool.Put(r)
}

func (rq *Request) Parse(r *http.Request) error {
	body := map[string]any{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return err
	}
	rq.ts = time.Now().UTC().Truncate(time.Second)
	rq.remoteIp = remoteIP(r)
	err = rq.parseUri(body)
	if err != nil {
		return err
	}
	rq.hostname = rq.uri.Host
	rq.userAgent = r.UserAgent()
	rq.parseParams(body)
	rq.parseProps(body)
	rq.parsePathname()
	rq.parseReferrer(body)
	rq.parseDomains(body)
	return rq.validate()
}

func (rq *Request) validate() error {
	if rq.eventName == "" {
		return errors.New("missing event name")
	}
	if rq.hostname == "" {
		return errors.New("missing host name")
	}
	if rq.pathname == "" {
		return errors.New("missing path name")
	}
	if len(rq.eventName) > 120 {
		return errors.New("event name too long")
	}
	if rq.uri.Scheme == "data" {
		return errors.New("data scheme not supported")
	}
	return nil
}

func (rq *Request) parsePathname() {
	if rq.uri == nil {
		return
	}
	path := rq.uri.Path
	if rq.hashMode == 1 && len(rq.uri.Fragment) > 0 {
		path += "#" + rq.uri.Fragment
	}
	rq.pathname = path
}

func (rq *Request) parseDomains(m map[string]any) {
	d, ok := m["d"]
	if !ok {
		d, ok = m["domain"]
	}
	if !ok {
		return
	}
	ds, _ := d.(string)
	ds = strings.TrimSpace(ds)
	rq.domains = strings.Split(ds, ",")
	for i := range rq.domains {
		rq.domains[i] = strings.TrimPrefix(
			strings.TrimSpace(rq.domains[i]),
			"www.",
		)
	}
}

func (rq *Request) parseProps(m map[string]any) {
	o, ok := m["m"]
	if !ok {
		o, ok = m["meta"]
	}
	if !ok {
		o, ok = m["p"]
	}
	if !ok {
		o, ok = m["props"]
	}
	if !ok {
		return
	}
	os, ok := o.(map[string]any)
	if !ok {
		return
	}
	rq.props = make(map[string]string)
	for k, v := range os {
		rq.props[k] = fmt.Sprint(v)
	}
}

const maxURI = 2_000

func (rq *Request) parseReferrer(m map[string]any) {
	r, ok := m["r"]
	if !ok {
		r, ok = m["referrer"]
	}
	_ = ok
	rs, _ := r.(string)
	if rs != "" {
		if len(rs) > maxURI {
			rs = rs[:maxURI]
		}
		rq.referrer = rs
	}
}

func (rq *Request) parseParams(m map[string]any) {
	if ts, ok := m["ts"]; ok {
		// Only used for random seed generation
		rq.ts = time.UnixMilli(int64(ts.(float64))).UTC().Truncate(time.Second)
	}
	n, ok := m["n"]
	if !ok {
		n, ok = m["name"]
	}
	_ = ok
	h, ok := m["h"]
	if !ok {
		h, ok = m["hashMode"]
	}
	_ = ok
	rq.eventName, _ = n.(string)
	rq.hashMode, _ = h.(int64)
}

func (rq *Request) parseUri(m map[string]any) error {
	u, ok := m["u"]
	if !ok {
		u, ok = m["url"]
	}
	_ = ok
	us, _ := u.(string)
	if us != "" {
		uri, err := url.Parse(us)
		if err != nil {
			return err
		}
		rq.uri = uri
	}

	return nil
}

var remoteIPHeaders = []string{
	"x-vince-ip", "cf-connecting-ip", "b-forwarded-for",
	"X-Real-IP", "X-Forwarded-For", "X-Client-IP",
}

func remoteIP(r *http.Request) string {
	var raw string
	for _, v := range remoteIPHeaders {
		if raw = r.Header.Get(v); raw != "" {
			break
		}
	}
	if raw == "" && r.RemoteAddr != "" {
		raw = r.RemoteAddr
	}
	var host string
	host, _, err := net.SplitHostPort(raw)
	if err != nil {
		host = raw
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "-"
	}
	return ip.String()
}
