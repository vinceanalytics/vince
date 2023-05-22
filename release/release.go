// Package release executes a go script that prepares a vince release. Releases
// follow semver versioning. Use environment variable VERSION to control which
// version to release.
//
//	VERSION = major | minor | patch
//
// This dictate which component for semver to increment. Say, the latest tag is
// v0.0.0 setting VERSION=major will increment major component yielding v1.0.0.
//
// PRERELEASE env var is used to set prerelease/build part of the version.
//
// To execute this run this on the root of the project.
//
//	VERSION=patch go generate ./release/prepare
package release

//go:generate go run gen/docs/main.go
//go:generate go run gen/tag/main.go
//go:generate go run gen/build/main.go
