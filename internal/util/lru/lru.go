/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package lru implements an LRU cache.
package lru

import (
	"container/list"
	"time"

	"github.com/vinceanalytics/vince/internal/models"
)

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	capacity int
	ttl      int64
	ll       *list.List
	cache    map[uint64]*list.Element
}

type entry struct {
	key   uint64
	value *models.Cached
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New(ttl time.Duration) *Cache {
	return &Cache{
		capacity: 1 << 20,
		ttl:      ttl.Milliseconds(),
		ll:       list.New(),
		cache:    make(map[uint64]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *Cache) Set(key uint64, value *models.Cached) {
	ee, ok := c.cache[key]
	if ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	if c.ll.Len() > c.capacity {
		now := time.Now().UTC().UnixMilli()
		for ele := c.ll.Back(); ele != nil; ele = c.ll.Back() {
			kv := ele.Value.(*entry)
			if now-kv.value.Start < c.ttl {
				break
			}
			c.ll.Remove(ele)
			delete(c.cache, kv.key)
		}

	}
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key uint64) (value *models.Cached, ok bool) {
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}
