package types

import (
	"errors"
	"time"

	"github.com/dball/constructive/internal/iterator"
)

type Datum struct {
	E ID
	A ID
	V Value
	T ID
}

type ID uint64

type Attr struct {
	ID    ID    `attr:"sys/db/id"`
	Ident Ident `attr:"sys/db/ident"`
	// TODO might it be useful to have a type for an Ident referred to by
	// an ID for unique attrs?
	Type        ID `attr:"sys/attr/type"`
	Cardinality ID `attr:"sys/attr/cardinality"`
	Unique      ID `attr:"sys/db/unique"`
}

type Value interface {
	IsEmpty() bool
}

type String string

// TODO should this be int64?
type Int int
type Bool bool
type Inst time.Time
type Float float64

func (x String) IsEmpty() bool { return string(x) == "" }
func (x Int) IsEmpty() bool    { return int64(x) == 0 }
func (x Bool) IsEmpty() bool   { return !bool(x) }
func (x Inst) IsEmpty() bool   { return time.Time(x).IsZero() }
func (x ID) IsEmpty() bool     { return uint64(x) == 0 }
func (x Float) IsEmpty() bool  { return float64(x) == 0 }

func Instant(s string) Inst {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return Inst(t.UTC())
}

type Void struct{}

func D(e ID, a ID, v Value, t ID) Datum {
	return Datum{e, a, v, t}
}

type LookupRef struct {
	A ARef
	V Value
}

type TempID string

type TxnID struct{}

type Ident string

type Claim struct {
	E EWriteRef
	A ARef
	V VRef
}

type EReadRef interface {
	IsEReadRef()
}

func (ID) IsEReadRef()        {}
func (LookupRef) IsEReadRef() {}
func (Ident) IsEReadRef()     {}

type EWriteRef interface {
	IsEWriteRef()
}

func (ID) IsEWriteRef()        {}
func (LookupRef) IsEWriteRef() {}
func (TempID) IsEWriteRef()    {}
func (TxnID) IsEWriteRef()     {}
func (Ident) IsEWriteRef()     {}

type ARef interface {
	IsARef()
}

func (ID) IsARef()        {}
func (LookupRef) IsARef() {}
func (Ident) IsARef()     {}

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

type Transaction struct {
	ID       ID
	NewIDs   map[TempID]ID
	Database Database
}

type Request struct {
	Claims []Claim
}

type Database interface {
	Select(selection Selection) *iterator.Iterator
	AttrByID(id ID) Attr
	AttrByIdent(ident Ident) Attr
	ResolveEReadRef(eref EReadRef) ID
	ResolveARef(aref ARef) ID
	ResolveLookupRef(ref LookupRef) ID
}

type IndexType int

const (
	IndexEAV IndexType = 1
	IndexAEV IndexType = 2
	IndexAVE IndexType = 3
)

type Selection struct {
	E ESel
	A ASel
	V VSel
}

type ESet map[ESel]Void

type ERange struct {
	Min EReadRef
	Max EReadRef
}

type ESel interface {
	IsESel()
}

func (ID) IsESel()        {}
func (LookupRef) IsESel() {}
func (Ident) IsESel()     {}
func (ESet) IsESel()      {}
func (ERange) IsESel()    {}

type ASet map[ASel]Void

type ARange struct {
	Min ARef
	Max ARef
}

type ASel interface {
	IsASel()
}

func (ID) IsASel()        {}
func (LookupRef) IsASel() {}
func (Ident) IsASel()     {}
func (ASet) IsASel()      {}
func (ARange) IsASel()    {}

type VSet map[VSel]Void

type VRange struct {
	Min Value
	Max Value
}

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

type Clock interface {
	Now() time.Time
}

type Connection interface {
	Read() Database
	Write(request Request) (Transaction, error)
	SetClock(clock Clock)
}

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

var ErrInvalidValue error = errors.New("invalid value")
var ErrInvalidUserIdent error = errors.New("users may not assert sys attrs")
var ErrAttrTypeChange error = errors.New("attr types may not change")
var ErrAttrUniqueChange error = errors.New("attr uniqueness may not change")
var ErrInvalidAttrCardinality error = errors.New("attr cardinality must be one or many")
var ErrAttrCardinalityChange error = errors.New("attr cardinality may not change")
var ErrInvalidAttr error = errors.New("invalid datum attr")
var ErrInvalidAttrUnique error = errors.New("attr uniqueness must be identity or value")
var ErrInvalidAttrType error = errors.New("attr type must be valid")
