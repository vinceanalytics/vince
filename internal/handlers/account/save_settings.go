package account

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/pkg/log"
)

func SaveSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	u.FullName = r.FormValue("full_name")
	models.Get(ctx).Save(u)
	if f, h, err := r.FormFile("profile_picture"); err != nil && !errors.Is(err, http.ErrMissingFile) {
		log.Get().Err(err).Msg("failed to open profile_picture file")
	} else {
		defer f.Close()
		if h.Size > (2 << 20) {
			log.Get().Debug().
				Str("size", fmt.Sprintf("%dmb", h.Size/int64(1<<20))).
				Msg("detected large file upload")
		} else {
			o := config.Get(ctx)
			x := o.Uploads.Dir
			if x == "" {
				x = filepath.Join(o.DataPath, "uploads")
			}
			os.MkdirAll(x, 0755)
			b, err := io.ReadAll(f)
			if err != nil {
				log.Get().Err(err).Msg("failed to read profile picture")
			} else {
				err = os.WriteFile(filepath.Join(x, u.Name), b, 0600)
				if err != nil {
					log.Get().Err(err).Msg("failed to write profile picture")
				}
			}
		}
	}
	http.Redirect(w, r, "/settings#profile", http.StatusFound)
}
