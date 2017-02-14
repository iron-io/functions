package datastore

import (
	"testing"

	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
	"github.com/iron-io/functions/api/models"
)

func TestDatastore(t *testing.T) {
	datastoretest.Test(t, NewMock())
}

func BenchmarkDatastore(b *testing.B) {
	datastoretest.Benchmark(b, func(*testing.B) (models.Datastore, func()) {
		return NewMock(), func() {}
	})
}
