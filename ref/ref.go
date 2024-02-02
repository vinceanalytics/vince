package ref

import (
	"bytes"
	_ "embed"
	"log/slog"
	"net/url"
	"sync"

	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/blevesearch/vellum"
	"github.com/blevesearch/vellum/levenshtein"
	"github.com/dgraph-io/ristretto"
	"github.com/vinceanalytics/staples/logger"
)

//go:generate go run gen/main.go

//go:embed refs.arrow
var arrowBytes []byte

var name *array.Dictionary
var nameData *array.String

//go:embed refs.fst
var fstBytes []byte

var fst *vellum.FST
var log *slog.Logger
var lb *levenshtein.LevenshteinAutomatonBuilder
var lbMu sync.Mutex
var cache *ristretto.Cache

func init() {
	rd, err := ipc.NewReader(bytes.NewReader(arrowBytes))
	if err != nil {
		logger.Fail(err.Error())
	}
	rd.Next()
	r := rd.Record()
	name = r.Column(0).(*array.Dictionary)
	nameData = name.Dictionary().(*array.String)
	fst, err = vellum.Load(fstBytes)
	if err != nil {
		logger.Fail(err.Error())
	}
	lb, err = levenshtein.NewLevenshteinAutomatonBuilder(2, false)
	if err != nil {
		logger.Fail(err.Error())
	}
	cache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	log = slog.Default().With(slog.String(
		"component", "referer",
	))
}

// Search use fuzz search to find a matching referral string.
func Search(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		log.Error("failed parsing url", slog.String("uri", uri), slog.String("err", err.Error()))
		return ""
	}
	if v, ok := cache.Get(u.Host); ok {
		return v.(string)
	}
	lbMu.Lock()
	dfa, err := lb.BuildDfa(u.Host, 2)
	if err != nil {
		lbMu.Unlock()
		log.Error("failed building dfal", slog.String("host", u.Host), slog.String("err", err.Error()))
		return ""
	}
	lbMu.Unlock()
	it, err := fst.Search(dfa, nil, nil)
	for err == nil {
		_, val := it.Current()
		r := nameData.Value(name.GetValueIndex(int(val)))
		cache.Set(u.Host, r, int64(len(r)))
		return r
	}
	return ""
}
