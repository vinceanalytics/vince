package packages

import (
	_ "embed"
)

//go:embed vince/src/index.ts
var VINCE []byte
