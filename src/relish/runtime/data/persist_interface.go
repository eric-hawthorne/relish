// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.
// Abstraction of persistence service for relish data.

package data

type DB interface {
	EnsureTypeTable(typ *RType) (err error)
	QueueStatements(statementGroup string)
	PersistSetAttr(obj RObject, attr *AttributeSpec, val RObject, attrHadValue bool) (err error)
	PersistAddToAttr(obj RObject, attr *AttributeSpec, val RObject, insertedIndex int) (err error)
	PersistRemoveFromAttr(obj RObject, attr *AttributeSpec, val RObject, removedIndex int) (err error)
	EnsurePersisted(obj RObject) (err error)
	EnsureAttributeAndRelationTables(t *RType) (err error)
	ObjectNameExists(name string) (found bool, err error)
	NameObject(obj RObject, name string)
	RecordPackageName(name string, shortName string)
	FetchByName(name string, radius int) (obj RObject, err error)
	Fetch(id int64, radius int) (obj RObject, err error)
	FetchAttribute(objId int64, obj RObject, attr *AttributeSpec, radius int) (val RObject, err error)

	/*
	Converts object and object-attribute query language expressions to SQL queries.

	e.g. vehicles/Car, "speed > 60"   ==> "select id from [vehicles/Vehicle] where speed > 60"
	*/
	OQLWhereToSQLSelect(typ *RType, oqlWhereCriteria string) (sqlSelectQuery String, err error)

	Close()
}
