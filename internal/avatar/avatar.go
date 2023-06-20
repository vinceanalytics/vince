package avatar

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/pkg/identicon"
	"github.com/vinceanalytics/vince/pkg/log"
)

// Generates and server avatar in png format. Accepts query params
//
//	u = string for username
//	s = an int for the image size
//
// Any size below 24 is set to 24 and an empty u defaults to a fallback avatar image.
//
// TODO:(gernest) Add caching for the images
func Serve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	var o bytes.Buffer
	u := query.Get("u")
	if u != "" {
		// check if we have uploads for this user
		o := config.Get(ctx)
		x := o.Uploads.Dir
		if x == "" {
			x = filepath.Join(o.DataPath, "uploads")
		}
		if b, err := os.Open(filepath.Join(x, u)); err == nil {
			defer b.Close()
			stat, _ := b.Stat()
			http.ServeContent(w, r, "avatar.png", stat.ModTime(), b)
			return
		}
	}
	err := create(query.Get("s"), query.Get("u"), &o)
	if err != nil {
		log.Get().Err(err).Msg("failed to generate avatar")
		render.ERROR(r.Context(), w, http.StatusNotFound)
		return
	}
	http.ServeContent(w, r, "avatar.png", time.Time{}, bytes.NewReader(o.Bytes()))
}

func create(size, uid string, o *bytes.Buffer) error {
	if size == "" {
		size = "24"
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
