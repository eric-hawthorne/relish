// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
    persist_data.go - sqlite persistence of relish objects

	Specific methods of the database abstraction layer for persisting relish objects and attribute assignments.


    `CREATE TABLE robject(
       id INTEGER PRIMARY KEY,
       id2 INTEGER, 
       idReversed BOOLEAN, --???
       typeName TEXT   -- Should be typeId because type should be another RObject!!!!
    )`

    Note that collection and mult-valued attribute persistence are handled by methods in persist_collection.go
*/

import (
	"errors"
	"fmt"
	. "relish/dbg"
	. "relish/runtime/data"
	"relish/rterr"
	"strconv"
	"strings"
	"time"
	// "relish/global_loader"
	"io"
)

const TIME_LAYOUT = "2006-01-02 15:04:05.000"

/*
   Persist the setting of an attribute to a value.
   Only applies to single-valued attributes.
   Assumes that the the obj is already persisted, but does not assume that the value is.
*/
func (db *SqliteDBThread) PersistSetAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, attrHadValue bool) (err error) {

	if attr.Part.Type.IsPrimitive {

		table := db.db.TableNameIfy(attr.WholeType.ShortName())

		if val.Type() == TimeType {
			attrName := attr.Part.Name
			attrLocName := attrName + "_loc"
			t := time.Time(val.(RTime))
			timeString := t.UTC().Format(TIME_LAYOUT)
			locationName := t.Location().String()
			err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET %s=?, %s=? WHERE id=?", table, attrName, attrLocName), timeString, locationName, obj.DBID())			
			if err != nil {
				obj.SetLoadNeeded()
			}    
		} else if val.Type() == MutexType || val.Type() == RWMutexType || val.Type() == OwnedMutexType {
			// skip persisting
		} else {
			valStr,args := db.db.primitiveAttrValSQL(val)
			stmt := Stmt(fmt.Sprintf("UPDATE %s SET %s=? WHERE id=?", table, attr.Part.Name))
			if valStr == "?" {
			   stmt.Args(args) 
			} else {
			   stmt.Arg(valStr)
			}	
			stmt.Arg(obj.DBID())
  	        err = db.ExecStatements(stmt)	
			if err != nil {
				obj.SetLoadNeeded()
			}      	        		
		}
	} else { // non-primitive value type

		err = db.EnsurePersisted(th, val)
		if err != nil {		
			return
		}

		table := db.db.TableNameIfy(attr.ShortName())

		if attrHadValue {
			err = db.ExecStatement(fmt.Sprintf("UPDATE %s SET id1=? WHERE id0=?", table), val.DBID(), obj.DBID())     
			if err != nil {
				obj.SetLoadNeeded() // Persistence of the change failed. Memory state is invalid.
			}                               
		} else {
			err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), obj.DBID(), val.DBID()) 
			if err != nil {
				obj.SetLoadNeeded()
			}     			
		}
	}
	return

}


/*
Return a string representation of the value, suitable for use as the argument in a sql "SET attrname=%s"
Works for String, Int, Int32, Float, Bool, Uint, Uint32
*/
func (db *SqliteDB) primitiveAttrValSQL(val RObject) (s string, args []interface{}) {

	switch val.(type) {
	case Int:
		s = strconv.FormatInt(int64(val.(Int)), 10)
	case Int32:
		s = strconv.FormatInt(int64(int32(val.(Int32))), 10)	
	case Uint:
	    s = strconv.FormatUint(uint64(val.(Uint)), 10)		
	case Uint32:
		s = strconv.FormatUint(uint64(uint32(val.(Uint32))), 10)		    											
	case String:
//		s = "'" + SqlStringValueEscape(string(val.(String))) + "'"		
		s = "?"
		args = append(args, SqlStringValueEscape(string(val.(String))))
	case Float:
		s = strconv.FormatFloat(float64(val.(Float)), 'G', -1, 64)				
	case Bool:
		boolVal := bool(val.(Bool))
		if boolVal {
			s = "1"
		} else {
			s = "0"
		}
	default:
		panic(fmt.Sprintf("I don't know how to create SQL for an attribute value of underlying type %v.", val.Type()))
	}
	return 
}












/*
   Persist the setting of an attribute to nil.
   Only applies to single-valued attributes.
   Assumes that the the obj is already persisted.
   Ok to call this even if there is no value of the attribute for the object.
*/
func (db *SqliteDBThread) PersistRemoveAttr(obj RObject, attr *AttributeSpec) (err error) {

	var stmt string

	if attr.Part.Type.IsPrimitive {

		table := db.db.TableNameIfy(attr.WholeType.ShortName())
		
		if attr.Part.Type == TimeType {
		   attrName := attr.Part.Name
		   attrLocName := attrName + "_loc"			
		   stmt = fmt.Sprintf("UPDATE %s SET %s=NULL,%s=NULL WHERE id=?", table, attrName, attrLocName)
		} else if attr.Part.Type == MutexType || attr.Part.Type == RWMutexType || attr.Part.Type == OwnedMutexType {
			// skip persisting        
        } else	{
		   stmt = fmt.Sprintf("UPDATE %s SET %s=NULL WHERE id=?", table, attr.Part.Name) 
	    }
	} else { // non-primitive value type

		table := db.db.TableNameIfy(attr.ShortName())

		// TODO create a map of prepared statements and look up the statement to use.

		stmt = fmt.Sprintf("DELETE FROM %s WHERE id0=?", table) // Ensure DBID?                                       
	}
	err = db.ExecStatement(stmt, obj.DBID())
	return

}

/*
   Store the object  in the local sqlite database for the first time.
   If object is stored in local db already, does nothing.
   The object may already have a uuid, in which case it was a remote object (or an object whose persisting failed), 
   or it may not, in which case it is a new object that has only lived in memory in this process.
   Asynchronous. DB errors will not happen and be reported til later. <- to be true eventually
   NOT HAPPY WITH THE UNCERTAINTY OF STORAGE IN THE FLAG VALUE !!!
*/
func (db *SqliteDBThread) EnsurePersisted(th InterpreterThread, obj RObject) (err error) {
    defer Un(Trace(PERSIST_TR2, "EnsurePersisted", obj))	

    // If being persisted in another transaction, just wait it out, and then return.
    pers := obj.(Persistable)
    tx := pers.Transaction()
	if tx != nil && tx != th.Transaction() {
		th.AllowGC()
		tx.RLock()
		th.DisallowGC()
		tx.RUnlock()
		return
	}

	if obj.IsBeingStored() {
		return
	}
	
	obj.SetBeingStored() 
    if th.Transaction() != nil {
   		obj.(Persistable).SetTransaction(th.Transaction())
   	}	

	if obj.HasUUID() {
		err = db.persistRemoteObject(th, obj)
	} else {
		err = db.persistNewObject(th, obj)
	}
	if err != nil {
		return
	}

	RT.Cache(obj) // Put in an in-memory object cache so that the runtime will only contain one object instance for each uuid.

	err = db.PersistAttributesAndRelations(th, obj)
	if err != nil {
	   return
   }
   
   if obj.IsCollection() {
      collection := obj.(RCollection)
      err = db.persistCollection(th, collection) 
   }
	
   if th.Transaction() == nil {
       obj.SetStoredLocally()
   	}	
	return
}

