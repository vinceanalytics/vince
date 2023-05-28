package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/pkg/log"
)

type Buffer struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	segments MultiEntry
	id       [16]byte
}

func (b *Buffer) AddEntry(e ...*Entry) {
	for _, v := range e {
		b.segments.Append(v)
		v.Release()
	}
}

func (b *Buffer) Init(uid, sid uint64, ttl time.Duration) *Buffer {
	binary.BigEndian.PutUint64(b.id[:8], uid)
	binary.BigEndian.PutUint64(b.id[8:], sid)
	return b
}

func (b *Buffer) Clone() *Buffer {
	o := bigBufferPool.Get().(*Buffer)
	o.segments.Copy(&b.segments)
	copy(o.id[:], b.id[:])
	return o
}

// Clones b and save this to the data store in a separate goroutine. b is reset.
func (b *Buffer) Save(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()
	clone := b.Clone()
	b.segments.Reset()
	go Save(ctx, clone)
}

func (b *Buffer) Reset() *Buffer {
	b.buf.Reset()
	b.segments.Reset()
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

func (b *Buffer) Register(ctx context.Context, e *Entry, prevUserId uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var s *Entry
	s = b.find(ctx, e, e.UserId)
	if s == nil {
		s = b.find(ctx, e, prevUserId)
	}
	if s != nil {
		// free e since we don't use it when doing updates
		defer e.Release()
		updated := s.Update(e)
		updated.Sign = 1
		s.Sign = -1
		b.AddEntry(updated, s)
		b.persist(ctx, updated.Clone())
		return
	}
	newSession := e.Session()
	b.AddEntry(newSession)
	b.persist(ctx, newSession.Clone())
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer)
	},
}

func (b *Buffer) find(ctx context.Context, e *Entry, userId uint64) *Entry {
	v, _ := caches.Session(ctx).Get(b.key(e.Domain, userId))
	if v != nil {
		return v.(*Entry)
	}
	return nil
}

func (b *Buffer) key(domain string, uid uint64) string {
	b.buf.Reset()
	b.buf.WriteString(domain)
	b.buf.WriteString(strconv.FormatUint(uid, 10))
	return b.buf.String()
}

func (b *Buffer) persist(ctx context.Context, s *Entry) {
	caches.Session(ctx).SetWithTTL(b.key(s.Domain, s.UserId), s, 1, 30*time.Minute)
}

func (b *Buffer) Build(ctx context.Context, f func(p Property, key string, ts uint64, sum *Sum) error) error {
	return b.segments.Build(ctx, f)
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
	Subdivision1Code       []string
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
	Duration               []float64
	Start                  []int64
	City                   []string
	PageViews              []int32
	Events                 []int32
	Sign                   []int32
	Hours                  []uint64
	IsBounce               []bool

	key [8]byte
	sum Sum
}

func (m *MultiEntry) Reset() {
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
	m.Subdivision1Code = m.Subdivision1Code[:0]
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
	m.Hours = m.Hours[:0]
	m.IsBounce = m.IsBounce[:0]
}

func (m *MultiEntry) Append(e *Entry) {
	m.UtmMedium = append(m.UtmMedium, e.UtmMedium)
	m.Referrer = append(m.Referrer, e.Referrer)
	m.Domain = append(m.Domain, e.Domain)
	m.ExitPage = append(m.ExitPage, e.ExitPage)
	m.EntryPage = append(m.EntryPage, e.EntryPage)
	m.Hostname = append(m.Hostname, e.Hostname)
	m.Pathname = append(m.Pathname, e.Pathname)
	m.UtmSource = append(m.UtmSource, e.UtmSource)
	m.ReferrerSource = append(m.ReferrerSource, e.ReferrerSource)
	m.CountryCode = append(m.CountryCode, e.CountryCode)
	m.Subdivision1Code = append(m.Subdivision1Code, e.Subdivision1Code)
	m.Subdivision2Code = append(m.Subdivision2Code, e.Subdivision2Code)
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
	m.Hours = append(m.Hours, e.Hours)
	m.IsBounce = append(m.IsBounce, e.IsBounce)
}

func (m *MultiEntry) Copy(e *MultiEntry) {
	m.UtmMedium = append(m.UtmMedium, e.UtmMedium...)
	m.Referrer = append(m.Referrer, e.Referrer...)
	m.Domain = append(m.Domain, e.Domain...)
	m.ExitPage = append(m.ExitPage, e.ExitPage...)
	m.EntryPage = append(m.EntryPage, e.EntryPage...)
	m.Hostname = append(m.Hostname, e.Hostname...)
	m.Pathname = append(m.Pathname, e.Pathname...)
	m.UtmSource = append(m.UtmSource, e.UtmSource...)
	m.ReferrerSource = append(m.ReferrerSource, e.ReferrerSource...)
	m.CountryCode = append(m.CountryCode, e.CountryCode...)
	m.Subdivision1Code = append(m.Subdivision1Code, e.Subdivision1Code...)
	m.Subdivision2Code = append(m.Subdivision2Code, e.Subdivision2Code...)
	m.TransferredFrom = append(m.TransferredFrom, e.TransferredFrom...)
	m.UtmCampaign = append(m.UtmCampaign, e.UtmCampaign...)
	m.OperatingSystem = append(m.OperatingSystem, e.OperatingSystem...)
	m.Browser = append(m.Browser, e.Browser...)
	m.UtmTerm = append(m.UtmTerm, e.UtmTerm...)
	m.Name = append(m.Name, e.Name...)
	m.ScreenSize = append(m.ScreenSize, e.ScreenSize...)
	m.BrowserVersion = append(m.BrowserVersion, e.BrowserVersion...)
	m.OperatingSystemVersion = append(m.OperatingSystemVersion, e.OperatingSystemVersion...)
	m.UtmContent = append(m.UtmContent, e.UtmContent...)
	m.UserId = append(m.UserId, e.UserId...)
	m.SessionId = append(m.SessionId, e.SessionId...)
	m.Timestamp = append(m.Timestamp, e.Timestamp...)
	m.Duration = append(m.Duration, e.Duration...)
	m.Start = append(m.Start, e.Start...)
	m.City = append(m.City, e.City...)
	m.PageViews = append(m.PageViews, e.PageViews...)
	m.Events = append(m.Events, e.Events...)
	m.Sign = append(m.Sign, e.Sign...)
	m.Hours = append(m.Hours, e.Hours...)
	m.IsBounce = append(m.IsBounce, e.IsBounce...)
}

