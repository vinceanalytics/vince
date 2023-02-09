package vince

import (
	"net/http"
	"time"

	"github.com/gernest/vince/models"
)

// - sets session timeouts
// - sets CurrentUser
// - update User.LastSeen
func (v *Vince) browser(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		if u, ok := session.Data[models.CurrentUserID]; ok {
			usr := &models.User{}
			if err := v.sql.First(usr, uint64(u.(int64))).Error; err != nil {
				xlg.Err(err).Msg("failed fetching current user")
			} else {
				now := time.Now().UTC()
				if sessionTimeout, ok := session.Data["_session_timeout_at"]; ok {
					timeout := time.Unix(sessionTimeout.(int64), 0)
					if now.After(timeout) {
						// drop everything
						session.Data = make(map[string]any)
					}
				} else {
					session.Data["_session_timeout_at"] = now.Add(24 * 7 * 2 * time.Hour)
					session.Save(w)
				}

				// update last seen
				if lastSeen, ok := session.Data["_last_seen"]; ok {
					ls := time.Unix(lastSeen.(int64), 0)
					if ls.Before(now.Add(-time.Hour)) {
						usr.LastSeen = now
						if err = v.sql.Model(usr).Update("last_seen", now).Error; err != nil {
							xlg.Err(err).Msg("failed to update last seen")
						} else {
							session.Data["_last_seen"] = now.Unix()
							session.Save(w)
						}
					}
				} else {
					session.Data["_last_seen"] = now.Unix()
					session.Save(w)
				}
				r = r.WithContext(
					models.SetCurrentUser(r.Context(), usr),
				)
			}
		}
		h.ServeHTTP(w, r)
	})
}
