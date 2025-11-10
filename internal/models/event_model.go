package models

import (
	"time"

	"github.com/vinceanalytics/vince/internal/util/xtime"
)

//go:generate go run gen/main.go
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
	Bounce           int8
	View             bool
	Session          bool
}

type Cached struct {
	EntryPage []byte
	Host      []byte
	ExitPage  []byte
	Start     int64
	Timestamp int64
	Bounce    int8
}

func (m *Model) Cached() *Cached {
	return &Cached{
		EntryPage: m.EntryPage,
		ExitPage:  m.ExitPage,
		Timestamp: m.Timestamp,
		Start:     m.Timestamp,
		Bounce:    m.Bounce,
	}
}

var maxSession = (15 * time.Minute).Milliseconds()

func (m *Model) Update(session *Cached) *Cached {
	// check if the session has already expied
	if m.Timestamp-session.Start >= maxSession {
		//drop existing session and create a new on
		return m.Cached()
	}
	if session.Bounce == 1 {
		session.Bounce, m.Bounce = -1, -1
	} else {
		session.Bounce, m.Bounce = 0, 0
	}
	m.Session = false
	if len(session.EntryPage) == 0 && m.View {
		m.EntryPage = m.Page
	} else {
		m.EntryPage = session.EntryPage
	}
	if m.View && len(session.Host) == 0 {
	} else {
		m.Host = session.Host
	}
	if m.View {
		m.ExitPage = m.Page
	} else {
		m.ExitPage = session.ExitPage
	}
	m.Duration = int64(xtime.UnixMilli(m.Timestamp).Sub(xtime.UnixMilli(session.Timestamp)))
	session.Timestamp = m.Timestamp
	return nil
}

type Agent struct {
	Device, Os, OsVersion, Browser, BrowserVersion []byte
}
