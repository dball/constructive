package types

import (
	"errors"
	"time"

	"github.com/dball/constructive/internal/iterator"
)

// Datum is the fundamental data model.
type Datum struct {
	// E is the entity id.
	E ID
	// A is the attribute id.
	A ID
	// V is the value.
	V Value
	// T is the transaction id.
	T ID
}

// IDs are issued by the system and never reused.
type ID uint64

// Attr is a convenient represention of the attributes of an attribute.
type Attr struct {
	// ID is the internal identifier of an attribute.
	ID ID `attr:"sys/db/id"`
	// Ident is the public identifier of an attribute.
	Ident Ident `attr:"sys/db/ident"`
	// TODO might it be useful to have a type for an Ident referred to by
	// an ID for unique attrs?
	// Type specifies the type of values to which the attribute refers.
	Type ID `attr:"sys/attr/type"`
	// Cardinality specifies the number of values an attribute may have on a given entity.
	Cardinality ID `attr:"sys/attr/cardinality"`
	// Unique specifies the uniqueness of the attribute's value.
	Unique ID `attr:"sys/db/unique"`
}

// Value is an immutable scalar.
type Value interface {
	IsEmpty() bool
}

// String is a string.
type String string

// TODO should this be int64?
// Int is a signed integer.
type Int int

// Bool is a boolean.
type Bool bool

// Inst is a instant in time.
type Inst time.Time

// Float is a floating-point number.
type Float float64

func (x String) IsEmpty() bool { return string(x) == "" }
func (x Int) IsEmpty() bool    { return int64(x) == 0 }
func (x Bool) IsEmpty() bool   { return !bool(x) }
func (x Inst) IsEmpty() bool   { return time.Time(x).IsZero() }
func (x ID) IsEmpty() bool     { return uint64(x) == 0 }
func (x Float) IsEmpty() bool  { return float64(x) == 0 }

// Instant is a convenience function to build an inst from an RFC3339 string. This
// will panic if the given string is not valid, and is intended only for constant
// literal values in source code.
func Instant(s string) Inst {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return Inst(t.UTC())
}

// Void is an empty struct, used for sets.
type Void struct{}

// D is a convenience function for building a datum.
func D(e ID, a ID, v Value, t ID) Datum {
	return Datum{e, a, v, t}
}

// LookupRef is a combination of a unique attribute reference and a value, which may
// resolve to an entity id in a database and can be used in place of an id in certain
// data structures.
type LookupRef struct {
	// A identifies the unique attribute.
	A ARef
	// V identifies the unique value.
	V Value
}

// TempID is a value that will be resolved to a system id when a claim is transacted.
type TempID string

// TxnID is a value that will be resolved to the transaction id when a claim is transacted.
type TxnID struct{}

// Ident is a globally unique system identifier for an entity, generally used for attributes.
type Ident string

// Claim is an assertion of or retraction of a datum.
type Claim struct {
	E       EWriteRef
	A       ARef
	V       VRef
	Retract bool
}

// EReadRef is a value that may resolve to a system id when reading from the database.
type EReadRef interface {
	IsEReadRef()
}

func (ID) IsEReadRef()        {}
func (LookupRef) IsEReadRef() {}
func (Ident) IsEReadRef()     {}

// EWriteRef is a value that may resolve to a system id when writing to the database.
type EWriteRef interface {
	IsEWriteRef()
}

func (ID) IsEWriteRef()        {}
func (LookupRef) IsEWriteRef() {}
func (TempID) IsEWriteRef()    {}
func (TxnID) IsEWriteRef()     {}
func (Ident) IsEWriteRef()     {}

// ARef is a value that may resolve to an attribute id within a database.
type ARef interface {
	IsARef()
}

func (ID) IsARef()        {}
func (LookupRef) IsARef() {}
func (Ident) IsARef()     {}

// VRef is a value that may resolve to a datum value when writing to the database.
type VRef interface {
	IsVRef()
}

func (ID) IsVRef()     {}
func (String) IsVRef() {}
func (Int) IsVRef()    {}
func (Bool) IsVRef()   {}
func (Inst) IsVRef()   {}
func (Float) IsVRef()  {}
func (TempID) IsVRef() {}

// TODO LookupRef as another vref type? How about Ident?

// Transaction is the result of successfully applying a request to the database.
type Transaction struct {
	// ID is the transaction's id.
	ID ID
	// NewIDs maps the TempIDs in the transacted claims to their resolved or assigned durable ids.
	NewIDs map[TempID]ID
	// Database is a snapshot of the database immediately after accepting the claims.
	Database Database
}

// Request is a request to record claims in the database.
type Request struct {
	Claims []Claim
}

