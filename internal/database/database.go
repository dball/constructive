package database

import (
	"sync"

	"github.com/dball/constructive/internal/index"
	"github.com/dball/constructive/internal/iterator"
	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
)

func OpenConnection() Connection {
	idx := index.BuildIndex().InitSys()
	return &BTreeConnection{
		idx:    idx,
		nextID: sys.FirstUserID,
	}
}

type BTreeDatabase struct {
	idx *index.BTreeIndex
}

func (db *BTreeDatabase) Select(selection Selection) *iterator.Iterator {
	return db.idx.Select(selection)
}

type BTreeConnection struct {
	lock   sync.Mutex
	idx    *index.BTreeIndex
	nextID ID
}

func (conn *BTreeConnection) Write(request Request) (txn Transaction, err error) {
	txn.NewIDs = make(map[TempID]ID)
	conn.lock.Lock()
	defer conn.lock.Unlock()
	newIdx := conn.idx.Clone()
	id := conn.nextID
	txn.ID = conn.allocID()
	for _, claim := range request.Claims {
		datum := Datum{
			E: conn.resolveEWriteRef(txn, claim.E),
			A: conn.resolveARef(claim.A),
			V: claim.V,
		}
		_, err = newIdx.Assert(datum)
		if err != nil {
			conn.nextID = id
			txn.ID = ID(0)
			txn.NewIDs = nil
			// TODO specify the failed claim
			return
		}
	}
	// TODO assert the system txn datums
	conn.idx = newIdx
	txn.Database = conn.Read()
	return
}

func (conn *BTreeConnection) Read() Database {
	return &BTreeDatabase{idx: conn.idx}
}

func (conn *BTreeConnection) allocID() (id ID) {
	id = conn.nextID
	if id == 0xffffffffffffffff {
		panic("id-space-exhausted")
	}
	conn.nextID++
	return
}

func (conn *BTreeConnection) resolveEWriteRef(txn Transaction, eref EWriteRef) ID {
	switch e := eref.(type) {
	case ID:
		return e
	case LookupRef:
		panic("TODO")
	case TempID:
		id, ok := txn.NewIDs[e]
		if !ok {
			id = conn.allocID()
			txn.NewIDs[e] = id
		}
		return id
	case TxnID:
		return txn.ID
	case Ident:
		return conn.idx.ResolveIdent(e)
	}
	return ID(0)
}

func (conn *BTreeConnection) resolveARef(aref ARef) ID {
	switch a := aref.(type) {
	case ID:
		return a
	case LookupRef:
		panic("TODO")
	case Ident:
		return conn.idx.ResolveIdent(a)
	}
	return ID(0)
}

/*
func (idx *BTreeIndex) Lookup(ref LookupRef) ID {
	var A ID
	switch a := ref.A.(type) {
	case ID:
		A = a
	case LookupRef:
		A = idx.Lookup(a)
	case Ident:
		c := Constraints{
			A: ,
		}
	}
}

func (idx *BTreeIndex) Resolve(sel Selection) (c Constraints) {
	switch e := sel.E.(type) {
	case ID:
		c.E = IDSet{e: void}
		return c
	case LookupRef:

	case Ident:
	case ESet:
	case ERange:
	default:
	}
	switch sel.A.(type) {
	case ID:
	case LookupRef:
	case Ident:
	case Attr:
	case ASet:
	case ARange:
	default:
	}
	panic("TODO")
}
*/
