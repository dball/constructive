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

func TestSelect(t *testing.T) {
	idx := BuildIndex()
	idx.InsertOne(D(1000, 1, String("Donald"), 100))
	idx.InsertOne(D(1001, 1, String("Stephen"), 100))
	seq := idx.Select(Selection{
		E: ID(1000),
		A: ID(1),
		V: String("Donald"),
	})
	count := 0
	var datum Datum
	for d := range seq.Values {
		datum = d
		count++
	}
	assert.Equal(t, 1, count)
	assert.Equal(t, D(ID(1000), ID(1), String("Donald"), ID(100)), datum)
}

func Test_doSearch(t *testing.T) {
	idx := BuildIndex()
	idx.InsertOne(D(1000, 1, String("Donald"), 100))
	idx.InsertOne(D(1001, 1, String("Stephen"), 100))
	search := rangeSearch{
		indexType:  IndexEAV,
		start:      D(ID(1000), ID(1), String("Donald"), ID(0)),
		ascending:  true,
		filter:     func(d Datum) bool { return true },
		terminator: func(d Datum) bool { return d.E > ID(1000) },
	}
	seq := idx.doSearch(search)
	count := 0
	var datum Datum
	for d := range seq.Values {
		datum = d
		count++
	}
	assert.Equal(t, 1, count)
	assert.Equal(t, D(ID(1000), ID(1), String("Donald"), ID(100)), datum)
}
