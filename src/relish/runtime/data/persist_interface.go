// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   persist_interface.go -  Abstraction of persistence service for relish data.
*/

import (
      "net/http"
   )   

type StatementGroup struct {
	Statements []*SqlStatement
}

type SqlStatement struct {
	Statement string
	Args []interface{}
}

/*
Return a new empty statement group.
*/
func Stmts() (group *StatementGroup) {
	group = &StatementGroup{}
	return
}

/*
Return a new statement group with a single sql statement in it.
*/
func Stmt(statement string) (group *StatementGroup) {
	group = &StatementGroup{}
	group.Add(statement)
	return
}

/*
Add a statement to a statement group.
*/
func (sg *StatementGroup) Add(statement string) {
   sg.Statements = append(sg.Statements, &SqlStatement{Statement: statement})
}

/*
Add an argument to the last-added statement in the statement group.
Must have at least one statement first before calling this.
*/
func (sg *StatementGroup) Arg(val interface{}) {
   stmt := sg.Statements[len(sg.Statements)-1]
   stmt.Args = append(stmt.Args, val)
}



func (sg *StatementGroup) Args(args []interface{}) {
   stmt := sg.Statements[len(sg.Statements)-1]
   if stmt.Args == nil {
      stmt.Args = args
   } else {
      for _,arg := range args {
         stmt.Args = append(stmt.Args, arg)         
      } 
   }
}


/*
Clears all arguments that have been added to the statement group.
Note: Only a single-statement statement group can be re-used with new args,
or more precisely, only the last statement in the statement group 
can have its arguments re-appended to after ClearArgs.
*/
func (sg *StatementGroup) ClearArgs() {
   for _,stmt := range sg.Statements {
      stmt.Args = nil
   }
}


type DB interface {
   GrabConnection(doingWrite bool) Connection
   ReleaseConnection(conn Connection)

   /*
   Returns a DBConnectionThread 
   */
   NewDBThread(th InterpreterThread) DBT  

   /*
   Returns a DBConnectionThread that is created upon creation of the db proxy. 
   */
   DefaultDBThread() DBT    
}

type DBT interface {


   EnsureObjectTable()
   EnsureObjectNameTable()
   EnsurePackageTable()

   EnsureTypeTable(typ *RType) (err error)

	 ExecStatements(statementGroup *StatementGroup) (err error)
	 ExecStatement(statement string, args ...interface{}) (err error)	
	 PersistSetAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, attrHadValue bool) (err error)
	 PersistAddToAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, insertedIndex int) (err error)
	 PersistRemoveFromAttr(obj RObject, attr *AttributeSpec, val RObject, removedIndex int) (err error)
   PersistRemoveAttr(obj RObject, attr *AttributeSpec) (err error) 	
   PersistClearAttr(obj RObject, attr *AttributeSpec) (err error)
   PersistSetAttrElement(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, index int) (err error) 
   
   PersistMapPut(th InterpreterThread, theMap Map, key RObject,val RObject, isNewKey bool) (err error)    
   // Note: Missing PersistRemoveFromMap  
   PersistSetCollectionElement(th InterpreterThread, coll IndexSettable, val RObject, index int) (err error)   
 	 PersistAddToCollection(th InterpreterThread, coll AddableCollection, val RObject, insertedIndex int) (err error)
 	 PersistRemoveFromCollection(coll RemovableCollection, val RObject, removedIndex int) (err error)
   PersistClearCollection(coll RemovableCollection) (err error)    
     
	 EnsurePersisted(th InterpreterThread, obj RObject) (err error)
	 EnsureAttributeAndRelationTables(t *RType) (err error)
	
	 ObjectNameExists(name string) (found bool, err error)
   ObjectNames(prefix string) (names []string, err error)  
	 NameObject(obj RObject, name string) (err error)
	 RenameObject(oldName string, newName string) (err error)	
	
   /*
   Deletes the object from the database as well as its canonical object name entry if any.

   Also removes the object from in-memory cache.
   TODO: Consider multiple in-memory caches when we have them!!

   Does NOT delete attribute / relation/ collection association table entries that may exist
   for the object in the database.

   Is a safe no-op for objects that are not stored locally.
   */
   Delete(obj RObject) (err error)
	 RecordPackageName(name string, shortName string)
	 FetchByName(name string, radius int) (obj RObject, err error)
	 Fetch(id int64, radius int) (obj RObject, err error)
   Refresh(obj RObject, radius int) (err error)
	 FetchAttribute(th InterpreterThread, objId int64, obj RObject, attr *AttributeSpec, radius int) (val RObject, err error)

	/*
	
	Given an object type and an OQL selection criteria clause in a string, set the argument collection to contain 
	the matching objects from the the database.

	e.g. of first two arguments: vehicles/Car, "speed > 60 order by speed desc"   

  If coll is non-nil, it is treated as a persistent collection that the selection conditions are filtering. 

	mayContainProxies will be true if the collection was fetched lazily from db.
	*/
	
    FetchN(typ *RType, oqlSelectionCriteria string, queryArgs []RObject, coll RCollection, radius int, objs *[]RObject) (mayContainProxies bool, err error) 

    /*
    Close the connection to the database.
    */
	 Close()


	/*
    Begins an immediate-mode database transaction.	
    Implementations may also first lock program access to the database to ensure a single goroutine at a time
    runs a database transaction and no other goroutines interact with the database at all during the transaction.
    Accepts InterpreterThread so it can AllowGC() while waiting to Begin the Transaction.    
	*/
	 BeginTransaction(transactionType string) (err error)

	/*
    Implementations may also unlock program access to the database to allow other goroutines to use the database.
    Accepts InterpreterThread so it can AllowGC() while waiting to use db.   
	*/
	 CommitTransaction() (err error)

	/*
    Implementations may also unlock program access to the database to allow other goroutines to use the database.
	*/
	 RollbackTransaction() (err error)
	
	/*
	Lock the dbMutex.
	Used to ensure exlusive access to db for single db reads / writes 
	for which we don't want to manually start a long-running transaction.
	(Or may also be used in multi-threaded extensions of the Begin,Commit,RollbackTransaction methods.)

	This method will block until no other goroutine is using the database.
  Accepts InterpreterThread so it can AllowGC() while waiting to use db.
	*/
	 UseDB()
	
	/*
	If this db connection or thread-of-connection has no further interest in owning the database,
	unlock the dbMutex.
	If this db connection or thread-of-connection still has an interest in owning the database,
	returns false. A series of calls to this should eventually return true meaning no further
	calls to it by this thread-of-connection are appropriate until UseDB() is called again.
	*/	
	 ReleaseDB() bool
}


