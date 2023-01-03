package ua

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

const meta = "\\.+*?()|[]{}^$#"

func IsRe(s string) bool {
	return strings.ContainsAny(s, meta)
}

func Clean(re string) string {
	rg := strings.Replace(re, `/`, `\/`, -1)
	rg = strings.Replace(rg, `++`, `+`, -1)
	rg = strings.Replace(rg, `\_`, `_`, -1)
	// if we find `\_` again, the original was `\\_`,
	// so restore that so the regex engine does not attempt to escape `_`
	rg = strings.Replace(rg, `\_`, `\\_`, -1)

	// only match if useragent begins with given regex or there is no letter before it
	return `(?:^|[^A-Z0-9-_]|[^A-Z0-9-]_|sprd-)(?:` + rg + ")"
}

func IsStdRe(s string) bool {
	r := Clean(s)
	_, err := regexp.Compile(r)
	return err == nil
}

func Read(name string, out any) error {
	path := filepath.Join(os.Getenv("UA_ROOT"), name)
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(f, out)
}
