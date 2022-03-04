package index

import (
	"testing"

	"github.com/dball/constructive/internal/iterator"
	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttrs(t *testing.T) {
	idx := BuildIndex().InitSys()
	t.Run("assert attrs", func(t *testing.T) {
		idx.Assert(D(500, sys.DbIdent, String("person/name"), 100))
		idx.Assert(D(500, sys.AttrType, sys.AttrTypeString, 100))
		idx.Assert(D(501, sys.DbIdent, String("person/age"), 100))
		idx.Assert(D(501, sys.AttrType, sys.AttrTypeInt, 100))
		// I reject your notion of linear time
		idx.Assert(D(501, sys.AttrCardinality, sys.AttrCardinalityMany, 100))
	})

	t.Run("cardinality one string attr", func(t *testing.T) {
		t.Run("accepts a string value", func(t *testing.T) {
			idx.Assert(D(1000, 500, String("Donald"), 101))
			sel := idx.Select(Selection{E: ID(1000)})
			assert.Equal(t, []Datum{D(1000, 500, String("Donald"), 101)}, slurp(sel))
		})
		t.Run("rejects another type", func(t *testing.T) {
			_, err := idx.Assert(D(1001, 500, Int(23), 102))
			assert.ErrorIs(t, err, ErrInvalidValue)
		})
		t.Run("replaces a string value", func(t *testing.T) {
			extant, err := idx.Assert(D(1000, 500, String("Stephen"), 102))
			require.NoError(t, err)
			assert.Equal(t, D(1000, 500, String("Donald"), 101), extant)
			sel := idx.Select(Selection{E: ID(1000)})
			assert.Equal(t, []Datum{D(1000, 500, String("Stephen"), 102)}, slurp(sel))
		})
	})

	t.Run("cardinality many int attr", func(t *testing.T) {
		t.Run("accepts an integer value", func(t *testing.T) {
			_, err := idx.Assert(D(1000, 501, Int(23), 101))
			require.NoError(t, err)
			sel := idx.Select(Selection{E: ID(1000), A: ID(501)})
			assert.Equal(t, []Datum{D(1000, 501, Int(23), 101)}, slurp(sel))
		})
		t.Run("adds another integer value", func(t *testing.T) {
			_, err := idx.Assert(D(1000, 501, Int(109), 102))
			require.NoError(t, err)
			sel := idx.Select(Selection{E: ID(1000), A: ID(501)})
			assert.Equal(t, []Datum{D(1000, 501, Int(23), 101), D(1000, 501, Int(109), 102)}, slurp(sel))
		})
		t.Run("adds another integer value", func(t *testing.T) {
			_, err := idx.Assert(D(1000, 501, Int(99), 103))
			require.NoError(t, err)
			sel := idx.Select(Selection{E: ID(1000), A: ID(501)})
			assert.Equal(t, []Datum{D(1000, 501, Int(23), 101), D(1000, 501, Int(99), 103), D(1000, 501, Int(109), 102)}, slurp(sel))
		})
	})
}

func TestInitSystem(t *testing.T) {
	idx := BuildIndex()
	idx.InitSys()
	datums := slurp(idx.Select(Selection{E: sys.DbIdent}))
	assert.Len(t, datums, 3)
}

func TestAssert(t *testing.T) {
	t.Skip("create 500 attr")
	idx := BuildIndex()
	t.Run("asserting a new datum returns the empty datum", func(t *testing.T) {
		extant, err := idx.Assert(D(1000, 500, Int(1), 100))
		require.NoError(t, err)
		assert.Equal(t, D(ID(0), ID(0), nil, ID(0)), extant)
	})
	t.Run("asserting a duplicate datum returns the extant datum", func(t *testing.T) {
		extant, err := idx.Assert(D(1000, 500, Int(1), 101))
		require.NoError(t, err)
		assert.Equal(t, D(ID(1000), ID(500), Int(1), 100), extant)
	})
	t.Run("asserting a new value returns the old datum", func(t *testing.T) {
		extant, err := idx.Assert(D(1000, 500, Int(5), 102))
		require.NoError(t, err)
		assert.Equal(t, D(ID(1000), ID(500), Int(1), 100), extant)
	})
	t.Run("asserting a newer value returns the old datum", func(t *testing.T) {
		extant, err := idx.Assert(D(1000, 500, Int(2), 103))
		require.NoError(t, err)
		assert.Equal(t, D(ID(1000), ID(500), Int(5), 102), extant)
	})
	t.Run("asserting an ident registers it in the cache", func(t *testing.T) {
		idx.Assert(D(1001, sys.DbIdent, String("person/name"), 104))
		assert.Equal(t, map[String]ID{"person/name": ID(1001)}, idx.idents)
	})
}

func TestSelect(t *testing.T) {
	t.Skip("create 500 attr")
	idx := BuildIndex()
	idx.Assert(D(1000, 500, String("Donald"), 100))
	idx.Assert(D(1001, 500, String("Stephen"), 100))
	iter := idx.Select(Selection{
		E: ID(1000),
		A: ID(500),
		V: String("Donald"),
	})
	datums := slurp(iter)
	require.Len(t, datums, 1)
	assert.Equal(t, D(ID(1000), ID(500), String("Donald"), ID(100)), datums[0])
}

func TestSelectLots(t *testing.T) {
	t.Skip("create 500-504 attrs")
	idx := BuildIndex().InitSys()
	es := []ID{1000, 1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009}
	as := []ID{500, 501, 502, 503, 504}
	v := String("a")
	tid := ID(100)
	for _, e := range es {
		for _, a := range as {
			idx.Assert(D(e, a, v, tid))
		}
	}
	t.Run("just one e", func(t *testing.T) {
		datums := slurp(idx.Select(Selection{E: ID(1005)}))
		assert.Len(t, datums, 5)
	})
	t.Run("an e and an a", func(t *testing.T) {
		datums := slurp(idx.Select(Selection{E: ID(1005), A: ID(502)}))
		assert.Len(t, datums, 1)
	})
	t.Run("two e in a set", func(t *testing.T) {
		datums := slurp(idx.Select(Selection{E: ESet{ID(1006): Void{}, ID(1004): Void{}}}))
		assert.Len(t, datums, 10)
	})
	t.Run("range of es", func(t *testing.T) {
		datums := slurp(idx.Select(Selection{E: ERange{Min: ID(1002), Max: ID(1004)}}))
		assert.Len(t, datums, 15)
	})
}

func TestIdents(t *testing.T) {
	idx := BuildIndex().InitSys()
	idx.Assert(D(ID(1000), sys.DbIdent, String("person/name"), ID(1000)))
	datums := slurp(idx.Select(Selection{E: Ident("person/name")}))
	assert.Equal(t, []Datum{D(ID(1000), sys.DbIdent, String("person/name"), ID(1000))}, datums)
}

func slurp(iter *iterator.Iterator) (datums []Datum) {
	for iter.Next() {
		datums = append(datums, iter.Value().(Datum))
	}
	return datums
}
