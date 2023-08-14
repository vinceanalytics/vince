package version

import (
	"bytes"
	_ "embed"
	"os"
	"runtime/debug"
	"strings"

	"github.com/urfave/cli/v3"
)

//go:embed VERSION.txt
var BuildVersion []byte

type Version struct {
	Commit string
	Date   string
	Dirty  bool
	Valid  bool
}

func (v Version) String() string {
	var s strings.Builder
	s.Write(bytes.TrimSpace(BuildVersion))
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

func VersionCmd() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "prints version information",
		Action: func(ctx *cli.Context) error {
			x := Build()
			os.Stdout.WriteString(x.String())
			return nil
		},
	}
}
