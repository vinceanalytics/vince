package staples

import (
	"sync"
	"time"
)

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
	Entry_Page      string
	Exit_Page       string
	Host            string
	Event           string
	Os              string
	OsVersion       string
	Path            string
	Referrer        string
	Referrer_Source string
	Region          string
	Screen          string
	Utm_Campaign    string
	Utm_Content     string
	Utm_Medium      string
	Utm_Source      string
	Utm_Term        string
}

var eventsPool = &sync.Pool{New: func() any { return new(Event) }}

func NewEvent() *Event {
	return eventsPool.Get().(*Event)
}

func (e *Event) Release() {
	*e = Event{}
	eventsPool.Put(e)
}

func (e *Event) TS() int64 { return e.Timestamp }

func (e *Event) Hit() {
	e.Entry_Page = e.Path
	e.Bounce = 1
	e.Session = 1
}

func (s *Event) Update(e *Event) {
	if s.Bounce == 1 {
		s.Bounce, e.Bounce = -1, -1
	} else {
		s.Bounce, e.Bounce = 0, 0
	}
	e.Session = 0
	e.Exit_Page = e.Path
	// Track duration since last visit.
	e.Duration = time.UnixMilli(e.Timestamp).Sub(time.UnixMilli(s.Timestamp)).Seconds()
	s.Timestamp = e.Timestamp
}
