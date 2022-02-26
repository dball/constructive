package index

import (
	"testing"

	. "github.com/dball/constructive/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	idx := BuildIndex()
	t.Run("inserting a new datum returns the empty datum", func(t *testing.T) {
		extant := idx.InsertOne(D(1000, 1, Int(1), 100))
		assert.Equal(t, D(ID(0), ID(0), nil, ID(0)), extant)
	})
	t.Run("inserting a duplicate datum returns the extant datum", func(t *testing.T) {
		extant := idx.InsertOne(D(1000, 1, Int(1), 101))
		assert.Equal(t, D(ID(1000), ID(1), Int(1), 100), extant)
	})
	t.Run("inserting a new value returns the old datum", func(t *testing.T) {
		extant := idx.InsertOne(D(1000, 1, Int(5), 102))
		assert.Equal(t, D(ID(1000), ID(1), Int(1), 100), extant)
	})
	t.Run("inserting a newer value returns the old datum", func(t *testing.T) {
		extant := idx.InsertOne(D(1000, 1, Int(2), 103))
		assert.Equal(t, D(ID(1000), ID(1), Int(5), 102), extant)
	})
}