/*
Abstraction of a connection to a SQL database.
*/
type Connection interface {
   // TODO Method signatures

    Prepare(sql string) (Statement, error)
  /*
  Return a previously prepared statement matching the sql string,
  and an indication of one was found or not.
  */
  PreparedStatement(sql string) (statement Statement, found bool)

  /*
  Save the prepared statement in the connection.
  */
  CachePreparedStatement(sql string, statement Statement)

  /*
  This connection is only used for reading from the database. Never writing.
  */
  IsReadOnly() bool 

  /*
  Asserts that this connection is only used for db reads.
  Default is it could be read or write.
  */
  SetReadOnly(isReadOnly bool) 

  Close() error

  Id() int  // A simple integer that is incremented each time a connection is created.
}

/* 
A SQL statement
*/
type Statement interface {

    Exec(args ...interface{}) error

    Query(args ...interface{}) error

    Scan(dst ...interface{}) error 
    
    Next() error      

    Reset() 

    // Close releases all resources associated with the prepared statement. This
    // method can be called at any point in the statement's life cycle.
    // [http://www.sqlite.org/c3ref/finalize.html]
   Close() error     
}

/*
A single "thread" of execution of the relish interpreter, complete with its own relish data stack (hidden).
This interface provides access to the package whose method is currently executing, 
to the actual RMethod that is currently executing, and to the DBThread which can execute
database queries in a multi-threaded context. 

Note that each InterpreterThread actually corresponds to one goroutine. Goroutines are lightweight userspace
threads in Go which may be cooperative coroutines or may be mapped onto separate processor threads and/or cores.
Typically, multiple goroutines will execute in a single OS thread, but if multiple cores are available and
Go is configured to use them, then groups of goroutines may be apportioned across multiple OS-threads and cores. 
*/
type InterpreterThread interface {
	/*
	The package context of the executing method.
	*/
	Package() *RPackage
	
	/*
	The executing method.
	*/
	Method() *RMethod

  /*
  Returns the package from which the currently executing method was called, 
  or nil if at stack bottom.
  */
  CallingPackage() *RPackage 

  /*
  Returns the method that called the currently executing method, 
  or nil if at stack bottom.
  */
  CallingMethod() *RMethod


   /*
    A db connection thread. Used to control access to the database in a multi-threaded environment,
    and to manage database transactions. Makes use of a connection pool, which may be smaller than
    the total number of DBThreads, so sometimes operations on the DBThread block until a real DB 
    connection is available for use.
  */
	DBT() DBT
	
	/*
	Will be "" unless we are in a stack-unrolling panic, in which case, should be the error message.
	*/
	Err() string
	
	// AllowGC(msg string)   // deadlock debug versions
	
	// DisallowGC(msg string)   // deadlock debug versions

  AllowGC()
  
  DisallowGC()
	
	GC()

	EvaluationContext() MethodEvaluationContext

  /*
  The currently active DB transaction which this thread started or is participating in.
  */
  Transaction() *RTransaction 

  SetTransaction(tx *RTransaction)

  /*
  Creates if necessary and returns an http client unique to this thread.
  One client per thead means that the client's Timeout parameter etc
  can be set independently of uses of an http client by a different thread.
  */
  HttpClient() *http.Client
}


