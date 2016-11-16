// Package routecache is meant to assist in resolving the most used routes at
// an application. Implemented as a LRU, it returns always its full context for
// iteration at the router handler.
package routecache

// based on groupcache's LRU

import (
	"container/list"

	"github.com/iron-io/functions/api/models"
)

const defaultMaxEntries = 100

// Cache holds an internal linkedlist for hotness management. It is not safe
// for concurrent use, must be guarded externally. Its zero value means nothing.
type Cache struct {
	maxentries int

	ll     *list.List
	cache  map[*models.Route]*list.Element
	values []*models.Route
}

type routecacheentry struct {
	route *models.Route
}

// New returns a route cache.
func New() *Cache {
	return &Cache{
		maxentries: defaultMaxEntries,
		ll:         list.New(),
		cache:      make(map[*models.Route]*list.Element),
	}
}

// Routes is an ordered slice with the hottest routes at the beginning, so to
// increase the likelihood of matching with incoming requests.
func (c *Cache) Routes() []*models.Route {
	return c.values
}

// Refresh updates internal linkedlist either adding a new route to the front,
// or moving it to the front when used. It will discard seldom used routes.
func (c *Cache) Refresh(route *models.Route) {
	if ee, ok := c.cache[route]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*routecacheentry).route = route
		c.updatevalues()
		return
	}
	ele := c.ll.PushFront(&routecacheentry{route})
	c.cache[route] = ele
	if c.maxentries != 0 && c.ll.Len() > c.maxentries {
		c.removeOldest()
		return
	}

	c.updatevalues()
}

func (c *Cache) updatevalues() {
	c.values = make([]*models.Route, 0, c.ll.Len())
	for e := c.ll.Front(); e != nil; e = e.Next() {
		route := e.Value.(*routecacheentry).route
		c.values = append(c.values, route)
	}
}

func (c *Cache) remove(key *models.Route) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

func (c *Cache) removeOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*routecacheentry)
	delete(c.cache, kv.route)
	c.updatevalues()
}
