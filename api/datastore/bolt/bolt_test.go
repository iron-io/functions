package bolt

import (
	"net/url"
	"os"
	"testing"

	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
	"github.com/iron-io/functions/api/models"
)

const tmpBolt = "/tmp/func_test_bolt.db"

func TestBolt(t *testing.T) {
	os.Remove(tmpBolt)
	u, err := url.Parse("bolt://" + tmpBolt)
	if err != nil {
		t.Fatalf("failed to parse url:", err)
	}
	ds, err := New(u)
	if err != nil {
		t.Fatalf("failed to create bolt datastore:", err)
	}
	datastoretest.Test(t, ds)
}

func BenchmarkBolt(b *testing.B) {
	u, err := url.Parse("bolt://" + tmpBolt)
	if err != nil {
		b.Fatalf("failed to parse url:", err)
	}

	datastoretest.Benchmark(b, func(*testing.B) (models.Datastore, func()) {
		os.Remove(tmpBolt)
		ds, err := New(u)
		if err != nil {
			b.Fatalf("failed to create bolt datastore:", err)
		}
		return ds, func() {os.Remove(tmpBolt)}
	})
}