package cache

import (
	"time"
	"unique"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/util/lru"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

var maxSession = (15 * time.Minute).Milliseconds()

type stringHandle = unique.Handle[string]

type value struct {
	Host      stringHandle
	EntryPage stringHandle
	ExitPage  stringHandle
	Start     int64
	Timestamp int64
	Bounce    int8
}

type Cache struct {
	base *lru.Cache[uint64, *value]
}

func New() *Cache {
	return &Cache{base: lru.New[uint64, *value](1 << 20)}
}

func (ca *Cache) Update(m *models.Model) {
	m.Hit()
	if cached, ok := ca.base.Get(m.Id); ok {
		if x := update(m, cached); x != nil {
			ca.base.Add(m.Id, x)
		}
		return
	}
	m.NewSession()
	ca.base.Add(m.Id, cached(m))
}

func update(m *models.Model, session *value) *value {
	// check if the session has already expied
	if m.Timestamp-session.Start >= maxSession {
		//drop existing session and create a new on
		return cached(m)
	}
	if session.Bounce == 1 {
		session.Bounce, m.Bounce = -1, -1
	} else {
		session.Bounce, m.Bounce = 0, 0
	}
	m.Session = false
	if session.EntryPage.Value() == "" && m.View {
		m.EntryPage = m.Page
	} else {
		m.EntryPage = []byte(session.EntryPage.Value())
	}
	if m.View && session.Host.Value() == "" {
	} else {
		m.Host = []byte(session.Host.Value())
	}
	if m.View {
		m.ExitPage = m.Page
	} else {
		m.ExitPage = []byte(session.ExitPage.Value())
	}
	m.Duration = int64(xtime.UnixMilli(m.Timestamp).Sub(xtime.UnixMilli(session.Timestamp)))
	session.Timestamp = m.Timestamp
	return nil
}

func cached(m *models.Model) *value {
	return &value{
		EntryPage: unique.Make(string(m.EntryPage)),
		ExitPage:  unique.Make(string(m.ExitPage)),
		Timestamp: m.Timestamp,
		Start:     m.Timestamp,
		Bounce:    m.Bounce,
	}
}
