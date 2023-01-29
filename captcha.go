package vince

import (
	"bytes"

	"github.com/dchest/captcha"
	"github.com/lestrrat-go/dataurl"
)

const captchaKey = "_captcha"

// generate a captcha image returns digits(solution) and data(data encoding of the image)
func createCaptcha() (digits []byte, data []byte, err error) {
	digits = captcha.RandomDigits(captcha.DefaultLen)
	img := captcha.NewImage("", digits, captcha.StdWidth, captcha.StdHeight)
	var b bytes.Buffer
	img.WriteTo(&b)
	data, err = dataurl.Encode(b.Bytes(), dataurl.WithBase64Encoding(true), dataurl.WithMediaType("image/png"))
	return
}