// Database is a read-only, stable snapshot of the database.
type Database interface {
	Select(selection Selection) *iterator.Iterator
	AttrByID(id ID) Attr
	AttrByIdent(ident Ident) Attr
	ResolveEReadRef(eref EReadRef) ID
	ResolveARef(aref ARef) ID
	ResolveLookupRef(ref LookupRef) ID
	Dump() interface{}
}

// IndexType is the type of index being used to store or query datums.
type IndexType int

const (
	// IndexEAV indexes datums by entity id, then attribute id, then value.
	IndexEAV IndexType = 1
	// IndexAEV indexes datums by attribute id, then entity id, then value.
	IndexAEV IndexType = 2
	// IndexAVE indexes datums by unique attribute id, then value, then entity. Non-unique attributes do not appear in this index.
	IndexAVE IndexType = 3
	// TODO IndexVAE for ref types would give us backwards references, almost certainly good and useful.
)

// Selection is a simple database query, where the constraints are scalars, scalar ranges, or sets of constraints.
type Selection struct {
	// E constrains the entity ids.
	E ESel
	// A constrains the attribute ids.
	A ASel
	// V constrains the values.
	V VSel
}

// ESet is a set of entity id constraints.
type ESet map[ESel]Void

// ERange is a range of entity ids, inclusive at both ends.
// TODO exclusive must be possible.
type ERange struct {
	Min EReadRef
	Max EReadRef
}

// ESel is an entity id constraint.
type ESel interface {
	IsESel()
}

func (ID) IsESel()        {}
func (LookupRef) IsESel() {}
func (Ident) IsESel()     {}
func (ESet) IsESel()      {}
func (ERange) IsESel()    {}

// ASet is a set of attribute id constraints.
type ASet map[ASel]Void

// ARange is a range of attribute ids, inclusive at both ends.
// TODO exclusive must be possible.
type ARange struct {
	Min ARef
	Max ARef
}

// ASel is an attribute id constraint.
type ASel interface {
	IsASel()
}

func (ID) IsASel()        {}
func (LookupRef) IsASel() {}
func (Ident) IsASel()     {}
func (ASet) IsASel()      {}
func (ARange) IsASel()    {}

// VSet is a set of value constraints.
type VSet map[VSel]Void

// VRange is a range of values, inclusive at both ends.
// TODO exclusive must be possible even more than ARange and VRange especially since Float is not enumerable.
type VRange struct {
	Min Value
	Max Value
}

// VSel is a value constraint.
type VSel interface {
	IsVSel()
}

func (ID) IsVSel()     {}
func (String) IsVSel() {}
func (Int) IsVSel()    {}
func (Bool) IsVSel()   {}
func (Inst) IsVSel()   {}
func (Float) IsVSel()  {}
func (VSet) IsVSel()   {}
func (VRange) IsVSel() {}

// Clock is a source of the current time.
type Clock interface {
	Now() time.Time
}

// Connection is a writable database.
type Connection interface {
	Read() Database
	Write(request Request) (Transaction, error)
	SetClock(clock Clock)
}

// ValueOf converts a value to a Value, if possible.
func ValueOf(x interface{}) (Value, error) {
	switch typed := x.(type) {
	case string:
		return String(typed), nil
	case int64:
		return Int(typed), nil
	case bool:
		return Bool(typed), nil
	case time.Time:
		return Inst(typed), nil
	case uint64:
		return ID(typed), nil
	case float64:
		return Float(typed), nil
	default:
		return nil, ErrInvalidValue
	}
}

// VSelValue converts a Value to a VSel.
// TODO could we just typecast and ignore panic by construction?
func VSelValue(v Value) (vsel VSel) {
	switch typed := v.(type) {
	case String:
		vsel = typed
	case Int:
		vsel = typed
	case Bool:
		vsel = typed
	case Inst:
		vsel = typed
	case ID:
		vsel = typed
	}
	return
}

//// The system runtime errors.

var ErrInvalidValue error = errors.New("invalid value")
var ErrInvalidUserIdent error = errors.New("users may not assert sys attrs")
var ErrAttrTypeChange error = errors.New("attr types may not change")
var ErrAttrUniqueChange error = errors.New("attr uniqueness may not change")
var ErrInvalidAttrCardinality error = errors.New("attr cardinality must be one or many")
var ErrAttrCardinalityChange error = errors.New("attr cardinality may not change")
var ErrInvalidAttr error = errors.New("invalid datum attr")
var ErrInvalidAttrUnique error = errors.New("attr uniqueness must be identity or value")
var ErrInvalidAttrType error = errors.New("attr type must be valid")
