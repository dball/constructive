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

## Acknowledgements

This library is an implementation of many of the ideas and features of the Datomic databases, albeit with no durability beyond the process runtime and no transaction history. In this respect, it is significantly also inspired by the Datascript library.

* [Datomic](https://www.datomic.com/)
* [Datascript](https://github.com/tonsky/datascript)

Constructive uses slightly different system idents than either, preferring a path hierachy.

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

An system ident uniquely identifies a datum by name, e.g. `person/name` or, indeed, `sys/db/ident`. Attributes are almost always referred to by their idents. Not all idents are attributes, only those with types. Idents are also the idiomatic way to represent enumerations, and have many uses beyond. The system reserves the `sys`
root, rejecting claims for such idents or about the entities to which they may refer. Users may use the remainder of the space as they see fit, though they're
recommended to use paths for consistency.

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

----

THESE ARE LIES this is aspirational, an experiment in documentation-driven development.

### sys/attr/ref/type

This qualifies the type of reference. It's only value is:

* `sys/attr/ref/type/component`

This specifies that the reference is to a component. A component entity is one whose existence is governed by a parent. A component ref attribute changes the
behavior of the system when applying transactions in the following ways:

If a claim to a tempid that resolves via identity uniqueness is about a scalar component ref, and the claimed value is a tempid, and the existing entity has a value
for the component ref, the value's tempid will resolve to the existing component.

### sys/attr/ref/type/component/key

If a similar claim is about a set component ref, this attribute governs the transaction behavior.

If this is not present, all datums about the existing components are retracted before considering the claims. This results in unnecessary writes if used liberally.
It is recommended that the component key be given, identifying the attribute whose value on the components is unique. If the key is present, all existing components
that lack a claim to their identity are fully retracted. 

----

## Structs

The primary use interface for constructive is intended to be structs, the dominant data structure in Golang.