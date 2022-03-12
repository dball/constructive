package sys

import (
	"strings"
	"time"

	. "github.com/dball/constructive/pkg/types"
)

const (
	DbId                = "sys/db/id"
	DbIdent             = ID(1)
	AttrType            = ID(2)
	AttrUnique          = ID(3)
	AttrCardinality     = ID(4)
	Tx                  = ID(5)
	TxAt                = ID(6)
	AttrUniqueIdentity  = ID(7)
	AttrUniqueValue     = ID(8)
	AttrCardinalityOne  = ID(9)
	AttrCardinalityMany = ID(10)
	AttrTypeRef         = ID(11)
	AttrTypeString      = ID(12)
	AttrTypeInt         = ID(13)
	AttrTypeBool        = ID(14)
	AttrTypeInst        = ID(15)
	AttrTypeFloat       = ID(16)
	FirstUserID         = ID(0x100000)
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
	{E: AttrTypeFloat, A: DbIdent, V: String("sys/attr/type/float"), T: Tx},
	{E: AttrCardinality, A: DbIdent, V: String("sys/attr/cardinality"), T: Tx},
	{E: AttrCardinality, A: AttrType, V: AttrTypeRef, T: Tx},
	{E: AttrCardinalityOne, A: DbIdent, V: String("sys/attr/cardinality/one"), T: Tx},
	{E: AttrCardinalityMany, A: DbIdent, V: String("sys/attr/cardinality/many"), T: Tx},
	{E: Tx, A: TxAt, V: Inst(epoch), T: Tx},
}

// This could be computed from Datums but this is smaller than the reducer code
var Attrs map[ID]Attr = map[ID]Attr{
	DbIdent:         {ID: DbIdent, Type: AttrTypeString, Unique: AttrUniqueIdentity, Ident: Ident("sys/db/ident")},
	AttrUnique:      {ID: AttrUnique, Type: AttrTypeRef, Ident: Ident("sys/attr/unique")},
	AttrType:        {ID: AttrType, Type: AttrTypeRef, Ident: Ident("sys/attr/type")},
	AttrCardinality: {ID: AttrCardinality, Type: AttrTypeRef, Ident: Ident("sys/attr/cardinality")},
	TxAt:            {ID: TxAt, Type: AttrTypeInst, Ident: Ident("sys/tx/at")},
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
	case AttrTypeFloat:
		_, ok = value.(Float)
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
	case AttrTypeFloat:
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
