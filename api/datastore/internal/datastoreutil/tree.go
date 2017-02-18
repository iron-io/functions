package datastoreutil

import (
	"context"
	"regexp"

	"github.com/iron-io/functions/api/models"
	"github.com/pkg/errors"
)

// A LocalDatastore extends LocalRouter with methods from models.Datastore required to complete the implementation.
// They will be called directly.
type LocalDatastore interface {
	LocalRouter

	InsertApp(ctx context.Context, app *models.App) (*models.App, error)
	GetApp(ctx context.Context, appName string) (*models.App, error)
	RemoveApp(ctx context.Context, appName string) error
	UpdateApp(ctx context.Context, app *models.App) (*models.App, error)
	Put(context.Context, []byte, []byte) error
	Get(context.Context, []byte) ([]byte, error)
}

// A LocalRouter is a local, tree based router, exposed through Nodes.
type LocalRouter interface {
	// MatchApps returns Apps for which match returns true.
	MatchApps(ctx context.Context, match func(*models.App) bool) ([]*models.App, error)
	// ViewAllAppNodes executes f for each App Node. f must only read from the Node, not write.
	ViewAllAppNodes(f func(Node) error) error
	// UpdateAppNode executes f for the appName App Node, or returns models.ErrRoutesNotFound if none exists.
	// f may call Node read or write methods.
	UpdateAppNode(appName string, f func(Node) error) error
	// CreateOrUpdateAppNode is identical to UpdateAppNode, except that the App Node will be created if absent,
	// rather than returning models.ErrRoutesNotFound.
	// Note that models.ErrAppsNotFound may still be returned if the App itself is not found.
	CreateOrUpdateAppNode(appName string, f func(Node) error) error
	// ViewAppNode executes f for the appName App Node, or returns models.ErrRoutesNotFound if none exists.
	// f must only read from the Node, not write.
	ViewAppNode(appName string, f func(Node) error) error
}

// NewLocalDatastore adapts ds into a models.Datastore.
func NewLocalDatastore(ds LocalDatastore) models.Datastore {
	return NewValidator(&localDatastore{ds})
}

// A router Node corresponding to a path segment.
// Example /test/path -> [test]->[path]
type Node interface {
	// /node
	Route() (*models.Route, error)
	HasRoute() bool
	SetRoute(*models.Route) error
	// /node/
	TrailingSlashRoute() (*models.Route, error)
	HasTrailingSlashRoute() bool
	SetTrailingSlashRoute(*models.Route) error

	// Child returns a child, or nil if none exists.
	// /node/child
	Child(k string) Node
	// ChildMore returns a child if one exists, and a flag indicating whether more children exist.
	ChildMore(k string) (child Node, more bool)
	// NewChild creates and returns a new child.
	NewChild(k string) (Node, error)
	// HasChildren returns true if children exist.
	HasChildren() bool
	// Remove removes this Node (and it's subtree) from its parent.
	Remove() error

	// Executes the func for this node and all ancestors. Route will never be nil.
	ForAll(func(*models.Route)) error
}

// A localDatastore adapts a LocalDatastore to complete the models.Datastore implementation.
type localDatastore struct {
	LocalDatastore
}

func (ds *localDatastore) GetApps(ctx context.Context, appFilter *models.AppFilter) ([]*models.App, error) {
	matches := func(*models.App) bool { return true }

	if appFilter != nil && appFilter.Name != "" {
		expr := SqlLikeToRegExp(appFilter.Name)
		r, err := regexp.Compile(expr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to compile filter regexp")
		}
		matches = func(a *models.App) bool {
			return r.Match([]byte(a.Name))
		}
	}

	return ds.MatchApps(ctx, matches)
}

