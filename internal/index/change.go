package index

import (
	. "github.com/dball/constructive/pkg/types"

	"github.com/dball/constructive/pkg/sys"
	"github.com/google/btree"
)

func (idx *BTreeIndex) Assert(assertion Datum) (conclusion Datum, err error) {
	if assertion.E == 0 {
		err = ErrInvalidValue
		return
	}
	if assertion.A == 0 {
		err = ErrInvalidValue
		return
	}
	// TODO validate T or nay
	attr, ok := idx.attrs[assertion.A]
	if !ok {
		err = ErrInvalidAttr
		return
	}
	if ok && !sys.ValidValue(attr.Type, assertion.V) {
		err = ErrInvalidValue
		return
	}
	if attr.Unique != 0 {
		iter := idx.Filter(IndexAVE, Datum{A: assertion.A, V: assertion.V})
		if iter.Next() {
			extant := iter.Value().(Datum)
			if extant.E == assertion.E {
				conclusion = extant
			} else {
				err = ErrInvalidValue
			}
			return
		}
	}
	switch assertion.A {
	case sys.AttrType:
		v := assertion.V.(ID)
		// TODO validate attr type is a valid attr type
		// TODO populate the Attr ident
		attr, ok := idx.attrs[assertion.E]
		if ok {
			if attr.Type != 0 && attr.Type != v {
				err = ErrAttrTypeChange
				return
			}
			attr.Type = v
			idx.attrs[assertion.E] = attr
		} else {
			idx.attrs[assertion.E] = Attr{ID: assertion.E, Type: v}
		}
	case sys.AttrUnique:
		v := assertion.V.(ID)
		if !sys.ValidUnique(v) {
			err = ErrInvalidAttrUnique
			return
		}
		attr, ok := idx.attrs[assertion.E]
		if ok {
			if attr.Unique != 0 && attr.Unique != v {
				err = ErrAttrUniqueChange
				return
			}
			attr.Unique = v
			idx.attrs[assertion.E] = attr
		} else {
			idx.attrs[assertion.E] = Attr{ID: assertion.E, Unique: v}
		}
	case sys.AttrCardinality:
		v := assertion.V.(ID)
		if !sys.ValidAttrCardinality(v) {
			err = ErrInvalidAttrCardinality
			return
		}
		attr, ok := idx.attrs[assertion.E]
		if ok {
			if attr.Cardinality != 0 && attr.Cardinality != v {
				err = ErrAttrCardinalityChange
				return
			}
			attr.Cardinality = v
			idx.attrs[assertion.E] = attr
		} else {
			idx.attrs[assertion.E] = Attr{ID: assertion.E, Cardinality: v}
		}
	case sys.DbIdent:
		ident := assertion.V.(String)
		if !sys.ValidUserIdent(ident) {
			err = ErrInvalidUserIdent
			return
		}
		idx.idents[ident] = assertion.E
		idx.identNames[assertion.E] = ident
	}
	if attr.Cardinality == sys.AttrCardinalityMany {
		conclusion, _ = idx.assertCardinalityMany(assertion)
	} else {
		conclusion, _ = idx.assertCardinalityOne(assertion)
	}
	return
}

// assertCardinalityOne ensures a datum for an attribute of cardinality one exists in the index.
// If no datum for the entity and attribute already existed, this inserts one and returns
// an empty datum. If one already exists with the same value, this returns that datum and
// makes no changes to the index. Otherwise, this removes and returns that datum and inserts
// the given datum.
//
// This performs no validation.
func (idx *BTreeIndex) assertCardinalityOne(d Datum) (d0 Datum, changed bool) {
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
		return Datum{}, true
	case Compare(d.V, extant.datum.V) == 0:
		return extant.datum, false
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
		return extant.datum, true
	}
}

func (idx *BTreeIndex) assertCardinalityMany(d Datum) (d0 Datum, changed bool) {
	floor := Datum{E: d.E, A: d.A, V: d.V}
	var extant Node
	idx.tree.AscendGreaterOrEqual(Node{IndexEAV, floor}, func(item btree.Item) bool {
		node := item.(Node)
		if node.kind != IndexEAV {
			return false
		}
		datum := node.datum
		if datum.E > d.E || datum.A > d.A {
			return false
		}
		switch Compare(datum.V, d.V) {
		case 0:
			extant = node
			return true
		case 1:
			return false
		default:
			return true
		}
	})
	if extant.datum.E != 0 {
		return extant.datum, false
	}
	node := Node{IndexEAV, d}
	idx.tree.ReplaceOrInsert(node)
	node.kind = IndexAEV
	idx.tree.ReplaceOrInsert(node)
	// TODO only for unique a
	node.kind = IndexAVE
	idx.tree.ReplaceOrInsert(node)
	return Datum{}, true
}
