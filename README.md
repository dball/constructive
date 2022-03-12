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

--

I contend this data model constitutes the simplest possible fundamental data model for general use. Simplifying by dropping the transaction component results in a system which has no general treatment of data attribution and is a huge problem for the industry.

This is not a great data model for representing observations about things that lack durable identities or where time and attribution are not important features.