package bolt

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/iron-io/functions/api/datastore/internal/datastoreutil"
	"github.com/iron-io/functions/api/models"

	"github.com/boltdb/bolt"
	"github.com/Sirupsen/logrus"
)

// A bolt backed datastoreutil.Datastore
type datastore struct {
	routesKey, appsKey, logsKey, extrasKey []byte
	db           *bolt.DB
	log          logrus.FieldLogger
}

// New returns a new bolt backed Datastore.
func New(url *url.URL) (models.Datastore, error) {
	dir := filepath.Dir(url.Path)
	log := logrus.WithFields(logrus.Fields{"db": url.Scheme, "dir": dir})
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.WithError(err).Errorln("Could not create data directory for db")
		return nil, err
	}
	log.WithFields(logrus.Fields{"path": url.Path}).Debug("Creating bolt db")
	db, err := bolt.Open(url.Path, 0655, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.WithError(err).Errorln("Error on bolt.Open")
		return nil, err
	}
	// I don't think we need a prefix here do we? Made it blank. If we do, we should call the query param "prefix" instead of bucket.
	bucketPrefix := ""
	if url.Query()["bucket"] != nil {
		bucketPrefix = url.Query()["bucket"][0]
	}
	routesBucketName := []byte(bucketPrefix + "routes")
	appsBucketName := []byte(bucketPrefix + "apps")
	logsBucketName := []byte(bucketPrefix + "logs")
	extrasBucketName := []byte(bucketPrefix + "extras") // todo: think of a better name
	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{routesBucketName, appsBucketName, logsBucketName, extrasBucketName} {
			_, err := tx.CreateBucketIfNotExists(name)
			if err != nil {
				log.WithError(err).WithFields(logrus.Fields{"name": name}).Error("create bucket")
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Errorln("Error creating bolt buckets")
		return nil, err
	}

	ds := &datastore{
		routesKey: routesBucketName,
		appsKey:   appsBucketName,
		logsKey:   logsBucketName,
		extrasKey: extrasBucketName,
		db:           db,
		log:          log,
	}
	log.WithFields(logrus.Fields{"prefix": bucketPrefix, "file": url.Path}).Debug("BoltDB initialized")

	return datastoreutil.NewLocalDatastore(ds), nil
}

func (ds *datastore) InsertApp(ctx context.Context, app *models.App) (*models.App, error) {
	appname := []byte(app.Name)

	err := ds.db.Update(func(tx *bolt.Tx) error {
		bIm := tx.Bucket(ds.appsKey)

		v := bIm.Get(appname)
		if v != nil {
			return models.ErrAppsAlreadyExists
		}

		buf, err := json.Marshal(app)
		if err != nil {
			return err
		}

		err = bIm.Put(appname, buf)
		if err != nil {
			return err
		}
		bjParent := tx.Bucket(ds.routesKey)
		_, err = bjParent.CreateBucketIfNotExists([]byte(app.Name))
		if err != nil {
			return err
		}
		return nil
	})

	return app, err
}

func (ds *datastore) UpdateApp(ctx context.Context, newapp *models.App) (*models.App, error) {
	var app *models.App
	appname := []byte(newapp.Name)

	err := ds.db.Update(func(tx *bolt.Tx) error {
		bIm := tx.Bucket(ds.appsKey)

		v := bIm.Get(appname)
		if v == nil {
			return models.ErrAppsNotFound
		}

		err := json.Unmarshal(v, &app)
		if err != nil {
			return err
		}

		//TODO changed from set to add, correct?
		app.UpdateConfig(newapp)

		buf, err := json.Marshal(app)
		if err != nil {
			return err
		}

		err = bIm.Put(appname, buf)
		if err != nil {
			return err
		}
		bjParent := tx.Bucket(ds.routesKey)
		_, err = bjParent.CreateBucketIfNotExists([]byte(app.Name))
		if err != nil {
			return err
		}
		return nil
	})

	return app, err
}
//TODO test that routes are deleted as well
func (ds *datastore) RemoveApp(ctx context.Context, appName string) error {
	err := ds.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(ds.appsKey).Delete([]byte(appName))
		if err != nil {
			return err
		}
		err = tx.Bucket(ds.routesKey).DeleteBucket([]byte(appName))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (ds *datastore) MatchApps(ctx context.Context, matches func(*models.App) bool) ([]*models.App, error) {
	res := []*models.App{}
	err := ds.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ds.appsKey)
		err2 := b.ForEach(func(key, v []byte) error {
			app := &models.App{}
			err := json.Unmarshal(v, app)
			if err != nil {
				return err
			}
			if matches(app) {
				res = append(res, app)
			}
			return nil
		})
		if err2 != nil {
			logrus.WithError(err2).Errorln("Couldn't get apps!")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ds *datastore) GetApp(ctx context.Context, name string) (*models.App, error) {
	var res *models.App
	err := ds.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ds.appsKey)
		v := b.Get([]byte(name))
		if v != nil {
			app := &models.App{}
			err := json.Unmarshal(v, app)
			if err != nil {
				return err
			}
			res = app
		} else {
			return models.ErrAppsNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ds *datastore) ViewAppNode(appName string, f func(datastoreutil.Node) error) error {
	return ds.db.View(func(tx *bolt.Tx) error {
		n := getNode(tx.Bucket(ds.routesKey), []byte(appName))
		if n == nil {
			return models.ErrRoutesNotFound
		}
		return f(n)
	})
}

func (ds *datastore) ViewAllAppNodes(f func(datastoreutil.Node) error) error {
	return ds.db.View(func(tx *bolt.Tx) error {
		rb := tx.Bucket(ds.routesKey)

		c := rb.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			n := getNode(rb, k)
			if n == nil {
				return models.ErrRoutesNotFound
			}
			if err := f(n); err != nil {
				return err
			}
		}
		return nil
	})
}
//TODO none of these do the app check first...
func (ds *datastore) UpdateAppNode(appName string, f func(datastoreutil.Node) error) error {
	return ds.db.Update(func(tx *bolt.Tx) error {
		n := getNode(tx.Bucket(ds.routesKey), []byte(appName))
		if n == nil {
			return models.ErrRoutesNotFound
		}

		return f(n)
	})
}

