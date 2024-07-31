package ref

import (
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"strings"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/blevesearch/vellum"
	"github.com/blevesearch/vellum/levenshtein"
)

//go:embed favicon
var faviconData embed.FS

var Favicon, _ = fs.Sub(faviconData, "favicon")

//go:embed refs.fst
var data []byte

type Re struct {
	cache *fastcache.Cache
	fst   *vellum.FST
	build *levenshtein.LevenshteinAutomatonBuilder
}

var base = must()

func Search(uri string) (string, error) {
	return base.Search(uri)
}

func must() *Re {
	re, err := New()
	if err != nil {
		panic(err)
	}
	return re
}

func New() (*Re, error) {
	fst, err := vellum.Load(data)
	if err != nil {
		return nil, err
	}
	lb, err := levenshtein.NewLevenshteinAutomatonBuilder(2, false)
	if err != nil {
		return nil, err
	}
	return &Re{
		fst:   fst,
		build: lb,
		cache: fastcache.New(16 << 20),
	}, nil
}

func (r *Re) Search(uri string) (string, error) {
	base, err := clean(uri)
	if err != nil {
		return "", err
	}
	if re := r.cache.Get(nil, []byte(base)); len(re) > 0 {
		return string(re), nil
	}
	dfa, err := r.build.BuildDfa(base, 2)
	if err != nil {
		return "", fmt.Errorf("building dfa%w", err)
	}
	it, err := r.fst.Search(dfa, nil, nil)
	if err != nil {
		return "", fmt.Errorf("searching dfa%w", err)
	}
	_, idx := it.Current()
	r.cache.Set([]byte(base), []byte(all_referrals[idx]))
	return all_referrals[idx], nil
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
