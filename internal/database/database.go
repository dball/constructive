package database

import (
	"github.com/dball/constructive/internal/index"
	"github.com/dball/constructive/internal/iterator"
	. "github.com/dball/constructive/pkg/types"
)

type BTreeDatabase struct {
	idx *index.BTreeIndex
}

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
