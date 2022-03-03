package database

import (
	"testing"

	"github.com/dball/constructive/pkg/sys"
	. "github.com/dball/constructive/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnection(t *testing.T) {
	conn := OpenConnection()
	txn, err := conn.Write(Request{
		Claims: []Claim{
			{E: TempID("name"), A: sys.DbIdent, V: String("person/name")},
			{E: TempID("name"), A: sys.AttrType, V: sys.AttrTypeString},
		}},
	)
	require.NoError(t, err)
	assert.Equal(t, map[TempID]ID{TempID("name"): sys.FirstUserID + 1}, txn.NewIDs)
	db0 := txn.Database
	name := txn.NewIDs[TempID("name")]
	txn, err = conn.Write(Request{
		Claims: []Claim{
			{E: TempID("donald"), A: name, V: String("Donald")},
			{E: TempID("stephen"), A: name, V: String("Stephen")},
		},
	})
	require.NoError(t, err)
	donald, ok := txn.NewIDs["donald"]
	require.True(t, ok)
	db1 := txn.Database
	assert.True(t, db1.Select(Selection{E: donald}).Next())
	assert.False(t, db0.Select(Selection{E: donald}).Next())
}