/*
   Persists (if necessary) each object which is an attribute of the argument object, or is related to the object.
   Then persists the attribute linkage table or relation table entry.
   Assumes that obj is persisted but has only just now been persisted.
   Note. Only persists relation entries in the forward direction.

   TODO WE STILL HAVE TO CATCH REFERENCE LOOPS !!! Or do we avoid them already?

   TODO NOT DOING RELATIONS YET !! JUST ATTRIBUTES !!!! !!!!!!!!! !!!!!!!!!!!!!!!!!!! !!!! !
*/
func (db *SqliteDBThread) PersistAttributesAndRelations(th InterpreterThread, obj RObject) (err error) {
	defer Un(Trace(PERSIST_TR2, "PersistAttributesAndRelations", obj))

	objTyp := obj.Type()

	for _, attr := range objTyp.Attributes {

        Logln(PERSIST2_,attr)

		if ! attr.Part.Type.IsPrimitive {  

			val, found := RT.AttrValue(th, obj, attr, false, true, true)
			if !found {
				continue
			}

			table := db.db.TableNameIfy(attr.ShortName())

			if attr.IsMultiValued() { // a collection of non-primitives, owned by the "whole" object		   
			   
				collection := val.(RCollection)
				isMap := collection.IsMap()
				if isMap {
					theMap := collection.(Map)
					for key := range theMap.Iter(nil) {
						val, _ := theMap.Get(key)
						err = db.EnsurePersisted(th, val)
						if err != nil {
							return
						}
						if attr.Part.CollectionType == "stringmap" || attr.Part.CollectionType == "orderedstringmap" {
							stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id0,id1,key1) VALUES(?,?,?)", table)) 
							stmt.Arg(obj.DBID())
							stmt.Arg(val.DBID())
			                stmt.Arg(SqlStringValueEscape(string(key.(String))))
							err = db.ExecStatements(stmt)
							if err != nil {
								return
							}							
						} else {
							// TODO - We do not know if the key is persisted. We don't know if the key is an integer!!!
							// !!!!!!!!!!!!!!!!!!!!!!!!
							// !!!! NOT DONE YET !!!!!!
							// !!!!!!!!!!!!!!!!!!!!!!!!
							err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table), obj.DBID(), val.DBID(), key.DBID()) 					 
							if err != nil {
								return
							}
						}
					}
				} else {
					i := 0
					for val := range collection.Iter(nil) {
						err = db.EnsurePersisted(th, val)
						if err != nil {
							return
						}
						if collection.IsOrdered() {
							err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table), obj.DBID(), val.DBID(), i)
							if err != nil {
								return
							}						
						} else { // unordered set 
							err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), obj.DBID(), val.DBID())		
							if err != nil {
								return
							}						
						}
						i++
					}
				}
			} else { // a single non-primitive value or independent collection of non-primitive element type.

				err = db.EnsurePersisted(th, val)
				if err != nil {
					return
				}
				err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), obj.DBID(), val.DBID())
				if err != nil {
					return
				}				
			}
		} else if attr.IsComplex() {  // multi-valued primitive type or independent collection of primitive element type

		    val, found := RT.AttrValue(th, obj, attr, false, true, true)
		    if !found {
		   	    continue
		    }
			
			table := db.db.TableNameIfy(attr.ShortName())	

		    if attr.IsMultiValued() { // a collection of primitive-type objects owned by the "whole" object

				// TODO
				// !!!!!!!!!!!!!!!!!!!!!!!!
				// !!!! NOT DONE YET !!!!!!
				// !!!!!!!!!!!!!!!!!!!!!!!!
				
				val, found := RT.AttrValue(th, obj, attr, false, true, true)
				if !found {
					continue
				}
							
				valCols,valVars := attr.Part.Type.DbCollectionColumnInsert()
								   
				collection := val.(RCollection)
				isMap := collection.IsMap()
				if isMap {
				   
				   // !!!!!!!!!!!!!!!!!!!!!!!!
				   // !!!! NOT DONE YET !!!!!!
				   // !!!!!!!!!!!!!!!!!!!!!!!!			   
				   
				   
					theMap := collection.(Map)
					for key := range theMap.Iter(nil) {
						val, _ := theMap.Get(key)

						if attr.Part.CollectionType == "stringmap" || attr.Part.CollectionType == "orderedstringmap" {
						    stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,key1) VALUES(?,%s,?)", table, valCols, valVars)) 
						    stmt.Arg(obj.DBID())	
						    valParts := db.db.primitiveValSQL(val) 
						    stmt.Args(valParts)								
							
			                stmt.Arg(SqlStringValueEscape(string(key.(String))))
			            
						    err = db.ExecStatements(stmt)
							if err != nil {
								return
							}


						} else {
							// TODO - We do not know if the key is persisted. We don't know if the key is an integer!!!
							// !!!!!!!!!!!!!!!!!!!!!!!!
							// !!!! NOT DONE YET !!!!!!
							// !!!!!!!!!!!!!!!!!!!!!!!!
												
	                        stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,ord1) VALUES(?,%s,?)", table, valCols, valVars))
						    stmt.Arg(obj.DBID())			
						    valParts := db.db.primitiveValSQL(val) 
						    stmt.Args(valParts)
			     		    stmt.Arg(key.DBID())	       						
							
						    err = db.ExecStatements(stmt)						
							if err != nil {
								return
							}										 
						}
					}
				} else {
					i := 0					
					for val := range collection.Iter(nil) {
						
						if collection.IsOrdered() {					   
							stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,ord1) VALUES(?,%s,?)", table, valCols, valVars))
						
						    stmt.Arg(obj.DBID())							
							valParts := db.db.primitiveValSQL(val) 
							stmt.Args(valParts)
							stmt.Arg(i)
							err = db.ExecStatements(stmt)	
							if err != nil {
								return
							}												
						} else { // unordered set 
						   
							stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s) VALUES(?,%s)", table, valCols, valVars))		
							
						    stmt.Arg(obj.DBID())					
						    valParts := db.db.primitiveValSQL(val) 
						    stmt.Args(valParts)						
							
							err = db.ExecStatements(stmt)	
					        if err != nil {
								return
							}												
						}
						i++
					}
				}			
			
			} else { // attr.IsIndependentCollection()  // independent collection of primitive element type

				err = db.EnsurePersisted(th, val)
				if err != nil {
					return
				}
				err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), obj.DBID(), val.DBID())
				if err != nil {
					return
				}				
			}	
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
			if ! attr.Part.Type.IsPrimitive {

				val, found := RT.AttrValue(th, obj, attr, false, true, true)
				if !found {
					continue
				}

				table := db.db.TableNameIfy(attr.ShortName())

				if attr.IsMultiValued() { // a collection of non-primitives, owned by the "whole" object	
					collection := val.(RCollection)
					isMap := collection.IsMap()
					if isMap {
						theMap := collection.(Map)
						for key := range theMap.Iter(nil) {
							val, _ := theMap.Get(key)
							err = db.EnsurePersisted(th, val)
							if err != nil {
								return
							}
							if attr.Part.CollectionType == "stringmap" || attr.Part.CollectionType == "orderedstringmap" {								
						        stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id0,id1,key1) VALUES(?,?,?)", table)) 
						        stmt.Arg(obj.DBID())
						        stmt.Arg(val.DBID())
				                stmt.Arg(SqlStringValueEscape(string(key.(String))))
								err = db.ExecStatements(stmt)
								if err != nil {
								   return
							    }										
							} else {
								// TODO - We do not know if the key is persisted. We don't know if the key is an integer!!!
								// !!!!!!!!!!!!!!!!!!!!!!!!
								// !!!! NOT DONE YET !!!!!!
								// !!!!!!!!!!!!!!!!!!!!!!!!
								err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table), obj.DBID(), val.DBID(), key.DBID())					 
								if err != nil {
								   return
							    }								
							}
						}
					} else {
						i := 0
						for val := range collection.Iter(nil) {
							err = db.EnsurePersisted(th, val)
							if err != nil {
								return
							}
							if collection.IsOrdered() {
								err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table), obj.DBID(), val.DBID(), i)
								if err != nil {
								   return
							    }	

							} else { // unordered set 
								err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), obj.DBID(), val.DBID())	
								if err != nil {
								   return
							    }										
							}
							i++
						}
					}
				} else { // a single non-primitive value or independent collection of non-primitive element type.

					err = db.EnsurePersisted(th, val)
					if err != nil {
						return
					}
					err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), obj.DBID(), val.DBID())
					if err != nil {
						return
					}					
				}
			} else if attr.IsComplex() {  // multi-valued primitive type or independent collection of primitive element type

			    val, found := RT.AttrValue(th, obj, attr, false, true, true)
			    if !found {
			   	    continue
			    }
				
				table := db.db.TableNameIfy(attr.ShortName())	

		        if attr.IsMultiValued() { // collection of primitives owned by the "whole" object

		   			valCols,valVars := attr.Part.Type.DbCollectionColumnInsert()
									   
		   			collection := val.(RCollection)
		   			isMap := collection.IsMap()
		   			if isMap {
					   
		   			   // !!!!!!!!!!!!!!!!!!!!!!!!
		   			   // !!!! NOT DONE YET !!!!!!
		   			   // !!!!!!!!!!!!!!!!!!!!!!!!			   
					   
					   
		   				theMap := collection.(Map)
		   				for key := range theMap.Iter(nil) {
		   					val, _ := theMap.Get(key)

		   					if attr.Part.CollectionType == "stringmap" || attr.Part.CollectionType == "orderedstringmap" {
		   						stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,key1) VALUES(?,%s,?)", table, valCols, valVars))
								stmt.Arg(obj.DBID())
		   					    valParts := db.db.primitiveValSQL(val) 
		   					    stmt.Args(valParts)								
								
		   		                stmt.Arg(SqlStringValueEscape(string(key.(String))))
				            
		   						err = db.ExecStatements(stmt)
								if err != nil {
									return
								}		


		   					} else {
		   						// TODO - We do not know if the key is persisted. We don't know if the key is an integer!!!
		   						// !!!!!!!!!!!!!!!!!!!!!!!!
		   						// !!!! NOT DONE YET !!!!!!
		   						// !!!!!!!!!!!!!!!!!!!!!!!!
													
		                       stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,ord1) VALUES(?,%s,?)", table, valCols, valVars))
							   stmt.Arg(obj.DBID())							
		   					   valParts := db.db.primitiveValSQL(val) 
		   					   stmt.Args(valParts)
				   			   stmt.Arg(key.DBID())         						
								
		   						err = db.ExecStatements(stmt)						
								if err != nil {
									return
								}													 
		   					}
		   				}
		   			} else {
		   				i := 0					
		   				for val := range collection.Iter(nil) {
							
		   					if collection.IsOrdered() {					   
		   						stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s,ord1) VALUES(?,%s,?)", table, valCols, valVars))
								
							    stmt.Arg(obj.DBID())									
		   						valParts := db.db.primitiveValSQL(val) 
		   						stmt.Args(valParts)
								stmt.Arg(i)
		   						err = db.ExecStatements(stmt)	
								if err != nil {
									return
								}

		   					} else { // unordered set 
							   
		   						stmt := Stmt(fmt.Sprintf("INSERT INTO %s(id,%s) VALUES(?,%s)", table, valCols, valVars))		
								stmt.Arg(obj.DBID())					
		   					    valParts := db.db.primitiveValSQL(val) 
		   					    stmt.Args(valParts)						
								
		   						err = db.ExecStatements(stmt)	
								if err != nil {
									return
								}		   											
		   					}
		   					i++
		   				}
		   			}							
				} else {  // attr.IsIndependentCollection()  // independent collection of primitive element type

					err = db.EnsurePersisted(th, val)
					if err != nil {
						return
					}
				    err = db.ExecStatement(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table), obj.DBID(), val.DBID())
					if err != nil {
						return
					}				    
				}
				
				   
			}
		}
	}
	return
}










