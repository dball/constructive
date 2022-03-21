package constructive

import (
	"github.com/dball/constructive/internal/database"
	"github.com/dball/constructive/internal/iterator"
	"github.com/dball/constructive/pkg/destruct"
	"github.com/dball/constructive/pkg/types"
)

// Connection is a writable constructive database.
type Connection interface {
	// Prepare asserts the schema required to record the given records and returns the transaction.
	Prepare(records ...interface{}) (Transaction, error)
	// Write atomically records the given records and returns the transaction.
	Write(records ...interface{}) (Transaction, error)
	// Erase atomically erases the given records and returns the transaction.
	Erase(records ...interface{}) (Transaction, error)
	// Read returns a snapshot of the database.
	Read() Database
}

type connection struct {
	connection types.Connection
}

func (conn connection) Prepare(records ...interface{}) (Transaction, error) {
	claims := destruct.Schema(records...)
	txn, err := conn.connection.Write(types.Request{Claims: claims})
	if err != nil {
		return Transaction{}, err
	}
	return Transaction{ID: txn.ID, NewIDs: txn.NewIDs, Database: db{txn.Database}}, nil
}

func (conn connection) Write(records ...interface{}) (Transaction, error) {
	claims := destruct.DestructOnlyData(records...)
	txn, err := conn.connection.Write(types.Request{Claims: claims})
	if err != nil {
		return Transaction{}, err
	}
	return Transaction{ID: txn.ID, NewIDs: txn.NewIDs, Database: db{txn.Database}}, nil
}

func (conn connection) Erase(records ...interface{}) (Transaction, error) {
	claims := destruct.DestructOnlyData(records...)
	for i, claim := range claims {
		claim.Retract = true
		claims[i] = claim
	}
	txn, err := conn.connection.Write(types.Request{Claims: claims})
	if err != nil {
		return Transaction{}, err
	}
	return Transaction{ID: txn.ID, NewIDs: txn.NewIDs, Database: db{txn.Database}}, nil
}

func (conn connection) Read() Database {
	return db{conn.connection.Read()}
}

func OpenConnection() Connection {
	return connection{connection: database.OpenConnection()}
}

// Transaction represents a successful recording of records or datums.
type Transaction struct {
	// ID is the id of the transaction entity.
	ID types.ID
	// NewIDs is a map of the ids assigned to or resolved for the temp ids in the request.
	NewIDs map[types.TempID]types.ID
	// Database is the snapshot of the database after transacting the request.
	Database Database
}

func (txn Transaction) Fetch(ref interface{}) {
	ok := txn.Database.FetchByID(ref, txn.ID)
	if !ok {
		panic("Corruption: transaction not found in database")
	}
}

// Database is a stable snapshot of data.
type Database interface {
	// Query returns an iterator of all records matching all of the selections, where the
	// records are instances of the exemplar with values corresponding to the fields' attrs.
	Query(exemplar interface{}, selections ...types.Selection) *iterator.Iterator
	// Fetch examines the struct value of the given ref and searches the database for
	// a unique record, using the entity id field, then any unique attr fields. Exactly
	// one of these must have a non-empty value, otherwise this returns false. If a match
	// is specified and found, the ref's struct's attr fields are set from the selected datums.
	Fetch(ref interface{}) bool
	FetchByID(ref interface{}, id types.ID) bool
	Dump() interface{}
}

type db struct {
	database types.Database
}

func (db db) Query(exemplar interface{}, selections ...types.Selection) *iterator.Iterator {
	panic("TODO actually do, but also: maybe this takes a slice of ref and selections that are structs with field tagged constraint values")
}

func (db db) Fetch(ref interface{}) bool {
	return destruct.Fetch(ref, db.database)
}

func (db db) FetchByID(ref interface{}, id types.ID) bool {
	return destruct.Construct(ref, db.database, id)
}

func (db db) Dump() interface{} {
	return db.database.Dump()
}
