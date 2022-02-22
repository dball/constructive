package types

import "time"

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

func (String) IsValue()
func (Int) IsValue()
func (Bool) IsValue()
func (Inst) IsValue()
func (ID) IsValue()

type Void struct{}

func D(e ID, a ID, v Value, t ID) Datum {
	return Datum{e, a, v, t}
}

type Attribute struct {
	ID          ID
	Name        string
	Type        string
	Cardinality string
	Unique      string
}

type LookupRef struct {
	A ARef
	V Value
}

type TempID string

type TxnID struct{}

type Attr string

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
func (Attr) IsARef()      {}

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
	Undefined IndexType = iota
	EAV
	AEV
	AVE
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

func (ID) IsESel()
func (LookupRef) ISESel()
func (Ident) IsESel()
func (ESet) IsESel()
func (ERange) IsESel()

type ASet map[ASel]Void

type ARange struct {
	Min ARef
	Max ARef
}

type ASel interface {
	IsASel()
}

func (ID) IsASel()
func (LookupRef) IsASel()
func (Ident) IsASel()  {}
func (Attr) IsASel()   {}
func (ASet) IsASel()   {}
func (ARange) IsASel() {}

type VSet map[VSel]Void

type VRange struct {
	Min Value
	Max Value
}

type VSel interface {
	IsVSel()
}

func (ID) IsVSel()
func (String) IsVSel()
func (Int) IsVSel()
func (Bool) IsVSel()
func (Inst) IsVSel()
func (VSet) IsVSel()
func (VRange) IsVSel()

type ReadError interface{}
type WriteError interface{}
type Result interface{}

type Connection interface {
	Read(func(db Database) Result) (Result, ReadError)
	Write(request Request) (Transaction, WriteError)
}
