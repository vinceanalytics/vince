package goals

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/params"
	"github.com/vinceanalytics/vince/internal/sessions"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	site := models.GetSite(ctx)
	id, _ := strconv.ParseUint(params.Get(ctx).Get("goal"), 10, 64)
	goal := models.GoalID(ctx, id)
	to := fmt.Sprintf("/%s/%s/settings#goals", u.Name, site.Domain)
	session, r := sessions.Load(r)
	if goal == nil {
		session.Fail("no such a goal")
	} else {
		if !models.DeleteGoal(ctx, site, goal) {
			session.Fail("failed to delete goal")
		} else {
			session.Success("Goal deleted successfully")
		}
	}
	session.Save(ctx, w)
	http.Redirect(w, r, to, http.StatusFound)
}
