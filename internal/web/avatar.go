package web

import (
	"bytes"
	"net/http"
	"strconv"
	"time"

	"github.com/vinceanalytics/vince/internal/identicon"
	"github.com/vinceanalytics/vince/internal/web/db"
)

func Avatar(db *db.Config, w http.ResponseWriter, r *http.Request) {
	size := r.PathValue("size")
	uid := r.PathValue("uid")
	var o bytes.Buffer
	err := create(size, uid, &o)
	if err != nil {
		db.HTMLCode(http.StatusInternalServerError, w, e500, nil)
		db.Logger().Error("generating avatar", "size", size, "uid", uid, "err", err)
		return
	}
	http.ServeContent(w, r, "avatar.png", time.Time{}, bytes.NewReader(o.Bytes()))
}

func create(size, uid string, o *bytes.Buffer) error {
	if size == "" {
		size = "24"
	}
	if _, err := strconv.ParseUint(uid, 10, 64); err != nil {
		// uid is supposed to be a valid user id which is a uint64. Fallback
		// to placeholder when given bad value.
		uid = ""
	}
	if uid == "" {
		uid = "vince"
	}
	g, err := identicon.New("vince", 10, 5)
	if err != nil {
		return err
	}
	i, err := g.Draw(uid)
	if err != nil {
		return err
	}
	sz, _ := strconv.Atoi(size)
	if sz < 24 {
		sz = 24
	}
	return i.Png(sz, o)
}
