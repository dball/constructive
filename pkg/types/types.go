package types

import (
	"errors"
	"time"
)

type Datum struct {
	E ID
	A ID
	V Value
	T ID
}

type ID uint64

type Value interface {
	IsValue()
}

type String string
type Int int
type Bool bool
type Inst time.Time

func (String) IsValue() {}
func (Int) IsValue()    {}
func (Bool) IsValue()   {}
func (Inst) IsValue()   {}
func (ID) IsValue()     {}

type Void struct{}

func D(e ID, a ID, v Value, t ID) Datum {
	return Datum{e, a, v, t}
}

type Attribute struct {
	ID          ID     `attr:"sys/db/id"`
	Name        String `attr:"sys/db/ident"`
	Type        ID     `attr:"sys/attr/type"`
	Cardinality ID     `attr:"sys/attr/cardinality"`
	Unique      ID     `attr:"sys/db/unique"`
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
	V Value
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

type Transaction struct {
	ID     ID
	NewIDs map[string]ID
	DB     Database
}

type Request struct {
	Claims []Claim
}

type Database interface {
	SelectSlice(index IndexType, selection Selection) []Datum
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
func (VSet) IsVSel()   {}
func (VRange) IsVSel() {}

type ReadError interface{}
type WriteError interface{}
type Result interface{}

type Connection interface {
	Read(func(db Database) Result) (Result, ReadError)
	Write(request Request) (Transaction, WriteError)
}

func ValueOf(x interface{}) (Value, error) {
	switch typed := x.(type) {
	case string:
		return String(typed), nil
	case int:
		return Int(typed), nil
	case bool:
		return Bool(typed), nil
	case time.Time:
		return Inst(typed), nil
	default:
		return nil, ErrInvalidValue
	}
}

var ErrInvalidValue error = errors.New("invalid value")
var ErrInvalidUserIdent error = errors.New("users may not assert sys attrs")
var ErrAttrTypeChange error = errors.New("attr types may not change")
var ErrAttrUniqueChange error = errors.New("attr uniqueness may not change")
var ErrInvalidAttrCardinality error = errors.New("attr cardinality must be one or many")
var ErrAttrCardinalityChange error = errors.New("attr cardinality may not change")
var ErrInvalidAttr error = errors.New("invalid datum attr")
