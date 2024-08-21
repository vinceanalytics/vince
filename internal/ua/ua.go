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

func Get(agent string) (a *v1.Agent) {
	if d := cache.Get(nil, []byte(agent)); d != nil {
		a = &v1.Agent{}
		proto.Unmarshal(d, a)
		return
	}
	a = ua2.Parse(agent)
	if a == nil {
		return
	}
	data, _ := proto.Marshal(a)
	cache.Set([]byte(agent), data)
	return
}
