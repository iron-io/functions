package datastore_test

import (
	"testing"

	"github.com/iron-io/functions/api/datastore"
	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
)

func TestMock(t *testing.T) {
	datastoretest.New(datastore.New("mock://"))(t)
}
