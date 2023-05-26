package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/pkg/log"
)

type Buffer struct {
	entries  []*Entry
	mu       sync.Mutex
	buf      bytes.Buffer
	segments Segments
	id       [16]byte
}

func (b *Buffer) Init(uid, sid uint64, ttl time.Duration) *Buffer {
	binary.BigEndian.PutUint64(b.id[:8], uid)
	binary.BigEndian.PutUint64(b.id[8:], sid)
	return b
}

func (b *Buffer) Clone() *Buffer {
	o := bigBufferPool.Get().(*Buffer)
	copy(o.id[:], b.id[:])
	o.entries = append(o.entries, b.entries...)
	return o
}

// Clones b and save this to the data store in a separate goroutine. b is reset.
func (b *Buffer) Save(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()
	clone := b.Clone()
	b.entries = b.entries[:0]
	go Save(ctx, clone)
}

func (b *Buffer) Reset() *Buffer {
	for _, e := range b.entries {
		e.Release()
	}
	b.buf.Reset()
	b.segments.Reset()
	b.entries = b.entries[:0]
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
		b.entries = append(b.entries, updated, s)
		b.persist(ctx, updated)
		return
	}
	newSession := e.Session()
	b.entries = append(b.entries, newSession)
	b.persist(ctx, newSession)
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
	return b.segments.Build(ctx, b.entries, f)
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

	key [4]byte
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

