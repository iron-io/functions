package datastore

import (
	"context"

	"github.com/iron-io/functions/api/datastore/internal/datastoreutil"
	"github.com/iron-io/functions/api/models"
)

// A datastore is an in-memory, map backed datastoreutil.Datastore
type datastore struct {
	apps            map[string]*models.App
	routesByAppName map[string]datastoreutil.Node
	metaData        map[string][]byte
}
//TODO we can't roll back, so double check that we definly have nil errors where we need them

// NewMock returns an in-memory, mock models.Datastore implementation, which is NOT safe for concurrent use.
//TODO make concurrent safe? (if RW mode leaks into interface defs anyways...)
func NewMock() models.Datastore {
	return datastoreutil.NewLocalDatastore(
		&datastore{make(map[string]*models.App), make(map[string]datastoreutil.Node), make(map[string][]byte)})
}

func (ds *datastore) GetApp(ctx context.Context, appName string) (app *models.App, err error) {
	//TODO read lock
	a, ok := ds.apps[appName]
	if ok {
		return a, nil
	}

	return nil, models.ErrAppsNotFound
}

func (ds *datastore) MatchApps(ctx context.Context, match func(*models.App) bool) ([]*models.App, error) {
	//TODO read lock
	var apps []*models.App
	for _, a := range ds.apps {
		if match(a) {
			apps = append(apps, a)
		}
	}
	return apps, nil
}

func (ds *datastore) InsertApp(ctx context.Context, app *models.App) (*models.App, error) {
	if _, ok := ds.apps[app.Name]; ok {
		return nil, models.ErrAppsAlreadyExists
	}
	ds.apps[app.Name] = app
	return app, nil
}
//TODO UpdateAppFunc instead
func (ds *datastore) UpdateApp(ctx context.Context, app *models.App) (*models.App, error) {
	//TODO lock
	a, err := ds.GetApp(ctx, app.Name)
	if err != nil {
		return nil, err
	}
	a.UpdateConfig(app)
	return a.Clone(), nil
}

//TODO delete all routes? #528
func (ds *datastore) RemoveApp(ctx context.Context, appName string) error {
	//TODO lock
	if _, ok := ds.apps[appName]; !ok {
		return models.ErrAppsNotFound
	}
	delete(ds.apps, appName)
	return nil
}

func (ds *datastore) UpdateAppNode(appName string, f func(datastoreutil.Node) error) error {
	if _, ok := ds.apps[appName]; !ok {
		return models.ErrAppsNotFound
	}

	n := ds.routesByAppName[appName]
	if n == nil {
		return models.ErrRoutesNotFound
	}

	return f(n)
}

func (ds *datastore) CreateOrUpdateAppNode(appName string, f func(datastoreutil.Node) error) error {
	//TODO write lock

	if _, ok := ds.apps[appName]; !ok {
		return models.ErrAppsNotFound
	}

	n := ds.routesByAppName[appName]
	if n == nil {
		n = &node{remove: func() error {
			delete(ds.routesByAppName, appName)
			return nil
		}}
		ds.routesByAppName[appName] = n
	}

	return f(n)
}

func (ds *datastore) ViewAppNode(appName string, f func(datastoreutil.Node) error) error {
	//TODO read lock

	if _, ok := ds.apps[appName]; !ok {
		return models.ErrAppsNotFound
	}

	n := ds.routesByAppName[appName]
	if n == nil {
		return models.ErrRoutesNotFound
	}

	return f(n)
}

func (ds *datastore) ViewAllAppNodes(f func(datastoreutil.Node) error) error {
	for _, n := range ds.routesByAppName {
		if err := f(n); err != nil {
			return err
		}
	}
	return nil
}

func (ds *datastore) Put(ctx context.Context, key, value []byte) error {
	ds.metaData[string(key)] = value
	return nil
}

func (ds *datastore) Get(ctx context.Context, key []byte) ([]byte, error) {
	return ds.metaData[string(key)], nil
}

//TODO doc
type node struct {
	route, trailingSlashRoute *models.Route

	children map[string]datastoreutil.Node

	remove func() error
}

func (n *node) HasRoute() bool {
	return n.route != nil
}

func (n *node) Route() (*models.Route, error) {
	return n.route, nil
}

func (n *node) SetRoute(r *models.Route) error {
	n.route = r
	return nil
}

func (n *node) HasTrailingSlashRoute() bool {
	return n.trailingSlashRoute != nil
}

func (n *node) TrailingSlashRoute() (*models.Route, error) {
	return n.trailingSlashRoute, nil
}

func (n *node) SetTrailingSlashRoute(r *models.Route) error {
	n.trailingSlashRoute = r
	return nil
}

func (n *node) Child(k string) datastoreutil.Node {
	if n.children == nil {
		return nil
	}
	return n.children[k]
}

func (n *node) ChildMore(k string) (datastoreutil.Node, bool) {
	if n.children == nil {
		return nil, false
	}
	child, ok := n.children[k]
	if ok {
		return child, len(n.children) > 1
	} else {
		return nil, len(n.children) > 0
	}
}

func (n *node) NewChild(k string) (datastoreutil.Node, error) {
	if n.children == nil {
		n.children = make(map[string]datastoreutil.Node)
	}
	c := &node{remove: func() error {
		delete(n.children, k)
		return nil
	}}
	n.children[k] = c
	return c, nil
}

func (n *node) Remove() error {
	return n.remove()
}

func (n *node) HasChildren() bool {
	return len(n.children) > 0
}

func (n *node) ForAll(f func(r *models.Route)) error {
	if n.route != nil {
		f(n.route)
	}
	if n.trailingSlashRoute != nil {
		f(n.trailingSlashRoute)
	}
	for _, child := range n.children {
		child.ForAll(f)
	}
	return nil
}