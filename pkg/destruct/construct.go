package destruct

import (
	"reflect"

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

func Construct2(ref interface{}, db Database, id ID) bool {
	refValue := reflect.ValueOf(ref)
	refType := reflect.TypeOf(ref)
	mutable := refValue.Elem()
	attrs := parseAttrFields(refType, db)
	if attrs.idIndex >= 0 {
		mutable.Field(attrs.idIndex).SetUint(uint64(id))
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
			mutable.Field(attrField.index).SetString(string(datum.V.(String)))
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
	var lookup LookupRef
	for _, field := range attrs.fields {
		if field.attr.Unique == 0 {
			continue
		}
		value := pluckFieldValue(field.attr, refValue.Field(field.index))
		if value.IsEmpty() {
			continue
		}
		if lookup.A != nil {
			return false
		}
		lookup.A = field.attr.ID
		lookup.V = value
	}
	switch id {
	case 0:
		switch lookup.A {
		case ID(0):
			return false
		default:
			id = db.ResolveLookupRef(lookup)
		}
	default:
		switch lookup.A {
		case ID(0):
		default:
			return false
		}
	}
	return Construct2(ref, db, id)
}
