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
