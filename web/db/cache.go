// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE_list file.

package db

import (
	"time"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
)

type LRU struct {
	size              int
	evictList         *List
	items             map[uint64]*CacheEntry
	ttl               time.Duration
	buckets           []bucket
	nextCleanupBucket uint8
}

const numBuckets = 100

func newCache(size uint, ttl time.Duration) *LRU {
	res := &LRU{
		size:      int(size),
		evictList: NewList[uint64, *v1.Model](),
		items:     make(map[uint64]*Entry[uint64, *v1.Model]),
		ttl:       ttl,
	}
	// initialize the buckets
	res.buckets = make([]bucket, numBuckets)
	for i := 0; i < numBuckets; i++ {
		res.buckets[i] = bucket{entries: make(map[uint64]*CacheEntry)}
	}
	return res
}

func (c *LRU) Add(key uint64, value *v1.Model) (evicted bool) {
	now := time.Now()

	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		c.removeFromBucket(ent) // remove the entry from its current bucket as expiresAt is renewed
		ent.Value = value
		ent.ExpiresAt = now.Add(c.ttl)
		c.addToBucket(ent)
		return false
	}

	// Add new item
	ent := c.evictList.PushFrontExpirable(key, value, now.Add(c.ttl))
	c.items[key] = ent
	c.addToBucket(ent) // adds the entry to the appropriate bucket and sets entry.expireBucket

	evict := c.size > 0 && c.evictList.Length() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return evict
}

func (c *LRU) Get(key uint64) (value *v1.Model, ok bool) {
	var ent *CacheEntry
	if ent, ok = c.items[key]; ok {
		// Expired item check
		if time.Now().After(ent.ExpiresAt) {
			return value, false
		}
		c.removeElement(ent)
		return ent.Value, true
	}
	return
}

func (c *LRU) removeOldest() {
	if ent := c.evictList.Back(); ent != nil {
		c.removeElement(ent)
	}
}

func (c *LRU) removeElement(e *CacheEntry) {
	c.evictList.Remove(e)
	delete(c.items, e.Key)
	c.removeFromBucket(e)
}

func (c *LRU) addToBucket(e *CacheEntry) {
	bucketID := (numBuckets + c.nextCleanupBucket - 1) % numBuckets
	e.ExpireBucket = bucketID
	c.buckets[bucketID].entries[e.Key] = e
	if c.buckets[bucketID].newestEntry.Before(e.ExpiresAt) {
		c.buckets[bucketID].newestEntry = e.ExpiresAt
	}
}

func (c *LRU) removeFromBucket(e *CacheEntry) {
	delete(c.buckets[e.ExpireBucket].entries, e.Key)
}

type bucket struct {
	entries     map[uint64]*CacheEntry
	newestEntry time.Time
}

type CacheEntry = Entry[uint64, *v1.Model]
type List = LruList[uint64, *v1.Model]

// Entry is an LRU Entry
type Entry[K comparable, V any] struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Entry[K, V]

	// The list to which this element belongs.
	list *LruList[K, V]

	// The LRU Key of this element.
	Key K

	// The Value stored with this element.
	Value V

	// The time this element would be cleaned up, optional
	ExpiresAt time.Time

	// The expiry bucket item was put in, optional
	ExpireBucket uint8
}

// PrevEntry returns the previous list element or nil.
func (e *Entry[K, V]) PrevEntry() *Entry[K, V] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// LruList represents a doubly linked list.
// The zero Value for LruList is an empty list ready to use.
type LruList[K comparable, V any] struct {
	root Entry[K, V] // sentinel list element, only &root, root.prev, and root.next are used
	len  int         // current list Length excluding (this) sentinel element
}

// Init initializes or clears list l.
func (l *LruList[K, V]) Init() *LruList[K, V] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// NewList returns an initialized list.
func NewList[K comparable, V any]() *LruList[K, V] { return new(LruList[K, V]).Init() }

// Length returns the number of elements of list l.
// The complexity is O(1).
func (l *LruList[K, V]) Length() int { return l.len }

// Back returns the last element of list l or nil if the list is empty.
func (l *LruList[K, V]) Back() *Entry[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero List Value.
func (l *LruList[K, V]) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *LruList[K, V]) insert(e, at *Entry[K, V]) *Entry[K, V] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Entry{Value: v, ExpiresAt: ExpiresAt}, at).
func (l *LruList[K, V]) insertValue(k K, v V, expiresAt time.Time, at *Entry[K, V]) *Entry[K, V] {
	return l.insert(&Entry[K, V]{Value: v, Key: k, ExpiresAt: expiresAt}, at)
}

// Remove removes e from its list, decrements l.len
func (l *LruList[K, V]) Remove(e *Entry[K, V]) V {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--

	return e.Value
}

// move moves e to next to at.
func (l *LruList[K, V]) move(e, at *Entry[K, V]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *LruList[K, V]) PushFront(k K, v V) *Entry[K, V] {
	l.lazyInit()
	return l.insertValue(k, v, time.Time{}, &l.root)
}

// PushFrontExpirable inserts a new expirable element e with Value v at the front of list l and returns e.
func (l *LruList[K, V]) PushFrontExpirable(k K, v V, expiresAt time.Time) *Entry[K, V] {
	l.lazyInit()
	return l.insertValue(k, v, expiresAt, &l.root)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *LruList[K, V]) MoveToFront(e *Entry[K, V]) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}

func (db *Config) append(e *v1.Model) {
	hit(e)
	if cached, _ := db.cache.Get(e.Id); cached != nil {
		update(cached, e)
		db.ts.Save(e)
		return
	}
	db.ts.Save(e)
	db.cache.Add(e.Id, e)
}
