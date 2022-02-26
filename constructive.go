package constructive

type Person struct {
	ID     uint   `attr:"sys/db/id"`
	Name   string `attr:"person/name"`
	Age    int    `attr:"person/age"`
	Active bool   `attr:"person/active"`
}

/*
func sketch() {
	// register the structs for which we have names
	conn := BuildConn(map[string]interface{}{
		"person": Person{},
	})
	// record takes a varargs of entities
	result := conn.Record(Person{Name: "Donald"}, Person{Name: "Stephen"})
	// selectall returns a slice of instances that will correspond to the type name
	// note the result returns the database after the transaction was applied
	people := result.db.SelectAll(Selection{A: "person/name", V: "Donald"}, "person")
	// selectone returns a one instance
	// note the conn returns the database at this moment in time
	person, ok := conn.Db().SelectOne(Selection{A: "person/name", V: "Donald"}, "person")
}
*/
