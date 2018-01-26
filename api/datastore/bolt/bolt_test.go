package bolt

import (
	"net/url"
	"os"
	"testing"

	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
)

const tmpBolt = "/tmp/bolt_fn_test.db"

func TestDatastore(t *testing.T) {
	u, err := url.Parse("bolt://" + tmpBolt)
	if err != nil {
		t.Fatalf("failed to parse url:", err)
	}
	ds, err := New(u)
	if err != nil {
		t.Fatalf("failed to create bolt datastore:", err)
	}
	datastoretest.Test(t, ds)
	os.Remove(tmpBolt)
}