/*
   Persist an independent collection (list, set, or map)
   Assumes the RObject core of the collection is already persisted.
*/
func (db *SqliteDBThread) persistCollection(th InterpreterThread, collection RCollection) (err error) {

   // Derive a collection table name from the collection's
   // collection-type and element type.
   //
   // Need to ensure the collection table exists in the db
   //
   // Return metadata about the collection, including the table name.
   // 
   table,isMap,isOrdered,keyType,elementType,err := db.EnsureCollectionTable(collection)
   if err != nil {
      return
   }	
   	
   if !elementType.IsPrimitive {
		if isMap {
			theMap := collection.(Map)		
			
			mapStmt := Stmt(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table))		
			stringMapStmt := Stmt(fmt.Sprintf("INSERT INTO %s(id0,id1,key1) VALUES(?,?,?)", table)) 	

			for key := range theMap.Iter(nil) {
				val, _ := theMap.Get(key)
				err = db.EnsurePersisted(th, val)
				if err != nil {
					return
				}
				stmt := mapStmt		
			    switch keyType {  
			      case StringType:
				   	stmt = stringMapStmt  
				    stmt.ClearArgs()					   		
		  		    keyStr := SqlStringValueEscape(string(key.(String)))   
		  		   	stmt.Arg(collection.DBID())
                    stmt.Arg(val.DBID()) 
		  		   	stmt.Arg(keyStr)
			   	case UintType:
				    stmt.ClearArgs()			   		
				   	stmt.Arg(collection.DBID())
                    stmt.Arg(val.DBID()) 				   	
				   	stmt.Arg(int64(uint64(key.(Uint))))   // val is actually the map key
			   	case Uint32Type:
				    stmt.ClearArgs()			   		
				   	stmt.Arg(collection.DBID())
                    stmt.Arg(val.DBID()) 				   	
				   	stmt.Arg(int(uint32(key.(Uint32))))   // val is actually the map key			   	
			   	case IntType:
				    stmt.ClearArgs()			   		
				   	stmt.Arg(collection.DBID())
                    stmt.Arg(val.DBID()) 				   	
				   	stmt.Arg(int64(key.(Int)))   // val is actually the map key
			   	case Uint32Type:
			   		stmt.ClearArgs()
				   	stmt.Arg(collection.DBID())
                    stmt.Arg(val.DBID()) 				   	
				   	stmt.Arg(int(key.(Int32)))   // val is actually the map key		   	   		          
		        default:
		         	stmt.ClearArgs()
				   	stmt.Arg(collection.DBID())
				   	stmt.Arg(val.DBID())  
				   	stmt.Arg(key.DBID())
		        }
				err = db.ExecStatements(stmt) 
				if err != nil {
					return
				}
			}
		} else {
			orderedStmt := Stmt(fmt.Sprintf("INSERT INTO %s(id0,id1,ord1) VALUES(?,?,?)", table))
			unorderedStmt := Stmt(fmt.Sprintf("INSERT INTO %s(id0,id1) VALUES(?,?)", table))	

			i := 0
			for val := range collection.Iter(nil) {
				err = db.EnsurePersisted(th, val)
				if err != nil {
					return
				}

				if isOrdered {
					stmt := orderedStmt
					stmt.ClearArgs()
					stmt.Arg(collection.DBID())
					stmt.Arg(val.DBID())
					stmt.Arg(i)
					err = db.ExecStatements(stmt)
					if err != nil {
						return
					}
				} else { // unordered set 
					stmt := unorderedStmt
					stmt.ClearArgs()
					stmt.Arg(collection.DBID())
					stmt.Arg(val.DBID())
					err = db.ExecStatements(stmt)	

					if err != nil {
						return
					}				
			    }
				i++
			}
		}

   } else { // a collection of primitive-type objects
   	// TODO
   	// !!!!!!!!!!!!!!!!!!!!!!!!
   	// !!!! NOT DONE YET !!!!!!
   	// !!!!!!!!!!!!!!!!!!!!!!!!
			
   	valCols,valVars := elementType.DbCollectionColumnInsert()
					   
   	if isMap {
	   
   	   // !!!!!!!!!!!!!!!!!!!!!!!!
   	   // !!!! NOT DONE YET !!!!!!
   	   // !!!!!!!!!!!!!!!!!!!!!!!!			   
	   
   		theMap := collection.(Map)
   			
   		stringMapStmt := Stmt(fmt.Sprintf("INSERT INTO %s(%s,id,key1) VALUES(%s,?,?)", table, valCols, valVars)) 
        mapStmt := Stmt(fmt.Sprintf("INSERT INTO %s(%s,id,ord1) VALUES(%s,?,?)", table, valCols, valVars))

   		for key := range theMap.Iter(nil) {
   			val, _ := theMap.Get(key)
   			valParts := db.db.primitiveValSQL(val) 

			stmt := mapStmt			 
		    switch keyType {  
		      case StringType:
			   	stmt = stringMapStmt  
			   	stmt.ClearArgs()	
			   	stmt.Args(valParts)	
	  		   	stmt.Arg(collection.DBID())			   	
	  		    keyStr := SqlStringValueEscape(string(key.(String)))   
	  		   	stmt.Arg(keyStr)
		   	case UintType:
		   		stmt.ClearArgs()
			   	stmt.Args(valParts)	
	  		   	stmt.Arg(collection.DBID())	                			   	
			   	stmt.Arg(int64(uint64(key.(Uint))))   // val is actually the map key
		   	case Uint32Type:
		   		stmt.ClearArgs()
			   	stmt.Args(valParts)	                	
	  		   	stmt.Arg(collection.DBID())	                			   	
			   	stmt.Arg(int(uint32(key.(Uint32))))   // val is actually the map key			   	
		   	case IntType:
		   		stmt.ClearArgs()
 			   	stmt.Args(valParts)	               
	  		   	stmt.Arg(collection.DBID())	                			   	
			   	stmt.Arg(int64(key.(Int)))   // val is actually the map key
		   	case Uint32Type:
		   		stmt.ClearArgs()
			   	stmt.Args(valParts)	
	  		   	stmt.Arg(collection.DBID())	                			   	
			   	stmt.Arg(int(key.(Int32)))   // val is actually the map key		   	   		          
	         default:
	         	stmt.ClearArgs()
			   	stmt.Args(valParts)				   	
	  		   	stmt.Arg(collection.DBID())				   	
			   	stmt.Arg(key.DBID())
	         }
			 err = db.ExecStatements(stmt) 
			 if err != nil {
				return
			 }
   		}
   	} else {


   		orderedStmt := Stmt(fmt.Sprintf("INSERT INTO %s(%s,id,ord1) VALUES(%s,?,?)", table, valCols, valVars))
   		unorderedStmt := Stmt(fmt.Sprintf("INSERT INTO %s(%s,id) VALUES(%s,?)", table, valCols, valVars))		

   		i := 0					
   		for val := range collection.Iter(nil) {
   			valParts := db.db.primitiveValSQL(val) 			
   			if isOrdered {					   
   				stmt := orderedStmt
   				stmt.ClearArgs()
   				stmt.Args(valParts)
    			stmt.Arg(collection.DBID())	
    			stmt.Arg(i)	  					
				err = db.ExecStatements(stmt) 
				if err != nil {
				   return
				}


   			} else { // unordered set 
   			    stmt := unorderedStmt	
   			    stmt.ClearArgs()
   			    stmt.Args(valParts)		
   			    stmt.Arg(collection.DBID())				
				err = db.ExecStatements(stmt) 
				if err != nil {
				   return
				}				
   			}
   			i++
   		}
   	}				
   }
   return
}

















/*
   NOT TESTED
   Persist a remote object which has not yet been persisted locally.
   TODO This does not check whether the object is already in the local db, and if it is,
   problems will occur!
   What should we do in that case?
   Check in this method and fetch and return the old object?
   Throw an exception?
   Right now it will probably create a second stored copy of the object with id reversed!!!
*/
func (db *SqliteDBThread) persistRemoteObject(th InterpreterThread, obj RObject) (err error) {

	obj.ClearIdReversed()
	id, id2 := obj.UUIDuint64s()

	dbid := int64(id)
	dbid2 := int64(id2)

	found, err := db.exists(dbid)
	if err != nil {
		return
	}
	if found {
		dbid, dbid2 = dbid2, dbid
		obj.SetIdReversed()
		found, err = db.exists(dbid)
		if err != nil {
			return
		}
		if found {
			rterr.Stop("Can't store remote object locally because of db id conflict.") // Chance of this is extremely small. Two 64 bit int random ids had to be present twice. 
		}
	}
	db.insert(th, obj, dbid, dbid2)
	return
}

/*
   Persist a newly created object. Creates a uuid for it.
*/
func (db *SqliteDBThread) persistNewObject(th InterpreterThread, obj RObject) (err error) {
	for { // loop til we create a first half of uuid which is unique as an object id in the local db. 
		// Should almost never require more than one attempt.
		var id,id2 uint64
		id, id2, err = obj.EnsureUUIDuint64s()
		if err != nil {
			return
		}
		dbid := int64(id)
		var found bool
		found, err = db.exists(dbid)
		if err != nil {
			return
		}
		if found {
			obj.RemoveUUID()
		} else {
			err = db.insert(th, obj, dbid, int64(id2))
            if err != nil {
            	return
            }
			// obj.SetStoredLocally() // We don't actually know if this is correct yet. The db statements may have failed. TODO!!! FIX

			break
		}
	}
	return
}

/*
   Returns true if the object is found in the database.
*/
func (db *SqliteDBThread) exists(id int64) (found bool, err error) {
	stmt := "SELECT count(*) FROM RObject where id=?"
	selectStmt, err := db.Prepare(stmt)
	if err != nil {
		return
	}

	defer selectStmt.Reset()

	err = selectStmt.Query(id)
    if err != nil {
    	if err == io.EOF {
    	   panic("exists: Why is there no result of a count(*) statement???")
        } else {
        	return
        }
    }

	var n int64

	err = selectStmt.Scan(&n)
	if err != nil {
		return
	}
	if n >= 1 {
		found = true
		if n > 1 {
			rterr.Stop("More than one object with same id in sqlite db!")
		}
	}

	return
}

