package destruct

import (
	"reflect"

	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
)

func Destruct(x interface{}) []Claim {
	typ := reflect.TypeOf(x)
	n := typ.NumField()
	claims := make([]Claim, 0, n)
	var id ID
	for i := 0; i < n; i++ {
		fieldType := typ.Field(i)
		attr, ok := fieldType.Tag.Lookup("attr")
		if !ok {
			continue
		}
		fieldValue := reflect.ValueOf(x).Field(i)
		if attr == sys.DbId {
			switch fieldType.Type.Kind() {
			case reflect.Uint:
				id = ID(fieldValue.Uint())
			default:
				// TODO error?
			}
			continue
		}
		var value Value
		switch fieldType.Type.Kind() {
		case reflect.Bool:
			value = Bool(fieldValue.Bool())
		case reflect.Int:
			value = Int(fieldValue.Int())
		case reflect.String:
			value = String(fieldValue.String())
		case reflect.Struct:
			// TODO time
		default:
			// TODO error?
			continue
		}
		claims = append(claims, Claim{E: ID(0), A: Ident(attr), V: value})
	}
	if id != ID(0) {
		for i := range claims {
			claims[i].E = id
		}
	}
	// else assign a tempid?
	return claims
}
