package web

import (
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

func System(sys *sys.Store) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/system/heap":
			err := sys.Heap(w, r)
			if err != nil {
				db.Logger().Error("serving heap graph", "err", err)
			}
		case "/system/requests":
			err := sys.Request(w, r)
			if err != nil {
				db.Logger().Error("serving requests graph", "err", err)
			}
		case "/system/duration":
			err := sys.Duration(w, r)
			if err != nil {
				db.Logger().Error("serving durations graph", "err", err)
			}
		case "/system":
			db.HTML(w, system, nil)
		default:
			db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
		}
	}
}