/*
   Gives the (already persisted) object an official name in the database.
*/
func (db *SqliteDBThread) NameObject(obj RObject, name string) (err error) {
	id := obj.DBID()
	name = SqlStringValueEscape(name)
	stmt := Stmt("INSERT INTO RName(name,id) VALUES(?,?);")
	stmt.Arg(name)
	stmt.Arg(id)
	err = db.ExecStatements(stmt)
	return
}



func (db *SqliteDBThread) RenameObject(oldName string, newName string) (err error) {
	oldName = SqlStringValueEscape(oldName)
	newName = SqlStringValueEscape(newName)	
	stmt := Stmt("UPDATE RName set name=? WHERE name=?;")
	stmt.Arg(newName)
	stmt.Arg(oldName)
	err = db.ExecStatements(stmt)
	return
}

/*
Returns true if an object has been named in the database with the argument name.
*/
func (db *SqliteDBThread) ObjectNameExists(name string) (found bool, err error) {
	name = SqlStringValueEscape(name)	
	stmt := "SELECT count(*) FROM RName where name=?"
	selectStmt, err := db.Prepare(stmt)
	if err != nil {
		return
	}

	defer selectStmt.Reset()  // Ensure the statement is not left open 	

	err = selectStmt.Query(name)
    if err != nil {
    	if err == io.EOF {
    	   panic("exists: Why is there no result of a count(*) statement???")
        } else {
        	return
        }
    }

	var n int64

	err = selectStmt.Scan(&n)
	if err != nil {
		return
	}
	found = (n > 0)

	return
}




func (db * SqliteDBThread) ObjectNames(prefix string) (names []string, err error) {
   if prefix == "" {
		stmt := "SELECT name FROM RName order by name"
		selectStmt, dbErr := db.Prepare(stmt)
		if dbErr != nil {
			err = dbErr
			return
		}

		defer selectStmt.Reset()  // Ensure the statement is not left open 

		err = selectStmt.Query()
		if err != nil && err != io.EOF {
			return
		}

        for ; err == nil ; err = selectStmt.Next() {   

			var name string

			err = selectStmt.Scan(&name)
			if err != nil {
				return
			}
			names = append(names,name)
		}
		if err == io.EOF {
		 err = nil 
		} else {
		 err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", stmt, err)   
		 return  
		} 

   } else {
        prefix = SqlStringValueEscape(prefix) + "%"
		stmt := "SELECT name FROM RName where name like ? order by name"
		selectStmt, dbErr := db.Prepare(stmt)
		if dbErr != nil {
			err = dbErr
			return
		}

		defer selectStmt.Reset()  // Ensure the statement is not left open 

		err = selectStmt.Query(prefix)
		if err != nil && err != io.EOF {
			return
		}

        for ; err == nil; err = selectStmt.Next() { 		

			var name string

			err = selectStmt.Scan(&name)
			if err != nil {
				return
			}
			names = append(names,name)
		}
		if err == io.EOF {
		 err = nil 
		} else {
		 err = fmt.Errorf("DB ERROR on query:\n%s\nDetail: %s\n\n", stmt, err)   
		 return  
		} 
   }
   return
}


/*
TODO Better error reporting
*/
func (db *SqliteDBThread)  Delete(obj RObject) (err error) {

    if ! obj.IsBeingStored() {
    	return
    }

    id := obj.DBID()

	stmts := Stmt("DELETE FROM RObject WHERE id=?;")
	stmts.Arg(id)
	stmts.Add("DELETE FROM RName WHERE id=?;")
	stmts.Arg(id)	


	objTyp := obj.Type()

	typeTable := db.db.TableNameIfy(objTyp.ShortName())
	stmts.Add(fmt.Sprintf("DELETE FROM %s WHERE id=?;",typeTable))	
	stmts.Arg(id)	
	for _, typ := range objTyp.Up {
		typeTable = db.db.TableNameIfy(typ.ShortName())
	    stmts.Add(fmt.Sprintf("DELETE FROM %s WHERE id=?;",typeTable))
	    stmts.Arg(id)
	}

	err = db.ExecStatements(stmts)
    if err != nil {
    	return
    }
	obj.ClearStoredLocally()
	obj.ClearBeingStored()

    RT.Uncache(obj)

	return
}

/*
   Given the local dbid of an object, retrieve and return the object stored in the database.
   If the object is a unit (structure), the radius determines how many related objects in the tree are also fetched.
   0 means fetch object only. 1 means referred-to objects, 2 means referred-to by referred to etc.

   When a collection of non-primitives is fetched, it is lazily fetched. Currently that means that proxy
   objects are created in memory for each element of the collection.
   So for example, if you have an object with a multi-valued attribute, and you want the actual other objects fetched
   along with the object, use a depth of 2.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDBThread) Fetch(id int64, radius int) (obj RObject, err error) {
	defer Un(Trace(PERSIST_TR, "Fetch", id, radius))
	var found bool
	obj, found = RT.GetObject(id)
	if found {
		// fmt.Printf("Fetch: found object %s in cache.\n",obj.Debug())
		return
	}
	stmt := "SELECT * FROM RObject where id=?"
	obj, err = db.fetch1(stmt, id, radius, fmt.Sprintf("id=%v", id), false)
	return 
}

/*
   TODO Refresh the state of the object from the object data stored in the database.
   If the object is a unit (structure), the radius determines how many related objects in the tree are also fetched.
   0 means fetch object only. 1 means referred-to objects, 2 means referred-to by referred to etc.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDBThread) Refresh(obj RObject , radius int) (err error) {
	defer Un(Trace(PERSIST_TR, "Refresh", obj.DBID(), radius))

	id := obj.DBID()
	errSuffix := fmt.Sprintf("id=%v", id)
    query := "SELECT * FROM RObject where id=?"

	selectStmt, err := db.Prepare(query)
	if err != nil {
		return
	}

	defer selectStmt.Reset() // Ensure statement is not left open

	err = selectStmt.Query(id)
    if err != nil {
    	if err == io.EOF {
		   panic(fmt.Sprintf("No object found in database with %s.", errSuffix))    		
        } else {
           return
        }
    }

	// var id int64
	var id2 int64
	var flags int
	var typeName string

	err = selectStmt.Scan(&id, &id2, &flags, &typeName)
	if err != nil {
		return
	}


   var isCollection bool 
      
   if typeName[0] == '[' {
        isCollection = true
   }

	// Now we have to store the unit64(id),uint64(id2),byte(flags) into the object.

	//   unit := obj.(*runit)
	//   (&(unit.robject)).RestoreIdsAndFlags(id,id2,flags)

	ob := obj.(Persistable)
	ob.RestoreIdsAndFlags(id, id2, flags)

	Logln(PERSIST2_, "id:", id, ", id2:", id2, ", flags:", flags, ", typeName:", typeName)

	oid, oid2 := obj.UUIDuint64s()

	Logln(PERSIST2_, "obj.id:", oid, ", obj.id2:", oid2, ", Flags():", obj.Flags(), ", obj.Type():", obj.Type())

	// fmt.Printf("fetch1: cache miss w. DBID %d. created object %s from db. Its DBID is %d\n",dbid,obj.Debug(),obj.DBID())


	
	// Now fetch and restore the values of the unary primitive attributes of the object.
	// These attribute values are stored in the db rows that represent the object in the db.
	// The object in the db consists of a single row in each of several database tables.
	// There is one table for each type the object conforms to (i.e. specific type and supertypes), 
	// and a single row in each such table identified by the object's dbid.

	err = db.fetchUnaryPrimitiveAttributeValues(id, obj)
	if err != nil {
		return
	}


	// Now fetch (at least proxies for) the non-primitive attributes (if we should do it now.)
	// Maybe this should be fully lazy. Wait until the attribute value is asked for.

	// TODO

	// THIS NEEDS TO DEPEND ON DEPTH

	// TODO

   if isCollection {
      collection := obj.(RCollection)
      if collection.ElementType().IsPrimitive {
   	   err = db.fetchPrimitiveValueCollection(collection, obj.DBID(), typeName)	   
         if err != nil {
         	return
         }	          
      } else {
      	err = db.fetchCollection(collection,  obj.DBID(), typeName, radius)	   
         if err != nil {
         	return
      	}
      }	   
   } else if radius > 0 {
		err = db.fetchUnaryNonPrimitiveAttributeValues(id, obj, radius-1)
		if err != nil {
			return
		}

		err = db.fetchMultiValuedNonPrimitiveAttributeValues(id, obj, radius-1)
		if err != nil {
			return
		}

		err = db.fetchMultiValuedPrimitiveAttributeValues(id, obj)
		if err != nil {
			return
		}
	}
	return
}






/*
   Given a name that some object has been 'dubbed' with, retrieve and return the object stored in the database.
   If the object is a unit (structure), the radius determines how many related objects in the tree are also fetched.
   0 means fetch object only. 1 means referred-to objects, 2 means referred-to by referred to etc.

   When a collection of non-primitives is fetched, it is lazily fetched. Currently that means that proxy
   objects are created in memory for each element of the collection.
   So for example, if you have an object with a multi-valued attribute, and you want the actual other objects fetched
   along with the object, use a depth of 2.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDBThread) FetchByName(name string, radius int) (obj RObject, err error) {
	defer Un(Trace(PERSIST_TR, "FetchByName", name, radius))
	name = SqlStringValueEscape(name)	
	stmt := "SELECT * FROM RObject WHERE id IN (SELECT id FROM RName WHERE name=?)"
	//fmt.Printf("FetchByName:  %s\n",name)	
	return db.fetch1(stmt, name, radius, fmt.Sprintf("name='%s'", name), true)
	
	
}

/*
   Give the dbid of an object, fetches the value of the specified attribute of the object from the database.
   The attribute should have a non-primitive value type or be multi-valued with primitive type of collection members.
*/
func (db *SqliteDBThread) FetchAttribute(th InterpreterThread, objId int64, obj RObject, attr *AttributeSpec, radius int) (val RObject, err error) {
	defer Un(Trace(PERSIST_TR, "FetchAttribute", objId, radius))

	if (attr.Part.CollectionType == "" && !attr.Part.Type.IsPrimitive) || attr.IsIndependentCollection() {
		err = db.fetchUnaryNonPrimitiveAttributeValue(objId, obj, attr, radius)
		if err != nil {
			return
		}
	} else if attr.Part.CollectionType != "" && !attr.Part.Type.IsPrimitive {
		err = db.fetchNonPrimitiveAttributeValueCollection(objId, obj, attr, radius)
		if err != nil {
			return
		}
	} else if attr.Part.CollectionType != "" && attr.Part.Type.IsPrimitive {
		err = db.fetchPrimitiveAttributeValueCollection(objId, obj, attr)
		if err != nil {
			return
		}
	}
	val, _ = RT.AttrVal(th, obj, attr)
	return
}

