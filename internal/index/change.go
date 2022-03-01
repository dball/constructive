package index

import (
	"strings"

	. "github.com/dball/constructive/pkg/types"

	"github.com/dball/constructive/pkg/sys"
	"github.com/google/btree"
)

func (idx *BTreeIndex) Assert(d Datum) Datum {
	attr, ok := idx.attrs[d.A]
	if ok && !sys.ValidValue(attr.Typ, d.V) {
		panic("invalid types are super uncool")
	}
	ident, ok := idx.identNames[d.A]
	if ok && strings.HasPrefix(string(ident), "sys/") {
		panic("users may not assert sys idents")
	}
	// TODO call assertOne or assertMany depending on a's cardinality
	// TODO if a is unique, enforce uniqueness
	switch d.A {
	case sys.AttrType:
		v := d.V.(ID)
		// TODO validate attr type is a valid attr type
		attr, ok := idx.attrs[d.E]
		if ok {
			if attr.Typ != 0 && attr.Typ != v {
				panic("attr type changes are not cool")
			}
			attr.Typ = v
		} else {
			idx.attrs[d.E] = sys.Attr{Typ: v}
		}
	case sys.AttrUnique:
		v := d.V.(ID)
		unique := v == sys.AttrUniqueGlobal
		attr, ok := idx.attrs[d.E]
		if ok {
			if attr.Unique && !unique {
				panic("attr unique changes are not cool")
			}
			attr.Unique = unique
		} else {
			idx.attrs[d.E] = sys.Attr{Unique: unique}
		}
	case sys.AttrCardinality:
		v := d.V.(ID)
		if !(v == sys.AttrCardinalityOne || v == sys.AttrCardinalityMany) {
			panic("invalid cardinality value")
		}
		many := v == sys.AttrCardinalityMany
		attr, ok := idx.attrs[d.E]
		if ok {
			if attr.Many && !many {
				panic("attr cardinality change from many to one is not cool")
			}
			attr.Many = many
		} else {
			idx.attrs[d.E] = sys.Attr{Many: many}
		}
	case sys.DbIdent:
		ident := d.V.(String)
		idx.idents[ident] = d.E
		idx.identNames[d.E] = ident
	}
	extant, _ := idx.assertOne(d)
	return extant
}

// assertOne ensures a datum for an attribute of cardinality one exists in the index.
// If no datum for the entity and attribute already existed, this inserts one and returns
// an empty datum. If one already exists with the same value, this returns that datum and
// makes no changes to the index. Otherwise, this removes and returns that datum and inserts
// the given datum.
//
// This performs no validation.
func (idx *BTreeIndex) assertOne(d Datum) (d0 Datum, changed bool) {
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
