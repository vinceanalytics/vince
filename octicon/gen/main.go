package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

var archiveLocation = "https://registry.npmjs.org/@primer/octicons/-/octicons-18.3.0.tgz"

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func getData() (map[string]octicon, error) {
	resp, err := http.Get(archiveLocation)
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err != nil {
			return nil, err
		}

		if header.Name != "package/build/data.json" {
			continue
		}

		var octicons map[string]octicon
		err = json.NewDecoder(tarReader).Decode(&octicons)
		return octicons, err
	}
}

func run() error {
	octicons, err := getData()
	if err != nil {
		return err
	}

	var names []string
	for name := range octicons {
		names = append(names, name)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	fmt.Fprint(&buf, `package octicon

import (
	"fmt"
	"html/template"
	"strings"
)

// Icon returns a string representing the named Octicon.
func Icon(name string, height int, extraClasses ...string) (string, bool) {
	switch name {
`)
	for _, name := range names {
		fmt.Fprintf(&buf, "	case %q:\n		return %v(height, extraClasses...)\n", name, kebabToCamelCase(name))
	}
	fmt.Fprintf(&buf, `	default:
		return "", false
	}
}

func IconTemplateFunc(name string, height int, extraClasses ...string) (template.HTML, error) {
	i, ok := Icon(name, height, extraClasses...)
	if !ok {
		return "", fmt.Errorf("unknown icon (%%s) or height (%%d)", name, height)
	}

	return template.HTML(i), nil
}
`)

	// Write all individual Octicon functions.
	for _, name := range names {
		generateAndWriteOcticon(&buf, octicons[name])
	}

	err = os.WriteFile("octicon.go", buf.Bytes(), 0777)
	return err
}

type octicon struct {
	Name    string
	Heights map[string]struct {
		Width int
		Path  string
	}
}

func generateAndWriteOcticon(w io.Writer, icon octicon) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, "// %s returns a string representing an %q Octicon.\n", kebabToCamelCase(icon.Name), icon.Name)
	fmt.Fprintf(w, "func %s(height int, rawExtraClasses ...string) (string, bool) {\n", kebabToCamelCase(icon.Name))
	fmt.Fprintf(w, "	extraClasses := strings.Join(rawExtraClasses, \" \")\n")
	fmt.Fprintf(w, "	if extraClasses != \"\" {\n")
	fmt.Fprintf(w, "		extraClasses = \" \" + extraClasses\n")
	fmt.Fprintf(w, "	}\n")
	fmt.Fprintf(w, "	switch height {\n")

	var heights []string
	for height := range icon.Heights {
		heights = append(heights, height)
	}
	sort.Strings(heights)

	for _, height := range heights {
		heightInfo := icon.Heights[height]
		fmt.Fprintf(w, "		case %s:\n", height)
		fmt.Fprintf(w, "			return fmt.Sprintf(%q, extraClasses), true\n", fmt.Sprintf(`<svg class="octicon octicon-%s%%s" height="%s" width="%d" viewbox="0 0 %s %d" aria-hidden="true">%s</svg>`, icon.Name, height, heightInfo.Width, height, heightInfo.Width, heightInfo.Path))
	}

	fmt.Fprintf(w, "		default:\n")
	fmt.Fprintf(w, "			return \"\", false")
	fmt.Fprint(w, " }")
	fmt.Fprintln(w, "}")
}

func kebabToCamelCase(kebab string) (camelCase string) {
	isToUpper := true
	for _, runeValue := range kebab {
		if isToUpper {
			camelCase += strings.ToUpper(string(runeValue))
			isToUpper = false
		} else {
			if runeValue == '-' {
				isToUpper = true
			} else {
				camelCase += string(runeValue)
			}
		}
	}
	return
}
