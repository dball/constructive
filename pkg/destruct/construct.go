package destruct

import (
	"reflect"

	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
)

type attr struct {
	id          ID
	ident       Ident
	typ         ID
	cardinality ID
}

func Construct(typ reflect.Type, database Database, id ID) interface{} {
	entity := reflect.New(typ)
	mutable := entity.Elem()
	n := typ.NumField()
	fieldIndexesByAttrName := make(map[Ident]int, n)
	for i := 0; i < n; i++ {
		field := typ.Field(i)
		attrName, ok := field.Tag.Lookup("attr")
		if !ok {
			continue
		}
		if attrName == sys.DbId {
			mutable.Field(i).SetUint(uint64(id))
			continue
		}
		fieldIndexesByAttrName[Ident(attrName)] = i
	}
	// TODO resolve attr ids by name - method on database
	// iter := database.Select(Selection{E: id})
	panic("TODO")
}
