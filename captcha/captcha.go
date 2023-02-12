package captcha

import (
	"strconv"
	"strings"
)

const Key = "_captcha"

func FormatCaptchaSolution(sol []byte) string {
	var s strings.Builder
	s.Grow(len(sol))
	for _, b := range sol {
		s.WriteString(strconv.FormatInt(int64(b), 10))
	}
	return s.String()
}