/*
   TODO Retrieve the object stored in the database.
   If the object is a unit (structure), the radius determines how many related objects in the tree are also fetched.
   0 means fetch object only. 1 means referred-to objects, 2 means referred-to by referred to etc.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDBThread) fetch1(query string, arg interface{}, radius int, errSuffix string, checkCache bool) (obj RObject, err error) {
	selectStmt, err := db.Prepare(query)
	if err != nil {
		return
	}

	defer selectStmt.Reset() // Ensure statement is not left open

	err = selectStmt.Query(arg)
    if err != nil {
    	if err == io.EOF {
		   err = fmt.Errorf("No object found in database with %s.", errSuffix)		
        }
           return
        }

	var id int64
	var id2 int64
	var flags int
	var typeName string

	err = selectStmt.Scan(&id, &id2, &flags, &typeName)
	if err != nil {
		return
	}

    var dbid int64  // Made it a var for later debug print. Undo this? put inside if with :=
	if checkCache {
		dbid = DBID(id, id2, flags)

		var found bool
		obj, found = RT.GetObject(dbid)
		if found {
		    // fmt.Printf("fetch1: using dbid %d found object %s in cache.\n",dbid, obj.Debug())			
			return
		}
	}

   var isCollection bool 
      
   if typeName[0] == '[' {
      isCollection = true
    	obj, err = RT.NewCollectionFromDB(typeName)
    	if err != nil {
    		return
    	}  
    	
    	// TODO Now load the collection using the typeName as the collection table name    
      
   } else {
   
      // If I am fetching a data object for which I have not loaded into runtime
      // the package in which its datatype is defined, I load that package now.
      // Note that the package must have been loaded previously in SOME run of the program
      // that used the same local database, so that the package's short name has been recorded 
      // in the artifact local database.
    
      TypesMutex.RLock()
      typ := RT.Typs[typeName]
      TypesMutex.RUnlock()
      if typ == nil {
   
         fmt.Println("typeName",typeName)
         pkgShortName := PackageShortName(typeName)
         fmt.Println("pkgShortName",pkgShortName)         
   //      localTypeName := LocalTypeName(typeName)   
         pkgFullName := RT.PkgShortNameToName[pkgShortName]
         fmt.Println("pkgFullName",pkgFullName)         
         originAndArtifact := OriginAndArtifact(pkgFullName) 
         packagePath := LocalPackagePath(pkgFullName)      
      
         // TODO Dubious values of version and mustBeFromShared here!!!
         err = RT.Loader.LoadRelishCodePackage(originAndArtifact,"",packagePath,false)
     
         if err != nil {
            return
         }
         
         TypesMutex.RLock()
         typ = RT.Typs[typeName]    
         TypesMutex.RUnlock()
          
         // Alternate strategy!!    
   	   // rterr.Stop("Can't summon object. The package which defines its type, '%s', has not been loaded into the runtime.",localTypeName) 
       }
       fullTypeName := typ.Name

   	obj, err = RT.NewObject(fullTypeName)
   	if err != nil {
   		return
   	}
   }

	// Now we have to store the unit64(id),uint64(id2),byte(flags) into the object.

	//   unit := obj.(*runit)
	//   (&(unit.robject)).RestoreIdsAndFlags(id,id2,flags)

	ob := obj.(Persistable)
	ob.RestoreIdsAndFlags(id, id2, flags)

	Logln(PERSIST2_, "id:", id, ", id2:", id2, ", flags:", flags, ", typeName:", typeName)

	oid, oid2 := obj.UUIDuint64s()

	Logln(PERSIST2_, "obj.id:", oid, ", obj.id2:", oid2, ", Flags():", obj.Flags(), ", obj.Type():", obj.Type())

	// fmt.Printf("fetch1: cache miss w. DBID %d. created object %s from db. Its DBID is %d\n",dbid,obj.Debug(),obj.DBID())

	
	RT.Cache(obj) // Put in an in-memory object cache so that the runtime will only contain one object instance for each uuid.
	// fmt.Println("fetch1: cached it.")

	
	
	// Now fetch and restore the values of the unary primitive attributes of the object.
	// These attribute values are stored in the db rows that represent the object in the db.
	// The object in the db consists of a single row in each of several database tables.
	// There is one table for each type the object conforms to (i.e. specific type and supertypes), 
	// and a single row in each such table identified by the object's dbid.

	err = db.fetchUnaryPrimitiveAttributeValues(id, obj)
	if err != nil {
		return
	}

	// Have to set this here before confirmed in order to avoid attribute or relation reference loops causing
	// infinite looping during fetching. 
	// TODO consider replacing with SetStoringLocally and a later SetStoredLocally

	obj.SetStoredLocally()

	// Now fetch (at least proxies for) the non-primitive attributes (if we should do it now.)
	// Maybe this should be fully lazy. Wait until the attribute value is asked for.

	// TODO

	// THIS NEEDS TO DEPEND ON DEPTH

	// TODO

   if isCollection {
      collection := obj.(RCollection)
      if collection.ElementType().IsPrimitive {
   	   err = db.fetchPrimitiveValueCollection(collection, obj.DBID(), typeName)	   
         if err != nil {
         	return
         }	          
      } else {
      	err = db.fetchCollection(collection,  obj.DBID(), typeName, radius)	   
         if err != nil {
         	return
      	}
      }	   
   } else if radius > 0 {
		err = db.fetchUnaryNonPrimitiveAttributeValues(id, obj, radius-1)
		if err != nil {
			return
		}

		err = db.fetchMultiValuedNonPrimitiveAttributeValues(id, obj, radius-1)
		if err != nil {
			return
		}

		err = db.fetchMultiValuedPrimitiveAttributeValues(id, obj)
		if err != nil {
			return
		}
	}
	return
}




/*
   TODO Retrieve a list of objects stored in the database.

   TODO THIS METHOD NEEDS TO HAVE A MUTEX LOCK!!!!!!!

   TODO THIS METHOD NEEDS TO FLUSH THE SQL STATEMENT QUEUE BEFORE IT RUNS THE SELECT QUERY!!!!!!!
*/
func (db *SqliteDBThread) fetchMultiple(query string, queryArgObjs []RObject, idsOnly bool, radius int, numPrimitiveAttrs int, errSuffix string, checkCache bool, objs *[]RObject) (err error) {

	Logln(PERSIST_, query)

	selectStmt, err := db.Prepare(query)
	if err != nil {
		return
	}

	defer selectStmt.Reset()
	
	var queryArgs []interface{}
	if len(queryArgObjs) > 0 {
		queryArgs = make([]interface{},len(queryArgObjs))
		for i,arg := range queryArgObjs {
			var convertedArg interface{}
			switch arg.(type) {
			case Bool:
               convertedArg = bool(arg.(Bool))  
			default:
			   convertedArg = fmt.Sprint(arg)
			}
			queryArgs[i] = convertedArg
		}
	}

	err = selectStmt.Query(queryArgs...)  // Exec(args...)
	if err != nil && err != io.EOF {
		return
	}

    for ; err == nil ; err = selectStmt.Next() {   
	
	    var obj RObject
		var id int64
			
	    if idsOnly {

			err = selectStmt.Scan(&id)
			if err != nil {
				return
			}
					
			if radius > 0 { // fetch the full objects
				obj, err = db.Fetch(id, radius-1)
				if err != nil {
					return
				}
			} else { // Just put proxy objects into the collection.
				obj = Proxy(id)
			}		
		
	    } else { // The query fetched all primitive attributes of the objects.
		
			var id2 int64
			var flags int
			var typeName string

			attrValsBytes1 := make([][]byte, numPrimitiveAttrs)

			attrValsBytes := make([]interface{}, numPrimitiveAttrs + 4)

	        attrValsBytes[0] = &id
	        attrValsBytes[1] = &id2
	        attrValsBytes[2] = &flags
	        attrValsBytes[3] = &typeName

	 		for i := 0; i < len(attrValsBytes1); i++ {
				attrValsBytes[i+4] = &attrValsBytes1[i]
			}


			err = selectStmt.Scan(attrValsBytes...)
		
			if err != nil {
				return
			}
		
			if checkCache {
				dbid := DBID(id, id2, flags)

				var found bool
				obj, found = RT.GetObject(dbid)
				if found {
		        	*objs = append(*objs, obj)				
					continue    // to next object in the resultset
				}
			}

                    TypesMutex.RLock()
		    typ := RT.Typs[typeName]
                    TypesMutex.RUnlock()
		    if typ == nil {
		   
		      pkgShortName := PackageShortName(typeName)  
		//      localTypeName := LocalTypeName(typeName)   
		      pkgFullName := RT.PkgShortNameToName[pkgShortName]
		      originAndArtifact := OriginAndArtifact(pkgFullName) 
		      packagePath := LocalPackagePath(pkgFullName)      
		      
		      // TODO Dubious values of version and mustBeFromShared here!!!
		      err = RT.Loader.LoadRelishCodePackage(originAndArtifact,"",packagePath,false)
		     
		      if err != nil {
		         return
		      }
		         
                    TypesMutex.RLock()
                    typ = RT.Typs[typeName]    		  
                    TypesMutex.RUnlock()
                     
		      // Alternate strategy!!    
			   // rterr.Stop("Can't summon object. The package which defines its type, '%s', has not been loaded into the runtime.",localTypeName) 
		    }
		    fullTypeName := typ.Name

		    obj, err = RT.NewObject(fullTypeName)
			if err != nil {
				return
			}

			// Now we have to store the unit64(id),uint64(id2),byte(flags) into the object.

			//   unit := obj.(*runit)
			//   (&(unit.robject)).RestoreIdsAndFlags(id,id2,flags)

			ob := obj.(Persistable)
			ob.RestoreIdsAndFlags(id, id2, flags)

			Logln(PERSIST2_, "id:", id, ", id2:", id2, ", flags:", flags, ", typeName:", typeName)

			oid, oid2 := obj.UUIDuint64s()

			Logln(PERSIST2_, "obj.id:", oid, ", obj.id2:", oid2, ", Flags():", obj.Flags(), ", obj.Type():", obj.Type())

            if numPrimitiveAttrs > 0 {
				// Now restore the values of the unary primitive attributes of the object.
				// These attribute values are stored in the db rows that represent the object in the db.
				// The object in the db consists of a single row in each of several database tables.
				// There is one table for each type the object conforms to (i.e. specific type and supertypes), 
				// and a single row in each such table identified by the object's dbid.
			
			    objTyp := obj.Type()
			    attrValsBytes = attrValsBytes[4:]
		        db.db.restoreAttrs(obj, objTyp, attrValsBytes)		
            }

			// Have to set this here before confirmed in order to avoid attribute or relation reference loops causing
			// infinite looping during fetching. 
			// TODO consider replacing with SetStoringLocally and a later SetStoredLocally
            
			obj.SetStoredLocally()
			
			
			RT.Cache(obj) // Put in an in-memory object cache so that the runtime will only contain one object instance for each uuid.
			// fmt.Printf("fetchMultiple: cached %s with DBID %d.\n",obj.Debug(), obj.DBID())			
			
			

			// Now fetch (at least proxies for) the non-primitive attributes (if we should do it now.)
			// Maybe this should be fully lazy. Wait until the attribute value is asked for.

			// TODO

			// THIS NEEDS TO DEPEND ON DEPTH

			// TODO IS THIS DEPTH CORRECT ????????
    
			if radius > 1 {
				err = db.fetchUnaryNonPrimitiveAttributeValues(id, obj, radius-2)
				if err != nil {
					return
				}

				err = db.fetchMultiValuedNonPrimitiveAttributeValues(id, obj, radius-2)
				if err != nil {
					return
				}

				err = db.fetchMultiValuedPrimitiveAttributeValues(id, obj)
				if err != nil {
					return
				}
			}
		}	
		
		*objs = append(*objs, obj)
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
Fetch from db and set unary primitive-valued attributes of an object.
*/
func (db *SqliteDBThread) fetchUnaryPrimitiveAttributeValues(id int64, obj RObject) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchUnaryPrimitiveAttributeValues", id))
	// TODO

	objTyp := obj.Type()

	// Create the select clause

    numPrimAttributeColumns := 0

	selectClause := "SELECT "
	sep := ""
	for _, attr := range objTyp.Attributes {
		if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
			if attr.Part.Type == TimeType {
			   selectClause += sep + attr.Part.Name	+ "," + attr.Part.Name + "_loc" 
			   numPrimAttributeColumns += 2							
			} else if attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
			   selectClause += sep + attr.Part.Name	+ "_r," + attr.Part.Name + "_i" 	
			   numPrimAttributeColumns += 2
			} else if attr.Part.Type == MutexType ||  attr.Part.Type == RWMutexType || attr.Part.Type == OwnedMutexType {
				// ignore
			} else {
			   selectClause += sep + attr.Part.Name
			   numPrimAttributeColumns ++
		    }
			sep = ","
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
				if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
				if attr.Part.Type == TimeType {
				   selectClause += sep + attr.Part.Name	+ "," + attr.Part.Name + "_loc" 
				   numPrimAttributeColumns += 2							
				} else if attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
				   selectClause += sep + attr.Part.Name	+ "_r," + attr.Part.Name + "_i" 	
				   numPrimAttributeColumns += 2
			    } else if attr.Part.Type == MutexType ||  attr.Part.Type == RWMutexType || attr.Part.Type == OwnedMutexType {
				   // ignore				   
				} else {
				   selectClause += sep + attr.Part.Name
				   numPrimAttributeColumns ++
			    }
				sep = ","				
			}
		}
	}

	if numPrimAttributeColumns > 0 {

		// Figure out the total number of primitive attributes in all the types combined.

		// numPrimitiveAttrs := objTyp.NumPrimitiveAttributes

		// Now for the object's type and all types in the upchain, we need to 
		// create a join statement on the type tables, then
		// collect the primitive attributes, in the order they were put into the type tables.  

		specificTypeTable := db.db.TableNameIfy(objTyp.ShortName())
		from := " FROM " + specificTypeTable

		for _, typ := range objTyp.Up {
			from += " JOIN " + db.db.TableNameIfy(typ.ShortName()) + " USING (id)"
			// numPrimitiveAttrs += typ.NumPrimitiveAttributes
		}

		where := fmt.Sprintf(" WHERE %s.id=?", specificTypeTable)

		stmt := selectClause + from + where

		Logln(PERSIST2_, "query:", stmt)

		selectStmt, queryErr := db.Prepare(stmt)
		if queryErr != nil {
			err = queryErr
			return
		}

		defer selectStmt.Reset()

		err = selectStmt.Query(id)
	    if err != nil {
	    	if err == io.EOF {
			   panic(fmt.Sprintf("No object found in database with id=%v", id))  		
	        } else {
	           return
	        }
	    }


		// Now construct a Scan call with [] *[]byte
		/*
		   attrValsBytes1 := make([]*[]byte,numPrimitiveAttrs)

		   attrValsBytes := make([]interface{},numPrimitiveAttrs)

		   for i := 0; i < len(attrValsBytes1); i++ {
		      attrValsBytes[i] = attrValsBytes1[i]	
		   }
		*/

		attrValsBytes1 := make([][]byte, numPrimAttributeColumns)

		attrValsBytes := make([]interface{}, numPrimAttributeColumns)

		for i := 0; i < len(attrValsBytes1); i++ {
			attrValsBytes[i] = &attrValsBytes1[i]
		}

		// have to make a slice whose length = the total number of attributes in all the types combined.
		// 
		err = selectStmt.Scan(attrValsBytes...)
		if err != nil {
			return
		}
		
	    db.db.restoreAttrs(obj, objTyp, attrValsBytes)
    }
	return
}

