package constructive

type Person struct {
	ID     uint   `attr:"sys/db/id"`
	Name   string `attr:"person/name"`
	Age    int    `attr:"person/age"`
	Active bool   `attr:"person/active"`
}

type Named struct {
	Name string `attr:"person/name"`
}
