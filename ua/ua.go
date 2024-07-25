package ua

import (
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/json"
	"fmt"
	"io"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/blevesearch/vellum"
	"github.com/blevesearch/vellum/levenshtein"
	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"google.golang.org/protobuf/proto"
)

//go:embed data
var data embed.FS

// All these variables are safe for concurrent use because we only perform reads
// after init
var (
	bot             = roaring64.NewDefaultBSI()
	browser         = roaring64.NewDefaultBSI()
	browser_version = roaring64.NewDefaultBSI()
	os              = roaring64.NewDefaultBSI()
	os_version      = roaring64.NewDefaultBSI()

	browser_translate         []string
	browser_version_translate []string
	os_translate              []string
	os_version_translate      []string

	fst *vellum.FST
	dfa *levenshtein.LevenshteinAutomatonBuilder

	cache = fastcache.New(64 << 20)
)

func init() {
	var r *gzip.Reader
	get := func(name string, f func(b io.Reader)) {
		b, err := data.ReadFile("data/" + name + ".gz")
		if err != nil {
			panic(fmt.Sprintf("reading file %s %v", name, err))
		}
		if r == nil {
			r, err = gzip.NewReader(bytes.NewReader(b))
			if err != nil {
				panic(fmt.Sprintf("reading gzip file %s %v", name, err))
			}
		} else {
			err = r.Reset(bytes.NewReader(b))
			if err != nil {
				panic(fmt.Sprintf("reading gzip file %s %v", name, err))
			}
		}
		f(r)
	}

	bsi := func(name string, m *roaring64.BSI) {
		get(name+".bsi", func(b io.Reader) {
			_, err := m.ReadFrom(b)
			if err != nil {
				panic(fmt.Sprintf("decoding bsi  %s %v", name, err))
			}
		})
	}
	translate := func(name string, bs *roaring64.BSI, m *[]string) {
		bsi(name, bs)
		get(name+"_translate.json", func(b io.Reader) {
			err := json.NewDecoder(b).Decode(m)
			if err != nil {
				panic(fmt.Sprintf("decoding translate file  %s %v", name, err))
			}
		})
	}
	bsi("bot", bot)
	translate("browser", browser, &browser_translate)
	translate("browser_version", browser_version, &browser_version_translate)
	translate("os", os, &os_translate)
	translate("os_version", os_version, &os_version_translate)

	get("fst", func(b io.Reader) {
		all, err := io.ReadAll(b)
		if err != nil {
			panic(fmt.Sprintf("reading fst %v", err))
		}
		fst, err = vellum.Load(all)
		if err != nil {
			panic(fmt.Sprintf("loading fst %v", err))
		}
		dfa, err = levenshtein.NewLevenshteinAutomatonBuilder(2, false)
		if err != nil {
			panic(fmt.Sprintf("building dfa %v", err))
		}
	})
}

func Get(agent string) (a *v1.Agent, err error) {
	if d := cache.Get(nil, []byte(agent)); d != nil {
		a = &v1.Agent{}
		err = proto.Unmarshal(d, a)
		return
	}
	var fa *levenshtein.DFA
	fa, err = dfa.BuildDfa(agent, 2)
	if err != nil {
		return
	}
	it, err := fst.Search(fa, nil, nil)
	for err == nil {
		_, val := it.Current()
		a = &v1.Agent{
			Bot:            isBot(val),
			Browser:        str(browser, browser_translate, val),
			BrowserVersion: str(browser_version, browser_version_translate, val),
			Os:             str(os, os_translate, val),
			OsVersion:      str(os_version, os_version_translate, val),
		}
		b, _ := proto.Marshal(a)
		cache.Set([]byte(agent), b)
		return
	}
	return
}

func isBot(id uint64) bool {
	_, ok := bot.GetValue(id)
	return ok
}

func str(b *roaring64.BSI, tr []string, id uint64) string {
	v, ok := b.GetValue(id)
	if !ok {
		return ""
	}
	return tr[v]
}
