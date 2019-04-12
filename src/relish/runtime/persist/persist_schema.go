// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
    persist_schema.go - sqlite persistence of relish data types and attribute-associations and relation-associations.

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
	"io"
)

/*
Adds the table to the database which stores the core of each RObject instance i.e. the object's
id (uuid actually) and type. Only creates the table if the table does not yet exist.
Should be called at first use of the db as part of  initializing it.
*/
func (db *SqliteDBThread) EnsureObjectTable() {
	s := `CREATE TABLE IF NOT EXISTS RObject(
           id INTEGER PRIMARY KEY,
           id2 INTEGER NOT NULL, 
           flags TINYINT NOT NULL, -- ??? is BOOLEAN a type in sqlite?
           typeName TEXT NOT NULL  -- Should be typeId because type should be another RObject!!!!
         )`
	err := db.ExecStatement(s)
	if err != nil {
		panic(fmt.Sprintf("db.ExecStatement(%s): db error: %s", s, err))
	}
}

/*
Adds a table to the database for a type, if the table does not yet exist.
Should be called after the type has been assigned all of its attribute specifications and
relation specifications.
*/
func (db *SqliteDBThread) EnsureTypeTable(typ *RType) (err error) {
	s := "CREATE TABLE IF NOT EXISTS " + db.db.TableNameIfy(typ.ShortName()) +
		"(id INTEGER PRIMARY KEY"

	// Loop over primitive-valued attributes - for each, make a column in the table - TBD

	for _, attr := range typ.Attributes {
		if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" && ! attr.IsTransient {
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
	err = db.ExecStatement(s)
	if err != nil {
		err = fmt.Errorf("db.ExecStatement(%s): db error: %s", s, err)
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
func (db *SqliteDBThread) EnsureObjectNameTable() {
	s := `CREATE TABLE IF NOT EXISTS RName(
	       name TEXT PRIMARY KEY,
           id INTEGER UNIQUE NOT NULL
         )`
	err := db.ExecStatement(s)
	if err != nil {
		panic(fmt.Sprintf("db.ExecStatement(%s): db error: %s", s, err))
	}
}



/*
Adds the table to the database which associates a full name of each code package with a local-site unique
short name for the package.

Only creates the table if the table does not yet exist.
Should be called at first use of the db as part of  initializing it.

*/
func (db *SqliteDBThread) EnsurePackageTable() {
	s := `CREATE TABLE IF NOT EXISTS RPackage(
	       name TEXT PRIMARY KEY,
           shortName TEXT UNIQUE NOT NULL
         )`
	err := db.ExecStatement(s)
	if err != nil {
		panic(fmt.Sprintf("db.ExecStatement(%s): db error: %s", s, err))
	}
	
	err = db.restorePackageNameMappings()
    if err != nil {
	   panic(fmt.Sprintf("restorePackageNameMappings: db error: %s", err))
   }	
}

/*
If packages have been already defined in this database, read the mappings between shortnames
of packages and full names of packages into the runtime, so they can be re-used during package generation.
*/
func (db *SqliteDBThread) restorePackageNameMappings() (err error) {

	query := "SELECT name,shortName FROM RPackage"

	selectStmt, err := db.dbt.conn.Prepare(query)
	if err != nil {
		return
	}

	defer selectStmt.Close()

	err = selectStmt.Query()
	if err != nil && err != io.EOF {
		return
	}

    for ; err == nil ; err = selectStmt.Next() {   

		var name string
		var shortName string
		err = selectStmt.Scan(&name,&shortName)
		if err != nil {
			return
		}
		RT.PkgNameToShortName[name] = shortName
		RT.PkgShortNameToName[shortName] = name
	}
	if err == io.EOF {
	   err = nil 
	} else {
	   err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", query, err)   
	   return  
	} 



	return
}

/*
Record in the db the mapping from the package name to shortName.
*/
func (db *SqliteDBThread)	RecordPackageName(name string, shortName string) {

	// TODO create a map of prepared statements and look up the statement to use.

	stmt := fmt.Sprintf("INSERT INTO RPackage(name,shortName) VALUES('%s','%s')", name, shortName)    

	err := db.ExecStatement(stmt)
    if err != nil {
    	panic(err)
    }
	RT.PkgNameToShortName[name] = shortName
	RT.PkgShortNameToName[shortName] = name	
	return	
}



/*
Ensure that DB tables exist to represent the non-primitive-valued attributes and the relations
of the type.
*/
func (db *SqliteDBThread) EnsureAttributeAndRelationTables(t *RType) (err error) {
    // fmt.Println("EnsureAttributeAndRelationTables", t)
	for _, attr := range t.Attributes {
		// fmt.Println(attr)
		if !attr.IsTransient {
			if attr.Part.Type.IsPrimitive {
				if attr.Part.ArityHigh != 1 { // a multi-valued primitive-type attribute
				//fmt.Println("arityHigh ",attr.Part.ArityHigh)
				   err = db.EnsureMultiValuedPrimitiveAttributeTable(attr)
				   if err != nil {
			  		  return
				   }					
				} else if attr.Part.CollectionType != "" {
					// fmt.Println("OOPS collection of primitives and arityHigh ",attr.Part.ArityHigh)
				    err = db.EnsureNonPrimitiveAttributeTable(attr)
				    if err != nil {
			  		  return
				    }					
				}
			} else {
			   err = db.EnsureNonPrimitiveAttributeTable(attr)
			   if err != nil {
		  		  return
			   }
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
  CollectionType string // "list","sortedlist", "set", "sortedset", "map", "stringmap","sortedmap","sortedstringmap",""
  OrderAttrName string   // What is this?
*/
func (db *SqliteDBThread) EnsureNonPrimitiveAttributeTable(attr *AttributeSpec) (err error) {

    //fmt.Println("EnsureNonPrimitiveAttributeTable", attr)
	s := "CREATE TABLE IF NOT EXISTS " + db.db.TableNameIfy(attr.ShortName()) + "("

	// Prepare Whole end

	s += "id0 INTEGER NOT NULL,\n"

	// Prepare Part

	s += "id1 INTEGER NOT NULL"

   if attr.Part.ArityHigh != 1 { // This is a multi-valued attribute.
   	switch attr.Part.CollectionType {
   	case "list", "sortedlist", "sortedset", "map", "sortedmap":
   		s += ",\nord1 INTEGER NOT NULL"
   	case "stringmap", "sortedstringmap":
   		s += ",\nkey1 TEXT NOT NULL"
   	}
   }

	s += ");"
	err = db.ExecStatement(s)
	if err != nil {
		err = fmt.Errorf("db.ExecStatement(%s): db error: %s", s, err)
	}
	return
}


/*
*/
func (db *SqliteDBThread) EnsureMultiValuedPrimitiveAttributeTable(attr *AttributeSpec) (err error) {
    //fmt.Println("EnsureMultiValuedPrimitiveAttributeTable", attr)	

	s := "CREATE TABLE IF NOT EXISTS " + db.db.TableNameIfy(attr.ShortName()) + "("

	// Prepare Whole end

	s += "id INTEGER NOT NULL,\n"
	
	// Prepare column for Part
		
	s += attr.Part.Type.DbCollectionColumnDef()	 
		
	// and add a sorting/ordering column if appropriate
		
    switch attr.Part.CollectionType {
    case "list", "sortedlist", "sortedset", "map", "sortedmap":   	
	s += ",\nord1 INTEGER NOT NULL"
    case "stringmap", "sortedstringmap":
	s += ",\nkey1 TEXT NOT NULL"
    }	

	s += ");"
	
	err = db.ExecStatement(s)
	if err != nil {
		err = fmt.Errorf("db.ExecStatement(%s): db error: %s", s, err)
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









func (db *SqliteDBThread) EnsureNonPrimitiveCollectionTable(table string, isMap bool, isOrdered bool, keyType *RType) (err error) {

	s := "CREATE TABLE IF NOT EXISTS " + table + "("

	// Prepare Whole end

	s += "id0 INTEGER NOT NULL,\n"

	// Prepare Part

	s += "id1 INTEGER NOT NULL"

	if keyType == StringType {	
		s += ",\nkey1 TEXT NOT NULL"
	} else if isMap || isOrdered {
		s += ",\nord1 INTEGER NOT NULL"
   }

	s += ");"
	err = db.ExecStatement(s)
	if err != nil {
		err = fmt.Errorf("db.ExecStatement(%s): db error: %s", s, err)
	}
	return
}


/*
*/
func (db *SqliteDBThread) EnsurePrimitiveCollectionTable(table string, isMap bool, isOrdered bool, keyType *RType, elementType *RType) (err error) {

	s := "CREATE TABLE IF NOT EXISTS " + table + "("

	// Prepare Whole end

	s += "id INTEGER NOT NULL,\n"
	
	// Prepare column for Part
		
	s += elementType.DbCollectionColumnDef()	 
		
	// and add a sorting/ordering column if appropriate
		
   if keyType == StringType {	
   	s += ",\nkey1 TEXT NOT NULL"
   } else if isMap || isOrdered {
   	s += ",\nord1 INTEGER NOT NULL"
   }

	s += ");"
	
	err = db.ExecStatement(s)
	if err != nil {
		err = fmt.Errorf("db.ExecStatement(%s): db error: %s", s, err)
	}
	return
}










/*
   Derive a collection table name from the collection's
   collection-type and element type.
   
   Need to ensure the collection table exists in the db

   Return metadata about the collection, including the table name.
*/
func (db *SqliteDBThread) EnsureCollectionTable(collection RCollection) (table string, isMap bool, isOrdered bool, keyType *RType, valType *RType, err error) {
        
   table, isMap, isOrdered, keyType, valType = db.db.TypeDescriptor(collection)
   
   if valType.IsPrimitive {
      err = db.EnsurePrimitiveCollectionTable(table, isMap, isOrdered, keyType, valType)      
   } else {
      err = db.EnsureNonPrimitiveCollectionTable(table, isMap, isOrdered, keyType)
   }
   return
}

/*
   valType is the ElementType() if the collection is not a map.
   valType is the ValType() if it is a map.
*/
func (db *SqliteDB) TypeDescriptor(collection RCollection) (table string, isMap bool, isOrdered bool, keyType *RType, valType *RType) {
   
   isMap = collection.IsMap()
   isSorting := collection.IsSorting()
   isOrdered = collection.IsOrdered()   

   var typeName string

   if isMap {
      // isStringMap = (    collection.(Map).KeyType() == StringType  )
      
      keyType = collection.(Map).KeyType()
      valType = collection.(Map).ValType()      

      
      typeName = collection.Type().ShortName()
      // fmt.Println(typeName, keyType.ShortName(), valType.ShortName(), valType.IsPrimitive)

      typeName = typeName[7:]
      
      
// IntType,Int32Type,UintType,Uint32Type    
      
      
   
      
      
   } else {  // The collection is not a map
      valType = collection.ElementType()   	
      typeName = valType.ShortName()

   }
   table = "["
   if isSorting {
      table += "sorted"
   }
   switch keyType {
   case StringType:
      table += "stringmap"   
   case IntType,Int32Type:
      table += "int64map"   
   case UintType,Uint32Type:
      table += "uint64map"   
   case nil:
      if collection.IsSet() {
         table += "set"
      } else {
         table += "list"
      }
   default:  
      table += "map"          
   }
   
   // TODO If a map, include the KeyType shortname in type name.
   table += "_of_" + typeName + "]"  
   return 
}



