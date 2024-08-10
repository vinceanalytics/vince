package web

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/internal/sys"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/db/plug"
)

func RequireSuper(h plug.Handler) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		if usr := db.CurrentUser(); usr != nil && usr.SuperUser {
			h(db, w, r)
			return
		}
		db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
	}
}

func SystemHeap(sys *sys.Store) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		err := sys.Heap(w, r)
		if err != nil {
			db.Logger().Error("serving heap graph", "err", err)
		}
	}
}

func SystemRequests(sys *sys.Store) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		err := sys.Request(w, r)
		if err != nil {
			db.Logger().Error("serving requests graph", "err", err)
		}
	}
}

func SystemDuration(sys *sys.Store) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		err := sys.Duration(w, r)
		if err != nil {
			db.Logger().Error("serving durations graph", "err", err)
		}
	}
}

func SystemStats(sys *sys.Store) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		size := float64(sys.Size()) / (1 << 10)
		db.HTML(w, system, map[string]any{
			"size": fmt.Sprintf("%.3f", size),
		})
	}
}

func SystemRest(sys *sys.Store) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		sys.Reset()
		http.Redirect(w, r, "/system/stats", http.StatusFound)
	}
}
