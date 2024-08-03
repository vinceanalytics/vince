package web

import (
	"net/http"

	"github.com/gernest/len64/web/db"
)

func UnimplementedStat(db *db.Config, w http.ResponseWriter, r *http.Request) {
}
