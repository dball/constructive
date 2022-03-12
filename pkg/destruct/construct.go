package destruct

import (
	"reflect"
	"time"

	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
)

type attrField struct {
	attr  Attr
	index int
}

type attrStruct struct {
	idIndex int
	fields  map[ID]attrField
}

func parseAttrFields(refType reflect.Type, db Database) (attrs attrStruct) {
	n := refType.NumField()
	attrs.idIndex = -1
	attrs.fields = make(map[ID]attrField, n)
	for i := 0; i < n; i++ {
		ident := ParseAttrField(refType.Field(i)).Ident
		switch ident {
		case "":
		case sys.DbId:
			attrs.idIndex = i
		default:
			attr := db.AttrByIdent(ident)
			if attr.ID == 0 {
				continue
			}
			attrs.fields[attr.ID] = attrField{attr: attr, index: i}
		}
	}
	return
}

func Construct(ref interface{}, db Database, id ID) bool {
	refValue := reflect.ValueOf(ref).Elem()
	refType := refValue.Type()
	attrs := parseAttrFields(refType, db)
	if attrs.idIndex >= 0 {
		refValue.Field(attrs.idIndex).SetUint(uint64(id))
	}
	found := false
	iter := db.Select(Selection{E: id})
	for iter.Next() {
		found = true
		datum := iter.Value().(Datum)
		attrField, ok := attrs.fields[datum.A]
		if !ok {
			continue
		}
		switch attrField.attr.Type {
		case sys.AttrTypeString:
			refValue.Field(attrField.index).SetString(string(datum.V.(String)))
		case sys.AttrTypeInt:
			refValue.Field(attrField.index).SetInt(int64(datum.V.(Int)))
		case sys.AttrTypeBool:
			refValue.Field(attrField.index).SetBool(bool(datum.V.(Bool)))
		case sys.AttrTypeRef:
			refValue.Field(attrField.index).SetUint(uint64(datum.V.(ID)))
		case sys.AttrTypeFloat:
			refValue.Field(attrField.index).SetFloat(float64(datum.V.(Float)))
		case sys.AttrTypeInst:
			refValue.Field(attrField.index).Set(reflect.ValueOf(time.Time(datum.V.(Inst))))
		default:
			panic("construct2 all the types")
		}
	}
	return found
}

func Fetch(ref interface{}, db Database) bool {
	refValue := reflect.ValueOf(ref).Elem()
	refType := refValue.Type()
	attrs := parseAttrFields(refType, db)
	var id ID
	if attrs.idIndex >= 0 {
		id = ID(refValue.Field(attrs.idIndex).Uint())
	}
	for _, field := range attrs.fields {
		if field.attr.Unique == 0 {
			continue
		}
		value := pluckFieldValue(field.attr, refValue.Field(field.index))
		if value.IsEmpty() {
			continue
		}
		fieldID := db.ResolveLookupRef(LookupRef{A: field.attr.ID, V: value})
		if fieldID == 0 {
			continue
		}
		if id == 0 {
			id = fieldID
			continue
		}
		if fieldID == id {
			continue
		}
		return false
	}
	return Construct(ref, db, id)
}
