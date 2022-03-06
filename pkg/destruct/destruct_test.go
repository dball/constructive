package destruct

import (
	"reflect"
	"testing"

	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
	"github.com/stretchr/testify/assert"
)

type Person struct {
	ID     uint   `attr:"sys/db/id"`
	Name   string `attr:"person/name"`
	Age    int    `attr:"person/age"`
	Active bool   `attr:"person/active"`
}

func Test_Schema(t *testing.T) {
	symCount = 0
	p := Person{}
	claims := Schema(reflect.TypeOf(p))
	expected := []Claim{
		{E: TempID("1"), A: sys.DbIdent, V: String("person/name")},
		{E: TempID("1"), A: sys.AttrType, V: sys.AttrTypeString},
		{E: TempID("2"), A: sys.DbIdent, V: String("person/age")},
		{E: TempID("2"), A: sys.AttrType, V: sys.AttrTypeInt},
		{E: TempID("3"), A: sys.DbIdent, V: String("person/active")},
		{E: TempID("3"), A: sys.AttrType, V: sys.AttrTypeBool},
	}
	assert.Equal(t, expected, claims)
}

func Test_Destruct(t *testing.T) {
	symCount = 0
	t.Run("identified person", func(t *testing.T) {
		p := Person{ID: 1, Name: "Donald", Age: 46, Active: true}
		claims := DestructOnlyData(p)
		expected := []Claim{
			{E: ID(1), A: Ident("person/name"), V: String("Donald")},
			{E: ID(1), A: Ident("person/age"), V: Int(46)},
			{E: ID(1), A: Ident("person/active"), V: Bool(true)},
		}
		assert.Equal(t, expected, claims)
	})
	t.Run("unidentified person", func(t *testing.T) {
		p := Person{Name: "Donald", Age: 46, Active: true}
		claims := DestructOnlyData(p)
		expected := []Claim{
			{E: TempID("1"), A: Ident("person/name"), V: String("Donald")},
			{E: TempID("1"), A: Ident("person/age"), V: Int(46)},
			{E: TempID("1"), A: Ident("person/active"), V: Bool(true)},
		}
		assert.Equal(t, expected, claims)
	})
}
