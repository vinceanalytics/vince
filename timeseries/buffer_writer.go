package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/vince/caches"
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

func (b *Buffer) Build(f func(p Property, key string, ts uint64, sum *Sum)) {
	b.segments.Build(b.entries, f)
}

type EntryItem struct {
	Index int
	Text  string
}

type Segments struct {
	ls   [City + 1][]int
	uniq roaring64.Bitmap
	sum  Sum
}

func (e *Segments) Reset() {
	for i := 0; i < len(e.ls); i++ {
		e.ls[i] = e.ls[i][:0]
	}
	e.uniq.Clear()
}

func (e *Segments) Build(ls []*Entry, f func(Property, string, uint64, *Sum)) {
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
		a := e.ls[i]
		p := Property(i)
		if i != 0 {
			// sort non Base properties
			sort.Slice(a, func(i, j int) bool {
				return p.Index(ls[a[i]]) < p.Index(ls[a[j]])
			})
		}
		chunk(e.ls[i], ls, func(u uint64, i []int) {
			e.compute(i, ls)
			f(p, p.Index(ls[i[0]]), u, &e.sum)
		})
	}
}

func chunk(a []int, ls []*Entry, f func(uint64, []int)) {
	if len(ls) < 2 {
		return
	}
	var pos int
	var last, curr uint64
	for i, v := range a {
		e := ls[v]
		curr = e.Hours
		if i > 0 && curr != last {
			f(ls[a[pos]].Hours, a[pos:i-1])
			pos = i
		}
		last = curr
	}
	if pos < len(ls)-1 {
		f(ls[a[pos]].Hours, a[pos:])
	}
}

func (e *Segments) compute(a []int, ls []*Entry) {
	e.uniq.Clear()
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
		if !e.uniq.Contains(ee.UserId) {
			visitors += 1
			e.uniq.Add(ee.UserId)
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
