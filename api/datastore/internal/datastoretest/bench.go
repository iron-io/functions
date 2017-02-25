package datastoretest

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/iron-io/functions/api/models"
)

func Benchmark(b *testing.B, newDS func(*testing.B) (models.Datastore, func())) {
	ctx := context.Background()
	withDS := func(f func(ds models.Datastore)) {
		ds, close := newDS(b)
		defer close()
		f(ds)
	}
	b.Run("GetRoute", func(b *testing.B) {
		b.Run("standard", func(b *testing.B) {
			for _, apps := range []int{100, 10, 1} {
				for _, routes := range []int{10, 100, 1000, 10000} {
					withDS(func(ds models.Datastore) {
						b.Run(fmt.Sprintf("apps=%d/routes=%d", apps, routes),
							standard(ctx, ds, apps, routes))
					})
				}
			}
		})

		b.Run("segmentCount", func(b *testing.B) {
			for _, cnt := range []int{1, 5, 10, 50, 100} {
				withDS(func(ds models.Datastore) {
					b.Run(fmt.Sprintf("segments=%d", cnt),
						segmentCount(ctx, cnt, ds))
				})
			}
		})

		b.Run("branchFactor", func(b *testing.B) {
			for _, cnt := range []int{1,5,10,50,100,500,1000,5000,10000} {
				withDS(func(ds models.Datastore) {
					b.Run(fmt.Sprintf("branches=%d", cnt),
						branchFactor(ctx, ds, cnt))
				})
			}
		})

		b.Run("segmentSize", func(b *testing.B) {
			for _, size := range []int{1, 5, 10, 50, 100, 200} {
				withDS(func(ds models.Datastore) {
					b.Run(fmt.Sprintf("chars=%d", size),
						segmentSize(ctx, size, ds))
				})
			}
		})
	})
}

func standard(ctx context.Context, ds models.Datastore, apps, routes int) func(*testing.B) {
	var cases []*models.Route
	var once sync.Once
	return func(b *testing.B) {
		once.Do(func() {
			cases = make([]*models.Route, routes, routes)
			for i := 0; i < apps; i++ {
				if _, err := ds.InsertApp(ctx, &models.App{Name: fmt.Sprintf("%#p", i)}); err != nil {
					b.Fatal("failed to insert app: ", err)
				}
			}

			wildcardReplacer := strings.NewReplacer(":","","*","")
			for i := 0; i < routes; i++ {
				appName := fmt.Sprintf("%#p", i%apps)
				path := generateRoute(uint64(i))

				_, err := ds.InsertRoute(ctx, &models.Route{
					AppName: appName,
					Path:    path,
				})
				if err != nil {
					b.Fatalf("failed to insert route i=%d %q: %s", i, path, err)
				}

				cases[i] = &models.Route{AppName: appName, Path: wildcardReplacer.Replace(path)}
			}
		})

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			c := cases[i%len(cases)]
			if _, err := ds.GetRoute(ctx, c.AppName, c.Path); err != nil {
				b.Fatal("failed to get route:", err)
			}
		}
	}
}

func branchFactor(ctx context.Context, ds models.Datastore, cnt int) func(*testing.B) {
	var paths []string
	var once sync.Once
	return func(b *testing.B) {
		once.Do(func() {
			if _, err := ds.InsertApp(ctx, &models.App{Name: "test"}); err != nil {
				b.Fatal("failed to insert app: ", err)
			}

			paths = make([]string, cnt, cnt)
			for i := 0; i < cnt; i++ {
				paths[i] = fmt.Sprintf("/%x", md5.Sum([]byte{byte(i >> 16), byte(i), byte(i >> 24), byte(i >> 8)}))
				if _, err := ds.InsertRoute(ctx, &models.Route{AppName: "test", Path: paths[i]}); err != nil {
					b.Fatalf("failed to insert route #%d: %s", cnt, err)
				}
			}
		})

		b.ResetTimer()

		for i:=0;i<b.N;i++ {
			path := paths[i % len(paths)]
			if _, err := ds.GetRoute(ctx, "test", path); err != nil {
				b.Fatalf("failed to get route %q: %s", path, err)
			}
		}
	}
}

func segmentCount(ctx context.Context, cnt int, ds models.Datastore) func(b *testing.B) {
	var path string
	var once sync.Once
	return func(b *testing.B) {
		once.Do(func() {
			if _, err := ds.InsertApp(ctx, &models.App{Name: "test"}); err != nil {
				b.Fatal("failed to insert app: ", err)
			}

			path = strings.Repeat("/test", cnt)
			if _, err := ds.InsertRoute(ctx, &models.Route{AppName: "test", Path: path}); err != nil {
				b.Fatalf("failed to insert route with %d segments: %s", cnt, err)
			}
		})

		b.ResetTimer()

		for i:=0;i<b.N;i++ {
			if _, err := ds.GetRoute(ctx, "test", path); err != nil {
				b.Fatal("failed to get route:", err)
			}
		}
	}
}

func segmentSize(ctx context.Context, size int, ds models.Datastore) func(b *testing.B) {
	var path string
	var once sync.Once
	return func(b *testing.B) {
		once.Do(func() {
			if _, err := ds.InsertApp(ctx, &models.App{Name: "test"}); err != nil {
				b.Fatal("failed to insert app: ", err)
			}

			path = strings.Repeat("/" + strings.Repeat("a", size), 5)
			if _, err := ds.InsertRoute(ctx, &models.Route{AppName: "test", Path: path}); err != nil {
				b.Fatalf("failed to insert route with %d character segments: %s", size, err)
			}
		})

		b.ResetTimer()

		for i:=0;i<b.N;i++ {
			if _, err := ds.GetRoute(ctx, "test", path); err != nil {
				b.Fatal("failed to get route:", err)
			}
		}
	}
}

// generateRoute generates a deterministic test route based on u.
// Routes are 32 bytes printed as hex, and split into 1, 2, 4, or 8 parts.
// ~1/5 of the parts after the first are parameters. ~1/2 of suffix parameters are catch all.
// Example: /61aab75e6134555e/:e6e8d0fa3ed7de93/20c98ae9da239a03/*8cad52b84387e5d5
// Verified conflict-free up to u=100000.
// TODO consider http://www.jandrewrogers.com/2015/05/27/metrohash/
func generateRoute(u uint64) string {
	s := make([]byte, 0, 32)
	for _, b := range md5.Sum([]byte{byte(u >> 24), byte(u >> 8), byte(u), byte(u >> 16)}) {
		s = append(s, b)
	}
	for _, b := range md5.Sum([]byte{byte(u), byte(u >> 16), byte(u >> 24), byte(u >> 8)}) {
		s = append(s, b)
	}

	parts := 1 << (u % 4)  // [1,2,4,8]
	each := len(s) / parts // [32,16,8,4]

	var b bytes.Buffer
	for i := 0; i < parts; i++ {
		b.WriteRune('/')

		name := s[i*each : (i+1)*each]

		nameSum := 0
		for j := 0; j < len(name); j++ {
			nameSum += int(name[j])
		}

		// if not first, 1/5 chance of param
		if i > 0 && nameSum%5 == 0 {
			// if last, 1/2 chance of *
			if i == parts-1 && u%2 == 0 {
				b.WriteRune('*')
			} else {
				b.WriteRune(':')
			}
		}
		fmt.Fprintf(&b, "%x", name)
	}

	return b.String()
}