func (ds *datastore) CreateOrUpdateAppNode(appName string, f func(datastoreutil.Node) error) error {
	return ds.db.Update(func(tx *bolt.Tx) error {
		n, err := getOrCreateNode(tx.Bucket(ds.routesKey), []byte(appName))
		if err != nil {
			return err
		}

		return f(n)
	})
}

func (ds *datastore) Put(ctx context.Context, key, value []byte) error {
	ds.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ds.extrasKey) // todo: maybe namespace by app?
		err := b.Put(key, value)
		return err
	})
	return nil
}

func (ds *datastore) Get(ctx context.Context, key []byte) ([]byte, error) {
	var ret []byte
	ds.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ds.extrasKey)
		ret = b.Get(key)
		return nil
	})
	return ret, nil
}

var (
	routeKey = []byte("route")
	trailingSlashRouteKey = []byte("route/")
	childrenKey = []byte("children")
)

// A bolt.Bucket backed datastoreutil.Node.
type node struct {
	b *bolt.Bucket

	remove func() error
}

// getNode gets the node key from parent, or returns nil if none exists.
func getNode(parent *bolt.Bucket, key []byte) datastoreutil.Node {
	b := parent.Bucket(key)
	if b == nil {
		return nil
	}

	return &node{b: b, remove: func() error {
		return parent.DeleteBucket(key)
	}}
}

// getOrCreateNode gets the node key from parent, creating the backing bucket if necessary.
func getOrCreateNode(parent *bolt.Bucket, key []byte) (datastoreutil.Node, error) {
	b, err := parent.CreateBucketIfNotExists(key)
	if err != nil {
		return nil, err
	}

	return &node{b: b, remove: func() error {
		return parent.DeleteBucket(key)
	}}, nil
}

// createNode gets the node key from parent, after creating the backing bucket.
func createNode(parent *bolt.Bucket, key []byte) (datastoreutil.Node, error) {
	b, err := parent.CreateBucket(key)
	if err != nil {
		return nil, err
	}

	return &node{b: b, remove: func() error {
		return parent.DeleteBucket(key)
	}}, nil
}

func (n *node) getRoute(key []byte) (*models.Route, error) {
	b := n.b.Get(key)
	if b == nil {
		return nil, nil
	}
	var r models.Route
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (n *node) setRoute(key []byte, r *models.Route) error {
	if r == nil {
		return n.b.Delete(key)
	}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if err := n.b.Put(key, b); err != nil {
		return err
	}
	return nil
}

func (n *node) HasRoute() bool {
	return n.b.Get(routeKey) != nil
}

func (n *node) Route() (*models.Route, error) {
	return n.getRoute(routeKey)
}

func (n *node) SetRoute(r *models.Route) error {
	return n.setRoute(routeKey, r)
}

func (n *node) HasTrailingSlashRoute() bool {
	return n.b.Get(trailingSlashRouteKey) != nil
}

func (n *node) TrailingSlashRoute() (*models.Route, error) {
	return n.getRoute(trailingSlashRouteKey)
}

func (n *node) SetTrailingSlashRoute(r *models.Route) error {
	return n.setRoute(trailingSlashRouteKey, r)
}

func (n *node) Child(k string) datastoreutil.Node {
	children := n.b.Bucket(childrenKey)
	if children == nil {
		return nil
	}
	return getNode(children, []byte(k))
}

func (n *node) ChildMore(k string) (datastoreutil.Node, bool) {
	children := n.b.Bucket(childrenKey)
	if children == nil {
		return nil, false
	}
	key := []byte(k)

	c := children.Cursor()
	ck, _ := c.First()
	if ck == nil {
		return nil, false
	} else if !bytes.Equal(ck, key) {
		return getNode(children, key), true
	}

	child := getNode(children, key)
	ck, _ = c.Next()
	return child, len(ck) > 0
}

func (n *node) NewChild(k string) (datastoreutil.Node, error) {
	children, err := n.b.CreateBucketIfNotExists(childrenKey)
	if err != nil {
		return nil, err
	}
	return createNode(children, []byte(k))
}

func (n *node) HasChildren() bool {
	children := n.b.Bucket(childrenKey)
	if children == nil {
		return false
	}
	k, _ := children.Cursor().First()
	return k != nil
}

func (n *node) ForAll(f func(*models.Route)) error {
	if r, err := n.Route(); err != nil {
		return err
	} else if r != nil {
		f(r)
	}
	if r, err := n.TrailingSlashRoute(); err != nil {
		return err
	} else if r != nil {
		f(r)
	}
	children := n.b.Bucket(childrenKey)
	if children != nil {
		c := children.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			n := getNode(children, k)
			if err := n.ForAll(f); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *node) Remove() error {
	return n.remove()
}