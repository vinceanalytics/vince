package ua

import (
	"bytes"
	_ "embed"
	"log/slog"
	"sync"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/blevesearch/vellum"
	"github.com/blevesearch/vellum/levenshtein"
	"github.com/dgraph-io/ristretto"
	"github.com/vinceanalytics/staples/staples/logger"
)

//go:generate go run gen/main.go device-detector/Tests/fixtures

//go:embed ua.arrow
var arrowBytes []byte

//go:embed ua.fst
var fstBytes []byte

var record Record
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
	for i := 0; i < int(r.NumCols()); i++ {
		a := r.Column(i)
		switch r.ColumnName(i) {
		case "isBot":
			record.IsBot = a.(*array.Int64)
		case "oSName":
			record.OSName = newStr(a)
		case "oSVersion":
			record.OSVersion = newStr(a)
		case "clientName":
			record.ClientName = newStr(a)
		case "clientVersion":
			record.ClientVersion = newStr(a)
		}
	}
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
		"component", "ua",
	))
}

type Record struct {
	IsBot         *array.Int64
	OSName        *Field
	OSVersion     *Field
	ClientName    *Field
	ClientVersion *Field
}

func (r *Record) Get(i int) *Model {
	return &Model{
		IsBot:          r.IsBot.Value(i) == 1,
		Os:             r.OSName.Get(i),
		OsVersion:      r.OSVersion.Get(i),
		Browser:        r.ClientName.Get(i),
		BrowserVersion: r.ClientVersion.Get(i),
	}
}

type Model struct {
	IsBot          bool
	Os             string
	OsVersion      string
	Browser        string
	BrowserVersion string
}

func (m *Model) Size() int {
	return 2 + len(m.Os) +
		len(m.OsVersion) + len(m.Browser) + len(m.BrowserVersion)
}

type Field struct {
	Dict *array.Dictionary
	Str  *array.String
}

func (f *Field) Get(i int) string {
	return f.Str.Value(f.Dict.GetValueIndex(i))
}

func newStr(a arrow.Array) *Field {
	d := a.(*array.Dictionary)
	return &Field{
		Dict: d,
		Str:  d.Dictionary().(*array.String),
	}
}

func Get(agent string) *Model {
	if v, ok := cache.Get(agent); ok {
		return v.(*Model)
	}
	lbMu.Lock()
	dfa, err := lb.BuildDfa(agent, 2)
	if err != nil {
		lbMu.Unlock()
		log.Error("failed building dfal", slog.String("agent", agent), slog.String("err", err.Error()))
		return &Model{}
	}
	lbMu.Unlock()
	it, err := fst.Search(dfa, nil, nil)
	for err == nil {
		_, val := it.Current()
		r := record.Get(int(val))
		cache.Set(agent, r, int64(r.Size()))
		return r
	}
	return &Model{}
}
