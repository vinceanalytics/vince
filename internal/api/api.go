package api

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bufbuild/protovalidate-go"
	"github.com/dgraph-io/ristretto"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/events"
	"github.com/vinceanalytics/vince/internal/geo"
	"github.com/vinceanalytics/vince/internal/guard"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/version"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	validator *protovalidate.Validator
)

func init() {
	validator, _ = protovalidate.New(protovalidate.WithFailFast(true))
}

type API struct {
	db     *db.DB
	geo    *geo.Geo
	log    *slog.Logger
	guard  guard.Guard
	events chan *v1.Data

	cache *ristretto.Cache
}

func New(db *db.DB, geo *geo.Geo, guard guard.Guard) *API {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     100 << 20, // 100MiB
		BufferItems: 64,
	})
	if err != nil {
		logger.Fail("setup events cache", "err", err)
	}

	return &API{
		db:     db,
		geo:    geo,
		log:    slog.Default().With("component", "api"),
		guard:  guard,
		events: make(chan *v1.Data, 4<<10),
		cache:  cache,
	}
}

func (a *API) Release() {
	a.cache.Close()
}

func (a *API) Start(ctx context.Context) {
	a.log.Info("Starting events processing loop")
	ts := time.NewTicker(time.Minute)
	defer ts.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-a.events:
			a.append(e)
		case <-ts.C:
			err := a.db.Save()
			if err != nil {
				logger.Fail("saving events", "err", err)
			}
		}
	}
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := NewParams(r)
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/v1/stats/realtime/visitors"):
		a.handleRealtime(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/v1/stats/aggregate"):
		a.handleAggregate(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/v1/stats/timeseries"):
		a.handleTimeseries(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/v1/stats/breakdown"):
		a.handleBreakdown(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/v1/event"):
		a.handleApiEvent(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/api/event"):
		a.handleEvent(w, r, params)
	case r.URL.Path == "/version":
		a.handleVersion(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (a *API) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	a.write(w, &v1.Version{
		Version: version.VERSION,
	})
}

func (a *API) handleApiEvent(w http.ResponseWriter, r *http.Request, _ Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var ev v1.Event
	err = protojson.Unmarshal(b, &ev)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = validator.Validate(&ev)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	e := events.Parse(a.log, a.geo, &ev)
	if e == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a.events <- e
	w.WriteHeader(http.StatusAccepted)
}

func (a *API) handleEvent(w http.ResponseWriter, r *http.Request, _ Params) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", http.MethodPost)
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	if !a.guard.Allow() {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var ev v1.Event
	err = protojson.Unmarshal(b, &ev)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !a.guard.Accept(ev.D) {
		w.Header().Set("x-vince-dropped", "1")
		w.WriteHeader(http.StatusOK)
		return
	}
	ev.Ip = remoteIP(r)
	ev.Ua = r.UserAgent()
	err = validator.Validate(&ev)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	e := events.Parse(a.log, a.geo, &ev)
	if e == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	a.events <- e
	w.WriteHeader(http.StatusAccepted)
}

var remoteIPHeaders = []string{
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

func (a *API) handleRealtime(w http.ResponseWriter, r *http.Request, params Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	req := &v1.Realtime_Request{
		SiteId: params.SiteID(),
	}
	res, err := a.db.Realtime(ctx, req)
	if err != nil {
		a.jsonErr(w, err.Error())
		return
	}
	a.write(w, res)
}

func (a *API) handleAggregate(w http.ResponseWriter, r *http.Request, params Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	req := &v1.Aggregate_Request{
		SiteId:  params.SiteID(),
		Period:  params.Period(ctx),
		Metrics: params.Metrics(ctx),
		Filters: params.Filters(ctx),
	}
	res, err := a.db.Aggregate(ctx, req)
	if err != nil {
		a.jsonErr(w, err.Error())
		return
	}
	a.write(w, res)
}

func (a *API) handleTimeseries(w http.ResponseWriter, r *http.Request, params Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	req := &v1.Timeseries_Request{
		SiteId:   params.SiteID(),
		Period:   params.Period(ctx),
		Metrics:  params.Metrics(ctx),
		Interval: params.Interval(ctx),
		Filters:  params.Filters(ctx),
	}
	res, err := a.db.Timeseries(ctx, req)
	if err != nil {
		a.jsonErr(w, err.Error())
		return
	}
	a.write(w, res)
}

func (a *API) handleBreakdown(w http.ResponseWriter, r *http.Request, params Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	req := &v1.BreakDown_Request{
		SiteId:   params.SiteID(),
		Period:   params.Period(ctx),
		Metrics:  params.Metrics(ctx),
		Filters:  params.Filters(ctx),
		Property: params.Property(ctx),
	}
	res, err := a.db.Breakdown(ctx, req)
	if err != nil {
		a.jsonErr(w, err.Error())
		return
	}
	a.write(w, res)
}

func (a *API) jsonErr(w http.ResponseWriter, msg string, code ...int) {
	c := http.StatusInternalServerError
	if len(code) > 0 {
		c = code[0]
	}
	w.WriteHeader(c)
	a.write(w, &v1.Error{Error: msg})
}

func (a *API) write(w http.ResponseWriter, msg proto.Message) {
	data, _ := protojson.Marshal(msg)
	_, err := w.Write(data)
	if err != nil {
		slog.Error("failed writing response data", "err", err)
	}
}

const DefaultSession = 30 * time.Minute

func (a *API) append(e *v1.Data) {
	events.Hit(e)
	if o, ok := a.cache.Get(e.Id); ok {
		cached := o.(*v1.Data)
		events.Update(cached, e)
		a.db.Append(e)
		return
	}
	a.db.Append(e)
	for range 5 {
		if a.cache.SetWithTTL(e.Id, e, int64(proto.Size(e)), DefaultSession) {
			return
		}
	}
}
