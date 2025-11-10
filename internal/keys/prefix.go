package keys

import v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"

var (
	DataPrefix         = []byte{byte(v1.Prefix_Data)}
	DataExistsPrefix   = []byte{byte(v1.Prefix_DataExists)}
	Cookie             = []byte{byte(v1.Prefix_Cookie)}
	SitePrefix         = []byte{byte(v1.Prefix_SITE)}
	AcmePrefix         = []byte{byte(v1.Prefix_Acme)}
	APIKeyNamePrefix   = []byte{byte(v1.Prefix_APIKeyName)}
	APIKeyHashPrefix   = []byte{byte(v1.Prefix_APIKeyHash)}
	AdminPrefix        = []byte{byte(v1.Prefix_ADMIN)}
	TranslateKeyPrefix = []byte{byte(v1.Prefix_TranslateKey)}
	TranslateIDPrefix  = []byte{byte(v1.Prefix_TranslateID)}
	TranslateSeqPrefix = []byte{byte(v1.Prefix_TranslateSeq)}
	CountryCodePrefix  = []byte{byte(v1.Prefix_CountryCode)}
	RegionCodePrefix   = []byte{byte(v1.Prefix_RegionCode)}
	RegionNamePrefix   = []byte{byte(v1.Prefix_RegionName)}
	CityNamePrefix     = []byte{byte(v1.Prefix_CityName)}
	CityCodePrefix     = []byte{byte(v1.Prefix_CityCode)}
	ReferralPrefix     = []byte{byte(v1.Prefix_Referral)}
)
