package routecache

import (
	"context"
	"errors"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/runner/common"
	"math"
	"sync"
)

type Cacher interface {
	PrimeCache(ctx context.Context)
	CacheGet(ctx context.Context, appname, path string) (*models.Route, bool)
	RefreshCache(ctx context.Context, appname string, route *models.Route)
	ResetCache(ctx context.Context, appname string, delta int)
}

type DefaultCacher struct {
	Datastore models.Datastore
	hotroutes map[string]*Cache
	mu        sync.Mutex // protects hotroutes
}

func NewDefaultCacher(ds models.Datastore) (Cacher, error) {
	if ds == nil {
		return nil, errors.New("Missing datastore")
	}
	return &DefaultCacher{
		Datastore: ds,
		hotroutes: make(map[string]*Cache),
	}, nil
}

func (c *DefaultCacher) PrimeCache(ctx context.Context) {
	log := common.Logger(ctx)
	log.Info("priming cache with known routes")
	apps, err := c.Datastore.GetApps(ctx, nil)
	if err != nil {
		log.WithError(err).Error("cannot prime cache - could not load application list")
		return
	}
	for _, app := range apps {
		routes, err := c.Datastore.GetRoutesByApp(ctx, app.Name, &models.RouteFilter{AppName: app.Name})
		if err != nil {
			log.WithError(err).WithField("appName", app.Name).Error("cannot prime cache - could not load routes")
			continue
		}

		entries := len(routes)
		// The idea here is to prevent both extremes: cache being too small that is ineffective,
		// or too large that it takes too much memory. Up to 1k routes, the cache will try to hold
		// all routes in the memory, thus taking up to 48K per application. After this threshold,
		// it will keep 1024 routes + 20% of the total entries - in a hybrid incarnation of Pareto rule
		// 1024+20% of the remaining routes will likelly be responsible for 80% of the workload.
		if entries > cacheParetoThreshold {
			entries = int(math.Ceil(float64(entries-1024)*0.2)) + 1024
		}
		c.hotroutes[app.Name] = New(entries)

		for i := 0; i < entries; i++ {
			c.RefreshCache(ctx, app.Name, routes[i])
		}
	}
	log.Info("cached prime")
}

// cacheParetoThreshold is both the mark from which the LRU starts caching only
// the most likely hot routes, and also as a stopping mark for the cache priming
// during start.
const cacheParetoThreshold = 1024

func (c *DefaultCacher) CacheGet(ctx context.Context, appname, path string) (*models.Route, bool) {
	c.mu.Lock()
	cache, ok := c.hotroutes[appname]
	if !ok {
		c.mu.Unlock()
		return nil, false
	}
	route, ok := cache.Get(path)
	c.mu.Unlock()
	return route, ok
}

func (c *DefaultCacher) RefreshCache(ctx context.Context, appname string, route *models.Route) {
	c.mu.Lock()
	cache := c.hotroutes[appname]
	cache.Refresh(route)
	c.mu.Unlock()
}

func (c *DefaultCacher) ResetCache(ctx context.Context, appname string, delta int) {
	c.mu.Lock()
	hr, ok := c.hotroutes[appname]
	if !ok {
		c.hotroutes[appname] = New(0)
		hr = c.hotroutes[appname]
	}
	c.hotroutes[appname] = New(hr.MaxEntries + delta)
	c.mu.Unlock()
}
