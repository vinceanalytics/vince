package models

type Model struct {
	ExitPage         []byte
	UtmTerm          []byte
	EntryPage        []byte
	Subdivision1Code []byte
	Event            []byte
	UtmSource        []byte
	UtmMedium        []byte
	Browser          []byte
	BrowserVersion   []byte
	Country          []byte
	Device           []byte
	Domain           []byte
	Subdivision2Code []byte
	UtmContent       []byte
	Page             []byte
	Host             []byte
	Os               []byte
	OsVersion        []byte
	UtmCampaign      []byte
	Referrer         []byte
	Source           []byte
	Timestamp        int64
	Duration         int64
	Id               uint64
	City             uint32
	Bounce           int32
	View             bool
	Session          bool
}
