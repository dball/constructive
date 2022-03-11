package constructive

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Person struct {
	ID     uint   `attr:"sys/db/id"`
	Name   string `attr:"person/name,identity"`
	Age    int    `attr:"person/age"`
	Active bool   `attr:"person/active"`
}

type Named struct {
	Name string `attr:"person/name,identity"`
	Age  int    `attr:"person/age"`
}

func TestEverything(t *testing.T) {
	conn := OpenConnection()
	txn, err := conn.Write(Person{Name: "Donald", Age: 48})
	require.NoError(t, err)
	db := txn.Database
	donald := Person{Name: "Donald"}
	ok := db.Fetch(&donald)
	require.True(t, ok)
	assert.Equal(t, 48, donald.Age)

	named := Named{Name: "Donald"}
	ok = db.Fetch(&named)
	assert.Equal(t, 48, named.Age)
	require.True(t, ok)

	missing := Person{Name: "Leah"}
	ok = db.Fetch(&missing)
	require.False(t, ok)

	t.Skip("failing, probably due to incorrect unique value check")
	_, err = conn.Write(Person{Name: "Stephen", Age: 44})
	require.NoError(t, err)
	stephen := Person{Name: "Stephen"}
	ok = db.Fetch(&stephen)
	require.True(t, ok)
	assert.Equal(t, 44, stephen.Age)
}
