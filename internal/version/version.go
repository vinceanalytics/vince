package version

import (
	_ "embed"
	"runtime/debug"
	"strings"
)

//go:embed VERSION.txt
var BuildVersion string

type Version struct {
	Commit string
	Date   string
	Dirty  bool
	Valid  bool
}

func (v Version) String() string {
	if v.Commit == "" && v.Date == "" {
		return BuildVersion
	}
	var s strings.Builder
	s.WriteString(BuildVersion)
	if !v.Valid {
		s.WriteString("-ERR-BuildInfo")
	} else {
		s.WriteString("-" + v.Date)
		commit := v.Commit
		if len(commit) > 9 {
			commit = commit[:9]
		}
		s.WriteString("-" + commit)
		if v.Dirty {
			s.WriteString("-dirty")
		}
	}
	return s.String()
}

func Build() Version {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return Version{}
	}
	v := Version{Valid: true}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			v.Commit = s.Value
		case "vcs.time":
			if len(s.Value) >= len("yyyy-mm-dd") {
				v.Date = s.Value[:len("yyyy-mm-dd")]
				v.Date = strings.ReplaceAll(v.Date, "-", "")
			}
		case "vcs.modified":
			v.Dirty = s.Value == "true"
		}
	}
	return v
}
