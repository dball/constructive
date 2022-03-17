package database

import (
	"github.com/dball/constructive/internal/index"
	"github.com/dball/constructive/internal/iterator"
	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
)

type BTreeDatabase struct {
	idx *index.BTreeIndex
}

var _ Database = &BTreeDatabase{}

func (db *BTreeDatabase) Select(selection Selection) *iterator.Iterator {
	return db.idx.Select(selection)
}

func (db *BTreeDatabase) AttrByID(id ID) Attr {
	return db.idx.AttrByID(id)
}

func (db *BTreeDatabase) AttrByIdent(ident Ident) Attr {
	return db.idx.AttrByIdent(ident)
}

func (db *BTreeDatabase) ResolveEReadRef(eref EReadRef) (id ID) {
	return db.idx.ResolveEReadRef(eref)
}

func (db *BTreeDatabase) ResolveARef(aref ARef) (id ID) {
	return db.idx.ResolveARef(aref)
}

func (db *BTreeDatabase) ResolveLookupRef(ref LookupRef) (id ID) {
	return db.idx.ResolveLookupRef(ref)
}

func (db *BTreeDatabase) Dump() interface{} {
	eavs := map[ID]map[Ident]interface{}{}
	var e ID
	iter := db.idx.Select(Selection{})
	for iter.Next() {
		datum := iter.Value().(Datum)
		attr := db.AttrByID(datum.A)
		if datum.E != e {
			e = datum.E
			eavs[datum.E] = map[Ident]interface{}{}
		}
		var v interface{}
		if attr.Cardinality != sys.AttrCardinalityMany {
			v = datum.V
		} else {
			vs, ok := eavs[e][attr.Ident].([]Value)
			if !ok {
				v = []Value{datum.V}
			} else {
				v = append(vs, datum.V)
			}
		}
		eavs[e][attr.Ident] = v
	}
	return eavs
}
