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
	Name   string `attr:"person/name,unique"`
	UUID   string `attr:"person/uuid,identity"`
	Age    int    `attr:"person/age"`
	Active bool   `attr:"person/active"`
}

type Character struct {
	Name string `attr:"player/name,identity"`
	// TODO component ref attr attr to indicate existence ownership
	Focus Skill `attr:"player/focus"`
	// TODO map of values for set cardinality
	// TODO slice of values for list cardinality
}

type Skill struct {
	Name string  `attr:"skill/name"`
	Rank float64 `attr:"skill/rank"`
}

func Test_Schema(t *testing.T) {
	symCount = 0
	t.Run("single entity", func(t *testing.T) {
		p := Person{}
		claims := typeSchema(reflect.TypeOf(p))
		expected := []Claim{
			{E: TempID("1"), A: sys.DbIdent, V: String("person/name")},
			{E: TempID("1"), A: sys.AttrType, V: sys.AttrTypeString},
			{E: TempID("1"), A: sys.AttrUnique, V: sys.AttrUniqueValue},
			{E: TempID("2"), A: sys.DbIdent, V: String("person/uuid")},
			{E: TempID("2"), A: sys.AttrType, V: sys.AttrTypeString},
			{E: TempID("2"), A: sys.AttrUnique, V: sys.AttrUniqueIdentity},
			{E: TempID("3"), A: sys.DbIdent, V: String("person/age")},
			{E: TempID("3"), A: sys.AttrType, V: sys.AttrTypeInt},
			{E: TempID("4"), A: sys.DbIdent, V: String("person/active")},
			{E: TempID("4"), A: sys.AttrType, V: sys.AttrTypeBool},
		}
		assert.Equal(t, expected, claims)
	})
	symCount = 0
	t.Run("single reference", func(t *testing.T) {
		claims := Schema(Character{})
		expected := []Claim{
			{E: TempID("1"), A: sys.DbIdent, V: String("player/name")},
			{E: TempID("1"), A: sys.AttrType, V: sys.AttrTypeString},
			{E: TempID("1"), A: sys.AttrUnique, V: sys.AttrUniqueIdentity},
			{E: TempID("2"), A: sys.DbIdent, V: String("player/focus")},
			{E: TempID("2"), A: sys.AttrType, V: sys.AttrTypeRef},
			{E: TempID("2"), A: sys.AttrRefType, V: sys.AttrRefTypeComponent},
			{E: TempID("3"), A: sys.DbIdent, V: String("skill/name")},
			{E: TempID("3"), A: sys.AttrType, V: sys.AttrTypeString},
			{E: TempID("4"), A: sys.DbIdent, V: String("skill/rank")},
			{E: TempID("4"), A: sys.AttrType, V: sys.AttrTypeFloat},
		}
		assert.Equal(t, expected, claims)
	})
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
