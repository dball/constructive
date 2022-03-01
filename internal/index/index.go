package index

import (
	"github.com/dball/constructive/internal/compare"
	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"

	"github.com/google/btree"
)

func BuildIndex() BTreeIndex {
	return BTreeIndex{
		tree:       *btree.New(16),
		idents:     make(map[String]ID, 256),
		identNames: make(map[ID]String, 256),
		attrs:      make(map[ID]sys.Attr, 256),
	}
}

func (idx *BTreeIndex) InitSys() {
	for _, datum := range sys.Datums {
		idx.assertOne(datum)
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
}

type BTreeIndex struct {
	tree       btree.BTree
	idents     map[String]ID
	identNames map[ID]String
	attrs      map[ID]sys.Attr
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
