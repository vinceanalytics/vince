package packages

import (
	_ "embed"
)

//go:embed packages/alerts/lib/index.js
var Alerts []byte

//go:embed packages/types/lib/index.js
var Types []byte
