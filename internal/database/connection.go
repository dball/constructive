package database

import (
	"sync"

	"github.com/dball/constructive/internal/index"
	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
)

func OpenConnection() Connection {
	idx := index.BuildIndex().InitSys()
	return &BTreeConnection{
		idx:    idx,
		nextID: sys.FirstUserID,
		clock:  SystemClock,
	}
}

type BTreeConnection struct {
	lock   sync.Mutex
	idx    *index.BTreeIndex
	nextID ID
	clock  Clock
}

func (conn *BTreeConnection) SetClock(clock Clock) {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	conn.clock = clock
}

// Write attempts to update the state of the connection to reflect the claims in the
// request. Claims assert or retract datums, though unlike datums, they may refer
// to values via Idents, LookupRefs, etc. Failure to resolve such references will
// result in rejected claims, as will claims about invalid attributes, values
// inconsistent with attribute types, etc. Empty values are not allowed in claims
// except for retractions, which specifically interpret a nil Value to retract
// all datums for that entity and attribute.
//
// Claims may specify entity and ref values as tempids. All claims about an tempid
// will resolve to an existing entity id if there is an identity unique attribute
// in the claims that corresponds to it. Otherwise, each distinct tempid is allocated
// a new id.
//
// Either all claims in a request are accepted, or all are rejected.
func (conn *BTreeConnection) Write(request Request) (txn Transaction, err error) {
	txn.NewIDs = make(map[TempID]ID)
	conn.lock.Lock()
	defer conn.lock.Unlock()
	newIdx := conn.idx.Clone()
	id := conn.nextID
	txn.ID = conn.allocID()
	// TODO since we want to apply claim sequentially, instead of these two passes, we should
	// resolve tempids first, probably with an unbounded iteration, then apply all claims.
	for _, claim := range request.Claims {
		v, ok := claim.V.(Value)
		if !ok {
			continue
		}
		a := conn.resolveARef(claim.A)
		// TODO if we have multiple claims for the same tempid, and the first one
		// is not for a unique identity attr, we must not allocate a new id until
		// we have checked the other claims for the same tempid.
		e := conn.resolveEWriteRef(txn, a, v, claim.E)
		if claim.Retract {
			if v != nil {
				err = newIdx.Retract(Datum{E: e, A: a, V: v})
			} else {
				iter := conn.idx.Select(Selection{E: e, A: a})
				for iter.Next() {
					err = newIdx.Retract(iter.Value().(Datum))
					if err != nil {
						break
					}
				}
			}
		} else {
			_, err = newIdx.Assert(Datum{E: e, A: a, V: v})
		}
		if err != nil {
			break
		}
	}
	if err != nil {
		for _, claim := range request.Claims {
			tempID, ok := claim.V.(TempID)
			if !ok {
				continue
			}
			v, ok := txn.NewIDs[tempID]
			if !ok {
				panic("claims with tempid values not used as entities in the request are invalid")
			}
			a := conn.resolveARef(claim.A)
			e := conn.resolveEWriteRef(txn, a, v, claim.E)
			_, err = newIdx.Assert(Datum{E: e, A: a, V: v})
		}
	}
	if err == nil {
		_, err = newIdx.Assert(Datum{E: txn.ID, A: sys.TxAt, V: Inst(conn.clock.Now()), T: txn.ID})
	}
	if err != nil {
		conn.nextID = id
		txn.ID = ID(0)
		txn.NewIDs = nil
		// TODO specify the failed claim
		return
	}
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

func (conn *BTreeConnection) resolveEWriteRef(txn Transaction, a ID, v Value, eref EWriteRef) ID {
	switch e := eref.(type) {
	case ID:
		return e
	case LookupRef:
		return conn.idx.ResolveLookupRef(e)
	case TempID:
		id, ok := txn.NewIDs[e]
		if ok {
			return id
		}
		attr := conn.idx.AttrByID(a)
		if attr.ID != 0 && attr.Unique == sys.AttrUniqueIdentity {
			iter := conn.idx.Filter(IndexAVE, Datum{A: a, V: v})
			if iter.Next() {
				d := iter.Value().(Datum)
				id = d.E
				txn.NewIDs[e] = id
				return id
			}
		}
		id = conn.allocID()
		txn.NewIDs[e] = id
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
