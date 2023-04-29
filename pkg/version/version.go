package version

import (
	"bytes"
	"fmt"
	"os"
	"runtime/debug"
	"text/tabwriter"
	"text/template"

	"github.com/urfave/cli/v2"
)

func Version() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "prints version information",
		Action: func(ctx *cli.Context) error {
			build, ok := debug.ReadBuildInfo()
			if !ok {
				return nil
			}
			var b bytes.Buffer
			tab := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
			fmt.Fprintf(tab, "Goversion\t%s\t\n", build.GoVersion)
			fmt.Fprintf(tab, "Package\t%s\t\n", build.Path)
			fmt.Fprintf(tab, "Version\t%s\t\n", build.Main.Version)
			for _, v := range build.Settings {
				if v.Value != "" {
					fmt.Fprintf(tab, "%s\t%s\t\n", v.Key, v.Value)
				}
			}
			tab.Flush()
			os.Stdout.Write(b.Bytes())
			return nil
		},
	}
}

var tpl = template.Must(template.New("v").Parse(`
GoVersion: 
`))
