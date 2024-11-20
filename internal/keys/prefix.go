package keys

const (
	base byte = 16 + iota
	data
	exists
	ops
	tr

	countryCode
	regionCode
	regionName
	cityName
	cityCode
	referral
)

var (
	DataPrefix         = []byte{data}
	DataExistsPrefix   = []byte{exists}
	Cookie             = []byte{ops, 0x00}
	SitePrefix         = []byte{ops, 0x01}
	AcmePrefix         = []byte{ops, 0x02}
	APIKeyNamePrefix   = []byte{ops, 0x03}
	APIKeyHashPrefix   = []byte{ops, 0x04}
	AdminPrefix        = []byte{ops, 0x05}
	TranslateKeyPrefix = []byte{tr, 0x00}
	TranslateIDPrefix  = []byte{tr, 0x01}
	TranslateSeqPrefix = []byte{tr, 0x02}

	CountryCodePrefix = []byte{countryCode}
	RegionCodePrefix  = []byte{regionCode}
	RegionNamePrefix  = []byte{regionName}
	CityNamePrefix    = []byte{cityName}
	CityCodePrefix    = []byte{cityCode}
	ReferralPrefix    = []byte{referral}
)
