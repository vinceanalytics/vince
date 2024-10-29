package keys

const (
	base byte = 16 + iota
	data
	ops
	tr
)

var (
	DataPrefix         = []byte{data}
	Cookie             = []byte{ops, 0x00}
	SitePrefix         = []byte{ops, 0x01}
	AcmePrefix         = []byte{ops, 0x02}
	APIKeyNamePrefix   = []byte{ops, 0x03}
	APIKeyHashPrefix   = []byte{ops, 0x04}
	AdminPrefix        = []byte{ops, 0x05}
	TranslateKeyPrefix = []byte{tr, 0x00}
	TranslateIDPrefix  = []byte{tr, 0x01}
	TranslateSeqPrefix = []byte{tr, 0x02}
)
