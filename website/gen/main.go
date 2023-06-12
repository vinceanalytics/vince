package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vinceanalytics/vince/pkg/version"
	"github.com/vinceanalytics/vince/tools"
)

const (
	repo = "vinceanalytics.github.io"
)

func main() {
	// build documentation
	tools.ExecPlain("npm", "run", "docs:build")
	// build blog
	tools.ExecPlain("npm", "run", "blog:build")
	tools.CopyDir("blog/.vitepress/dist/", "docs/.vitepress/dist/blog/")
	// We also deploy v8s helm chart as part of the documentation website under
	// the /charts/ path
	//
	// - [0] Package our charts: go generate ./chart will package the chart and
	//   save the package in chart/charts. Note that the packaged charts are not
	//   tracked under version control
	// - [1] Copy packaged charts to generated site
	// - [2] Copy site to docs/vinceanalytics.github.io : This is the github pages
	//   repository for our site.
	// - [3] Generate helm repo index: This step is delayed because we need to include
	//  all helm packages that have already been released in the index.
	//  docs/vinceanalytics.github.io is a local clone of your remote github pages.
	//  All packages are commited to this repo

	root := tools.RootVince()
	//[0]
	tools.ExecPlainWithWorkingPath(
		root,
		"go", "generate", "./chart")
	//[1]
	tools.CopyDir("chart/charts", "website/docs/.vitepress/dist/", root)
	//[2]
	tools.CopyDir(
		"docs/.vitepress/dist/",
		repo,
	)
	//[3]
	// Index from the repo charts file, we need old helm releases to show up on the
	// index as well.
	tools.ExecPlainWithWorkingPath(
		filepath.Join(repo, "charts"),
		"helm", "repo", "index", ".", "--url", "https://vinceanalytics.github.io/charts/",
	)
	commit()
	if m := os.Getenv("RUN"); m != "" {
		tools.ExecPlain("npm", "run", m)
	}
}

func commit() {
	if os.Getenv("SITE") != "" {
		tools.ExecPlain(
			"git", "pull",
		)
		tools.ExecPlain(
			"git", "add", ".",
		)
		rev := tools.ExecCollect(
			"git", "rev-parse", "--short", "HEAD",
		)
		msg := fmt.Sprintf("Build %s-%s", string(version.BuildVersion), rev)
		tools.ExecPlain(
			"git", "commit", "-m", msg,
		)
		tools.ExecPlain(
			"git", "push",
		)
	}
}
