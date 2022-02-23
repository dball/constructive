package destruct

import (
	"testing"

	. "github.com/dball/constructive/pkg/types"
	"github.com/stretchr/testify/assert"
)

type Person struct {
	ID     uint   `attr:"sys/db/id"`
	Name   string `attr:"person/name"`
	Age    int    `attr:"person/age"`
	Active bool   `attr:"person/active"`
}

func Test_Destruct(t *testing.T) {
	p := Person{ID: 1, Name: "Donald", Age: 46, Active: true}
	claims := Destruct(p)
	expected := []Claim{
		{E: ID(1), A: Ident("person/name"), V: String("Donald")},
		{E: ID(1), A: Ident("person/age"), V: Int(46)},
		{E: ID(1), A: Ident("person/active"), V: Bool(true)},
	}
	assert.Equal(t, expected, claims)
}