/*
Given the result of a scan of a select result row, restore object attribute values for an object in the runtime.
*/
func (db *SqliteDB) restoreAttrs(obj RObject, objTyp *RType, attrValsBytes []interface{}) {
	defer Un(Trace(PERSIST_TR2, "restoreAttrs of a", objTyp.Name))

	// Now go through the attrValsBytes and interpret each according to the datatype of each primitive
	// attribute, and set the primitive attributes using the runtime SetAttrVal method.

	i := 0
	var val RObject
	var nonNil bool 

	for _, attr := range objTyp.Attributes {
		if attr.Part.Type.IsPrimitive  && attr.Part.CollectionType == "" && attr.Part.Type != MutexType && attr.Part.Type != RWMutexType  && attr.Part.Type != OwnedMutexType {
			valByteSlice := *(attrValsBytes[i].(*[]byte))
			if attr.Part.Type == TimeType || attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
				i++
				valByteSlice2 := *(attrValsBytes[i].(*[]byte))
     		    nonNil = convertAttrValTwoFields(valByteSlice, valByteSlice2, attr, &val) 				
			} else {
			   nonNil = convertAttrVal(valByteSlice, attr, &val) 
		    }
		    if nonNil {
		   		RT.RestoreAttrNonLocking(obj, attr, val)			    
		    }
		    i++
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
			if attr.Part.Type.IsPrimitive  && attr.Part.CollectionType == "" && attr.Part.Type != MutexType && attr.Part.Type != RWMutexType  && attr.Part.Type != OwnedMutexType {
				valByteSlice := *(attrValsBytes[i].(*[]byte))
				if attr.Part.Type == TimeType || attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
					i++
					valByteSlice2 := *(attrValsBytes[i].(*[]byte))
	     		    nonNil = convertAttrValTwoFields(valByteSlice, valByteSlice2, attr, &val) 				
				} else {
				   nonNil = convertAttrVal(valByteSlice, attr, &val) 
			    }
			    if nonNil {
			   		RT.RestoreAttrNonLocking(obj, attr, val)			    
			    }
			    i++				
			}
		}
	}
}



