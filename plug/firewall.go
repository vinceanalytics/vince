package plug

import (
	"context"
	"net/http"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/firewall"
	"github.com/vinceanalytics/vince/render"
)

func Firewall(ctx context.Context) Plug {
	var pass firewall.List
	if w := config.Get(ctx).Firewall; w.Enabled {
		if len(w.AllowIP) > 0 {
			pass = append(pass, firewall.Negate(firewall.IP(w.AllowIP)))
		}
		if len(w.BlockIP) > 0 {
			pass = append(pass, firewall.Negate(firewall.IP(w.BlockIP)))
		}
	}
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !pass.Allow(r) {
				render.ERROR(r.Context(), w, http.StatusNotFound)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
