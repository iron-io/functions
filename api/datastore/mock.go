package datastore

import (
	"context"
	"regexp"

	"github.com/iron-io/functions/api/datastore/internal/datastoreutil"
	"github.com/iron-io/functions/api/models"

	"github.com/pkg/errors"
)

type Mock struct {
	apps            map[string]*models.App
	routesByAppName map[string]*node
	metaData        map[string][]byte
}

// NewMock returns an in-memory, mock models.Datastore implementation, which is NOT safe for concurrent use.
func NewMock() *Mock {
	return &Mock{make(map[string]*models.App), make(map[string]*node), make(map[string][]byte)}
}

func (m *Mock) GetApp(ctx context.Context, appName string) (app *models.App, err error) {
	if appName == "" {
		return nil, models.ErrDatastoreEmptyAppName
	}
	for _, a := range m.apps {
		if a.Name == appName {
			return a, nil
		}
	}

	return nil, models.ErrAppsNotFound
}

func (m *Mock) GetApps(ctx context.Context, appFilter *models.AppFilter) ([]*models.App, error) {
	matches := func(*models.App) bool { return true }

	if appFilter == nil || appFilter.Name != "" {
		expr := datastoreutil.SqlLikeToRegExp(appFilter.Name)
		r, err := regexp.Compile(expr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to compile filter regexp")
		}
		matches = func(a *models.App) bool {
			return r.Match([]byte(a.Name))
		}
	}

	var apps []*models.App
	for _, a := range m.apps {
		if matches(a) {
			apps = append(apps, a)
		}
	}
	return apps, nil
}

func (m *Mock) InsertApp(ctx context.Context, app *models.App) (*models.App, error) {
	if app == nil {
		return nil, models.ErrDatastoreEmptyApp
	}
	if app.Name == "" {
		return nil, models.ErrDatastoreEmptyAppName
	}
	if _, ok := m.apps[app.Name]; ok {
		return nil, models.ErrAppsAlreadyExists
	}
	m.apps[app.Name] = app
	return app, nil
}

func (m *Mock) UpdateApp(ctx context.Context, app *models.App) (*models.App, error) {
	if app == nil {
		return nil, models.ErrDatastoreEmptyApp
	}
	a, err := m.GetApp(ctx, app.Name)
	if err != nil {
		return nil, err
	}
	if app.Config != nil {
		if a.Config == nil {
			a.Config = map[string]string{}
		}
		for k, v := range app.Config {
			a.Config[k] = v
		}
	}
	return a, nil
}

func (m *Mock) RemoveApp(ctx context.Context, appName string) error {
	if appName == "" {
		return models.ErrDatastoreEmptyAppName
	}
	if _, ok := m.apps[appName]; !ok {
		return models.ErrAppsNotFound
	}
	delete(m.apps, appName)
	return nil
}

func (m *Mock) GetRoute(ctx context.Context, appName, routePath string) (*models.Route, error) {
	if appName == "" {
		return nil, models.ErrDatastoreEmptyAppName
	}
	if routePath == "" {
		return nil, models.ErrDatastoreEmptyRoutePath
	}

	if _, ok := m.apps[appName]; !ok {
		return nil, models.ErrAppsNotFound
	}

	n := m.routesByAppName[appName]
	if n == nil {
		return nil, models.ErrRoutesNotFound
	}

	var r *models.Route

	if parts, trailingSlash := datastoreutil.SplitPath(routePath); len(parts)==0 {
		if trailingSlash {
			r = n.trailingSlashRoute
		} else {
			r = n.route
		}
	} else {
		a := n.match(parts)
		if trailingSlash {
			r = a.trailingSlashRoute
		} else {
			r = a.route
		}
	}

	if r == nil {
		return nil, models.ErrRoutesNotFound
	}

	return r, nil
}

func (m *Mock) GetRoutes(ctx context.Context, routeFilter *models.RouteFilter) (routes []*models.Route, err error) {
	for _, n := range m.routesByAppName {
		n.forAll(func(r *models.Route) {
			if (routeFilter.Path == "" || r.Path == routeFilter.Path) && (routeFilter.AppName == "" || r.AppName == routeFilter.AppName) {
				routes = append(routes, r)
			}
		})
	}
	return
}

func (m *Mock) GetRoutesByApp(ctx context.Context, appName string, routeFilter *models.RouteFilter) (routes []*models.Route, err error) {
	if appName == "" {
		return nil, models.ErrDatastoreEmptyAppName
	}
	n := m.routesByAppName[appName]
	n.forAll(func(r *models.Route) {
		if r.AppName == appName && (routeFilter.Path == "" || r.Path == routeFilter.Path) && (routeFilter.AppName == "" || r.AppName == routeFilter.AppName) {
			routes = append(routes, r)
		}
	})
	return
}

