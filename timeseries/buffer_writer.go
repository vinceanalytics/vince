package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/pkg/entry"
	"github.com/gernest/vince/pkg/log"
)

const SessionTime = time.Minute * 10

type Buffer struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	segments MultiEntry
	id       [16]byte
}

func (b *Buffer) AddEntry(e ...*entry.Entry) {
	for _, v := range e {
		b.segments.append(v)
	}
}

func (b *Buffer) Init(uid, sid uint64, ttl time.Duration) *Buffer {
	binary.BigEndian.PutUint64(b.id[:8], uid)
	binary.BigEndian.PutUint64(b.id[8:], sid)
	return b
}

func (b *Buffer) Reset() *Buffer {
	b.buf.Reset()
	b.segments.reset()
	return b
}

func (b *Buffer) UID() uint64 {
	return binary.BigEndian.Uint64(b.id[:8])
}

func (b *Buffer) SID() uint64 {
	return binary.BigEndian.Uint64(b.id[8:])
}

func (b *Buffer) Release() {
	b.Reset()
	bigBufferPool.Put(b)
}

func NewBuffer(uid, sid uint64, ttl time.Duration) *Buffer {
	return bigBufferPool.Get().(*Buffer).Init(uid, sid, ttl)
}

func (b *Buffer) Register(ctx context.Context, e *entry.Entry, prevUserId uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var s *entry.Entry
	x := caches.Session(ctx)

	if o, ok := x.Get(b.key(e.Domain, e.UserId)); ok {
		s = o.(*entry.Entry)
	}
	if s == nil {
		if o, ok := x.Get(b.key(e.Domain, prevUserId)); ok {
			s = o.(*entry.Entry)
		}
	}

	if s != nil {
		// There are cases where the key is expired but still present in the cache
		// waiting for eviction.
		//
		// We make sure the key is not expired yet before updating it.
		if ttl, ok := x.GetTTL(b); ttl != 0 && ok {
			// free e since we don't use it when doing updates
			defer e.Release()
			old := s.Clone()
			// Update modifies s which is still in cache. It is illegal for a session to
			// happen concurrently.
			updated := s.Update(e)
			old.Sign = -1
			b.AddEntry(updated, old)
			old.Release()
			return
		}
	}
	newSession := e.Session()
	b.AddEntry(newSession)
	x.SetWithTTL(b.key(newSession.Domain, newSession.UserId), newSession, 1, SessionTime)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer)
	},
}

func (b *Buffer) key(domain string, uid uint64) string {
	b.buf.Reset()
	b.buf.WriteString(domain)
	b.buf.WriteString(strconv.FormatUint(uid, 10))
	return b.buf.String()
}

func (b *Buffer) Build(ctx context.Context, f func(p Property, key string, sum *Sum) error) error {
	return b.segments.build(ctx, f)
}

type MultiEntry struct {
	UtmMedium              []string
	Referrer               []string
	Domain                 []string
	ExitPage               []string
	EntryPage              []string
	Hostname               []string
	Pathname               []string
	UtmSource              []string
	ReferrerSource         []string
	CountryCode            []string
	Region                 []string
	Subdivision2Code       []string
	TransferredFrom        []string
	UtmCampaign            []string
	OperatingSystem        []string
	Browser                []string
	UtmTerm                []string
	Name                   []string
	ScreenSize             []string
	BrowserVersion         []string
	OperatingSystemVersion []string
	UtmContent             []string
	UserId                 []uint64
	SessionId              []uint64
	Timestamp              []int64
	Duration               []time.Duration
	Start                  []int64
	City                   []string
	PageViews              []int32
	Events                 []int32
	Sign                   []int32
	IsBounce               []bool

	key [8]byte
}

func (m *MultiEntry) reset() {
	m.UtmMedium = m.UtmMedium[:0]
	m.Referrer = m.Referrer[:0]
	m.Domain = m.Domain[:0]
	m.ExitPage = m.ExitPage[:0]
	m.EntryPage = m.EntryPage[:0]
	m.Hostname = m.Hostname[:0]
	m.Pathname = m.Pathname[:0]
	m.UtmSource = m.UtmSource[:0]
	m.ReferrerSource = m.ReferrerSource[:0]
	m.CountryCode = m.CountryCode[:0]
	m.Region = m.Region[:0]
	m.Subdivision2Code = m.Subdivision2Code[:0]
	m.TransferredFrom = m.TransferredFrom[:0]
	m.UtmCampaign = m.UtmCampaign[:0]
	m.OperatingSystem = m.OperatingSystem[:0]
	m.Browser = m.Browser[:0]
	m.UtmTerm = m.UtmTerm[:0]
	m.Name = m.Name[:0]
	m.ScreenSize = m.ScreenSize[:0]
	m.BrowserVersion = m.BrowserVersion[:0]
	m.OperatingSystemVersion = m.OperatingSystemVersion[:0]
	m.UtmContent = m.UtmContent[:0]
	m.UserId = m.UserId[:0]
	m.SessionId = m.SessionId[:0]
	m.Timestamp = m.Timestamp[:0]
	m.Duration = m.Duration[:0]
	m.Start = m.Start[:0]
	m.City = m.City[:0]
	m.PageViews = m.PageViews[:0]
	m.Events = m.Events[:0]
	m.Sign = m.Sign[:0]
	m.IsBounce = m.IsBounce[:0]
}

