package keys

var (
	DataPrefix         = []byte{0x00}
	OpsPrefix          = []byte{0x01}
	Cookie             = []byte{0x01, 0x00}
	SitePrefix         = []byte{0x01, 0x01}
	TranslateKeyPrefix = []byte{0x02, 0x00}
	TranslateIDPrefix  = []byte{0x02, 0x01}
	TranslateSeqPrefix = []byte{0x02, 0x02}
)

func Data(key []byte) []byte {
	return append(DataPrefix, key...)
}

func Ops(key []byte) []byte {
	return append(OpsPrefix, key...)
}

func Site(key []byte) []byte {
	return append(SitePrefix, key...)
}

func TranslateKey(key []byte) []byte {
	return append(TranslateKeyPrefix, key...)
}

func TranslateID(key []byte) []byte {
	return append(TranslateIDPrefix, key...)
}

func TranslateSeq(key []byte) []byte {
	return append(TranslateSeqPrefix, key...)
}
