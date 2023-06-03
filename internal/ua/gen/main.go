package main

import (
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/ua/gen/bot"
	"github.com/vinceanalytics/vince/internal/ua/gen/client"
	"github.com/vinceanalytics/vince/internal/ua/gen/device"
	"github.com/vinceanalytics/vince/internal/ua/gen/index"
	uos "github.com/vinceanalytics/vince/internal/ua/gen/os"
	"github.com/vinceanalytics/vince/internal/ua/gen/vendorfragment"
	"github.com/vinceanalytics/vince/tools"
)

const (
	repo = "git@github.com:matomo-org/device-detector.git"
	dir  = "device-detector"
)

func main() {
	root := tools.RootVince()
	tools.EnsureRepo(
		filepath.Join(root, "internal", "ua"),
		repo, dir,
	)
	rootRegex := filepath.Join(root, "/internal/ua", dir, "regexes")
	bot.Make(rootRegex)
	client.Make(rootRegex)
	device.Make(rootRegex)
	uos.Make(rootRegex)
	vendorfragment.Make(rootRegex)
	index.Make()
}