func (m *Mock) InsertRoute(ctx context.Context, route *models.Route) (*models.Route, error) {
	if route == nil {
		return nil, models.ErrDatastoreEmptyRoute
	}
	if route.AppName == "" {
		return nil, models.ErrDatastoreEmptyAppName
	}
	if route.Path == "" {
		return nil, models.ErrRoutesMissingNew
	}

	namelessParts, trailingSlash := datastoreutil.SplitPath(datastoreutil.StripParamNames(route.Path))

	node, err := m.getOrCreateNode(ctx, route.AppName, namelessParts)
	if err != nil {
		return nil, err
	}

	if trailingSlash {
		if node.trailingSlashRoute != nil {
			return nil, models.ErrRoutesAlreadyExists
		}
		node.trailingSlashRoute = route.Clone()
	} else {
		if node.route != nil {
			return nil, models.ErrRoutesAlreadyExists
		}
		node.route = route.Clone()
	}

	return route, nil
}

func (m *Mock) UpdateRoute(ctx context.Context, newroute *models.Route) (*models.Route, error) {
	route, err := m.getRoute(ctx, newroute.AppName, newroute.Path)
	if err != nil {
		return nil, err
	}

	clone := route.Clone()
	clone.Update(newroute)

	if err := clone.Validate(); err != nil {
		return nil, err
	}

	route.Update(clone)

	return clone, nil
}

func (m *Mock) RemoveRoute(ctx context.Context, appName, routePath string) error {
	if appName == "" {
		return models.ErrDatastoreEmptyAppName
	}
	if routePath == "" {
		return models.ErrDatastoreEmptyRoutePath
	}

	if _, ok := m.apps[appName]; !ok {
		return models.ErrAppsNotFound
	}

	n := m.routesByAppName[appName]
	if n == nil {
		return models.ErrRoutesNotFound
	}

	prune := func() {
		delete(m.routesByAppName, appName)
	}
	pathParts, trailingSlash := datastoreutil.SplitPath(datastoreutil.StripParamNames(routePath))
	for _, r := range pathParts {
		child := n.child(r)
		if child == nil {
			return models.ErrRoutesNotFound
		}
		if n.route != nil || n.trailingSlashRoute != nil || len(n.children) > 1 {
			// advance prune past n
			prune = func() {
				delete(n.children, r)
			}
		}
		n = child
	}

	if trailingSlash {
		if n.trailingSlashRoute == nil {
			return models.ErrRoutesNotFound
		}
		n.trailingSlashRoute = nil
	} else {
		if n.route == nil {
			return models.ErrRoutesNotFound
		}
		n.route = nil
	}
	if n.route == nil && n.trailingSlashRoute == nil && len(n.children) == 0 {
		// node no longer in use, so prune back as far as possible
		prune()
	}

	return nil
}

func (m *Mock) Put(ctx context.Context, key, value []byte) error {
	if key == nil {
		return models.ErrDatastoreEmptyKey
	}
	m.metaData[string(key)] = value
	return nil
}

func (m *Mock) Get(ctx context.Context, key []byte) ([]byte, error) {
	if key == nil {
		return nil, models.ErrDatastoreEmptyKey
	}
	return m.metaData[string(key)], nil
}

func (m *Mock) getRoute(ctx context.Context, appName, routePath string) (*models.Route, error) {
	if _, ok := m.apps[appName]; !ok {
		return nil, models.ErrAppsNotFound
	}

	nameless := datastoreutil.StripParamNames(routePath)

	node := m.routesByAppName[appName]
	if node == nil {
		return nil, models.ErrRoutesNotFound
	}

	pathParts, trailingSlash := datastoreutil.SplitPath(nameless)

	for _, r := range pathParts {
		child := node.child(r)
		if child == nil {
			return nil, models.ErrRoutesNotFound
		}
		node = child
	}

	var route *models.Route
	if trailingSlash {
		route = node.trailingSlashRoute
	} else {
		route = node.route
	}

	if route == nil {
		return nil, models.ErrRoutesNotFound
	}

	return route, nil
}

func (m *Mock) getOrCreateNode(ctx context.Context, appName string, pathParts []string) (*node, error) {
	if _, ok := m.apps[appName]; !ok {
		return nil, models.ErrAppsNotFound
	}

	n := m.routesByAppName[appName]
	if n == nil {
		n = &node{}
		m.routesByAppName[appName] = n
	}

	for _, r := range pathParts {
		child := n.child(r)
		if child == nil {
			if len(n.children) > 0 {
				// Check for conflicts
				if r == ":" || r == "*" {
					return nil, models.ErrRoutesCreate
				}
				if n.child(":") != nil || n.child("*") != nil {
					return nil, models.ErrRoutesCreate
				}
			}

			child = &node{}
			n.addChild(r, child)
		}
		n = child
	}

	return n, nil
}

type node struct {
	route, trailingSlashRoute *models.Route

	children map[string]*node
}

func (n *node) child(k string) *node {
	if n.children == nil {
		return nil
	}
	return n.children[k]
}

func (n *node) addChild(k string, c *node) {
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	n.children[k] = c
}

func (n *node) forAll(f func(r *models.Route)) {
	f(n.route)
	f(n.trailingSlashRoute)
	for _, child := range n.children {
		child.forAll(f)
	}
}

// match returns an ancestor matching path
// Required: len(path) > 0
func (n *node) match(path []string) *node {
	for _, m := range []string{path[0], ":", "*"} {
		if child := n.child(m); child != nil {
			if len(path) == 1 || m == "*" {
				return child
			}
			return child.match(path[1:])
		}
	}
	return nil
}
