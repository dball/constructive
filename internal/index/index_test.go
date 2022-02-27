package index

import (
	"fmt"
	"testing"

	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestAssert(t *testing.T) {
	idx := BuildIndex()
	t.Run("asserting a new datum returns the empty datum", func(t *testing.T) {
		extant := idx.Assert(D(1000, 500, Int(1), 100))
		assert.Equal(t, D(ID(0), ID(0), nil, ID(0)), extant)
	})
	t.Run("asserting a duplicate datum returns the extant datum", func(t *testing.T) {
		extant := idx.Assert(D(1000, 500, Int(1), 101))
		assert.Equal(t, D(ID(1000), ID(500), Int(1), 100), extant)
	})
	t.Run("asserting a new value returns the old datum", func(t *testing.T) {
		extant := idx.Assert(D(1000, 500, Int(5), 102))
		assert.Equal(t, D(ID(1000), ID(500), Int(1), 100), extant)
	})
	t.Run("asserting a newer value returns the old datum", func(t *testing.T) {
		extant := idx.Assert(D(1000, 500, Int(2), 103))
		assert.Equal(t, D(ID(1000), ID(500), Int(5), 102), extant)
	})
	t.Run("asserting an ident registers it in the cache", func(t *testing.T) {
		idx.Assert(D(1001, sys.DbIdent, String("person/name"), 104))
		assert.Equal(t, map[String]ID{"person/name": ID(1001)}, idx.idents)
	})
}

func TestSelect(t *testing.T) {
	idx := BuildIndex()
	idx.Assert(D(1000, 500, String("Donald"), 100))
	idx.Assert(D(1001, 500, String("Stephen"), 100))
	seq := idx.Select(Selection{
		E: ID(1000),
		A: ID(500),
		V: String("Donald"),
	})
	count := 0
	var datum Datum
	for d := range seq.Values {
		datum = d
		count++
	}
	assert.Equal(t, 1, count)
	assert.Equal(t, D(ID(1000), ID(500), String("Donald"), ID(100)), datum)
}

func TestSelectLots(t *testing.T) {
	idx := BuildIndex()
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
	idx := BuildIndex()
	idx.Assert(D(ID(1000), sys.DbIdent, String("person/name"), ID(1000)))
	datums := slurp(idx.Select(Selection{E: Ident("person/name")}))
	assert.Equal(t, []Datum{D(ID(1000), sys.DbIdent, String("person/name"), ID(1000))}, datums)
}

func slurp(seq Seq) (datums []Datum) {
	for datum := range seq.Values {
		fmt.Println("Received", datum)
		datums = append(datums, datum)
	}
	return datums
}