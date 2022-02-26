package index

import (
	"github.com/dball/constructive/internal/compare"
	"github.com/dball/constructive/internal/ids"
	. "github.com/dball/constructive/pkg/types"

	"github.com/google/btree"
)

type Seq struct {
	Close  chan<- Void
	Values <-chan Datum
}

type BTreeIndex struct {
	tree btree.BTree
}

type Node struct {
	kind  IndexType
	datum Datum
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

func (idx *BTreeIndex) Select(c Constraints) Seq {
	var indexType IndexType
	if c.E != nil {
		indexType = IndexEAV
	} else if c.A != nil {
		indexType = IndexAVE
	}
	filters := make([]func(d Datum) bool, 0, 2)
	datums := make(chan Datum, 16)
	cls := make(chan Void)
	seq := Seq{Values: datums, Close: cls}
	iter := func(item btree.Item) bool {
		node := item.(Node)
		datum := node.datum
		pass := true
		for _, filter := range filters {
			if !filter(datum) {
				pass = false
				break
			}
		}
		if pass {
			select {
			case datums <- datum:
			case <-cls:
				return false
			}
		}
		return true
	}
	switch indexType {
	case IndexEAV:
		if c.V != nil {
			// TODO construct and append filter from VSel
		}
		if c.E == nil {
			go idx.tree.Ascend(iter)
			return seq
		}
	}
	panic("TODO")
}

// The Constraints builder assumes the responsibility of ensuring the
// components are satisfiable by the index, and that sets have been
// converted to ranges if desirable.
type Constraints struct {
	E ids.Constraint
	A ids.Constraint
	V VSel
}

// Filters filter datums.
// TODO how do we indicate and handle open/closedness on the bounds?
type Filter struct {
	Pred func(datum Datum) bool
	Min  Value
	Max  Value
}

var matchesNone = Filter{Pred: func(datum Datum) bool { return false }}

func BuildVSelFilter(vsel VSel) Filter {
	switch typed := vsel.(type) {
	case ID:
		return Filter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case String:
		return Filter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case Int:
		return Filter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case Bool:
		return Filter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case Inst:
		return Filter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case VSet:
		filters := make([]Filter, 0, len(typed))
		for v, _ := range typed {
			filters = append(filters, BuildVSelFilter(v))
		}
		return Filter{
			Pred: func(datum Datum) bool {
				for _, filter := range filters {
					if filter.Pred(datum) {
						return true
					}
				}
				return false
			},
		}
	case VRange:
		// min and max
		// min and no max
		// max and no min

		// id, string, int, bool, inst

		// we may apply the filter to many datums, so type switches that don't
		// depend on the datum should be performed outside the Pred method
		var exemplar Value
		if typed.Min != nil {
			exemplar = typed.Min
		} else {
			exemplar = typed.Max
		}
		if exemplar == nil {
			return matchesNone
		}
		var ok bool
		switch exemplar.(type) {
		case ID:
			var min ID
			var max ID
			if typed.Min != nil {
				min, ok = typed.Min.(ID)
				if !ok {
					return matchesNone
				}
			}
			if typed.Max != nil {
				max, ok = typed.Max.(ID)
				if !ok {
					return matchesNone
				}
			}
			if typed.Min != nil {
				if typed.Max != nil {
					return Filter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
						id, ok := datum.V.(ID)
						return ok && id >= min && id <= max
					}}
				}
				return Filter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
					id, ok := datum.V.(ID)
					return ok && id >= min
				}}
			}
			return Filter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
				id, ok := datum.V.(ID)
				return ok && id <= max
			}}
		case String:
			var min String
			var max String
			if typed.Min != nil {
				min, ok = typed.Min.(String)
				if !ok {
					return matchesNone
				}
			}
			if typed.Max != nil {
				max, ok = typed.Max.(String)
				if !ok {
					return matchesNone
				}
			}
			if typed.Min != nil {
				if typed.Max != nil {
					return Filter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
						id, ok := datum.V.(String)
						return ok && id >= min && id <= max
					}}
				}
				return Filter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
					id, ok := datum.V.(String)
					return ok && id >= min
				}}
			}
			return Filter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
				id, ok := datum.V.(String)
				return ok && id <= max
			}}
		case Int:
		case Inst:
		case Bool:
		}
	}
	return matchesNone
}
