package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"sort"
	"strconv"
	"sync"
	"time"

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

type EntryItem struct {
	Index int
	Text  string
}

type Segments struct {
	ls [City + 1][]EntryItem
}

func (e *Segments) Reset() {
	for i := 0; i < len(e.ls); i++ {
		e.ls[i] = e.ls[i][:0]
	}
}

func (e *Segments) Build(ls []*Entry) {
	for i, v := range ls {
		if v.Name != "" {
			e.ls[Event] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.Name,
			})
		}
		if v.Pathname != "" {
			e.ls[Page] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.Pathname,
			})
		}

		if v.EntryPage != "" {
			e.ls[EntryPage] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.EntryPage,
			})
		}

		if v.ExitPage != "" {
			e.ls[ExitPage] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.ExitPage,
			})
		}

		if v.Referrer != "" {
			e.ls[Referrer] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.Referrer,
			})
		}

		if v.ScreenSize != "" {
			e.ls[UtmDevice] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.ScreenSize,
			})
		}

		if v.UtmMedium != "" {
			e.ls[UtmMedium] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.UtmMedium,
			})
		}

		if v.UtmSource != "" {
			e.ls[UtmSource] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.UtmSource,
			})
		}

		if v.UtmCampaign != "" {
			e.ls[UtmCampaign] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.UtmCampaign,
			})
		}

		if v.UtmContent != "" {
			e.ls[UtmContent] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.UtmContent,
			})
		}

		if v.UtmTerm != "" {
			e.ls[UtmTerm] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.UtmTerm,
			})
		}

		if v.OperatingSystem != "" {
			e.ls[Os] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.OperatingSystem,
			})
		}

		if v.OperatingSystemVersion != "" {
			e.ls[OsVersion] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.OperatingSystemVersion,
			})
		}

		if v.Browser != "" {
			e.ls[UtmBrowser] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.Browser,
			})
		}

		if v.BrowserVersion != "" {
			e.ls[BrowserVersion] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.BrowserVersion,
			})
		}

		if v.Subdivision1Code != "" {
			e.ls[Region] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.Subdivision1Code,
			})
		}

		if v.CountryCode != "" {
			e.ls[Country] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.CountryCode,
			})
		}

		if v.City != "" {
			e.ls[City] = append(e.ls[Event], EntryItem{
				Index: i,
				Text:  v.City,
			})
		}

	}
	for i := 0; i < len(e.ls); i++ {
		a := e.ls[i]
		sort.Slice(a, func(i, j int) bool {
			return a[i].Text < a[j].Text
		})
	}
}
