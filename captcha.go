package vince

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/log"
	"github.com/lestrrat-go/dataurl"
)

const captchaKey = "_captcha"

func saveCaptcha(w http.ResponseWriter, r *http.Request, session *SessionContext) *http.Request {
	solution := captcha.RandomDigits(captcha.DefaultLen)
	img := captcha.NewImage("", solution, captcha.StdWidth, captcha.StdHeight)
	var b bytes.Buffer
	img.WriteTo(&b)
	data, err := dataurl.Encode(b.Bytes(), dataurl.WithBase64Encoding(true), dataurl.WithMediaType("image/png"))
	if err != nil {
		log.Get(r.Context()).Err(err).Msg("failed to encode captcha image")
		return r
	}
	session.Data[captchaKey] = formatCaptchaSolution(solution)
	session.Save(w)
	return r.WithContext(templates.SetCaptcha(r.Context(),
		template.HTMLAttr(fmt.Sprintf("src=%q", string(data))),
	))
}
