package avatar

import (
	"bytes"
	"net/http"
	"strconv"
	"time"

	"github.com/gernest/vince/pkg/identicon"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/render"
)

// Generates and server avatar in png format. Accepts query params
//
//	u = for user id which is a uint64
//	s = an int for the image size
//
// Any size below 24 is set to 24 and an empty u defaults to a fallback avatar image.
//
// TODO:(gernest) Add caching for the images
func Serve(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var o bytes.Buffer
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
