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
package release

//go:generate go run gen/docs/main.go
//go:generate go run gen/tag/main.go
