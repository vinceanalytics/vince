package lru

import (
	"container/list"
)

type LRU[T any] struct {
	size  int
	ll    *list.List
	cache map[uint64]*list.Element
}

type entry[T any] struct {
	key   uint64
	value T
}

func New[T any](maxEntries int) *LRU[T] {
	return &LRU[T]{
		size:  maxEntries,
		ll:    list.New(),
		cache: make(map[uint64]*list.Element),
	}
}

func (c *LRU[T]) Add(key uint64, value T) {
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry[T]).value = value
		return
	}
	ele := c.ll.PushFront(&entry[T]{key, value})
	c.cache[key] = ele
	if c.size != 0 && c.ll.Len() > c.size {
		c.removeOldest()
	}
}

func (c *LRU[T]) Get(key uint64) (value T, ok bool) {
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry[T]).value, true
	}
	return
}

func (c *LRU[T]) Remove(key uint64) {
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

func (c *LRU[T]) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRU[T]) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry[T])
	delete(c.cache, kv.key)
}
