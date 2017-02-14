package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/iron-io/functions/api/datastore/internal/datastoreutil"
	"github.com/iron-io/functions/api/models"

	"github.com/Sirupsen/logrus"
	"github.com/lib/pq"
)

const routesTableCreate = `
CREATE TABLE IF NOT EXISTS routes (
	app_name character varying(256) NOT NULL,
	path text NOT NULL,
	nameless_path text NOT NULL,
	image character varying(256) NOT NULL,
	format character varying(16) NOT NULL,
	maxc integer NOT NULL,
	memory integer NOT NULL,
	timeout integer NOT NULL,
	type character varying(16) NOT NULL,
	headers text NOT NULL,
	config text NOT NULL,
	PRIMARY KEY (app_name, path)
);`

const routesTablePathIndex = `CREATE UNIQUE INDEX IF NOT EXISTS routes_path_idx ON routes (app_name, path);`
const routesTableNamelessPathIndex = `CREATE UNIQUE INDEX IF NOT EXISTS routes_path_idx ON routes (app_name, nameless_path);`

const appsTableCreate = `CREATE TABLE IF NOT EXISTS apps (
    name character varying(256) NOT NULL PRIMARY KEY,
	config text NOT NULL
);`

const appTableNameIndex = `CREATE UNIQUE INDEX IF NOT EXISTS apps_name_idx ON apps (name);`

const extrasTableCreate = `CREATE TABLE IF NOT EXISTS extras (
    key character varying(256) NOT NULL PRIMARY KEY,
	value character varying(256) NOT NULL
);`

const extrasTableKeyIndex = `CREATE UNIQUE INDEX IF NOT EXISTS extras_key_idx ON extras (key);`

const routeSelector = `SELECT app_name, path, image, format, maxc, memory, type, timeout, headers, config FROM routes`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

type rowQuerier interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

type PostgresDatastore struct {
	db *sql.DB
}

func New(url *url.URL) (models.Datastore, error) {
	db, err := sql.Open("postgres", url.String())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	maxIdleConns := 30 // c.MaxIdleConnections
	db.SetMaxIdleConns(maxIdleConns)
	logrus.WithFields(logrus.Fields{"max_idle_connections": maxIdleConns}).Info("Postgres dialed")

	pg := &PostgresDatastore{
		db: db,
	}

	for _, v := range []string{routesTableCreate, appsTableCreate, extrasTableCreate,
		routesTablePathIndex, routesTableNamelessPathIndex, appTableNameIndex, extrasTableKeyIndex} {
		_, err = db.Exec(v)
		if err != nil {
			return nil, err
		}
	}

	return pg, nil
}

func (ds *PostgresDatastore) InsertApp(ctx context.Context, app *models.App) (*models.App, error) {
	var cbyte []byte
	var err error

	if app == nil {
		return nil, models.ErrDatastoreEmptyApp
	}

	if app.Name == "" {
		return nil, models.ErrDatastoreEmptyAppName
	}

	if app.Config != nil {
		cbyte, err = json.Marshal(app.Config)
		if err != nil {
			return nil, err
		}
	}

	_, err = ds.db.Exec(`INSERT INTO apps (name, config) VALUES ($1, $2);`,
		app.Name,
		string(cbyte),
	)

	if err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Code == "23505" {
			return nil, models.ErrAppsAlreadyExists
		}
		return nil, err
	}

	return app, nil
}

func (ds *PostgresDatastore) UpdateApp(ctx context.Context, app *models.App) (*models.App, error) {
	if app == nil {
		return nil, models.ErrAppsNotFound
	}

	cbyte, err := json.Marshal(app.Config)
	if err != nil {
		return nil, err
	}
	//TODO the mock impl adds the config values to the existing map
	res, err := ds.db.Exec(`
	  UPDATE apps SET
		config = $2
	  WHERE name = $1
	  RETURNING *;
	`,
		app.Name,
		string(cbyte),
	)

	if err != nil {
		return nil, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, models.ErrAppsNotFound
	}

	return app, nil
}

func (ds *PostgresDatastore) RemoveApp(ctx context.Context, appName string) error {
	if appName == "" {
		return models.ErrDatastoreEmptyAppName
	}

	_, err := ds.db.Exec(`
	  DELETE FROM apps
	  WHERE name = $1
	`, appName)

	if err != nil {
		return err
	}

	return nil
}

func (ds *PostgresDatastore) GetApp(ctx context.Context, name string) (*models.App, error) {
	if name == "" {
		return nil, models.ErrDatastoreEmptyAppName
	}

	row := ds.db.QueryRow("SELECT name, config FROM apps WHERE name=$1", name)

	var resName string
	var config string
	err := row.Scan(&resName, &config)

	res := &models.App{
		Name: resName,
	}

	json.Unmarshal([]byte(config), &res.Config)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrAppsNotFound
		}
		return nil, err
	}

	return res, nil
}

