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

type op struct {
	e       ID
	a       ID
	v       Value
	retract bool
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

	// The real question: are request claims processed serially or with unspecified ordering?
	// The latter allows us flexibility in resolving claims about existing datums, but it
	// admits ambiguous outcomes unless we are able to refuse such inputs.
	// Ambiguous claim sets would seem to include:
	// * Multiple asserts of the same eav are irrelevant, possibly indicative of errors?
	// * An assert and a retract for the same eav
	// * An assert of an eav and a retract for the same ea
	//
	// First we resolve all attrs and reject non-attrs. We reject nil value asserts.
	// We resolve all e and v lookup refs and reject any missing.
	//
	// Then we resolve all identity attr e tempids. Note we reject the same tempid resolving to
	// multiple ids. Then we allocate ids for the remainder and populate all v tempids, rejecting
	// if there are any remaining (?).
	//
	// We could explicit expand ea retracts if we wanted to.
	// But regardless, we then reject incompatible assert/retracts.
	//
	// Then we apply the changes in whatever order we feel like. Note that a tree for each
	// index would allow them to be processed concurrently, probably a good runtime optimization.
	// Another level of obvious partitioning would be segregating the system and maybe schema datoms, the transaction
	// datums, and the user datums.
	claims := make([]*Claim, 0, len(request.Claims))
	for i := range request.Claims {
		claim := request.Claims[i]
		err = conn.resolveClaimRefs(txn.ID, &claim)
		if err != nil {
			break
		}
		claims = append(claims, &claim)
	}
	// Resolve entity tempids
	if err == nil {
		for _, claim := range claims {
			a := claim.A.(ID)
			attr := conn.idx.AttrByID(a)
			if attr.Unique == sys.AttrUniqueIdentity {
				tempid, ok := claim.E.(TempID)
				if ok {
					v, ok := claim.V.(Value)
					if !ok {
						err = ErrInvalidClaim
						break
					}
					iter := conn.idx.Filter(IndexAVE, Datum{A: a, V: v})
					if iter.Next() {
						d := iter.Value().(Datum)
						extant, ok := txn.NewIDs[tempid]
						if ok {
							if extant != d.E {
								err = ErrInvalidClaim
								break
							}
						} else {
							claim.E = d.E
							txn.NewIDs[tempid] = d.E
						}
					}
				}
			}
		}
	}
	// Allocate entity tempids
	if err == nil {
		for _, claim := range claims {
			tempid, ok := claim.E.(TempID)
			if !ok {
				continue
			}
			extant, ok := txn.NewIDs[tempid]
			if ok {
				claim.E = extant
			} else {
				id := conn.allocID()
				claim.E = id
				txn.NewIDs[tempid] = id
			}
		}
	}
	// Resolve value tempids
	if err == nil {
		for _, claim := range claims {
			tempid, ok := claim.V.(TempID)
			if !ok {
				continue
			}
			extant, ok := txn.NewIDs[tempid]
			if !ok {
				err = ErrInvalidClaim
				break
			}
			claim.V = extant
		}
	}
	if err == nil {
		ops := make([]op, 0, len(claims))
		// Expand ea retractions
		for _, claim := range claims {
			e := claim.E.(ID)
			a := claim.A.(ID)
			if claim.V != nil {
				ops = append(ops, op{e, a, claim.V.(Value), claim.Retract})
				continue
			}
			iter := conn.idx.Select(Selection{E: e, A: a})
			for iter.Next() {
				ops = append(ops, op{e, a, iter.Value().(Datum).V, true})
			}
		}
		// TODO validate op consistency
		for _, op := range ops {
			if !op.retract {
				_, err = newIdx.Assert(Datum{E: op.e, A: op.a, V: op.v})
			} else {
				err = newIdx.Retract(Datum{E: op.e, A: op.a, V: op.v})
			}
			if err != nil {
				// TODO specify the failed op
				break
			}
		}
	}
	if err == nil {
		_, err = newIdx.Assert(Datum{E: txn.ID, A: sys.TxAt, V: Inst(conn.clock.Now()), T: txn.ID})
	}
	if err != nil {
		conn.nextID = id
		txn.ID = ID(0)
		txn.NewIDs = nil
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

// resolveClaimRefs resolves the refs in the given claim
func (conn *BTreeConnection) resolveClaimRefs(txnID ID, claim *Claim) error {
	switch e := claim.E.(type) {
	case LookupRef:
		id := conn.idx.ResolveLookupRef(e)
		if id == 0 {
			return ErrInvalidClaim
		}
		claim.E = id
	case Ident:
		id := conn.idx.ResolveIdent(e)
		if id == 0 {
			return ErrInvalidClaim
		}
		claim.E = id
	case TxnID:
		claim.E = txnID
	case ID:
		if e == 0 {
			return ErrInvalidClaim
		}
	}
	switch a := claim.A.(type) {
	case LookupRef:
		id := conn.idx.ResolveLookupRef(a)
		if id == 0 {
			return ErrInvalidClaim
		}
		claim.A = id
	case Ident:
		id := conn.idx.ResolveIdent(a)
		if id == 0 {
			return ErrInvalidClaim
		}
		claim.A = id
	case ID:
		if a == 0 {
			return ErrInvalidClaim
		}
	}
	switch v := claim.V.(type) {
	case LookupRef:
		id := conn.idx.ResolveLookupRef(v)
		if id == 0 {
			return ErrInvalidClaim
		}
		claim.V = id
	case Ident:
		id := conn.idx.ResolveIdent(v)
		if id == 0 {
			return ErrInvalidClaim
		}
		claim.V = id
	case ID:
		if v == 0 {
			return ErrInvalidClaim
		}
	case nil:
		if !claim.Retract {
			return ErrInvalidClaim
		}
	}
	return nil
}
