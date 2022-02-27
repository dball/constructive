package index

import (
	"fmt"

	"github.com/dball/constructive/internal/compare"
	"github.com/dball/constructive/internal/ids"
	. "github.com/dball/constructive/pkg/types"

	"github.com/google/btree"
)

func BuildIndex() BTreeIndex {
	return BTreeIndex{tree: *btree.New(16)}
}

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
		return extant.datum
	}
}

type predicate func(Datum) bool

type rangeSearch struct {
	indexType  IndexType
	start      Datum
	ascending  bool
	filter     predicate
	terminator predicate
}

// TODO if this returned a seq of range searches, we could thread any close
// signal from the search consumer back to the id seqs and be maximally lazy.
func buildRangeSearches(c Constraints) []rangeSearch {
	var es ids.Seq
	if c.E != nil {
		es = c.E.Seq()
	}
	var as ids.Seq
	if c.A != nil {
		as = c.A.Seq()
	}
	var indexType IndexType
	searchCount := 1
	switch {
	case c.E != nil:
		indexType = IndexEAV
		searchCount *= es.Length
		if as.Length != 0 {
			searchCount *= as.Length
		}
		var filter predicate
		if c.V != nil {
			filter = buildValueFilter(c.V).Pred
		}
		searches := make([]rangeSearch, 0, searchCount)
		for e := range es.Values {
			// The way golang closures and range blocks work is the gift that keeps on taking.
			e := e
			if c.A != nil {
				for a := range as.Values {
					a := a
					search := rangeSearch{
						indexType:  indexType,
						start:      Datum{E: e, A: a},
						ascending:  true,
						filter:     filter,
						terminator: func(d Datum) bool { return d.E > e || d.A > a },
					}
					searches = append(searches, search)
				}
			} else {
				search := rangeSearch{
					indexType:  indexType,
					start:      Datum{E: e},
					ascending:  true,
					filter:     filter,
					terminator: func(d Datum) bool { return d.E > e },
				}
				searches = append(searches, search)
			}
		}
		return searches
	case c.A != nil:
		indexType = IndexAVE
		searchCount *= as.Length
		if es.Length != 0 {
			searchCount *= es.Length
		}
		panic("TODO")
	default:
		panic("TODO")
	}
}

func (idx *BTreeIndex) doSearch(search rangeSearch) Seq {
	start := Node{kind: search.indexType, datum: search.start}
	if !search.ascending {
		panic("TODO")
	}
	datums := make(chan Datum)
	cls := make(chan Void)
	seq := Seq{Values: datums, Close: cls}
	iter := func(item btree.Item) bool {
		select {
		case <-cls:
			close(datums)
			return false
		default:
		}
		node := item.(Node)
		datum := node.datum
		fmt.Println("inspecting node", datum)
		if node.kind != search.indexType || search.terminator != nil && search.terminator(datum) {
			fmt.Println("terminating")
			close(datums)
			return false
		}
		if search.filter == nil || search.filter(datum) {
			fmt.Println("accepted node", datum)
			datums <- datum
		} else {
			fmt.Println("rejected node", datum)
		}
		return true
	}
	go idx.tree.AscendGreaterOrEqual(start, iter)
	return seq
}

func (idx *BTreeIndex) Select(sel Selection) Seq {
	c := idx.buildConstraints(sel)
	searches := buildRangeSearches(c)
	datums := make(chan Datum, 16)
	cls := make(chan Void)
	seq := Seq{Values: datums, Close: cls}
	go func() {
	OUTER:
		// If we don't care to retain sequentiality, these could be concurrent
		for _, search := range searches {
			sseq := idx.doSearch(search)
			for {
				fmt.Println("select waiting for datum")
				select {
				case datum, ok := <-sseq.Values:
					if !ok {
						fmt.Println("sseq is closed")
						continue OUTER
					}
					select {
					case datums <- datum:
						fmt.Println("conveyed datum", datum)
					case <-cls:
						fmt.Println("closed while trying to write datum")
						sseq.Close <- Void{}
						break OUTER
					}
				case <-cls:
					fmt.Println("closed while trying to read datum")
					sseq.Close <- Void{}
					break OUTER
				}
			}
		}
		fmt.Println("closing select datums")
		close(datums)
	}()
	return seq
}