func scanApp(scanner rowScanner, app *models.App) error {
	var configStr string

	err := scanner.Scan(
		&app.Name,
		&configStr,
	)

	json.Unmarshal([]byte(configStr), &app.Config)

	return err
}

func (ds *PostgresDatastore) GetApps(ctx context.Context, filter *models.AppFilter) ([]*models.App, error) {
	res := []*models.App{}

	filterQuery, args := buildFilterAppQuery(filter)
	rows, err := ds.db.Query(fmt.Sprintf("SELECT DISTINCT * FROM apps %s", filterQuery), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var app models.App
		err := scanApp(rows, &app)

		if err != nil {
			if err == sql.ErrNoRows {
				return res, nil
			}
			return res, err
		}
		res = append(res, &app)
	}

	if err := rows.Err(); err != nil {
		return res, err
	}
	return res, nil
}

func (ds *PostgresDatastore) InsertRoute(ctx context.Context, route *models.Route) (*models.Route, error) {
	if route == nil {
		return nil, models.ErrDatastoreEmptyRoute
	}

	hbyte, err := json.Marshal(route.Headers)
	if err != nil {
		return nil, err
	}

	cbyte, err := json.Marshal(route.Config)
	if err != nil {
		return nil, err
	}
	err = ds.Tx(ctx, func(tx *sql.Tx) error {
		same, err := tx.QueryContext(ctx, `SELECT DISTINCT 1 FROM routes WHERE app_name=$1 AND path=$2`,
			route.AppName, route.Path)
		if err != nil {
			return err
		}
		defer same.Close()
		if same.Next() {
			return models.ErrRoutesAlreadyExists
		}

		namelessPath := datastoreutil.StripParamNames(route.Path)
		conflicts, err := tx.QueryContext(ctx, `SELECT 1 FROM routes WHERE app_name=$1 AND nameless_path~$2 LIMIT 1`,
			route.AppName, conflictRegexp(namelessPath))
		if err != nil {
			return err
		}
		defer conflicts.Close()
		if conflicts.Next() {
			return models.ErrRoutesCreate
		}

		_, err = tx.Exec(`
		INSERT INTO routes (
			app_name,
			path,
			nameless_path,
			image,
			format,
			maxc,
			memory,
			type,
			timeout,
			headers,
			config
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
			route.AppName,
			route.Path,
			namelessPath,
			route.Image,
			route.Format,
			route.MaxConcurrency,
			route.Memory,
			route.Type,
			route.Timeout,
			string(hbyte),
			string(cbyte),
		)
		return err
	})

	if err != nil {
		return nil, err
	}
	return route, nil
}

func (ds *PostgresDatastore) Tx(ctx context.Context, f func(*sql.Tx) error) error {
	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	err = f(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (ds *PostgresDatastore) UpdateRoute(ctx context.Context, route *models.Route) (*models.Route, error) {
	if route == nil {
		return nil, models.ErrDatastoreEmptyRoute
	}

	hbyte, err := json.Marshal(route.Headers)
	if err != nil {
		return nil, err
	}

	cbyte, err := json.Marshal(route.Config)
	if err != nil {
		return nil, err
	}

	res, err := ds.db.Exec(`
		UPDATE routes SET
			image = $3,
			format = $4,
			memory = $5,
			maxc = $6,
			type = $7,
			timeout = $8,
			headers = $9,
			config = $10
		WHERE app_name = $1 AND path = $2;`,
		route.AppName,
		route.Path,
		route.Image,
		route.Format,
		route.Memory,
		route.MaxConcurrency,
		route.Type,
		route.Timeout,
		string(hbyte),
		string(cbyte),
	)

	if err != nil {
		return nil, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, models.ErrRoutesNotFound
	}

	return route, nil
}

func (ds *PostgresDatastore) RemoveRoute(ctx context.Context, appName, routePath string) error {
	if appName == "" {
		return models.ErrDatastoreEmptyAppName
	}

	if routePath == "" {
		return models.ErrDatastoreEmptyRoutePath
	}

	res, err := ds.db.Exec(`
		DELETE FROM routes
		WHERE path = $1 AND app_name = $2
	`, routePath, appName)

	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return models.ErrRoutesRemoving
	}

	return nil
}

func scanRoute(scanner rowScanner, route *models.Route) error {
	var headerStr string
	var configStr string

	err := scanner.Scan(
		&route.AppName,
		&route.Path,
		&route.Image,
		&route.Format,
		&route.Memory,
		&route.MaxConcurrency,
		&route.Type,
		&route.Timeout,
		&headerStr,
		&configStr,
	)

	if headerStr == "" {
		return models.ErrRoutesNotFound
	}

	json.Unmarshal([]byte(headerStr), &route.Headers)
	json.Unmarshal([]byte(configStr), &route.Config)

	return err
}

func (ds *PostgresDatastore) GetRoute(ctx context.Context, appName, routePath string) (*models.Route, error) {
	if appName == "" {
		return nil, models.ErrDatastoreEmptyAppName
	}

	if routePath == "" {
		return nil, models.ErrDatastoreEmptyRoutePath
	}

	var route models.Route

	row := ds.db.QueryRow(fmt.Sprintf("%s WHERE app_name=$1 AND nameless_path~$2", routeSelector), appName, pathRegexp(routePath))
	err := scanRoute(row, &route)

	if err == sql.ErrNoRows {
		return nil, models.ErrRoutesNotFound
	} else if err != nil {
		return nil, err
	}
	return &route, nil
}

// The pathRegexp function returns a regexp to match namessless paths which accept path.
func pathRegexp(path string) string {
	if path == "/" {
		return `^(/|/\*)$`
	}
	var b bytes.Buffer
	b.WriteRune('^')
	parts, trailingSlash := datastoreutil.SplitPath(path)
	for _, p := range parts {
		fmt.Fprintf(&b, `/(\*|((:|%s)`, p)
	}
	if trailingSlash {
		b.WriteRune('/')
	}
	for range parts {
		b.WriteString("))")
	}
	b.WriteRune('$')

	return b.String()
}

// The pathRegexp function returns a regexp to match nameless paths with conflict with namelessPath.
func conflictRegexp(namelessPath string) string {
	if namelessPath == "/" {
		return `^/$`
	}
	var b bytes.Buffer
	b.WriteRune('^')
	parts, trailingSlash := datastoreutil.SplitPath(namelessPath)
	for _, p := range parts {
		switch p {
		case "*":
			fmt.Fprint(&b, `/((.+`)
		case ":":
			fmt.Fprint(&b, `/(([^:](.*))|(:`)
		default:
			fmt.Fprintf(&b, `/(\*|:(.*)|((:|%s)`, p)
		}
	}
	if trailingSlash {
		b.WriteRune('/')
	}
	for range parts {
		b.WriteString("))")
	}
	b.WriteRune('$')

	return b.String()
}

func (ds *PostgresDatastore) GetRoutes(ctx context.Context, filter *models.RouteFilter) ([]*models.Route, error) {
	res := []*models.Route{}
	filterQuery, args := buildFilterRouteQuery(filter)
	rows, err := ds.db.Query(fmt.Sprintf("%s %s", routeSelector, filterQuery), args...)
	// todo: check for no rows so we don't respond with a sql 500 err
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var route models.Route
		err := scanRoute(rows, &route)
		if err != nil {
			continue
		}
		res = append(res, &route)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (ds *PostgresDatastore) GetRoutesByApp(ctx context.Context, appName string, filter *models.RouteFilter) ([]*models.Route, error) {
	res := []*models.Route{}
	filter.AppName = appName
	filterQuery, args := buildFilterRouteQuery(filter)
	rows, err := ds.db.Query(fmt.Sprintf("%s %s", routeSelector, filterQuery), args...)
	// todo: check for no rows so we don't respond with a sql 500 err
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var route models.Route
		err := scanRoute(rows, &route)
		if err != nil {
			continue
		}
		res = append(res, &route)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func buildFilterAppQuery(filter *models.AppFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	if filter.Name != "" {
		return "WHERE name LIKE $1", []interface{}{filter.Name}
	}

	return "", nil
}

func buildFilterRouteQuery(filter *models.RouteFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}
	var b bytes.Buffer
	var args []interface{}

	where := func(colOp, val string) {
		if val != "" {
			args = append(args, val)
			if len(args) == 1 {
				fmt.Fprintf(&b, "WHERE %s $1", colOp)
			} else {
				fmt.Fprintf(&b, " AND %s $%d", colOp, len(args))
			}
		}
	}

	where("path =", filter.Path)
	where("app_name =", filter.AppName)
	where("image =", filter.Image)

	return b.String(), args
}

func (ds *PostgresDatastore) Put(ctx context.Context, key, value []byte) error {
	if key == nil || len(key) == 0 {
		return models.ErrDatastoreEmptyKey
	}

	_, err := ds.db.Exec(`
	    INSERT INTO extras (
			key,
			value
		)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET
			value = $2;
		`, string(key), string(value))

	if err != nil {
		return err
	}

	return nil
}

func (ds *PostgresDatastore) Get(ctx context.Context, key []byte) ([]byte, error) {
	if key == nil || len(key) == 0 {
		return nil, models.ErrDatastoreEmptyKey
	}

	row := ds.db.QueryRow("SELECT DISTINCT value FROM extras WHERE key=$1", key)

	var value string
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return []byte(value), nil
}
