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

import "container/list"

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache[Key comparable, Value any] struct {
	ll         *list.List
	cache      map[Key]*list.Element
	MaxEntries int
}

type entry[Key, Value any] struct {
	key   Key
	value Value
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New[Key comparable, Value any](maxEntries int) *Cache[Key, Value] {
	return &Cache[Key, Value]{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[Key]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *Cache[Key, Value]) Add(key Key, value Value) {
	if c.cache == nil {
		c.cache = make(map[Key]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry[Key, Value]).value = value
		return
	}
	ele := c.ll.PushFront(&entry[Key, Value]{key, value})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[Key, Value]) Get(key Key) (value Value, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry[Key, Value]).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *Cache[Key, Value]) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache[Key, Value]) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache[Key, Value]) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry[Key, Value])
	delete(c.cache, kv.key)
}

// Len returns the number of items in the cache.
func (c *Cache[Key, Value]) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *Cache[Key, Value]) Clear() {
	c.ll = nil
	c.cache = nil
}
