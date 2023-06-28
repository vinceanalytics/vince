package names

import (
	"github.com/dlclark/regexp2"
)

var owner = regexp2.MustCompile(`^(?<owner>[a-z\d](?:[a-z\d]|-(?=[a-z\d])){0,38})$`, regexp2.ECMAScript)
var site = regexp2.MustCompile(`^(?<site>\b((?=[a-z0-9-]{1,63}\.)(xn--)?[a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,63}\b)$`, regexp2.ECMAScript)

func Owner(s string) bool {
	ok, _ := owner.MatchString(s)
	return ok
}

func Site(s string) bool {
	ok, _ := site.MatchString(s)
	return ok
}