// Chunk finds same m.Hours values and call f with the index range. m.Hours are
// guaranteed to be sorted in ascending order.
func (m *MultiEntry) Chunk(f func(m *MultiEntry, start, end int) error) error {
	if len(m.Hours) < 2 {
		return nil
	}
	var pos int
	var last, curr uint64
	for i, v := range m.Hours {
		curr = v
		if i > 0 && curr != last {
			err := f(m, pos, i)
			if err != nil {
				return err
			}
			pos = i
		}
		last = curr
	}
	if pos < len(m.Hours) {
		return f(m, pos, len(m.Hours))
	}
	return nil
}

type computed struct {
	signSum, bounce, views, events, visitors int32
	duration                                 float64
	uniq                                     func(uint64) bool
	done                                     func()
}

func (c *computed) Sum(sum *Sum) {
	sum.BounceRate = uint32(math.Round(float64(c.bounce) / float64(c.signSum) * 100))
	sum.Visits = uint32(c.signSum)
	sum.Views = uint32(c.views)
	sum.Events = uint32(c.events)
	sum.Visitors = uint32(c.visitors)
	sum.VisitDuration = uint32(math.Round(c.duration / float64(c.signSum)))
	sum.ViewsPerVisit = uint32(math.Round(float64(c.views) / float64(c.signSum)))
}

func (m *MultiEntry) Build(ctx context.Context, f func(p Property, key string, ts uint64, sum *Sum) error) error {
	// We capitalize on badger Transaction to globally track unique visitors in
	// this entries batch.
	//
	// txn holds visible user_id seen on Base property. This ensure we correctly account
	// for all buffered entries.
	//
	// seen function will use this for Base property. All other properties create a new
	// transaction before calling m.Compute and the transaction is discarded there
	// after to ensure we only commit Base user_id but still be able to correctly
	// detect unique visitors within m.Compute calls (Which operate on unique entry keys
	// over the same hour window)
	txn := GetUnique(ctx).NewTransaction(true)
	mls := newMetaList()
	defer func() {
		err := txn.Commit()
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to commit transaction for unique index")
		}
		txn.Discard()
		mls.Release()
	}()
	uniq := seen(ctx, txn, m.key[:], mls)
	return m.Chunk(func(m *MultiEntry, start, end int) error {
		for i := Base; i <= City; i++ {
			err := m.Compute(i, start, end, PickProp(i), uniq, func(u uint64, s1 string, s2 *Sum) error {
				return f(i, s1, u, s2)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *MultiEntry) Compute(
	prop Property,
	start, end int,
	pick func(*MultiEntry, int) (string, bool),
	seen seenFunc,
	f func(uint64, string, *Sum) error,
) error {
	seg := make(map[string]*computed)
	defer func() {
		for _, v := range seg {
			v.done()
		}
	}()
	for i := start; i < end; i++ {
		key, ok := pick(m, i)
		if !ok {
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
		e.signSum += m.Sign[i]
		if m.IsBounce[i] {
			e.bounce += m.Sign[i]
		}
		e.views += m.PageViews[i] * m.Sign[i]
		e.events += m.Events[i] * m.Sign[i]
		if !e.uniq(m.UserId[i]) {
			e.visitors += 1
		}
		e.duration += m.Duration[i] * float64(m.Sign[i])
	}
	h := m.Hours[start]
	for k, v := range seg {
		v.Sum(&m.sum)
		err := f(h, k, &m.sum)
		if err != nil {
			return err
		}
	}
	return nil
}

func PickProp(p Property) func(m *MultiEntry, i int) (string, bool) {
	switch p {
	case Base:
		return func(m *MultiEntry, i int) (string, bool) {
			return BaseKey, true
		}
	case Event:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.Name[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case Page:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.Pathname[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case EntryPage:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.EntryPage[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case ExitPage:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.ExitPage[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case Referrer:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.Referrer[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case UtmMedium:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.UtmMedium[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case UtmSource:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.UtmSource[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case UtmCampaign:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.UtmCampaign[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case UtmContent:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.UtmContent[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case UtmTerm:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.UtmTerm[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case UtmDevice:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.ScreenSize[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case UtmBrowser:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.Browser[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case BrowserVersion:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.BrowserVersion[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case Os:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.OperatingSystem[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case OsVersion:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.OperatingSystemVersion[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case Country:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.CountryCode[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case Region:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.Subdivision1Code[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	case City:
		return func(m *MultiEntry, i int) (string, bool) {
			key := m.City[i]
			if key == "" {
				return "", false
			}
			return key, true
		}
	default:
		panic("Unknown property value")
	}
}

func seen(ctx context.Context, txn *badger.Txn, buf []byte, mls *metaList) seenFunc {
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
					log.Get(ctx).Err(err).Msg("failed to get key from unique index")
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
