package plug

import (
	"context"
	"net/http"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/firewall"
	"github.com/gernest/vince/render"
)

func Firewall(ctx context.Context) Plug {
	var pass firewall.Wall = firewall.Pass{}
	if w := config.Get(ctx).Firewall; w != nil {
		var list firewall.List
		if len(w.Allow) > 0 {
			var ip []string
			for _, m := range w.Allow {
				switch e := m.Match.(type) {
				case *config.Firewall_Match_Ip:
					ip = append(ip, e.Ip.Address)
				}
			}
			if len(ip) > 0 {
				list = append(list, firewall.IP(ip))
			}
		}
		if len(w.Block) > 0 {
			var ip []string
			for _, m := range w.Allow {
				switch e := m.Match.(type) {
				case *config.Firewall_Match_Ip:
					ip = append(ip, e.Ip.Address)
				}
			}
			if len(ip) > 0 {
				list = append(list, firewall.Negate(firewall.IP(ip)))
			}
		}
		if len(list) > 0 {
			pass = list
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
