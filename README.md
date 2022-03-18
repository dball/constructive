# Constructive

Constructive is an experiment in indexed data storage in Golang, specifically focused on storing and querying structs.

## Motivations

* I often have large numbers of structs in a collection that I need to search in various ways.
* I often need to efficiently apply changes to the collection without affecting current readers.
* I often need to convert from one type of struct to another with very similar or identical data.
* I am convinced the datum rdf model is widely applicable and underused.

## Constraints

* I only care about local storage at this time. Durable data storage and interchange are goals, but not immediate.
* I care most about correct behavior, then API usability and stability, then performance, then memory efficiency.
* I do not care about being able to go back in history at this time. The data model readily supports it, but it would require a more sophisticated index to be practical.

## Theory

The datum model is an extension of the RDF model. There are four components to a datum:

### Entity

An entity is a thing with an identity. Its characteristics may change over time, but its identity persists. I am an entity.

### Attribute

An attribute is a property of an entity. Attributes have global identities and are themselves entities. "Person's given name" is an attribute.

### Value

A value is an observation, a claim, a fact, a reading. "Donald" is a value which happens to be my current given name.

### Transaction

A transaction is an entity that asserts a set of datums are true at a point in time. When I register an account, my name, credentials, etc. are collectively recorded in a transaction. The transaction necessarily records the time but may also include other audit details like the IP address of the host from which I created the account.

----

I contend this data model constitutes the simplest possible fundamental data model for general use. Simplifying by dropping the transaction component results in a system which has no general treatment of data attribution and is a huge problem for the industry.

This is not a great data model for representing observations about things that lack durable identities or where time and attribution are not important features. It is particularly valuable when working with data combined from diverse sources, where being able to track the provenance of data consistently is important. Joining
datasets either by sharing attributes or asserting a specific relationship between attributes can be readily expressed and often straightforwardly implemented.

## Attributes

Attributes are entities that have at least two system attributes, ident and type, and are governed by others. The system attribute values, once asserted, may neither be asserted anew with new values nor retracted.

### sys/db/ident

An system ident uniquely identifies a datum by name, e.g. `person/name` or, indeed, `sys/db/ident`. Attributes are almost always referred to by their idents. Not all idents are attributes, only those with types. Idents are the idiomatic way to represent enumerations.

### sys/attr/type

This identifies the type of value to which the attribute refers, one of:

* `sys/attr/type/string`
* `sys/attr/type/inst` a moment in time
* `sys/attr/type/int`
* `sys/attr/type/float`
* `sys/attr/type/ref` a reference to an entity
* `sys/attr/type/bool`

Nil is not a valid value for any type. The absence of a value is represented by the absence of the datum. An affirmation of a value's absence should be represented by another attribute if the zero value is valid in the use domain.

### sys/attr/cardinality

This specifies the number of values to which the attribute may refer. The valid cardinality values are:

* `sys/attr/cardinality/one`
* `sys/attr/cardinality/many`

Cardinality one, a scalar, is assumed in the absence of a cardinality attribute. Cardinality many uses set semantics.

### sys/attr/unique

This specifies that the attribute's value is unique in the database, only one entity may assert it. It has two values:

* `sys/attr/unique/identity`
* `sys/attr/unique/value`

Both enforce the uniqueness constraint. The only difference is that when asserting claims, if a tempid is used in a claim for this attribute, and an entity already asserts the claimed value, the tempid will resolve to the extant entity for identity uniqueness. By contrast, a value uniqueness attribute will cause the claim to be rejected.

## Structs

The primary use interface for constructive is intended to be structs, the dominant data structure in Golang.