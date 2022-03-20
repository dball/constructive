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
	inst := Instant("2020-03-11T12:00:00Z")
	conn.SetClock(BuildFixedClock(inst))
	txn, err := conn.Write(Request{
		Claims: []Claim{
			{E: TempID("name"), A: sys.DbIdent, V: String("person/name")},
			{E: TempID("name"), A: sys.AttrType, V: sys.AttrTypeString},
			{E: TempID("name"), A: sys.AttrUnique, V: sys.AttrUniqueIdentity},
		}},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, txn.NewIDs["name"])
	db0 := txn.Database
	t.Log("db0", db0.Dump())
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

	txnDatums := db1.Select(Selection{E: txn.ID})
	require.True(t, txnDatums.Next())
	txnDatum, ok := txnDatums.Value().(Datum)
	require.True(t, ok)
	assert.Equal(t, Datum{E: txn.ID, A: sys.TxAt, V: inst, T: txn.ID}, txnDatum)
	assert.False(t, txnDatums.Next())

	txn, err = conn.Write(Request{
		Claims: []Claim{
			{E: LookupRef{A: Ident("person/name"), V: String("Donald")}, A: Ident("person/name"), V: String("Donald"), Retract: true},
		},
	})
	require.NoError(t, err)
	assert.False(t, txn.Database.Select(Selection{E: donald}).Next())
}

func TestTempIDs(t *testing.T) {
	t.Run("identity uniqueness resolves to an extant entity", func(t *testing.T) {
		conn := OpenConnection()
		txn, err := conn.Write(Request{
			Claims: []Claim{
				{E: TempID("name"), A: sys.DbIdent, V: String("person/name")},
				{E: TempID("name"), A: sys.AttrType, V: sys.AttrTypeString},
				{E: TempID("name"), A: sys.AttrUnique, V: sys.AttrUniqueValue},
				{E: TempID("age"), A: sys.DbIdent, V: String("person/age")},
				{E: TempID("age"), A: sys.AttrType, V: sys.AttrTypeInt},
			}},
		)
		require.NoError(t, err)
		name := txn.NewIDs[TempID("name")]
		age := txn.NewIDs[TempID("age")]
		txn, err = conn.Write(Request{
			Claims: []Claim{
				{E: TempID("me"), A: name, V: String("Donald")},
				{E: TempID("me"), A: age, V: Int(7)},
			},
		})
		require.NoError(t, err)
		donald := txn.NewIDs[TempID("me")]
		assert.NotEmpty(t, donald)
		txn, err = conn.Write(Request{
			Claims: []Claim{
				{E: TempID("me"), A: name, V: String("Donald")},
				{E: TempID("me"), A: age, V: Int(17)},
			},
		})
		assert.Error(t, err)
	})
	t.Run("value uniqueness does not resolve to an extant entity", func(t *testing.T) {
		conn := OpenConnection()
		txn, err := conn.Write(Request{
			Claims: []Claim{
				{E: TempID("name"), A: sys.DbIdent, V: String("person/name")},
				{E: TempID("name"), A: sys.AttrType, V: sys.AttrTypeString},
				{E: TempID("name"), A: sys.AttrUnique, V: sys.AttrUniqueValue},
				{E: TempID("age"), A: sys.DbIdent, V: String("person/age")},
				{E: TempID("age"), A: sys.AttrType, V: sys.AttrTypeInt},
			}},
		)
		require.NoError(t, err)
		name := txn.NewIDs[TempID("name")]
		age := txn.NewIDs[TempID("age")]
		txn, err = conn.Write(Request{
			Claims: []Claim{
				{E: TempID("me"), A: name, V: String("Donald")},
				{E: TempID("me"), A: age, V: Int(7)},
			},
		})
		require.NoError(t, err)
		donald := txn.NewIDs[TempID("me")]
		assert.NotEmpty(t, donald)
		txn, err = conn.Write(Request{
			Claims: []Claim{
				{E: TempID("me"), A: name, V: String("Donald")},
				{E: TempID("me"), A: age, V: Int(17)},
			},
		})
		assert.ErrorIs(t, err, ErrInvalidValue)
	})
	t.Run("temp id values also resolve", func(t *testing.T) {
		conn := OpenConnection()
		txn, err := conn.Write(Request{
			Claims: []Claim{
				{E: TempID("name"), A: sys.DbIdent, V: String("node/name")},
				{E: TempID("name"), A: sys.AttrType, V: sys.AttrTypeString},
				{E: TempID("link"), A: sys.DbIdent, V: String("node/link")},
				{E: TempID("link"), A: sys.AttrType, V: sys.AttrTypeRef},
			},
		})
		require.NoError(t, err)
		name := txn.NewIDs[TempID("name")]
		link := txn.NewIDs[TempID("link")]
		txn, err = conn.Write(Request{
			Claims: []Claim{
				{E: TempID("head"), A: name, V: String("head")},
				{E: TempID("head"), A: link, V: TempID("tail")},
				{E: TempID("tail"), A: name, V: String("tail")},
			},
		})
		require.NoError(t, err)
	})
}

func TestRetract(t *testing.T) {
	conn := OpenConnection()
	_, err := conn.Write(Request{
		Claims: []Claim{
			{E: TempID("name"), A: sys.DbIdent, V: String("person/name")},
			{E: TempID("name"), A: sys.AttrType, V: sys.AttrTypeString},
			{E: TempID("name"), A: sys.AttrUnique, V: sys.AttrUniqueIdentity},
			{E: TempID("age"), A: sys.DbIdent, V: String("person/age")},
			{E: TempID("age"), A: sys.AttrType, V: sys.AttrTypeInt},
		}},
	)
	require.NoError(t, err)
	txn, err := conn.Write(Request{
		Claims: []Claim{
			{E: TempID("me"), A: Ident("person/name"), V: String("Donald")},
			{E: TempID("me"), A: Ident("person/age"), V: Int(7)},
		},
	})
	require.NoError(t, err)
	donald := txn.NewIDs[TempID("me")]
	txn, err = conn.Write(Request{
		Claims: []Claim{
			{E: donald, A: Ident("person/name"), Retract: true},
			{E: donald, A: Ident("person/age"), Retract: true},
		},
	})
	require.NoError(t, err)
	assert.False(t, txn.Database.Select(Selection{E: donald}).Next())
}