/*
Fetch from the db the values of non-primitive-typed unary attributes of the object.
Set each attribute to the fetched object.
*/
func (db *SqliteDBThread) fetchUnaryNonPrimitiveAttributeValues(id int64, obj RObject, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchUnaryNonPrimitiveAttributeValues", id))

	objTyp := obj.Type()

	for _, attr := range objTyp.Attributes {
		if (attr.Part.CollectionType == "" && !attr.Part.Type.IsPrimitive) || attr.IsIndependentCollection() {
			err = db.fetchUnaryNonPrimitiveAttributeValue(id, obj, attr, radius)
			if err != nil {
				return
			}
		}
	}

	for _, typ := range objTyp.Up {
		for _, attr := range typ.Attributes {
			if (attr.Part.CollectionType == "" && !attr.Part.Type.IsPrimitive) || attr.IsIndependentCollection() {
				err = db.fetchUnaryNonPrimitiveAttributeValue(id, obj, attr, radius)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

/*
Fetch from the db the object value of a non-primitive-typed unary attribute of the object.
Set the attribute to the fetched object.  
*/
func (db *SqliteDBThread) fetchUnaryNonPrimitiveAttributeValue(objId int64, obj RObject, attr *AttributeSpec, radius int) (err error) {
	defer Un(Trace(PERSIST_TR2, "fetchUnaryNonPrimitiveAttributeValue", objId, attr.ShortName()))

	// First query for the id of the attribute value object

	query := fmt.Sprintf("SELECT id1 FROM %s WHERE id0=?", db.db.TableNameIfy(attr.ShortName()))
    // fmt.Println(query)  // debug
	selectStmt, err := db.Prepare(query)
	if err != nil {
		// panic(err) // debug
		return
	}

	defer selectStmt.Reset()

	err = selectStmt.Query(objId)
    if err != nil {
    	if err == io.EOF {
		   // TODO NOT SURE IF THIS SHOULD BE AN ERROR!!!!!!    		
		   err = errors.New(fmt.Sprintf("Object %s has no value for attribute %s", objId, attr.Part.Name))		
        }  else {
		   err = errors.New(fmt.Sprintf("Error fetching value Object %s attribute %s: %s", objId, attr.Part.Name, err))	        	
        }
        return
    }

	var id1 int64
	err = selectStmt.Scan(&id1)
	if err != nil {
		// panic(err) // debug			
		return
	}

	// Then fetch the object with the id
	val, err := db.Fetch(id1, radius)
	if err != nil {
		// panic(err) // debug				
		return
	}
	RT.RestoreAttr(obj, attr, val) 

	return
}

/*
Convert the byte slice which was returned as a column-value of a databse result row into
a primitive-type RObject. Sets the val argument to the new RObject.
If the value from the database was NULL (empty string in numeric fields), does not
set the val argument, and returns false.

TODO NOT HANDLING NULLS PROPERLY HERE YET !!!!!!!!!

*/
func convertAttrVal(valByteSlice []byte, attr *AttributeSpec, val *RObject) (nonNullValueFound bool) {
	switch attr.Part.Type {
	case IntType:
		if len(valByteSlice) > 0 {
			x, err := strconv.ParseInt(string(valByteSlice), 10, 64)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " as int64: " + err.Error()))
			}
			*val = Int(x)
			nonNullValueFound = true
		}

	case Int32Type:
		if len(valByteSlice) > 0 {
			x, err := strconv.Atoi(string(valByteSlice))
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " as int: " + err.Error()))
			}
			*val = Int32(x)
			nonNullValueFound = true
		}

	case FloatType:
		if len(valByteSlice) > 0 {
			x, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " as float64: " + err.Error()))
			}
			*val = Float(x)
			nonNullValueFound = true
		}

	case BoolType:
		*val = Bool(string(valByteSlice) == "1")
		nonNullValueFound = true		

	case StringType:
		*val = String(SqlStringValueUnescape(string(valByteSlice)))
		nonNullValueFound = true // Shoot!! How do we distinguish between not set and empty string?

	default:
		panic(fmt.Sprintf("I don't know how to restore a type %v attribute.", attr.Part.Type))
	}
	return
}


var extraTimeZones map[string]string = map[string]string{  
	"pst" : "America/Vancouver",
	"pdt" : "America/Vancouver",
	"mst" : "America/Calgary",
	"mdt" : "America/Calgary",	
	"cst" : "America/Winnipeg",
	"cdt" : "America/Winnipeg",
	"est" : "America/Toronto",
	"edt" : "America/Toronto",	
	"ast" : "America/Halifax",
	"adt" : "America/Halifax",			
}

/*
Kludge to get around that on some platforms, common short timezone names are unknown.
Seem to be getting this on Mac OS X 10.9.4 all of a sudden.
Should move this to a util/time_util package.
*/
func LoadLocation(locationName string) (location *time.Location, err error) {

	location, err = time.LoadLocation(locationName)
	if err != nil {
       name := strings.ToLower(locationName)
       var locationFound bool
       locationName, locationFound = extraTimeZones[name]
       if locationFound {
       	  location, err = time.LoadLocation(locationName)
       }
	}		
    return
}

/*
Convert the byte slice which was returned as a column-value of a databse result row into
a primitive-type RObject. Sets the val argument to the new RObject.
If the value from the database was NULL (empty string in numeric fields), does not
set the val argument, and returns false.

TODO NOT HANDLING NULLS PROPERLY HERE YET !!!!!!!!!

*/
func convertAttrValTwoFields(valByteSlice []byte, valByteSlice2 []byte, attr *AttributeSpec, val *RObject) (nonNullValueFound bool) {
	switch attr.Part.Type {
	case TimeType:
		if len(valByteSlice) > 0 {
			timeString := string(valByteSlice)
			var locationName string
	  	    if len(valByteSlice2) > 0 {
			   locationName = string(valByteSlice2)		       
		    }		
			timeUTC, err := time.Parse(TIME_LAYOUT, timeString) 
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " as Time: " + err.Error()))
			}
			location, err := LoadLocation(locationName)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + " time.Location: " + err.Error()))
			}			
            *val = RTime(timeUTC.In(location))	
			nonNullValueFound = true
		}



	case ComplexType:
		if len(valByteSlice) > 0 {
			r, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + "_r as float64: " + err.Error()))
			}
		    if len(valByteSlice2) == 0 {
				panic(errors.New("attr " + attr.Part.Name + " imaginary part is null"))			
			}
			
			i, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + "_i as float64: " + err.Error()))
			}			
				
			*val = Complex(complex(r,i))
			nonNullValueFound = true
		}
		
	case Complex32Type:
		if len(valByteSlice) > 0 {
			r, err := strconv.ParseFloat(string(valByteSlice), 32)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + "_r as float32: " + err.Error()))
			}
		    if len(valByteSlice2) == 0 {
				panic(errors.New("attr " + attr.Part.Name + " imaginary part is null"))			
			}
			
			i, err := strconv.ParseFloat(string(valByteSlice), 32)
			if err != nil {
				panic(errors.New("attr " + attr.Part.Name + "_i as float32: " + err.Error()))
			}			
				
			r32 := float32(r)
			i32 := float32(i)	
			*val = Complex32(complex(r32,i32))
			nonNullValueFound = true
		}		

	default:
		panic(fmt.Sprintf("I don't know how to restore a type %v attribute.", attr.Part.Type))
	}
	return
}



/*
Convert the byte slice which was returned as a column-value of a databse result row into
a primitive-type RObject. Sets the val argument to the new RObject.
If the value from the database was NULL (empty string in numeric fields), does not
set the val argument, and returns false.

TODO NOT HANDLING NULLS PROPERLY HERE YET !!!!!!!!!

*/
func convertVal(valByteSlice []byte, typ *RType, errPrefix string, val *RObject) (nonNullValueFound bool) {
	switch typ {
	case IntType:
		if len(valByteSlice) > 0 {
			x, err := strconv.ParseInt(string(valByteSlice), 10, 64)
			if err != nil {
				panic(errors.New(errPrefix + " as int64: " + err.Error()))
			}
			*val = Int(x)
			nonNullValueFound = true
		}

	case Int32Type:
		if len(valByteSlice) > 0 {
			x, err := strconv.Atoi(string(valByteSlice))
			if err != nil {
				panic(errors.New(errPrefix + " as int: " + err.Error()))
			}
			*val = Int32(x)
			nonNullValueFound = true
		}

	case FloatType:
		if len(valByteSlice) > 0 {
			x, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New(errPrefix + " as float64: " + err.Error()))
			}
			*val = Float(x)
			nonNullValueFound = true
		}

	case BoolType:
		*val = Bool(string(valByteSlice) == "1")
		nonNullValueFound = true		

	case StringType:
		*val = String(SqlStringValueUnescape(string(valByteSlice)))
		nonNullValueFound = true // Shoot!! How do we distinguish between not set and empty string?

	default:
		panic(fmt.Sprintf("I don't know how to restore a type %v primitive value.", typ))
	}
	return
}



