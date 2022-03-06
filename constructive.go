package constructive

import (
	"github.com/dball/constructive/internal/iterator"
	"github.com/dball/constructive/pkg/types"
)

// Record is a struct with fields that have attr tags that control its
// marshaling into and out of a constructive database.
type Record interface{}

// Connection is a writable constructive database.
type Connection interface {
	types.Connection
	// Record records the given records and returns the successful transaction
	// or error.
	Record(records ...Record) (Transaction, error)
	// Stable returns a stable snapshot of the database.
	Stable() Database
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

// Database is a stable snapshot of data.
type Database interface {
	types.Database
	// Query returns an iterator of all records matching all of the selections, where the
	// records are instances of the exemplar with values corresponding to the fields' attrs.
	Query(exemplar Record, selections ...types.Selection) *iterator.Iterator
	// Fetch returns the first record matching the non-empty values of the attr fields in the exemplar.
	// The exemplar must have an entity id or at least one unique attr field. If multiple such fields
	// are given and resolve to different entities, this returns false.
	Fetch(exemplar Record) (Record, bool)
	// FetchByID returns the record with the given entity id.
	FetchByID(exemplar Record, id types.ID) (Record, bool)
}