func (m *MultiEntry) append(e *entry.Entry) {
	m.UtmMedium = append(m.UtmMedium, e.UtmMedium)
	m.Referrer = append(m.Referrer, e.Referrer)
	m.Domain = append(m.Domain, e.Domain)
	m.ExitPage = append(m.ExitPage, e.ExitPage)
	m.EntryPage = append(m.EntryPage, e.EntryPage)
	m.Hostname = append(m.Hostname, e.Hostname)
	m.Pathname = append(m.Pathname, e.Pathname)
	m.UtmSource = append(m.UtmSource, e.UtmSource)
	m.ReferrerSource = append(m.ReferrerSource, e.ReferrerSource)
	m.CountryCode = append(m.CountryCode, e.Country)
	m.Region = append(m.Region, e.Region)
	m.TransferredFrom = append(m.TransferredFrom, e.TransferredFrom)
	m.UtmCampaign = append(m.UtmCampaign, e.UtmCampaign)
	m.OperatingSystem = append(m.OperatingSystem, e.OperatingSystem)
	m.Browser = append(m.Browser, e.Browser)
	m.UtmTerm = append(m.UtmTerm, e.UtmTerm)
	m.Name = append(m.Name, e.Name)
	m.ScreenSize = append(m.ScreenSize, e.ScreenSize)
	m.BrowserVersion = append(m.BrowserVersion, e.BrowserVersion)
	m.OperatingSystemVersion = append(m.OperatingSystemVersion, e.OperatingSystemVersion)
	m.UtmContent = append(m.UtmContent, e.UtmContent)
	m.UserId = append(m.UserId, e.UserId)
	m.SessionId = append(m.SessionId, e.SessionId)
	m.Timestamp = append(m.Timestamp, e.Timestamp)
	m.Duration = append(m.Duration, e.Duration)
	m.Start = append(m.Start, e.Start)
	m.City = append(m.City, e.City)
	m.PageViews = append(m.PageViews, e.PageViews)
	m.Events = append(m.Events, e.Events)
	m.Sign = append(m.Sign, e.Sign)
	m.IsBounce = append(m.IsBounce, e.IsBounce)
}

type computed struct {
	Sum
	uniq func(uint64) bool
	done func()
}