/*
Convert the byte slice which was returned as a column-value of a databse result row into
a primitive-type RObject. Sets the val argument to the new RObject.
If the value from the database was NULL (empty string in numeric fields), does not
set the val argument, and returns false.

TODO NOT HANDLING NULLS PROPERLY HERE YET !!!!!!!!!

*/
func convertValTwoFields(valByteSlice []byte, valByteSlice2 []byte, typ *RType, errPrefix string, val *RObject) (nonNullValueFound bool) {
	switch typ {
	case TimeType:
		if len(valByteSlice) > 0 {
			timeString := string(valByteSlice)
			var locationName string
	  	    if len(valByteSlice2) > 0 {
			   locationName = string(valByteSlice2)		       
		    }		
			timeUTC, err := time.Parse(TIME_LAYOUT, timeString) 
			if err != nil {
				panic(errors.New(errPrefix + " as Time: " + err.Error()))
			}
			location, err := LoadLocation(locationName)
			if err != nil {
				panic(errors.New(errPrefix + " time.Location: " + err.Error()))
			}			
            *val = RTime(timeUTC.In(location))	
			nonNullValueFound = true
		}



	case ComplexType:
		if len(valByteSlice) > 0 {
			r, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New(errPrefix + "_r as float64: " + err.Error()))
			}
		    if len(valByteSlice2) == 0 {
				panic(errors.New(errPrefix + " imaginary part is null"))			
			}
			
			i, err := strconv.ParseFloat(string(valByteSlice), 64)
			if err != nil {
				panic(errors.New(errPrefix + "_i as float64: " + err.Error()))
			}			
				
			*val = Complex(complex(r,i))
			nonNullValueFound = true
		}
		
	case Complex32Type:
		if len(valByteSlice) > 0 {
			r, err := strconv.ParseFloat(string(valByteSlice), 32)
			if err != nil {
				panic(errors.New(errPrefix +  "_r as float32: " + err.Error()))
			}
		    if len(valByteSlice2) == 0 {
				panic(errors.New(errPrefix + " imaginary part is null"))			
			}
			
			i, err := strconv.ParseFloat(string(valByteSlice), 32)
			if err != nil {
				panic(errors.New(errPrefix +  "_i as float32: " + err.Error()))
			}			
				
			r32 := float32(r)
			i32 := float32(i)	
			*val = Complex32(complex(r32,i32))
			nonNullValueFound = true
		}		

	default:
		panic(fmt.Sprintf("I don't know how to restore a type %v primitive value.", typ))
	}
	return
}
















/*
   Persist an robject for the first time. 
   TODO need a special case for a collection.

   1. inserts a row into the RObject table.
   2. For the object's type and each type in the up chain of the type, create a row in the type's table.

*/
func (db *SqliteDBThread) insert(th InterpreterThread, obj RObject, dbid, dbid2 int64) (err error) {

   var stmtStr string
   var args []interface{}
   var stmt *StatementGroup
   if obj.IsCollection() {
      collection := obj.(RCollection)
      collectionTypeDescriptor,_,_,_,_ := db.db.TypeDescriptor(collection)
	  stmt = Stmt("INSERT INTO RObject(id,id2,flags,typeName) VALUES(?,?,?,?);")
	  stmt.Arg(dbid)   
	  stmt.Arg(dbid2)  
	  stmt.Arg(int(obj.Flags()))
	  stmt.Arg(collectionTypeDescriptor)
   } else {
	  stmt = Stmt("INSERT INTO RObject(id,id2,flags,typeName) VALUES(?,?,?,?);")
	  stmt.Arg(dbid)   
	  stmt.Arg(dbid2)  
	  stmt.Arg(int(obj.Flags()))
	  stmt.Arg(obj.Type().ShortName())


	  stmtStr,args = db.db.instanceInsertStatement(th, obj.Type(), obj)
	  stmt.Add(stmtStr)
	  stmt.Args(args)
	  for _, typ := range obj.Type().Up {
	   	stmtStr,args = db.db.instanceInsertStatement(th, typ, obj)
	   	stmt.Add(stmtStr)
	   	stmt.Args(args)		
	  }
   }

   err = db.ExecStatements(stmt)
   return
}

/*
   Return the SQL INSERT statement that inserts a row in the type's instance-persistence table.
   This table is named with the type's short name, and contains columns to represent the primitive-valued attributes
   that are defined for the type.

   Note: Must begin with ";"
*/
func (db *SqliteDB) instanceInsertStatement(th InterpreterThread, t *RType, obj RObject) (string,[]interface{}) {

	table := db.TableNameIfy(t.ShortName())
	primitiveAttrVals,args := db.primitiveAttrValsSQL(th, t, obj)
	s := fmt.Sprintf("INSERT INTO %s VALUES(%s);", table, primitiveAttrVals)
	return s, args
}

/*
Return a comma separated string representation of the values (for the argument object) of 
the primitive=valued attributes defined in the type.

Return a string with the correct number of ?s for the type's primitive attributes + the id of the object.
Also return a list of the values to be inserted into a new row for the type: the id of the object and the attribute
values.
*/
func (db *SqliteDB) primitiveAttrValsSQL(th InterpreterThread, t *RType, obj RObject) (s string, args []interface{}) {
    s = "?"
    args = append(args,obj.DBID())

	for _, attr := range t.Attributes {
		if attr.Part.Type.IsPrimitive && attr.Part.CollectionType == "" {
			val, found := RT.AttrVal(th, obj, attr)
			s += ","
			if found {
				switch val.(type) {
				case Int:
					s += strconv.FormatInt(int64(val.(Int)), 10)
				case Int32:
					s += strconv.FormatInt(int64(int32(val.(Int32))), 10)	
				case Uint:
				    s += strconv.FormatUint(uint64(val.(Uint)), 10)		
				case Uint32:
					s += strconv.FormatUint(uint64(uint32(val.(Uint32))), 10)		
				case RTime:
					t := time.Time(val.(RTime))
					timeString := t.UTC().Format(TIME_LAYOUT)
					locationName := t.Location().String()
					s += "'" + timeString + "','" + locationName + "'"															
				case String:
					// s += "'" + SqlStringValueEscape(string(val.(String))) + "'"
					s += "?"
					args = append(args, SqlStringValueEscape(string(val.(String))))
				case Float:
					s += strconv.FormatFloat(float64(val.(Float)), 'G', -1, 64)
					
				case Complex:
					c := complex128(val.(Complex))
					r := real(c)
					i := imag(c)
					s += strconv.FormatFloat(r, 'G', -1, 64) + "," + strconv.FormatFloat(i, 'G', -1, 64)
				case Complex32:
					c := complex64(val.(Complex32))
					r := float64(real(c))
					i := float64(imag(c))					
					s += strconv.FormatFloat(r, 'G', -1, 32) + "," + strconv.FormatFloat(i, 'G', -1, 32)								
				case Bool:
					boolVal := bool(val.(Bool))
					if boolVal {
						s += "1"
					} else {
						s += "0"
					}
				case *Mutex,*RWMutex:
				   // Ignore these transient-type attributes 
					s = s[:len(s)-1] // take off the comma
				default:
					panic(fmt.Sprintf("I don't know how to create SQL for an attribute value of underlying type %v.", val.Type()))
				}
			} else if attr.Part.Type == TimeType || attr.Part.Type == ComplexType || attr.Part.Type == Complex32Type {
				s += "NULL,NULL"
			} else if attr.Part.Type == MutexType || attr.Part.Type == RWMutexType || attr.Part.Type == OwnedMutexType {
				// Ignore these transient-type attributes 
				s = s[:len(s)-1] // take off the comma				
			} else {
				s += "NULL"
			}
		}
	}
	return 

}




/*
Used by primitive-valued collection insert and set clauses.
Does not handle nil values.
*/
func (db *SqliteDB) primitiveValSQL(val RObject) (args []interface{}) {
   switch val.(type) {
   case Int:
   	args = []interface{}{strconv.FormatInt(int64(val.(Int)), 10)}
   case Int32:
   	args = []interface{}{strconv.FormatInt(int64(int32(val.(Int32))), 10)}
   case Uint:
       args = []interface{}{strconv.FormatUint(uint64(val.(Uint)), 10)}
   case Uint32:
   	args = []interface{}{strconv.FormatUint(uint64(uint32(val.(Uint32))), 10)}	
   case RTime:
   	t := time.Time(val.(RTime))
   	timeString := t.UTC().Format(TIME_LAYOUT)
   	locationName := t.Location().String()
   	args = []interface{}{timeString,locationName}   															
   case String:
   	args = []interface{}{SqlStringValueEscape(string(val.(String)))}
   case Float:
   	args = []interface{}{strconv.FormatFloat(float64(val.(Float)), 'G', -1, 64)}
	
   case Complex:
   	c := complex128(val.(Complex))
   	r := real(c)
   	i := imag(c)
   	args = []interface{}{strconv.FormatFloat(r, 'G', -1, 64),strconv.FormatFloat(i, 'G', -1, 64)}
   case Complex32:
   	c := complex64(val.(Complex32))
   	r := float64(real(c))
   	i := float64(imag(c))					
   	args = []interface{}{strconv.FormatFloat(r, 'G', -1, 32),strconv.FormatFloat(i, 'G', -1, 32)}							
   case Bool:
   	boolVal := bool(val.(Bool))
   	if boolVal {
   		args = []interface{}{"1"}
   	} else {
   		args = []interface{}{"0"}
   	}
   default:
   	panic(fmt.Sprintf("I don't know how to create SQL for a value of underlying type %v.", val.Type()))
   }
	return 
}


func SqlStringValueEscape(s string) string {
	s = strconv.Quote(s)	
    // Use QuoteToASCII(s) instead???
	
	s = strings.Replace(s, `\"`, `"`, -1)
//	s = strings.Replace(s, "'", "''", -1)
	s = s[1:len(s)-1] // Strip surrounding double-quotes
	return s
}

/* Old version, prior to passing args to sqlite3 exec
func SqlStringValueEscape(s string) string {
	s = strconv.Quote(s)	
    // Use QuoteToASCII(s) instead???
	
	s = strings.Replace(s, `\"`, `"`, -1)
	s = strings.Replace(s, "'", "''", -1)
	s = s[1:len(s)-1] // Strip surrounding double-quotes
	return s
}
*/

func SqlStringValueUnescape(s string) string {

	
	s = strings.Replace(s, `"`, `\"`, -1)
	s, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		panic(err)
	}
	return s
}
