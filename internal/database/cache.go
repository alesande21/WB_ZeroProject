package database

import (
	entity2 "WB_ZeroProject/internal/entity"
	_ "github.com/patrickmn/go-cache"
	"sync"
)

type AllCache struct {
	*cache
}

type cache struct {
	orders map[entity2.OrderId]entity2.Order
	sync.RWMutex
}

func NewCache() *AllCache {
	orders := make(map[entity2.OrderId]entity2.Order)
	c := cache{
		orders:  orders,
		RWMutex: sync.RWMutex{},
	}
	return &AllCache{&c}
}

func (c *AllCache) Get(k string) (*entity2.Order, bool) {
	c.RLock()
	defer c.RUnlock()
	order, found := c.orders[k]
	if !found {
		return nil, false
	}
	return &order, true
}

func (c *AllCache) Set(k string, value entity2.Order) {
	c.Lock()
	c.orders[k] = value
	c.Unlock()
}

func (c *AllCache) ItemCount() int {
	c.RLock()
	n := len(c.orders)
	c.RUnlock()
	return n
}

func (c *AllCache) Delete(k string) bool {
	c.Lock()
	defer c.Unlock()
	if _, found := c.orders[k]; found {
		delete(c.orders, k)
		return true
	}
	return false
}
