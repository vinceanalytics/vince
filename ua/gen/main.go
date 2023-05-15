package main

import (
	"os"
	"path/filepath"

	"github.com/gernest/vince/tools"
	"github.com/gernest/vince/ua/gen/bot"
	"github.com/gernest/vince/ua/gen/client"
	"github.com/gernest/vince/ua/gen/device"
	"github.com/gernest/vince/ua/gen/index"
	uos "github.com/gernest/vince/ua/gen/os"
	"github.com/gernest/vince/ua/gen/vendorfragment"
)

const (
	repo = "git@github.com:matomo-org/device-detector.git"
	dir  = "device-detector"
)

func main() {
	root := tools.RootVince()
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			println(">  downloading device-detector")
			tools.ExecPlain("git", "clone", repo)
		} else {
			tools.Exit(err.Error())
		}
	} else {
		// make sure we are up to date
		println(">  updating device-detector")
		tools.ExecPlainWithWorkingPath(
			filepath.Join(root, "ua", dir),
			"git", "pull",
		)
	}
	rootRegex := filepath.Join(root, "ua", dir, "regexes")
	bot.Make(rootRegex)
	client.Make(rootRegex)
	device.Make(rootRegex)
	uos.Make(rootRegex)
	vendorfragment.Make(rootRegex)
	index.Make()
}
