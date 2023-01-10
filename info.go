package vince

import (
	"net/http"
	"runtime/debug"
)

func (v *Vince) info(w http.ResponseWriter, r *http.Request) {
	ServeJSON(w, GetBuildInfo())
}

type BuildInfo struct {
	Version   string
	Commit    string
	CreatedAt string
}

func GetBuildInfo() BuildInfo {
	i, _ := debug.ReadBuildInfo()
	var commit string
	var created string
	for _, e := range i.Settings {
		switch e.Key {
		case "vcs.revision":
			commit = e.Value
		case "vcs.time":
			created = e.Value
		}
	}
	return BuildInfo{
		Version:   i.Main.Version,
		Commit:    commit,
		CreatedAt: created,
	}
}
