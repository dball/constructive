package destruct

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
)

var symCount uint64

func Schema(typ reflect.Type) []Claim {
	n := typ.NumField()
	claims := make([]Claim, 0, n)
	for i := 0; i < n; i++ {
		field := typ.Field(i)
		attrTag, ok := field.Tag.Lookup("attr")
		parts := strings.Split(attrTag, ",")
		attrIdent := parts[0]
		var attrUnique ID
		if len(parts) > 1 {
			switch parts[1] {
			case "identity":
				attrUnique = sys.AttrUniqueIdentity
			case "unique":
				attrUnique = sys.AttrUniqueValue
			}
		}
		if !ok {
			continue
		}
		if attrTag == sys.DbId {
			continue
		}
		var attrType ID
		switch field.Type.Kind() {
		case reflect.Bool:
			attrType = sys.AttrTypeBool
		case reflect.Int:
			attrType = sys.AttrTypeInt
		case reflect.String:
			attrType = sys.AttrTypeString
		case reflect.Struct:
			panic("TODO time")
		default:
			panic("TODO what even")
		}
		symCount++
		e := TempID(fmt.Sprintf("%d", symCount))
		claims = append(claims,
			Claim{E: e, A: sys.DbIdent, V: String(attrIdent)},
			Claim{E: e, A: sys.AttrType, V: attrType},
		)
		if attrUnique > 0 {
			claims = append(claims, Claim{E: e, A: sys.AttrUnique, V: attrUnique})
		}
	}
	return claims
}

func Destruct(xs ...interface{}) []Claim {
	return destruct(true, xs)
}

func DestructOnlyData(xs ...interface{}) []Claim {
	return destruct(false, xs)
}

func destruct(schema bool, xs []interface{}) []Claim {
	var claims []Claim
	var types []reflect.Type
	for _, x := range xs {
		typ := reflect.TypeOf(x)
		n := typ.NumField()
		if claims == nil {
			data := len(xs) * n
			if schema {
				data += n
			}
			claims = make([]Claim, 0, data)
		}
		if schema {
			done := false
			for _, t := range types {
				if typ == t {
					done = true
					break
				}
			}
			if !done {
				claims = append(claims, Schema(typ)...)
				types = append(types, typ)
			}
		}
		var id ID
		xclaims := make([]Claim, 0, n)
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
			xclaims = append(xclaims, Claim{E: ID(0), A: Ident(attr), V: value})
		}
		if id != ID(0) {
			for i := range xclaims {
				xclaims[i].E = id
			}
		} else {
			symCount++
			fmt.Println("assign symcount", symCount)
			// TODO note we could use a different TempID type here and keep the whole string domain available to our callers
			tempID := TempID(fmt.Sprintf("%d", symCount))
			for i := range xclaims {
				xclaims[i].E = tempID
			}
		}
		claims = append(claims, xclaims...)
	}
	return claims
}
