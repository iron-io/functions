package bolt_test

import (
	"os"
	"testing"

	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
	"github.com/iron-io/functions/api/datastore"
)

const tmpBolt = "/tmp/func_test_bolt.db"

func TestBolt(t *testing.T) {
	os.Remove(tmpBolt)
	datastoretest.New(datastore.New("bolt://" + tmpBolt))(t)
}
