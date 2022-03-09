package index

import (
	"github.com/dball/constructive/internal/compare"
	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"

	"github.com/google/btree"
)

func BuildIndex() *BTreeIndex {
	return &BTreeIndex{
		tree:       *btree.New(16),
		idents:     make(map[String]ID, 256),
		identNames: make(map[ID]String, 256),
		attrs:      make(map[ID]Attr, 256),
	}
}

func (idx *BTreeIndex) AttrByID(id ID) Attr {
	return idx.attrs[id]
}

func (idx *BTreeIndex) AttrByIdent(ident Ident) Attr {
	return idx.attrs[idx.idents[String(ident)]]
}

func (idx *BTreeIndex) InitSys() *BTreeIndex {
	for _, datum := range sys.Datums {
		idx.assertCardinalityOne(datum)
	}
	for id, attr := range sys.Attrs {
		idx.attrs[id] = attr
	}
	iter := func(item btree.Item) bool {
		node := item.(Node)
		if node.kind != IndexEAV {
			return false
		}
		datum := node.datum
		if datum.A == sys.DbIdent {
			value := datum.V.(String)
			idx.idents[value] = datum.E
			idx.identNames[datum.E] = value
		}
		return true
	}
	idx.tree.Ascend(iter)
	return idx
}

func (idx *BTreeIndex) Clone() *BTreeIndex {
	return &BTreeIndex{
		tree: *idx.tree.Clone(),
		// TODO these either need to be guaranteed to be backwards compatible or need to be copied or similar
		idents:     idx.idents,
		identNames: idx.identNames,
		attrs:      idx.attrs,
	}
}

type BTreeIndex struct {
	tree       btree.BTree
	idents     map[String]ID
	identNames map[ID]String
	attrs      map[ID]Attr
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

func (idx *BTreeIndex) ResolveEReadRef(eref EReadRef) (id ID) {
	switch e := eref.(type) {
	case ID:
		id = e
	case Ident:
		id = idx.ResolveIdent(e)
	case LookupRef:
		a := idx.ResolveARef(e.A)
		v := VSelValue(e.V)
		iter := idx.Select(Selection{A: a, V: v})
		if iter.Next() {
			datum := iter.Value().(Datum)
			id = datum.E
		}
	}
	return
}

func (idx *BTreeIndex) ResolveARef(aref ARef) (id ID) {
	switch a := aref.(type) {
	case ID:
		id = idx.ResolveEReadRef(a)
	case Ident:
		id = idx.ResolveEReadRef(a)
	case LookupRef:
		id = idx.ResolveEReadRef(a)
	}
	return
}