func (m *MultiEntry) Compute(
	start, end int,
	pick func(*MultiEntry, int) (string, bool),
	seen func(uint64) bool,
	f func(uint64, string, *Sum) error,
) error {
	seg := make(map[string]*computed)
	for i := start; i < end; i++ {
		key, ok := pick(m, i)
		if !ok {
			continue
		}
		e, ok := seg[key]
		if !ok {
			e = &computed{}
			seg[key] = e
		}
		e.signSum += m.Sign[i]
		if m.IsBounce[i] {
			e.bounce += m.Sign[i]
		}
		e.views += m.PageViews[i] * m.Sign[i]
		e.events += m.Events[i] * m.Sign[i]
		if !seen(m.UserId[i]) {
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

func PickProp(ctx context.Context, p Property) func(m *MultiEntry, i int) (string, bool) {
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

type Segments struct {
	ls  [City + 1][]int
	key [4]byte
	sum Sum
}

func (e *Segments) Reset() {
	for i := 0; i < len(e.ls); i++ {
		e.ls[i] = e.ls[i][:0]
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

func (e *Segments) Build(ctx context.Context, ls []*Entry, f func(Property, string, uint64, *Sum) error) error {
	// We capitalize on badger Transaction to globally track unique visitors in
	// this entries batch.
	//
	// txn holds visible user_id seen on Base property. This ensure we correctly account
	// for all buffered entries.
	//
	// seen function will use this for Base property. All other properties create a new
	// transaction before calling e.compute and the transaction is discarded there
	// after to ensure we only commit Base user_id but still be able to correctly
	// detect unique visitors within e.compute calls (Which operate on unique entry keys
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
	uniq := seen(ctx, txn, e.key[:], mls)
	for i, v := range ls {
		e.ls[Base] = append(e.ls[Base], i)
		if v.Name != "" {
			e.ls[Event] = append(e.ls[Event], i)
		}
		if v.Pathname != "" {
			e.ls[Page] = append(e.ls[Event], i)
		}
		if v.EntryPage != "" {
			e.ls[EntryPage] = append(e.ls[Event], i)
		}

		if v.ExitPage != "" {
			e.ls[ExitPage] = append(e.ls[Event], i)
		}

		if v.Referrer != "" {
			e.ls[Referrer] = append(e.ls[Event], i)
		}

		if v.ScreenSize != "" {
			e.ls[UtmDevice] = append(e.ls[Event], i)
		}

		if v.UtmMedium != "" {
			e.ls[UtmMedium] = append(e.ls[Event], i)
		}

		if v.UtmSource != "" {
			e.ls[UtmSource] = append(e.ls[Event], i)
		}

		if v.UtmCampaign != "" {
			e.ls[UtmCampaign] = append(e.ls[Event], i)
		}

		if v.UtmContent != "" {
			e.ls[UtmContent] = append(e.ls[Event], i)
		}

		if v.UtmTerm != "" {
			e.ls[UtmTerm] = append(e.ls[Event], i)
		}

		if v.OperatingSystem != "" {
			e.ls[Os] = append(e.ls[Event], i)
		}

		if v.OperatingSystemVersion != "" {
			e.ls[OsVersion] = append(e.ls[Event], i)
		}

		if v.Browser != "" {
			e.ls[UtmBrowser] = append(e.ls[Event], i)
		}

		if v.BrowserVersion != "" {
			e.ls[BrowserVersion] = append(e.ls[Event], i)
		}

		if v.Subdivision1Code != "" {
			e.ls[Region] = append(e.ls[Event], i)
		}

		if v.CountryCode != "" {
			e.ls[Country] = append(e.ls[Event], i)
		}

		if v.City != "" {
			e.ls[City] = append(e.ls[Event], i)
		}
	}
	for i := 0; i < len(e.ls); i++ {
		p := Property(i)
		err := chunk(e.ls[i], ls, func(hs []int) error {
			// First we group all entries by .Hours
			if p != Base {
				// sort non Base properties before chunking them by property.
				sort.Slice(hs, func(i, j int) bool {
					return p.Index(ls[hs[i]]) < p.Index(ls[hs[j]])
				})
			}
			// Then we group  hs by Property.Index and emitting each observed
			// Property.Index to the callback.
			return chunkPropKey(p, hs, ls, func(cp []int) error {
				// The cp list contains entries
				// - Occurring in the same hour
				// - Belonging to the same Property.Index
				el := ls[cp[0]]
				uq, done := uniq(p)
				e.compute(uq, cp, ls)
				done()
				return f(p, p.Index(el), el.Hours, &e.sum)
			})
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type seenFunc func(Property) (func(uint64) bool, func())

func chunk(a []int, ls []*Entry, f func([]int) error) error {
	if len(ls) < 2 {
		return nil
	}
	var pos int
	var last, curr uint64
	for i, v := range a {
		e := ls[v]
		curr = e.Hours
		if i > 0 && curr != last {
			f(a[pos:i])
			pos = i
		}
		last = curr
	}
	if pos < len(ls) {
		return f(a[pos:])
	}
	return nil
}

func chunkPropKey(p Property, a []int, ls []*Entry, f func([]int) error) error {
	if len(ls) < 2 {
		return nil
	}
	if p == Base {
		return f(a)
	}

	var pos int
	var last, curr uint64
	for i, v := range a {
		e := ls[v]
		curr = e.Hours
		if i > 0 && curr != last {
			f(a[pos:i])
			pos = i
		}
		last = curr
	}
	if pos < len(ls) {
		return f(a[pos:])
	}
	return nil
}

func (e *Segments) compute(seen func(uint64) bool, a []int, ls []*Entry) {
	e.sum = Sum{}
	sum := &e.sum

	var signSum, bounce, views, events, visitors int32
	var duration float64
	for _, i := range a {
		ee := ls[i]
		signSum += ee.Sign
		bounce += ee.Bounce() * ee.Sign
		views += ee.PageViews * ee.Sign
		events += ee.Events * ee.Sign
		if !seen(ee.UserId) {
			visitors += 1
		}
		duration += ee.Duration * float64(ee.Sign)
	}
	sum.BounceRate = uint32(math.Round(float64(bounce) / float64(signSum) * 100))
	sum.Visits = uint32(signSum)
	sum.Views = uint32(views)
	sum.Events = uint32(events)
	sum.Visitors = uint32(visitors)
	sum.VisitDuration = uint32(math.Round(duration / float64(signSum)))
	sum.ViewsPerVisit = uint32(math.Round(float64(views) / float64(signSum)))
}

func (p Property) Index(e *Entry) string {
	switch p {
	case Base:
		return BaseKey
	case Event:
		return e.Name
	case Page:
		return e.Pathname
	case EntryPage:
		return e.EntryPage
	case ExitPage:
		return e.ExitPage
	case Referrer:
		return e.Referrer
	case UtmMedium:
		return e.UtmMedium
	case UtmSource:
		return e.UtmSource
	case UtmCampaign:
		return e.UtmCampaign
	case UtmContent:
		return e.UtmContent
	case UtmTerm:
		return e.UtmTerm
	case UtmDevice:
		return e.ScreenSize
	case UtmBrowser:
		return e.Browser
	case BrowserVersion:
		return e.BrowserVersion
	case Os:
		return e.OperatingSystem
	case OsVersion:
		return e.OperatingSystemVersion
	case Country:
		return e.CountryCode
	case Region:
		return e.Referrer
	case City:
		return e.City
	default:
		return ""
	}
}
