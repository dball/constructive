package destruct

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
)

func ParseAttrTag(tag string) (attr Attr) {
	parts := strings.Split(tag, ",")
	attr.Ident = Ident(parts[0])
	if len(parts) > 1 {
		switch parts[1] {
		case "identity":
			attr.Unique = sys.AttrUniqueIdentity
		case "unique":
			attr.Unique = sys.AttrUniqueValue
		}
	}
	return
}

var timeType = reflect.TypeOf(time.Time{})

func ParseAttrField(field reflect.StructField) (attr Attr) {
	tag, ok := field.Tag.Lookup("attr")
	if !ok {
		return
	}
	attr = ParseAttrTag(tag)
	if attr.Ident == sys.DbId {
		return
	}
	switch field.Type.Kind() {
	case reflect.Bool:
		attr.Type = sys.AttrTypeBool
	case reflect.Int:
		attr.Type = sys.AttrTypeInt
	case reflect.String:
		attr.Type = sys.AttrTypeString
	case reflect.Float64:
		attr.Type = sys.AttrTypeFloat
	case reflect.Struct:
		if timeType == field.Type {
			attr.Type = sys.AttrTypeInst
		} else {
			panic("TODO struct ref")
		}
	default:
		panic("TODO what even")
	}
	return
}

var symCount uint64

func Schema(typ reflect.Type) []Claim {
	n := typ.NumField()
	claims := make([]Claim, 0, n)
	for i := 0; i < n; i++ {
		attr := ParseAttrField(typ.Field(i))
		if attr.Ident == "" || attr.Ident == sys.DbId {
			continue
		}
		symCount++
		e := TempID(fmt.Sprintf("%d", symCount))
		claims = append(claims,
			Claim{E: e, A: sys.DbIdent, V: String(attr.Ident)},
			Claim{E: e, A: sys.AttrType, V: attr.Type},
		)
		if attr.Unique > 0 {
			claims = append(claims, Claim{E: e, A: sys.AttrUnique, V: attr.Unique})
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

func pluckFieldValue(attr Attr, refValue reflect.Value) Value {
	switch attr.Type {
	case sys.AttrTypeString:
		return String(refValue.String())
	case sys.AttrTypeBool:
		return Bool(refValue.Bool())
	case sys.AttrTypeInt:
		return Int(refValue.Int())
	case sys.AttrTypeRef:
		return ID(refValue.Uint())
	case sys.AttrTypeInst:
		panic("TODO inst")
	case sys.AttrTypeFloat:
		return Float(refValue.Float())
	}
	panic("TODO whatttt")
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
			attr := ParseAttrField(fieldType)
			if attr.Ident == "" {
				continue
			}
			fieldValue := reflect.ValueOf(x).Field(i)
			if attr.Ident == sys.DbId {
				switch fieldType.Type.Kind() {
				case reflect.Uint:
					id = ID(fieldValue.Uint())
				default:
					// TODO error?
				}
				continue
			}
			// TODO parseFieldValue?
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
			case reflect.Float64:
				value = Float(fieldValue.Float())
			default:
				// TODO error?
				continue
			}
			if attr.Unique != 0 && value.IsEmpty() {
				continue
			}
			xclaims = append(xclaims, Claim{E: ID(0), A: attr.Ident, V: value})
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
