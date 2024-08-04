package ua

import (
	"github.com/VictoriaMetrics/fastcache"
	"github.com/gernest/vice"
	"github.com/gernest/vice/pkg/bot"
	"github.com/gernest/vice/pkg/browser"
	"github.com/gernest/vice/pkg/device"
	"github.com/gernest/vice/pkg/os"
	v1 "github.com/vinceanalytics/vince/gen/go/len64/v1"
	"google.golang.org/protobuf/proto"
)

var (
	root  = must(vice.New(get))
	cache = fastcache.New(16 << 10)
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func get(id uint64) *v1.Agent {
	return &v1.Agent{
		Bot:            bot.GetBot(id),
		Browser:        browser.GetName(id),
		BrowserVersion: browser.GetVersion(id),
		Os:             os.GetName(id),
		OsVersion:      os.GetVersion(id),
		Device:         deviceMapping[device.GetType(id)],
	}
}

func Warm() {
	_ = get(0)
}

func Get(agent string) (a *v1.Agent, err error) {
	if d := cache.Get(nil, []byte(agent)); d != nil {
		a = &v1.Agent{}
		err = proto.Unmarshal(d, a)
		return
	}
	a, err = root.Get(agent)
	if err != nil {
		return nil, err
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
