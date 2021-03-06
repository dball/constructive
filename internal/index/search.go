package index

import (
	"github.com/dball/constructive/internal/ids"
	"github.com/dball/constructive/internal/iterator"
	. "github.com/dball/constructive/pkg/types"
	"github.com/google/btree"
)

type predicate func(Datum) bool

type rangeSearch struct {
	indexType  IndexType
	start      Datum
	ascending  bool
	filter     predicate
	terminator predicate
}

// TODO if this returned an iterator of range searches, we could thread any close
// signal from the search consumer back to the id iteraors and be maximally lazy.
func buildRangeSearches(c Constraints) []rangeSearch {
	var indexType IndexType
	searchCount := 1
	switch {
	case c.E != nil:
		// TODO we should prefer and use IndexAEV if we have both A and E and A is smaller
		indexType = IndexEAV
		searchCount *= c.E.Size()
		if c.A != nil {
			searchCount *= c.A.Size()
		}
		var filter predicate
		if c.V != nil {
			filter = buildValueFilter(c.V).Pred
		}
		searches := make([]rangeSearch, 0, searchCount)
		es := c.E.Iterator()
		for es.Next() {
			e := es.Value().(ID)
			if c.A != nil {
				as := c.A.Iterator()
				for as.Next() {
					a := as.Value().(ID)
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
		searchCount *= c.A.Size()
		searches := make([]rangeSearch, 0, searchCount)
		as := c.A.Iterator()
		for as.Next() {
			a := as.Value().(ID)
			switch c.V.(type) {
			case VRange:
				panic("TODO filter on value range")
			case VSet:
				panic("TODO range search on min and max with filter on set values")
			default:
				v := c.V.(Value)
				search := rangeSearch{
					indexType:  indexType,
					start:      Datum{A: a, V: v},
					ascending:  true,
					terminator: func(d Datum) bool { return d.A > a || Compare(v, d.V) != 0 },
				}
				searches = append(searches, search)
			}
		}
		return searches
	case c.V != nil:
		panic("TODO range search on v would require a vae index")
	default:
		search := rangeSearch{
			indexType: IndexEAV,
			ascending: true,
		}
		return []rangeSearch{search}
	}
}

type btreeRangeSearch struct {
	rangeSearch
	idx *BTreeIndex
}

func (search btreeRangeSearch) Each(accept iterator.Accept) {
	start := Node{kind: search.indexType, datum: search.start}
	if !search.ascending {
		panic("TODO")
	}
	treeIter := func(item btree.Item) bool {
		node := item.(Node)
		datum := node.datum
		if node.kind != search.indexType || search.terminator != nil && search.terminator(datum) {
			return false
		}
		if search.filter == nil || search.filter(datum) {
			if !accept(datum) {
				return false
			}
		}
		return true
	}
	search.idx.tree.AscendGreaterOrEqual(start, treeIter)
}

func (search btreeRangeSearch) Iterator() *iterator.Iterator {
	return iterator.BuildIterator(search)
}

func (idx *BTreeIndex) Select(sel Selection) *iterator.Iterator {
	c := idx.buildConstraints(sel)
	searches := buildRangeSearches(c)
	iterators := make(iterator.Iterators, 0, len(searches))
	for _, search := range searches {
		iterators = append(iterators, *btreeRangeSearch{rangeSearch: search, idx: idx}.Iterator())
	}
	return iterator.BuildIterator(&iterators)
}

func (idx *BTreeIndex) SelectOne(sel Selection) (datum Datum) {
	iter := idx.Select(sel)
	if iter == nil || !iter.Next() {
		return
	}
	return iter.Value().(Datum)
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

type btreeFilter struct {
	indexType IndexType
	start     Datum
	idx       *BTreeIndex
}

func (filter btreeFilter) Each(accept iterator.Accept) {
	start := Node{kind: filter.indexType, datum: filter.start}
	var treeIter func(item btree.Item) bool
	switch filter.indexType {
	case IndexAVE:
		switch {
		case filter.start.A == 0:
			treeIter = func(item btree.Item) bool {
				node := item.(Node)
				if node.kind != IndexAVE {
					return false
				}
				if !accept(node.datum) {
					return false
				}
				return true
			}
		case filter.start.V == nil:
			treeIter = func(item btree.Item) bool {
				node := item.(Node)
				if node.kind != IndexAVE || node.datum.A != filter.start.A {
					return false
				}
				if !accept(node.datum) {
					return false
				}
				return true
			}
		case filter.start.E == 0:
			treeIter = func(item btree.Item) bool {
				node := item.(Node)
				if node.kind != IndexAVE || node.datum.A != filter.start.A || Compare(node.datum.V, filter.start.V) != 0 {
					return false
				}
				if !accept(node.datum) {
					return false
				}
				return true
			}
		}
	default:
		panic("TODO")
	}
	filter.idx.tree.AscendGreaterOrEqual(start, treeIter)
}

func (filter btreeFilter) Iterator() *iterator.Iterator {
	return iterator.BuildIterator(filter)
}

func (idx *BTreeIndex) Filter(typ IndexType, d Datum) *iterator.Iterator {
	return btreeFilter{idx: idx, indexType: typ, start: d}.Iterator()
}

func (idx *BTreeIndex) ResolveIdent(ident Ident) ID {
	return idx.idents[String(ident)]
}

func (idx *BTreeIndex) resolveESel(sel ESel) ids.Constraint {
	switch e := sel.(type) {
	case ID:
		return ids.Scalar(e)
	case LookupRef:
		panic("TODO")
	case Ident:
		id := idx.ResolveIdent(e)
		if id == 0 {
			return nil
		}
		return ids.Scalar(id)
	case ESet:
		r := make(ids.Set, len(e))
		for esel := range e {
			es := idx.resolveESel(esel).Iterator()
			for es.Next() {
				r[es.Value().(ID)] = Void{}
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
			id := idx.ResolveIdent(min)
			if id == 0 {
				return nil
			}
			return ids.Scalar(id)
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
			id := idx.ResolveIdent(max)
			if id == 0 {
				return nil
			}
			return ids.Scalar(id)
		case nil:
		default:
			panic("TODO")
		}
		return r
	case nil:
		return nil
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
	case Float:
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
			panic("TODO int")
		case Float:
			panic("TODO float")
		case Inst:
			panic("TODO inst")
		case Bool:
			panic("TODO bool")
		default:
			return matchesNoValue
		}
	default:
		return matchesNoValue
	}
}