func (ds *localDatastore) GetRoute(ctx context.Context, appName, routePath string) (*models.Route, error) {
	var route *models.Route
	err := ds.ViewAppNode(appName, func(n Node) error {
		parts, trailingSlash := SplitPath(routePath)
		n = match(n, parts)
		if n == nil {
			return models.ErrRoutesNotFound
		}

		var err error
		if trailingSlash {
			route, err = n.TrailingSlashRoute()
		} else {
			route, err = n.Route()
		}
		if err != nil {
			return err
		}
		if route == nil {
			return models.ErrRoutesNotFound
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return route, nil
}

func (ds *localDatastore) GetRoutes(ctx context.Context, filter *models.RouteFilter) ([]*models.Route, error) {
	var routes []*models.Route

	match := func(*models.Route) bool { return true }

	if filter != nil {
		match = func(r *models.Route) bool {
			return (filter.Path == "" || r.Path == filter.Path) &&
				(filter.AppName == "" || r.AppName == filter.AppName) &&
				(filter.Image == "" || r.Image == filter.Image)
		}
	}

	err := ds.ViewAllAppNodes(func(n Node) error {
		n.ForAll(func(r *models.Route) {
			if match(r) {
				routes = append(routes, r)
			}
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return routes, nil
}

func (ds *localDatastore) GetRoutesByApp(ctx context.Context, appName string, routeFilter *models.RouteFilter) ([]*models.Route, error) {
	var routes []*models.Route

	if routeFilter != nil && routeFilter.Path != "" {
		r, err := ds.GetRoute(ctx, appName, routeFilter.Path)
		if err != nil {
			return nil, err
		}
		if routeFilter.Image == "" || r.Image == routeFilter.Image {
			routes = append(routes, r)
		}

		return routes, nil
	}

	err := ds.ViewAppNode(appName, func(n Node) error {
		n.ForAll(func(r *models.Route) {
			if routeFilter == nil || (routeFilter.Image == "" || r.Image == routeFilter.Image) {
				routes = append(routes, r)
			}
		})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return routes, nil
}

func (ds *localDatastore) InsertRoute(ctx context.Context, route *models.Route) (*models.Route, error) {
	err := ds.CreateOrUpdateAppNode(route.AppName, func(n Node) error {
		parts, trailingSlash := SplitPath(StripParamNames(route.Path))

		for _, p := range parts {
			child := n.Child(p)
			if child == nil {
				if n.HasChildren() {
					// Check for conflicts
					if p == ":" || p == "*" {
						return models.ErrRoutesCreate
					}
					if n.Child(":") != nil || n.Child("*") != nil {
						return models.ErrRoutesCreate
					}
				}
				var err error
				child, err = n.NewChild(p)
				if err != nil {
					return err
				}
			}
			n = child
		}

		if trailingSlash {
			if n.HasTrailingSlashRoute() {
				return models.ErrRoutesAlreadyExists
			}
			return n.SetTrailingSlashRoute(route)
		} else {
			if n.HasRoute() {
				return models.ErrRoutesAlreadyExists
			}
			return n.SetRoute(route)
		}
	})
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (ds *localDatastore) UpdateRoute(ctx context.Context, newroute *models.Route) (*models.Route, error) {
	err := ds.UpdateAppNode(newroute.AppName, func(n Node) error {
		parts, trailingSlash := SplitPath(StripParamNames(newroute.Path))
		for _, r := range parts {
			child := n.Child(r)
			if child == nil {
				return models.ErrRoutesNotFound
			}
			n = child
		}

		var getRoute func() (*models.Route, error)
		var setRoute func(*models.Route) error
		if trailingSlash {
			getRoute = n.TrailingSlashRoute
			setRoute = n.SetTrailingSlashRoute
		} else {
			getRoute = n.Route
			setRoute = n.SetRoute
		}

		route, err := getRoute()
		if err != nil {
			return err
		}
		if route == nil {
			return models.ErrRoutesNotFound
		}
		route.Update(newroute)
		if err := route.Validate(); err != nil {
			return err
		}
		newroute = route.Clone()
		return setRoute(route)
	})
	if err != nil {
		return nil, err
	}
	return newroute, nil
}

func (ds *localDatastore) RemoveRoute(ctx context.Context, appName, routePath string) error {
	return ds.UpdateAppNode(appName, func(n Node) error {
		var prune Node
		parts, trailingSlash := SplitPath(StripParamNames(routePath))
		for _, r := range parts {
			child, more := n.ChildMore(r)
			if child == nil {
				return models.ErrRoutesNotFound
			}
			if more || n.HasRoute() || n.HasTrailingSlashRoute() {
				// n in use, so don't prune
				prune = nil
			} else if prune == nil {
				// n no longer in use, prune from here
				prune = n
			}
			n = child
		}

		if trailingSlash {
			if !n.HasTrailingSlashRoute() {
				return models.ErrRoutesNotFound
			}
			if err := n.SetTrailingSlashRoute(nil); err != nil {
				return err
			}
		} else {
			if !n.HasRoute() {
				return models.ErrRoutesNotFound
			}
			if err := n.SetRoute(nil); err != nil {
				return err
			}
		}
		if !n.HasRoute() && !n.HasTrailingSlashRoute() && !n.HasChildren() {
			// n no longer in use, so prune back as far as possible
			if prune == nil {
				prune = n
			}
			prune.Remove()
		}
		return nil
	})
}
// match follows path through the tree rooted at n, and returns a matching Node, or nil if none is found.
// An empty path returns n.
func match(n Node, path []string) Node {
	for _, p := range path {
		if child, m := childMatch(n, p); child != nil {
			n = child
			if m == "*" {
				break
			}
			continue
		}
		return nil
	}
	return n
}

// childMatch returns a child node matching m and it's key (m or a wildcard), or nil for no match.
func childMatch(n Node, m string) (Node, string) {
	for _, m := range []string{m, ":", "*"} {
		c := n.Child(m)
		if c != nil {
			return c, m
		}
	}
	return nil, ""
}