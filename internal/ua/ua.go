package ua

import (
	"github.com/VictoriaMetrics/fastcache"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/ua2"
	"google.golang.org/protobuf/proto"
)

var (
	cache = fastcache.New(16 << 10)
)

func Get(agent string) (a *v1.Agent, err error) {
	if d := cache.Get(nil, []byte(agent)); d != nil {
		a = &v1.Agent{}
		err = proto.Unmarshal(d, a)
		return
	}
	o := ua2.Parse(agent)
	a = &v1.Agent{
		Bot:            o.Bot,
		Browser:        o.Browser,
		BrowserVersion: o.BrowserVersion,
		Os:             o.Os,
		OsVersion:      o.OsVersion,
		Device:         o.Device,
	}
	data, _ := proto.Marshal(a)
	cache.Set([]byte(agent), data)
	return
}

var deviceMapping = map[string]string{
	"smartphone":            "Mobile",
	"feature phone":         "Mobile",
	"portable media player": "Mobile",
	"phablet":               "Mobile",
	"wearable":              "Mobile",
	"camera":                "Mobile",
	"car browser":           "Tablet",
	"tablet":                "Tablet",
	"tv":                    "Desktop",
	"console":               "Desktop",
	"desktop":               "Desktop",
}
