// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
    sqlite persistence of relish objects

	Specific methods of the database abstraction layer for persisting relish types and attribute and relation specifications.



    `CREATE TABLE robject(
       id INTEGER PRIMARY KEY,
       id2 INTEGER, 
       idReversed BOOLEAN, --???
       typeName TEXT   -- Should be typeId because type should be another RObject!!!!
    )`

*/

import (
	"fmt"
	. "relish/runtime/data"
)

/*
Adds the table to the database which stores the core of each RObject instance i.e. the object's
id (uuid actually) and type. Only creates the table if the table does not yet exist.
Should be called at first use of the db as part of  initializing it.
*/
func (db *SqliteDB) EnsureObjectTable() {
	s := `CREATE TABLE IF NOT EXISTS RObject(
           id INTEGER PRIMARY KEY,
           id2 INTEGER NOT NULL, 
           flags TINYINT NOT NULL, -- ??? is BOOLEAN a type in sqlite?
           typeName TEXT NOT NULL  -- Should be typeId because type should be another RObject!!!!
         )`
	err := db.conn.Exec(s)
	if err != nil {
		panic(fmt.Sprintf("conn.Exec(%s): db error: %s", s, err))
	}
}

/*
Adds a table to the database for a type, if the table does not yet exist.
Should be called after the type has been assigned all of its attribute specifications and
relation specifications.
*/
func (db *SqliteDB) EnsureTypeTable(typ *RType) (err error) {
	s := "CREATE TABLE IF NOT EXISTS " + db.TableNameIfy(typ.ShortName()) +
		"(id INTEGER PRIMARY KEY"

	// Loop over primitive-valued attributes - for each, make a column in the table - TBD

	for _, attr := range typ.Attributes {
		if attr.Part.Type.IsPrimitive {
			s += ",\n" + attr.Part.DbColumnDef()

		}
	}

	// What about relations? Do separately.

	/*
	   someAttributeName_ID INTEGER PRIMARY KEY,
	   someIntAttribute INTEGER,
	   someFloatAttribute FLOAT,
	   someBooleanAttribute BOOLEAN,
	   someStringAttribute TEXT 
	*/

	s += ");"
	err = db.conn.Exec(s)
	if err != nil {
		err = fmt.Errorf("conn.Exec(%s): db error: %s", s, err)
	}
	return
}

/*
Adds the table to the database which associates a unique name to each specially dubbed RObject instance.
RELISH's local persistence model uses persistence by reachability. Special objects are "dubbed" with
an official name. These objects are, in the dubbing operation, made persistent. Other objects are
made persistent if they are referred to, directly or indirectly, by a persistent object. i.e. 
persistence is contagious, via object attribute and relation linkage with already-persistent objects.

Only creates the table if the table does not yet exist.
Should be called at first use of the db as part of  initializing it.

*/
func (db *SqliteDB) EnsureObjectNameTable() {
	s := `CREATE TABLE IF NOT EXISTS RName(
	       name TEXT PRIMARY KEY,
           id INTEGER UNIQUE NOT NULL
         )`
	err := db.conn.Exec(s)
	if err != nil {
		panic(fmt.Sprintf("conn.Exec(%s): db error: %s", s, err))
	}
}



/*
Adds the table to the database which associates a full name of each code package with a local-site unique
short name for the package.

Only creates the table if the table does not yet exist.
Should be called at first use of the db as part of  initializing it.

*/
func (db *SqliteDB) EnsurePackageTable() {
	s := `CREATE TABLE IF NOT EXISTS RPackage(
	       name TEXT PRIMARY KEY,
           shortName TEXT UNIQUE NOT NULL
         )`
	err := db.conn.Exec(s)
	if err != nil {
		panic(fmt.Sprintf("conn.Exec(%s): db error: %s", s, err))
	}
}




/*
Ensure that DB tables exist to represent the non-primitive-valued attributes and the relations
of the type.
*/
func (db *SqliteDB) EnsureAttributeAndRelationTables(t *RType) (err error) {

	for _, attr := range t.Attributes {
		if !attr.IsTransient && !attr.Part.Type.IsPrimitive {
			err = db.EnsureAttributeTable(attr)
			if err != nil {
				return
			}
		}
	}

	for _, rel := range t.ForwardRelations {
		if !rel.IsTransient {
			err = db.EnsureRelationTable(rel)
			if err != nil {
				return
			}
		}
	}
	return
}

/*
Name string
  Type *RType
  ArityLow int32
  ArityHigh int32
  CollectionType string // "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap", ""
  OrderAttrName string   // What is this?

  // End 0

  // 0,1,or set or attribute
  id0 INTEGER

  // list or sortedset 
  id0 INTEGER, ord0 INTEGER

  // End 1    can also have a map or string map as the collection type

  // 0,1,or set or attribute
  id1 INTEGER

  // list or sortedset or non-string-map
  id1 INTEGER, ord1 INTEGER

  // string-map 
  id1 INTEGER, key1 TEXT

*/
func (db *SqliteDB) EnsureRelationTable(rel *RelationSpec) (err error) {

	s := "CREATE TABLE IF NOT EXISTS " + db.TableNameIfy(rel.ShortName()) + "("

	// Prepare End[0]

	s += "id0 INTEGER NOT NULL,\n"

	switch rel.End[0].CollectionType {
	case "list", "sortedlist", "sortedset":
		s += "ord0 INTEGER NOT NULL,\n"
	}

	// Prepare End[1]

	s += "id1 INTEGER NOT NULL"

	switch rel.End[1].CollectionType {
	case "list", "sortedlist", "sortedset", "map", "sortedmap":
		s += ",\nord1 INTEGER NOT NULL"
	case "stringmap", "sortedstringmap":
		s += ",\nkey1 TEXT NOT NULL"
	}

	s += ");"
	err = db.conn.Exec(s)
	if err != nil {
		err = fmt.Errorf("conn.Exec(%s): db error: %s", s, err)
	}
	return
}

/*
Name string
  Type *RType
  ArityLow int32
  ArityHigh int32
  CollectionType string // "list","sortedlist", "set", "sortedset", "map", "stringmap","sortedmap","sortedstringmap",""
  OrderAttrName string   // What is this?
*/
func (db *SqliteDB) EnsureAttributeTable(attr *AttributeSpec) (err error) {

	s := "CREATE TABLE IF NOT EXISTS " + db.TableNameIfy(attr.ShortName()) + "("

	// Prepare Whole end

	s += "id0 INTEGER NOT NULL,\n"

	// Prepare Part

	s += "id1 INTEGER NOT NULL"

	switch attr.Part.CollectionType {
	case "list", "sortedlist", "sortedset", "map", "sortedmap":
		s += ",\nord1 INTEGER NOT NULL"
	case "stringmap", "sortedstringmap":
		s += ",\nkey1 TEXT NOT NULL"
	}

	s += ");"
	err = db.conn.Exec(s)
	if err != nil {
		err = fmt.Errorf("conn.Exec(%s): db error: %s", s, err)
	}
	return
}

// maybe use RObject interface instead of robject

/*
   Make a fully qualified type name into a legal table name in sqlite.
*/
func (db *SqliteDB) TableNameIfy(typeName string) string {
	return "[" + typeName + "]" // TODO Implement substitutions as needed.
}

/*
   Make a fully qualified type name from the corresponding table name in sqlite.
*/
func (db *SqliteDB) TypeNameIfy(tableName string) string {
	return tableName[1 : len(tableName)-1] // TODO Implement substitutions as needed.
}