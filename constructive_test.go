package constructive

import (
	"testing"

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
	conn := OpenConnection()
	_, err := conn.Write(Person{Name: "Donald", Age: 48}, Person{Name: "Stephen", Age: 44})
	require.NoError(t, err)
}
