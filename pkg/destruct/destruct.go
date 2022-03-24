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
			attr.Type = sys.AttrTypeRef
		}
	default:
		panic("Invalid attr field type")
	}
	return
}

var symCount uint64

func typeSchema(typ reflect.Type) []Claim {
	n := typ.NumField()
	claims := make([]Claim, 0, n)
	for i := 0; i < n; i++ {
		field := typ.Field(i)
		attr := ParseAttrField(field)
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
		if attr.Type == sys.AttrTypeRef {
			// TODO component should be the default, but we must allow foreign references
			claims = append(claims, Claim{E: e, A: sys.AttrRefType, V: sys.AttrRefTypeComponent})
			// TODO we need a types-that-have-been-schematized collection to prevent infinite cycles
			claims = append(claims, typeSchema(field.Type)...)
		}
	}
	return claims
}

// TODO remove destruct with schema option

func Destruct(xs ...interface{}) []Claim {
	return destruct(true, xs)
}

func DestructOnlyData(xs ...interface{}) []Claim {
	return destruct(false, xs)
}

func Schema(xs ...interface{}) []Claim {
	var claims []Claim
	var types []reflect.Type
	for _, x := range xs {
		typ := reflect.TypeOf(x)
		n := typ.NumField()
		if claims == nil {
			data := len(xs) * n
			claims = make([]Claim, 0, data)
		}
		done := false
		for _, t := range types {
			if typ == t {
				done = true
				break
			}
		}
		if !done {
			claims = append(claims, typeSchema(typ)...)
			types = append(types, typ)
		}
	}
	return claims
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
				claims = append(claims, typeSchema(typ)...)
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
			var vref VRef
			switch fieldType.Type.Kind() {
			case reflect.Bool:
				vref = Bool(fieldValue.Bool())
			case reflect.Int:
				vref = Int(fieldValue.Int())
			case reflect.String:
				vref = String(fieldValue.String())
			case reflect.Struct:
				v := fieldValue.Interface()
				switch typed := v.(type) {
				case time.Time:
					vref = Inst(typed)
				default:
					xxclaims := destruct(schema, []interface{}{typed})
					if len(xxclaims) == 0 {
						panic("TODO do we assign a tempid to the ref or what")
					}
					switch ee := xclaims[0].E.(type) {
					case ID:
						vref = ee
					case TempID:
						vref = ee
					default:
						panic("TODO struct ref has an unexpected e type")
					}
				}
			case reflect.Float64:
				vref = Float(fieldValue.Float())
			default:
				// TODO error?
				continue
			}
			v, ok := vref.(Value)
			if ok && attr.Unique != 0 && v.IsEmpty() {
				continue
			}
			xclaims = append(xclaims, Claim{E: ID(0), A: attr.Ident, V: vref})
		}
		if id != ID(0) {
			for i := range xclaims {
				xclaims[i].E = id
			}
		} else {
			symCount++
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
