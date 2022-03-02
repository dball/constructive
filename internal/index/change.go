package index

import (
	. "github.com/dball/constructive/pkg/types"

	"github.com/dball/constructive/pkg/sys"
	"github.com/google/btree"
)

func (idx *BTreeIndex) Assert(assertion Datum) (conclusion Datum, err error) {
	attr, ok := idx.attrs[assertion.A]
	if ok && !sys.ValidValue(attr.Typ, assertion.V) {
		return conclusion, ErrInvalidValue
	}
	// TODO call assertOne or assertMany depending on a's cardinality
	// TODO if a is unique, enforce uniqueness
	switch assertion.A {
	case sys.AttrType:
		v := assertion.V.(ID)
		// TODO validate attr type is a valid attr type
		attr, ok := idx.attrs[assertion.E]
		if ok {
			if attr.Typ != 0 && attr.Typ != v {
				return conclusion, ErrAttrTypeChange
			}
			attr.Typ = v
		} else {
			idx.attrs[assertion.E] = sys.Attr{Typ: v}
		}
	case sys.AttrUnique:
		v := assertion.V.(ID)
		unique := sys.ValidUnique(v)
		attr, ok := idx.attrs[assertion.E]
		if ok {
			if attr.Unique && !unique {
				return conclusion, ErrAttrUniqueChange
			}
			attr.Unique = unique
		} else {
			idx.attrs[assertion.E] = sys.Attr{Unique: unique}
		}
	case sys.AttrCardinality:
		v := assertion.V.(ID)
		if !sys.ValidAttrCardinality(v) {
			return conclusion, ErrInvalidAttrCardinality
		}
		many := v == sys.AttrCardinalityMany
		attr, ok := idx.attrs[assertion.E]
		if ok {
			if attr.Many && !many {
				return conclusion, ErrAttrCardinalityChange
			}
			attr.Many = many
		} else {
			idx.attrs[assertion.E] = sys.Attr{Many: many}
		}
	case sys.DbIdent:
		ident := assertion.V.(String)
		if !sys.ValidUserIdent(ident) {
			return conclusion, ErrInvalidUserIdent
		}
		idx.idents[ident] = assertion.E
		idx.identNames[assertion.E] = ident
	}
	extant, _ := idx.assertCardinalityOne(assertion)
	return extant, nil
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
