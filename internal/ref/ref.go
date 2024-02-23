package ref

import (
	"bytes"
	_ "embed"
	"log/slog"
	"math/rand"
	"net/url"
	"strings"
	"sync"

	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/blevesearch/vellum"
	"github.com/blevesearch/vellum/levenshtein"
	"github.com/dgraph-io/ristretto"
	"github.com/vinceanalytics/vince/internal/logger"
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
	host := clean(uri)
	if v, ok := cache.Get(host); ok {
		return v.(string)
	}
	lbMu.Lock()
	dfa, err := lb.BuildDfa(host, 2)
	if err != nil {
		lbMu.Unlock()
		log.Error("failed building dfa", slog.String("host", host), slog.String("err", err.Error()))
		return ""
	}
	lbMu.Unlock()
	it, err := fst.Search(dfa, nil, nil)
	for err == nil {
		_, val := it.Current()
		r := nameData.Value(name.GetValueIndex(int(val)))
		cache.Set(host, r, int64(len(r)))
		return r
	}
	return ""
}

func clean(r string) string {
	if strings.HasPrefix(r, "http://") || strings.HasPrefix(r, "https://") {
		u, err := url.Parse(r)
		if err != nil {
			log.Error("failed parsing url", slog.String("uri", r), slog.String("err", err.Error()))
			return ""
		}
		return u.Host
	}
	return r
}

func Random(count int) (o []string) {
	top := fst.Len() - 1
	from := rand.Intn(top)
	if from+count > top {
		from -= count
	}
	o = make([]string, 0, count)
	it, err := fst.Iterator(nil, nil)
	end := from + count
	var n int
	for err == nil && end > 0 {
		end--
		key, _ := it.Current()
		if n > from {
			o = append(o, "https://"+string(key))
		}
		err = it.Next()
		n++
	}
	return
}