func (idx *BTreeIndex) buildConstraints(sel Selection) Constraints {
	c := Constraints{
		E: idx.resolveESel(sel.E),
	}
	// TODO also resolveASel
	switch a := sel.A.(type) {
	case ID:
		c.A = ids.Scalar(a)
	case nil:
	default:
		panic("TODO")
	}
	c.V = sel.V
	return c
}

func (idx *BTreeIndex) resolveESel(sel ESel) ids.Constraint {
	switch e := sel.(type) {
	case ID:
		return ids.Scalar(e)
	case LookupRef:
		panic("TODO")
	case Ident:
		panic("TODO")
	case ESet:
		r := make(ids.Set, len(e))
		for esel := range e {
			seq := idx.resolveESel(esel).Seq()
			for id := range seq.Values {
				r[id] = Void{}
			}
		}
		return r
	case ERange:
		r := ids.Range{}
		switch min := e.Min.(type) {
		case ID:
			r.Min = min
		case LookupRef:
			panic("TODO")
		case Ident:
			panic("TODO")
		case nil:
		default:
			panic("TODO")
		}
		switch max := e.Max.(type) {
		case ID:
			r.Max = max
		case LookupRef:
			panic("TODO")
		case Ident:
			panic("TODO")
		case nil:
		default:
			panic("TODO")
		}
		return r
	default:
		panic("TODO nope")
	}
}

// The Constraints builder assumes the responsibility of ensuring the
// components are satisfiable by the index, and that sets have been
// converted to ranges if desirable.
type Constraints struct {
	E ids.Constraint
	A ids.Constraint
	V VSel
}

// TODO how do we indicate and handle open/closedness on the bounds?
type ValueFilter struct {
	Pred predicate
	Min  Value
	Max  Value
}

var matchesNoValue = ValueFilter{Pred: func(datum Datum) bool { return false }}

func buildValueFilter(vsel VSel) ValueFilter {
	switch typed := vsel.(type) {
	case ID:
		return ValueFilter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case String:
		return ValueFilter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case Int:
		return ValueFilter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case Bool:
		return ValueFilter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case Inst:
		return ValueFilter{Pred: func(datum Datum) bool { return datum.V == typed }, Min: typed, Max: typed}
	case VSet:
		filters := make([]ValueFilter, 0, len(typed))
		for v, _ := range typed {
			filters = append(filters, buildValueFilter(v))
		}
		return ValueFilter{
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
		var exemplar Value
		if typed.Min != nil {
			exemplar = typed.Min
		} else {
			exemplar = typed.Max
		}
		if exemplar == nil {
			return matchesNoValue
		}
		var ok bool
		switch exemplar.(type) {
		case ID:
			var min ID
			var max ID
			if typed.Min != nil {
				min, ok = typed.Min.(ID)
				if !ok {
					return matchesNoValue
				}
			}
			if typed.Max != nil {
				max, ok = typed.Max.(ID)
				if !ok {
					return matchesNoValue
				}
			}
			if typed.Min != nil {
				if typed.Max != nil {
					return ValueFilter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
						id, ok := datum.V.(ID)
						return ok && id >= min && id <= max
					}}
				}
				return ValueFilter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
					id, ok := datum.V.(ID)
					return ok && id >= min
				}}
			}
			return ValueFilter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
				id, ok := datum.V.(ID)
				return ok && id <= max
			}}
		case String:
			var min String
			var max String
			if typed.Min != nil {
				min, ok = typed.Min.(String)
				if !ok {
					return matchesNoValue
				}
			}
			if typed.Max != nil {
				max, ok = typed.Max.(String)
				if !ok {
					return matchesNoValue
				}
			}
			if typed.Min != nil {
				if typed.Max != nil {
					return ValueFilter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
						id, ok := datum.V.(String)
						return ok && id >= min && id <= max
					}}
				}
				return ValueFilter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
					id, ok := datum.V.(String)
					return ok && id >= min
				}}
			}
			return ValueFilter{Min: typed.Min, Max: typed.Max, Pred: func(datum Datum) bool {
				id, ok := datum.V.(String)
				return ok && id <= max
			}}
		case Int:
			panic("TODO")
		case Inst:
			panic("TODO")
		case Bool:
			panic("TODO")
		default:
			return matchesNoValue
		}
	default:
		return matchesNoValue
	}
}
