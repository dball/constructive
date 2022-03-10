package constructive

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Person struct {
	ID     uint   `attr:"sys/db/id"`
	Name   string `attr:"person/name"`
	Age    int    `attr:"person/age"`
	Active bool   `attr:"person/active"`
}

type Named struct {
	Name string `attr:"person/name"`
}

func TestEverything(t *testing.T) {
	t.Skip("TODO")
	conn := OpenConnection()
	txn, err := conn.Write(Person{Name: "Donald", Age: 48})
	require.NoError(t, err)
	db := txn.Database
	donald := Person{Name: "Donald"}
	ok := db.Fetch(&donald)
	require.True(t, ok)
	assert.Equal(t, 48, donald.Age)
}
