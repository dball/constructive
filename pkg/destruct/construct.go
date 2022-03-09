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

func Construct(typ reflect.Type, db Database, id ID) interface{} {
	entity := reflect.New(typ)
	mutable := entity.Elem()
	n := typ.NumField()
	attrFields := make(map[ID]attrField)
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
		attr := db.AttrByIdent(Ident(attrName))
		if attr.ID == 0 {
			continue
		}
		attrFields[attr.ID] = attrField{attr: attr, index: i}
	}
	found := false
	iter := db.Select(Selection{E: id})
	for iter.Next() {
		found = true
		datum := iter.Value().(Datum)
		attrField, ok := attrFields[datum.A]
		if !ok {
			continue
		}
		switch attrField.attr.Type {
		case sys.AttrTypeString:
			mutable.Field(attrField.index).SetString(string(datum.V.(String)))
		}
	}
	if !found {
		// TODO or would we prefer an empty value, probably with no id?
		return nil
	}
	return entity.Interface()
}

func Construct2(ref interface{}, db Database, id ID) bool {
	refValue := reflect.ValueOf(ref)
	refType := reflect.TypeOf(ref)
	mutable := refValue.Elem()
	n := refType.NumField()
	attrFields := make(map[ID]attrField)
	for i := 0; i < n; i++ {
		field := refType.Field(i)
		attrName, ok := field.Tag.Lookup("attr")
		if !ok {
			continue
		}
		if attrName == sys.DbId {
			mutable.Field(i).SetUint(uint64(id))
			continue
		}
		attr := db.AttrByIdent(Ident(attrName))
		if attr.ID == 0 {
			continue
		}
		attrFields[attr.ID] = attrField{attr: attr, index: i}
	}
	found := false
	iter := db.Select(Selection{E: id})
	for iter.Next() {
		found = true
		datum := iter.Value().(Datum)
		attrField, ok := attrFields[datum.A]
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
	panic("TODO")
	// refValue := reflect.ValueOf(ref)
	// refType := reflect.TypeOf(ref)
	// find sys/db/id and any unique attr fields
}
