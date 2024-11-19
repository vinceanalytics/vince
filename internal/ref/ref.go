package ref

import (
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"strings"

	"github.com/vinceanalytics/vince/internal/location"
)

//go:embed favicon
var faviconData embed.FS

var Favicon, _ = fs.Sub(faviconData, "favicon")

func Search(lo *location.Location, uri string) ([]byte, error) {
	base, err := clean(uri)
	if err != nil {
		return nil, err
	}
	m := lo.GetReferral(base)
	if m != nil {
		return m, nil
	}
	return []byte(base), nil
}

func clean(r string) (string, error) {
	if strings.HasPrefix(r, "http://") || strings.HasPrefix(r, "https://") {
		u, err := url.Parse(r)
		if err != nil {
			return "", fmt.Errorf("cleaning referer uri%w", err)
		}
		return u.Host, nil
	}
	return r, nil
}
