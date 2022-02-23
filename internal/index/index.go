package index

import (
	"github.com/dball/constructive/internal/compare"
	. "github.com/dball/constructive/pkg/types"

	"github.com/google/btree"
)

type BTreeIndex struct {
	tree btree.BTree
}

type Node struct {
	kind  IndexType
	datum Datum
}

type Iterator interface {
	Next() Datum
	HasNext() bool
}

func compareKind(kind IndexType) compare.Fn {
	switch kind {
	case IndexEAV:
		return compare.EAV
	case IndexAEV:
		return compare.AEV
	case IndexAVE:
		return compare.AVE
	default:
		return nil
	}
}

func (n1 Node) Less(than btree.Item) bool {
	n2 := than.(Node)
	if n1.kind < n2.kind {
		return true
	} else if n1.kind > n2.kind {
		return false
	}
	return (compareKind(n1.kind))(n1.datum, n2.datum) < 0
}

func (idx *BTreeIndex) InsertOne(d Datum) Datum {
	floor := Datum{E: d.E, A: d.A}
	ceiling := Datum{E: d.E, A: d.A + 1}
	var extant Node
	idx.tree.AscendRange(Node{IndexEAV, floor}, Node{IndexEAV, ceiling}, func(item btree.Item) bool {
		extant = item.(Node)
		return false
	})
	switch {
	case extant.datum.E == 0:
		node := Node{IndexEAV, d}
		idx.tree.ReplaceOrInsert(node)
		node.kind = IndexAEV
		idx.tree.ReplaceOrInsert(node)
		// TODO only for unique a
		node.kind = IndexAVE
		idx.tree.ReplaceOrInsert(node)
		return Datum{}
	case Compare(d.V, extant.datum.V) == 0:
		return extant.datum
	default:
		node := Node{IndexEAV, extant.datum}
		idx.tree.Delete(node)
		node.kind = IndexAEV
		idx.tree.Delete(node)
		node.kind = IndexAVE
		idx.tree.Delete(node)

		node.datum = d
		node.kind = IndexEAV
		idx.tree.ReplaceOrInsert(node)
		node.kind = IndexAEV
		idx.tree.ReplaceOrInsert(node)
		node.kind = IndexAVE
		idx.tree.ReplaceOrInsert(node)
		return d
	}
}

var void struct{}

func (idx *BTreeIndex) Select(c Constraints) Iterator {
	// TODO maybe Iterator should be a chan of datums
	switch c.IndexType {
	case IndexEAV:
		switch e := c.E.(type) {
		case IDScalar:
			floor := Node{kind: IndexEAV, datum: Datum{E: ID(e)}}
			ceiling := Node{kind: IndexEAV, datum: Datum{E: ID(e) + 1}}
			idx.tree.AscendRange(ceiling, floor, iterator)
		case IDSet:
			// TODO for each e, idx. ascend range from e to e+1
		case IDRange:
			floor := Node{kind: IndexEAV, datum: Datum{E: e.min}}
			ceiling := Node{kind: IndexEAV, datum: Datum{E: e.max + 1}}
			idx.tree.AscendRange(ceiling, floor, iterator)
		}
	}
}

// The Constraints builder assumes the responsibility of ensuring the
// components are satisfiable by the index, and that sets have been
// converted to ranges if desirable.
type Constraints struct {
	IndexType IndexType
	E         IDConstraint
	A         IDConstraint
	V         VSel
}

type IDConstraint interface {
	IsIDConstraint()
}

type IDScalar ID
type IDSet map[ID]Void
type IDRange struct {
	min ID
	max ID
}

func (IDScalar) IsIDConstraint() {}
func (IDSet) IsIDConstraint()    {}
func (IDRange) IsIDConstraint()  {}
