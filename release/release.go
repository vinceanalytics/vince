// Package release executes a go scripts that prepares a vince release.
// When preparing release you can follow these step
//
//	DOCS=true go generate ./release
//
// Generates any automatic documentation, such as man pages and completions
// which are packaged with the release.
//
// Commit  any changes on the files from this step before going to the next one.
//
//	VERSION=(major |minor|patch) go generate ./release
//
// Creates annotated tag and push to remote.
//
//	BUILD=true go generate ./release
//
// Builds the binaries/docker images/ apk packages/brew formula and upload them
// to github.
//
//	DOWNLOAD=true go generate ./release
//
// Generate download page based on the build metadata. Make sure you commit the
// changes before moving to the next step.
//
//	SITE=true go generate ./release
//
// Generates documentation ,blog and helm repository. The generated static site
// is pushed to github that deploy the site on https://vinceanalytics.github.io
package release

//go:generate go run gen/docs/main.go
//go:generate go run gen/tag/main.go
//go:generate go run gen/build/main.go
//go:generate go run gen/publish/main.go
