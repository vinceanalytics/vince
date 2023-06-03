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
	root := tools.RootVince()
	if os.Getenv("SITE") != "" {
		// make sure we have latest changes to the main repo
		website := filepath.Join(root, "website")
		git := filepath.Join(website, repo)
		tools.ExecPlainWithWorkingPath(
			git,
			"git", "pull",
		)

		tools.ExecPlainWithWorkingPath(
			website,
			"npm", "run", "docs:build",
		)
		tools.ExecPlainWithWorkingPath(
			website,
			"npm", "run", "blog:build",
		)
		from := "blog/.vitepress/dist/"
		to := "docs/.vitepress/dist/blog/"
		tools.CopyDir(from, to, website)

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

		//[0]
		tools.ExecPlainWithWorkingPath(root,
			"go", "generate", "./chart",
		)
		//[1]
		tools.CopyDir("chart/charts", "website/docs/.vitepress/dist/", root)
		//[2]
		tools.CopyDir(
			"docs/.vitepress/dist/",
			repo,
			website)
		//[3]
		tools.ExecPlainWithWorkingPath(
			filepath.Join(git, "charts"),
			"helm", "repo", "index", ".", "--url", "https://vinceanalytics.github.io/charts/",
		)

		// commit changes to the repository.
		tools.ExecPlainWithWorkingPath(
			git,
			"git", "pull",
		)
		tools.ExecPlainWithWorkingPath(
			git,
			"git", "add", ".",
		)
		rev := tools.ExecCollect(
			"git", "rev-parse", "--short", "HEAD",
		)
		msg := fmt.Sprintf("Build %s-%s", string(version.BuildVersion), rev)
		tools.ExecPlainWithWorkingPath(
			git,
			"git", "commit", "-m", msg,
		)

		tools.ExecPlainWithWorkingPath(
			git,
			"git", "push",
		)
	}

}
