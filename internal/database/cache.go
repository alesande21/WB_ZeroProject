package database

import (
	entity2 "WB_ZeroProject/internal/entity"
	_ "github.com/patrickmn/go-cache"
	"sync"
)

type allCache struct {
	*cache
}

type cache struct {
	orders map[entity2.OrderId]entity2.Order
	sync.RWMutex
}

func newCache() *allCache {
	orders := make(map[entity2.OrderId]entity2.Order)
	c := cache{
		orders:  orders,
		RWMutex: sync.RWMutex{},
	}
	return &allCache{&c}
}

func (c *allCache) Get(k string) (*entity2.Order, bool) {

	return nil, false
}

/*
func (c *cache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	// "Inlining" of get and Expired
	item, found := c.items[k]
	if !found {
		c.mu.RUnlock()
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			c.mu.RUnlock()
			return nil, false
		}
	}
	c.mu.RUnlock()
	return item.Object, true
}
*/

/*
func (c *cache) Set(k string, x interface{}, d time.Duration) {
	// "Inlining" of set
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
	// TODO: Calls to mu.Unlock are currently not deferred because defer
	// adds ~200 ns (as of go1.)
	c.mu.Unlock()
}
*/

func (c *allCache) Set(k string, value entity2.Order) bool {

	return false
}
