package timeseries

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/spec"
)

type Buffer struct {
	mu       sync.Mutex
	segments MultiEntry
	id       [16]byte
	ttl      time.Duration
}

func (b *Buffer) AddEntry(e ...*entry.Entry) {
	for _, v := range e {
		b.segments.append(v)
	}
}

func (b *Buffer) Init(uid, sid uint64, ttl time.Duration) *Buffer {
	binary.BigEndian.PutUint64(b.id[:8], uid)
	binary.BigEndian.PutUint64(b.id[8:], sid)
	b.ttl = ttl
	return b
}

func (b *Buffer) hasEntries() bool {
	return len(b.segments.Timestamp) > 0
}

func (b *Buffer) Reset() *Buffer {
	b.segments.reset()
	return b
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
	cacheKey := e.UserId
	if o, ok := x.Get(cacheKey); ok {
		s = o.(*entry.Entry)
	}
	if s == nil {
		cacheKey = prevUserId
		if o, ok := x.Get(cacheKey); ok {
			s = o.(*entry.Entry)
		}
	}
	if s != nil {
		s.Update(e)
		e.Release()
		return
	}

	// Generate a new session based on event e. This includes creating a new
	// session id and associate it with the user id. Sessions allows us to track
	// bounce rate without fingerprinting the user.
	//
	// Note that this is best case estimates.
	e.Session()
	x.SetWithTTL(e.UserId, e, 1, b.ttl)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer)
	},
}

func (b *Buffer) build(ctx context.Context, f func(p spec.Property, key string, sum *Aggregate) error) error {
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
	Timestamp              []int64
	Duration               []time.Duration
	Start                  []int64
	City                   []string
	PageViews              []uint16
	Events                 []uint16
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
	m.Timestamp = m.Timestamp[:0]
	m.Duration = m.Duration[:0]
	m.Start = m.Start[:0]
	m.City = m.City[:0]
	m.PageViews = m.PageViews[:0]
	m.Events = m.Events[:0]
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
	m.Timestamp = append(m.Timestamp, e.Timestamp)
	m.Duration = append(m.Duration, e.Duration)
	m.Start = append(m.Start, e.Start)
	m.City = append(m.City, e.City)
	m.PageViews = append(m.PageViews, e.PageViews)
	m.Events = append(m.Events, e.Events)
	m.IsBounce = append(m.IsBounce, e.IsBounce)
}

type computed struct {
	Aggregate
	uniq func(uint64) bool
	done func()
}

func (m *MultiEntry) build(ctx context.Context, f func(p spec.Property, key string, sum *Aggregate) error) error {
	// We capitalize on badger Transaction to globally track unique visitors in
	// this entries batch.
	//
	// txn holds visible user_id seen on Base spec. This ensure we correctly account
	// for all buffered entries.
	//
	// seen function will use this for Base spec. All other properties create a new
	// transaction before calling m.Compute and the transaction is discarded there
	// after to ensure we only commit Base user_id but still be able to correctly
	// detect unique visitors within m.Compute calls.
	txn := Unique(ctx).NewTransaction(true)
	mls := newTxnBufferList()
	defer func() {
		err := txn.Commit()
		if err != nil {
			log.Get().Err(err).Msg("failed to commit transaction for unique index")
		}
		txn.Discard()
		mls.release()
	}()
	uniq := seen(ctx, txn, m.key[:], mls)
	for i := spec.Base; i <= spec.City; i++ {
		err := m.compute(i, choose(i), uniq, func(s1 string, s2 *Aggregate) error {
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
	prop spec.Property,
	pick func(*MultiEntry, int) string,
	seen seenFunc,
	f func(string, *Aggregate) error,
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
		e.Visits += 1
		if m.IsBounce[i] {
			e.BounceRates += 1
		}
		e.Views += m.PageViews[i]
		e.Events += m.Events[i]
		if !e.uniq(m.UserId[i]) {
			e.Visitors += 1
		}
		e.VisitDurations += m.Duration[i]
	}
	for k, v := range seg {
		// Visit duration is average. Compute the average before calling f. Avoid
		// division by zero bugs.
		if len(m.Timestamp) > 0 {
			v.Aggregate.VisitDurations = v.Aggregate.VisitDurations / time.Duration(len(m.Timestamp))
		}
		err := f(k, &v.Aggregate)
		if err != nil {
			return err
		}
	}
	return nil
}

// returns a function that chooses a key from MultiEntry based on Property p.
// All keys are strings, empty keys are ignored. Base property uses __root__ as
// its key.
func choose(p spec.Property) func(m *MultiEntry, i int) string {
	switch p {
	case spec.Base:
		return func(m *MultiEntry, i int) string {
			return spec.BaseKey
		}
	case spec.Event:
		return func(m *MultiEntry, i int) string {
			return m.Name[i]
		}
	case spec.Page:
		return func(m *MultiEntry, i int) string {
			return m.Pathname[i]
		}
	case spec.EntryPage:
		return func(m *MultiEntry, i int) string {
			return m.EntryPage[i]
		}
	case spec.ExitPage:
		return func(m *MultiEntry, i int) string {
			return m.ExitPage[i]
		}
	case spec.Referrer:
		return func(m *MultiEntry, i int) string {
			return m.ReferrerSource[i]
		}
	case spec.UtmMedium:
		return func(m *MultiEntry, i int) string {
			return m.UtmMedium[i]
		}
	case spec.UtmSource:
		return func(m *MultiEntry, i int) string {
			return m.UtmSource[i]
		}
	case spec.UtmCampaign:
		return func(m *MultiEntry, i int) string {
			return m.UtmCampaign[i]
		}
	case spec.UtmContent:
		return func(m *MultiEntry, i int) string {
			return m.UtmContent[i]
		}
	case spec.UtmTerm:
		return func(m *MultiEntry, i int) string {
			return m.UtmTerm[i]
		}
	case spec.UtmDevice:
		return func(m *MultiEntry, i int) string {
			return m.ScreenSize[i]
		}
	case spec.UtmBrowser:
		return func(m *MultiEntry, i int) string {
			return m.Browser[i]
		}
	case spec.BrowserVersion:
		return func(m *MultiEntry, i int) string {
			return m.BrowserVersion[i]
		}
	case spec.Os:
		return func(m *MultiEntry, i int) string {
			return m.OperatingSystem[i]
		}
	case spec.OsVersion:
		return func(m *MultiEntry, i int) string {
			return m.OperatingSystemVersion[i]
		}
	case spec.Country:
		return func(m *MultiEntry, i int) string {
			return m.CountryCode[i]
		}
	case spec.Region:
		return func(m *MultiEntry, i int) string {
			return m.Region[i]
		}
	case spec.City:
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
	return func(p spec.Property) (func(uint64) bool, func()) {
		if p == spec.Base {
			return use(txn), func() {}
		}
		x := Unique(ctx).NewTransaction(true)
		return use(x), func() {
			x.Discard()
		}
	}
}

type seenFunc func(spec.Property) (func(uint64) bool, func())
