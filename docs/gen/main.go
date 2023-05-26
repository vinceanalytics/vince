package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gernest/vince/tools"
	"moul.io/http2curl"
)

const (
	host   = "http://localhost:8080"
	bearer = "$VINCE_BOOTSTRAP_KEY"
	domain = "vince.example.com"
)

func main() {
	createSite()
}

func createSite() {
	b, _ := json.Marshal(map[string]string{
		"domain": domain,
	})
	r, _ := http.NewRequest(http.MethodPost, host+"/api/v1/sites", bytes.NewReader(b))
	r.Header.Set("Authorization", "Bearer "+bearer)

	write("guide/files/create_site.sh", r)
}

func write(path string, r *http.Request) {
	cmd, _ := http2curl.GetCurlCommand(r)
	tools.WriteFile(path, []byte(breakDown(cmd.String())))
}

func breakDown(a string) string {
	a = strings.ReplaceAll(a, "' -d '", "' \\\n\t-d '")
	a = strings.ReplaceAll(a, "' -H '", "' \\\n\t-H '")
	return a
}
