package sys

import (
	"time"

	. "github.com/dball/constructive/pkg/types"
)

const (
	DbId                = "sys/db/id"
	Tx                  = ID(1)
	TxAt                = ID(2)
	DbIdent             = ID(3)
	DbUnique            = ID(4)
	DbUniqueGlobal      = ID(5)
	AttrType            = ID(6)
	AttrTypeRef         = ID(7)
	AttrTypeString      = ID(8)
	AttrTypeInt         = ID(9)
	AttrTypeBool        = ID(10)
	AttrTypeInst        = ID(11)
	AttrCardinality     = ID(12)
	AttrCardinalityOne  = ID(13)
	AttrCardinalityMany = ID(14)

	FirstUserID = ID(0x100000)
)

var epoch time.Time

var Datums []Datum = []Datum{
	{E: DbIdent, A: DbIdent, V: String("sys/db/ident"), T: Tx},
	{E: DbIdent, A: AttrType, V: AttrTypeString, T: Tx},
	{E: DbIdent, A: DbUnique, V: DbUniqueGlobal, T: Tx},
	{E: DbUnique, A: DbIdent, V: String("sys/db/unique"), T: Tx},
	{E: DbUnique, A: AttrType, V: AttrTypeRef, T: Tx},
	{E: DbUniqueGlobal, A: DbIdent, V: String("sys/db/unique/global"), T: Tx},
	{E: TxAt, A: DbIdent, V: String("sys/tx/at"), T: Tx},
	{E: TxAt, A: AttrType, V: AttrTypeInst, T: Tx},
	{E: AttrType, A: DbIdent, V: String("sys/attr/type"), T: Tx},
	{E: AttrType, A: AttrType, V: AttrTypeRef, T: Tx},
	{E: AttrTypeRef, A: DbIdent, V: String("sys/attr/type/ref"), T: Tx},
	{E: AttrTypeString, A: DbIdent, V: String("sys/attr/type/string"), T: Tx},
	{E: AttrTypeInst, A: DbIdent, V: String("sys/attr/type/inst"), T: Tx},
	{E: AttrTypeInt, A: DbIdent, V: String("sys/attr/type/int"), T: Tx},
	{E: Tx, A: TxAt, V: Inst(epoch), T: Tx},
}
