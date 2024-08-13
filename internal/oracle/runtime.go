package oracle

import (
	"time"
)

func date(ts int64) string {
	yy, mm, dd := time.UnixMilli(ts).Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
}
