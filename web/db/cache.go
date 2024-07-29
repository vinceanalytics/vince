package db

import (
	"container/list"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
)

type cache struct {
	size  int
	ll    *list.List
	cache map[uint64]*list.Element
}

type entry struct {
	key   uint64
	value *v1.Model
}

func newCache(maxEntries int) *cache {
	return &cache{
		size:  maxEntries,
		ll:    list.New(),
		cache: make(map[uint64]*list.Element),
	}
}

func (c *cache) Add(key uint64, value *v1.Model) {
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	if c.size != 0 && c.ll.Len() > c.size {
		c.removeOldest()
	}
}

func (c *cache) Get(key uint64) (value *v1.Model, ok bool) {
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

func (c *cache) Remove(key uint64) {
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

func (c *cache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
}

func (db *Config) append(e *v1.Model) {
	hit(e)
	if cached, ok := db.cache.Get(e.Id); ok {
		update(cached, e)
		db.ts.Save(e)
		return
	}
	db.ts.Save(e)
	db.cache.Add(e.Id, e)
}
