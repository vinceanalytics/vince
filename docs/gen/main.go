package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gernest/vince/tools"
	"moul.io/http2curl"
)

func main() {
	createSite()
}

const host = "http://localhost:8080"
const bearer = "$VINCE_BOOTSTRAP_KEY"

func createSite() {
	b, _ := json.Marshal(map[string]string{
		"domain": "vince.example.com",
	})
	r, _ := http.NewRequest(http.MethodPost, host+"/api/v1/sites/", bytes.NewReader(b))
	r.Header.Set("Authorization", "Bearer "+bearer)

	write("guide/files/create_site.sh", r)
}

func write(path string, r *http.Request) {
	cmd, _ := http2curl.GetCurlCommand(r)
	tools.WriteFile(path, []byte(cmd.String()))
}
