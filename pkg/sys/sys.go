package sys

import (
	"strings"
	"time"

	. "github.com/dball/constructive/pkg/types"
)

const (
	DbId                = "sys/db/id"
	Tx                  = ID(1)
	TxAt                = ID(2)
	DbIdent             = ID(3)
	AttrUnique          = ID(4)
	AttrUniqueIdentity  = ID(5)
	AttrUniqueValue     = ID(6)
	AttrType            = ID(7)
	AttrTypeRef         = ID(8)
	AttrTypeString      = ID(9)
	AttrTypeInt         = ID(10)
	AttrTypeBool        = ID(11)
	AttrTypeInst        = ID(12)
	AttrCardinality     = ID(13)
	AttrCardinalityOne  = ID(14)
	AttrCardinalityMany = ID(15)

	FirstUserID = ID(0x100000)
)

var epoch time.Time

var Datums []Datum = []Datum{
	{E: DbIdent, A: DbIdent, V: String("sys/db/ident"), T: Tx},
	{E: DbIdent, A: AttrType, V: AttrTypeString, T: Tx},
	{E: DbIdent, A: AttrUnique, V: AttrUniqueIdentity, T: Tx},
	{E: AttrUnique, A: DbIdent, V: String("sys/attr/unique"), T: Tx},
	{E: AttrUnique, A: AttrType, V: AttrTypeRef, T: Tx},
	{E: AttrUniqueIdentity, A: DbIdent, V: String("sys/attr/unique/identity"), T: Tx},
	{E: AttrUniqueValue, A: DbIdent, V: String("sys/attr/unique/value"), T: Tx},
	{E: TxAt, A: DbIdent, V: String("sys/tx/at"), T: Tx},
	{E: TxAt, A: AttrType, V: AttrTypeInst, T: Tx},
	{E: AttrType, A: DbIdent, V: String("sys/attr/type"), T: Tx},
	{E: AttrType, A: AttrType, V: AttrTypeRef, T: Tx},
	{E: AttrTypeRef, A: DbIdent, V: String("sys/attr/type/ref"), T: Tx},
	{E: AttrTypeString, A: DbIdent, V: String("sys/attr/type/string"), T: Tx},
	{E: AttrTypeInst, A: DbIdent, V: String("sys/attr/type/inst"), T: Tx},
	{E: AttrTypeInt, A: DbIdent, V: String("sys/attr/type/int"), T: Tx},
	{E: AttrCardinality, A: DbIdent, V: String("sys/attr/cardinality"), T: Tx},
	{E: AttrCardinality, A: AttrType, V: AttrTypeRef, T: Tx},
	{E: AttrCardinalityOne, A: DbIdent, V: String("sys/attr/cardinality/one"), T: Tx},
	{E: AttrCardinalityMany, A: DbIdent, V: String("sys/attr/cardinality/many"), T: Tx},
	{E: Tx, A: TxAt, V: Inst(epoch), T: Tx},
}

type Attr struct {
	Typ    ID
	Many   bool
	Unique ID
}

// This could be computed from Datums but this is smaller than the reducer code
var Attrs map[ID]Attr = map[ID]Attr{
	DbIdent:         {Typ: AttrTypeString, Unique: AttrUniqueIdentity},
	AttrUnique:      {Typ: AttrTypeRef},
	AttrType:        {Typ: AttrTypeRef},
	AttrCardinality: {Typ: AttrTypeRef},
	TxAt:            {Typ: AttrTypeInst},
}

func ValidValue(typ ID, value Value) (ok bool) {
	switch typ {
	case AttrTypeRef:
		_, ok = value.(ID)
	case AttrTypeString:
		_, ok = value.(String)
	case AttrTypeInt:
		_, ok = value.(Int)
	case AttrTypeBool:
		_, ok = value.(Bool)
	case AttrTypeInst:
		_, ok = value.(Inst)
	}
	return
}

func ValidUnique(id ID) bool {
	switch id {
	case AttrUniqueIdentity:
	case AttrUniqueValue:
	default:
		return false
	}
	return true
}

func ValidAttrType(id ID) bool {
	switch id {
	case AttrTypeRef:
	case AttrTypeString:
	case AttrTypeInt:
	case AttrTypeBool:
	case AttrTypeInst:
	default:
		return false
	}
	return true
}

func ValidAttrCardinality(id ID) bool {
	switch id {
	case AttrCardinalityOne:
	case AttrCardinalityMany:
	default:
		return false
	}
	return true
}

func ValidUserIdent(value String) bool {
	return !strings.HasPrefix(string(value), "sys/")
}
