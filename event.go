package staples

type Event struct {
	Timestamp int64
	ID        int64
	// When a new session is established for the first time we set Bounce to 1, if
	// a user visits another page within the same session for the first time Bounce
	// is set to -1, any subsequent visits within the session sets Bounce to 0.
	//
	// This allows effective calculation of bounce rate by just summing the Bounce
	// column with faster math.Int64.Sum.
	//
	// NOTE: Bounce is calculated per session. We simply want to know if a user
	// stayed and browsed the website.
	Bounce          int64
	Session         int64
	Duration        float64
	Browser         string
	Browser_Version string
	City            string
	Country         string
	Domain          string
	EntryPage       string
	ExitPage        string
	Host            string
	Event           string
	Os              string
	OsVersion       string
	Path            string
	Referrer        string
	ReferrerSource  string
	Region          string
	Screen          string
	UtmCampaign     string
	UtmContent      string
	UtmMedium       string
	UtmSource       string
	UtmTerm         string
}
