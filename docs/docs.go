package docs

import (
	"embed"
	"encoding/json"
	"sync"
	"time"
)

//go:generate go run gen/main.go

//go:embed  manifest.json
var FS embed.FS

var modTime = &sync.Map{}

func init() {
	f, err := FS.Open("manifest.json")
	if err != nil {
		panic("failed to open manifest file " + err.Error())
	}
	defer f.Close()
	o := []struct {
		Path string
		Mod  time.Time
	}{}
	err = json.NewDecoder(f).Decode(&o)
	if err != nil {
		panic("failed to decode manifest file " + err.Error())
	}
	for _, v := range o {
		modTime.Store(v.Path, v.Mod)
	}
}

func ModTime(path string) time.Time {
	v, ok := modTime.Load(path)
	if !ok {
		return time.Time{}
	}
	return v.(time.Time)
}
