package web

import (
	"net/http"
	"strings"

	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/db/plug"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func CurrentVisitorsAPI(db *db.Config, w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("site_id")
	visitors, err := db.Get().CurrentVisitors(domain)
	if err != nil {
		db.Logger().Error("retrieving current visitors", "domain", domain, "err", err)
	}
	db.JSON(w, visitors)
}

func AgggregatesAPI(db *db.Config, w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("site_id")
	params := query.New(r.URL.Query())

	stats, err := db.Get().Stats(domain, params.Start(), params.End(), params.Interval(), params.Filter(), params.Metrics())
	if err != nil {
		db.Logger().Error("reading top stats", "err", err)
		stats = &ro2.Stats{}
	}
	stats.Compute()
	db.JSON(w, stats)
}

func AuthorizeAPI(h plug.Handler) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("authorization")
		token := strings.TrimPrefix(auth, "Bearer ")
		token = strings.TrimSpace(token)
		if token == "" {
			http.Error(w,
				"Missing site ID. Please provide the required site_id parameter with your request.",
				http.StatusUnauthorized,
			)
			return
		}

		if !db.Get().ValidAPIKkey(token) {
			http.Error(w,
				"Invalid API key or site ID. Please make sure you're using a valid API key with access to the site you've requested.",
				http.StatusUnauthorized,
			)
			return
		}
		if r.URL.Query().Get("site_id") == "" {
			http.Error(w,
				"Missing site ID. Please provide the required site_id parameter with your request.",
				http.StatusUnauthorized,
			)
			return
		}
		h(db, w, r)
	}
}
