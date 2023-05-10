package secrets

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateAPIKey() string {
	k := make([]byte, 64)
	rand.Read(k)
	return base64.URLEncoding.EncodeToString(k)
}