func (m *MultiEntry) build(ctx context.Context, f func(p Property, key string, sum *Sum) error) error {
	// We capitalize on badger Transaction to globally track unique visitors in
	// this entries batch.
	//
	// txn holds visible user_id seen on Base property. This ensure we correctly account
	// for all buffered entries.
	//
	// seen function will use this for Base property. All other properties create a new
	// transaction before calling m.Compute and the transaction is discarded there
	// after to ensure we only commit Base user_id but still be able to correctly
	// detect unique visitors within m.Compute calls.
	txn := GetUnique(ctx).NewTransaction(true)
	mls := newTxnBufferList()
	defer func() {
		err := txn.Commit()
		if err != nil {
			log.Get().Err(err).Msg("failed to commit transaction for unique index")
		}
		txn.Discard()
		mls.Release()
	}()
	uniq := seen(ctx, txn, m.key[:], mls)
	for i := Base; i <= City; i++ {
		err := m.compute(i, choose(i), uniq, func(s1 string, s2 *Sum) error {
			return f(i, s1, s2)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// On the given start ... end key range, calculate Sum for Property prop , group
// by result of calling pick
//   - if  key != "" then  key will be used as group key. Each unique key will have
//     unique Sum
//   - if  key == "" the key is ignored.
//
// seen is used to compute unique users in each grouped key over the range.
//
// f is called on each key:Sum result. The order of the keys is not guaranteed.
func (m *MultiEntry) compute(
	prop Property,
	pick func(*MultiEntry, int) string,
	seen seenFunc,
	f func(string, *Sum) error,
) error {
	seg := make(map[string]*computed)
	defer func() {
		for _, v := range seg {
			v.done()
		}
	}()
	for i := 0; i < len(m.Timestamp); i++ {
		key := pick(m, i)
		if key == "" {
			continue
		}
		e, ok := seg[key]
		if !ok {
			uq, done := seen(prop)
			e = &computed{
				uniq: uq,
				done: done,
			}
			seg[key] = e
		}
		// When a new Event is received sign is set to 1, when an event is in the same
		// session a new Event with updated details is created with sign 1 and a clone
		// of the found Event is updated with sign -1. At any given time an Entry
		// exists with up to date stats fro the session.
		//
		// This means we can correctly track different measurements depending on the
		// Sign of the Entry
		e.Visits += float64(m.Sign[i])
		var bounce int32
		if m.IsBounce[i] {
			bounce = 1
		}
		e.BounceRate += float64(bounce * m.Sign[i])
		e.Views += float64(m.PageViews[i] * m.Sign[i])
		e.Events += float64(m.Events[i] * m.Sign[i])
		if !e.uniq(m.UserId[i]) {
			e.Visitors += 1
		}
		e.VisitDuration += float64(m.Duration[i] * time.Duration(m.Sign[i]))
	}
	for k, v := range seg {
		err := f(k, &v.Sum)
		if err != nil {
			return err
		}
	}
	return nil
}

// returns a function that chooses a key from MultiEntry based on Property p.
// All keys are strings, empty keys are ignored. Base property uses __root__ as
// its key.
func choose(p Property) func(m *MultiEntry, i int) string {
	switch p {
	case Base:
		return func(m *MultiEntry, i int) string {
			return BaseKey
		}
	case Event:
		return func(m *MultiEntry, i int) string {
			return m.Name[i]
		}
	case Page:
		return func(m *MultiEntry, i int) string {
			return m.Pathname[i]
		}
	case EntryPage:
		return func(m *MultiEntry, i int) string {
			return m.EntryPage[i]
		}
	case ExitPage:
		return func(m *MultiEntry, i int) string {
			return m.ExitPage[i]
		}
	case Referrer:
		return func(m *MultiEntry, i int) string {
			return m.Referrer[i]
		}
	case UtmMedium:
		return func(m *MultiEntry, i int) string {
			return m.UtmMedium[i]
		}
	case UtmSource:
		return func(m *MultiEntry, i int) string {
			return m.UtmSource[i]
		}
	case UtmCampaign:
		return func(m *MultiEntry, i int) string {
			return m.UtmCampaign[i]
		}
	case UtmContent:
		return func(m *MultiEntry, i int) string {
			return m.UtmContent[i]
		}
	case UtmTerm:
		return func(m *MultiEntry, i int) string {
			return m.UtmTerm[i]
		}
	case UtmDevice:
		return func(m *MultiEntry, i int) string {
			return m.ScreenSize[i]
		}
	case UtmBrowser:
		return func(m *MultiEntry, i int) string {
			return m.Browser[i]
		}
	case BrowserVersion:
		return func(m *MultiEntry, i int) string {
			return m.BrowserVersion[i]
		}
	case Os:
		return func(m *MultiEntry, i int) string {
			return m.OperatingSystem[i]
		}
	case OsVersion:
		return func(m *MultiEntry, i int) string {
			return m.OperatingSystemVersion[i]
		}
	case Country:
		return func(m *MultiEntry, i int) string {
			return m.CountryCode[i]
		}
	case Region:
		return func(m *MultiEntry, i int) string {
			return m.Region[i]
		}
	case City:
		return func(m *MultiEntry, i int) string {
			return m.City[i]
		}
	default:
		panic("Unknown property value")
	}
}

func seen(ctx context.Context, txn *badger.Txn, buf []byte, mls *txnBufferList) seenFunc {
	use := func(x *badger.Txn) func(uint64) bool {
		return func(u uint64) bool {
			binary.BigEndian.PutUint64(buf, u)
			_, err := x.Get(buf)
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					b := mls.Get()
					b.Write(buf)
					txn.Set(b.Bytes(), []byte{})
				} else {
					log.Get().Err(err).Msg("failed to get key from unique index")
				}
				return false
			}
			return true
		}
	}
	return func(p Property) (func(uint64) bool, func()) {
		if p == Base {
			return use(txn), func() {}
		}
		x := GetUnique(ctx).NewTransaction(true)
		return use(x), func() {
			x.Discard()
		}
	}
}

type seenFunc func(Property) (func(uint64) bool, func())
