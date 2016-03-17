// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
   persist.go - sqlite-specific implementation of generic persistence operations

   This file handles definition of the database abstraction layer type, and basic generic persistence operations.
   See also persist_schema.go, and persist_data.go, which add specific methods to the database abstraction layer
   (for persisting relish types and attribute and relation specifications, 
   and for persisting relish objects and attribute assignments, respectively)


   `CREATE TABLE robject(
      id INTEGER PRIMARY KEY,
      id2 INTEGER, 
      idReversed BOOLEAN, --???
      typeName TEXT   -- Should be typeId because type should be another RObject!!!!
   )`

*/

import (
  sqlite "github.com/mxk/go-sqlite/sqlite3"


	"fmt"
	. "relish/dbg"
	. "relish/runtime/data"
	"relish/params"
  "os"
)



func (db *SqliteDB) NewDBThread(th InterpreterThread) DBT {
   dbti := &SqliteDBThread{db:db}
   dbt := &DBThread{db: db, dbti: dbti, th: th}
   dbti.dbt = dbt
   return dbt
}


// TODO: Move this to persistence package and have it get a reference to a SqliteDBThread, which 
// is most of the methods of SqliteDB




/*
    Has a reference to the DB. 
    Executes DB queries in a serialized fashion in a multi-threaded environment.
    Also manages database transactions.
*/
type DBThread struct {
	db DB  // the database connection-managing object

	dbti DBT  // the database-type specific implementation of a db connection thread

	acquiringDbLock bool  // This thread is in the process of acquiring and locking the dbMutex 
	                      // (but may still be blocked waiting for the mutex to be unlocked by another thread)
	
	dbLockOwnershipDepth int  // How many nested claims has this thread currently made for ownership of the dbMutex
                              // If > 0, this thread owns and has locked the dbMutex.
                              // Note: thread "ownership" of dbMutex is an abstract concept imposed by this DBThread s/w,
                              // because Go Mutexes are not inherently owned by any particular goroutine.	

  conn Connection // SQL db connection

  isReadOnlyTransaction bool  // Experiment

  th InterpreterThread
}

/*
Grabs the dbMutex when it can (blocking until then) then executes a BEGIN IMMEDIATE TRANSACTION sql statement.
Does not unlock the dbMutex or release this thread's ownership of the mutex. 
Use CommitTransaction or RollbackTransaction to do that.
*/
func (dbt * DBThread) BeginTransaction(transactionType string) (err error) {
   Logln(PERSIST2_,"DBThread.BeginTransaction") 	

   if transactionType == "DEFERRED" {
       dbt.isReadOnlyTransaction = true
   }
   
   dbt.UseDB()	

   dbt.isReadOnlyTransaction = false  // was only needed to influence UseDB()

   err = dbt.dbti.BeginTransaction(transactionType) 
   if err != nil {
   	   dbt.ReleaseDB()
   }
   return
}

/*
Executes a COMMIT TRANSACTION sql statement. If it succeeds, unlocks the dbMutex and releases this thread's ownership
of the mutex.
If it fails (returns a non-nil error), does not unlock the dbMutex or release this thread's ownership of the mutex.

In the error case, the correct behaviour is to either retry the commit, do a rollback, or just call ReleaseDB to
unlock the dbMutex and release this thread's ownership of the mutex.
*/
func (dbt * DBThread) CommitTransaction() (err error) {
    Logln(PERSIST2_,"DBThread.CommitTransaction") 		
	err = dbt.dbti.CommitTransaction()
	if err == nil {
	   dbt.ReleaseDB()
    }
   return
}

/*
Executes a ROLLBACK TRANSACTION sql statement. If it succeeds, unlocks the dbMutex and releases this thread's ownership
of the mutex.
If it fails (returns a non-nil error), does not unlock the dbMutex or release this thread's ownership of the mutex.

In the error case, the correct behaviour is to either retry the rollback, or just call ReleaseDB to
unlock the dbMutex and release this thread's ownership of the mutex.
*/
func (dbt * DBThread) RollbackTransaction() (err error) {
    Logln(PERSIST2_,"DBThread.RollbackTransaction") 	
	err = dbt.dbti.RollbackTransaction()
	if err == nil {
		dbt.ReleaseDB()
	}
	return
}

/*
If the thread does not already own the dbMutex, lock the mutex and
flag that this thread owns it.
Used to ensure exlusive access to db for single db reads / writes 
for which we don't want to manually start a long-running transaction.

This method will block until no other DBThread is using the database.
*/
func (dbt * DBThread) UseDB() {
   Logln(PERSIST2_,"DBThread.UseDB when ownership level is",dbt.dbLockOwnershipDepth) 		
   if dbt.acquiringDbLock {  // Umm, shouldn't this be impossible? The same thread is blocked further inside this method.
      return	
   }	
   if dbt.dbLockOwnershipDepth == 0 {
   	   dbt.acquiringDbLock = true

       if dbt.th != nil {
          dbt.th.AllowGC()          
       }
       dbt.conn = dbt.db.GrabConnection(! dbt.isReadOnlyTransaction)
       if dbt.th != nil {
          dbt.th.DisallowGC()
       }
      

   	   // dbt.db.UseDB()
       
       dbt.acquiringDbLock = false      	
   }
   dbt.dbLockOwnershipDepth++
   Logln(PERSIST2_,"DBThread.UseDB: Set ownership level to",dbt.dbLockOwnershipDepth)    
}

/*

Remove one level of interest of this thread in the dbMutex.
If we have lost all interest in it, and
if the thread owns the dbMutex, unlock the mutex and
flag that this thread no longer owns it.
Returns false if this thread still has an interest in and lock on the dbMutex.
*/	
func (dbt * DBThread) ReleaseDB() bool {
    Logln(PERSIST2_,"DBThread.ReleaseDB when ownership level is",dbt.dbLockOwnershipDepth) 		
    if dbt.dbLockOwnershipDepth > 0 {
	   dbt.dbLockOwnershipDepth--
       Logln(PERSIST2_,"DBThread.ReleaseDB: Set ownership level to",dbt.dbLockOwnershipDepth)  	
	   if dbt.dbLockOwnershipDepth == 0 {
        dbt.db.ReleaseConnection(dbt.conn)
         
        dbt.conn = nil

		  // dbt.db.ReleaseDB()	
	   } else {
	      return false	
	   }
    }	
    return true
}


func (dbt * DBThread) EnsureObjectTable() {
   dbt.UseDB()	
   dbt.dbti.EnsureObjectTable()
   dbt.ReleaseDB()  
}

func (dbt * DBThread) EnsureObjectNameTable() {
   dbt.UseDB()	
   dbt.dbti.EnsureObjectNameTable()
   dbt.ReleaseDB()  
}

func (dbt * DBThread) EnsurePackageTable() {
   dbt.UseDB()	
   dbt.dbti.EnsurePackageTable()
   dbt.ReleaseDB()  
}

func (dbt * DBThread) EnsureTypeTable(typ *RType) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.EnsureTypeTable(typ)
   dbt.ReleaseDB()  
   return 
}

func (dbt * DBThread) ExecStatements(statementGroup *StatementGroup) (err error) {
   dbt.UseDB()
   err = dbt.dbti.ExecStatements(statementGroup)
   dbt.ReleaseDB()
   return
}

func (dbt * DBThread) ExecStatement(statement string, args ...interface{}) (err error) {
   dbt.UseDB()
   err = dbt.dbti.ExecStatement(statement, args...)
   dbt.ReleaseDB()
   return
}

func (dbt * DBThread) PersistSetAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, attrHadValue bool) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistSetAttr(th, obj, attr, val, attrHadValue)
   dbt.ReleaseDB()
   return
}

func (dbt * DBThread) PersistAddToAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, insertedIndex int) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistAddToAttr(th, obj, attr, val, insertedIndex)
   dbt.ReleaseDB()
   return 
}

func (dbt * DBThread) PersistRemoveFromAttr(obj RObject, attr *AttributeSpec, val RObject, removedIndex int) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistRemoveFromAttr(obj, attr, val, removedIndex)
   dbt.ReleaseDB()
   return   
}

func (dbt * DBThread) PersistRemoveAttr(obj RObject, attr *AttributeSpec) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistRemoveAttr(obj, attr)
   dbt.ReleaseDB() 
   return  
}

func (dbt * DBThread) PersistClearAttr(obj RObject, attr *AttributeSpec) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistClearAttr(obj, attr)
   dbt.ReleaseDB()
   return
}


func (dbt * DBThread) PersistSetAttrElement(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, index int) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistSetAttrElement(th, obj, attr, val, index)
   dbt.ReleaseDB()
   return   
}



      
func (dbt * DBThread) PersistMapPut(th InterpreterThread, theMap Map, key RObject,val RObject, isNewKey bool) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistMapPut(th, theMap, key, val, isNewKey)
   dbt.ReleaseDB()
   return   
}
      
      
func (dbt * DBThread) PersistSetCollectionElement(th InterpreterThread, coll IndexSettable, val RObject, index int) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistSetCollectionElement(th, coll, val, index)
   dbt.ReleaseDB()
   return   
}
  
func (dbt * DBThread) PersistAddToCollection(th InterpreterThread, coll AddableCollection, val RObject, insertedIndex int) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistAddToCollection(th, coll, val, insertedIndex)
   dbt.ReleaseDB()
   return   
}

func (dbt * DBThread) PersistRemoveFromCollection(coll RemovableCollection, val RObject, removedIndex int) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistRemoveFromCollection(coll, val, removedIndex)
   dbt.ReleaseDB()
   return   
}

func (dbt * DBThread) PersistClearCollection(coll RemovableCollection) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.PersistClearCollection(coll)
   dbt.ReleaseDB()
   return   
}








func (dbt * DBThread) EnsurePersisted(th InterpreterThread, obj RObject) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.EnsurePersisted(th, obj)
   dbt.ReleaseDB() 
   return  
}

func (dbt * DBThread) EnsureAttributeAndRelationTables(t *RType) (err error) {
   dbt.UseDB()	
   err = dbt.dbti.EnsureAttributeAndRelationTables(t)
   dbt.ReleaseDB()
   return 
}

func (dbt * DBThread) ObjectNameExists(name string) (found bool, err error) {
   dbt.UseDB()
   found,err = dbt.dbti.ObjectNameExists(name)
   dbt.ReleaseDB()
   return
}

func (dbt * DBThread) ObjectNames(prefix string) (names []string, err error) {
   dbt.UseDB()
   names,err = dbt.dbti.ObjectNames(prefix)
   dbt.ReleaseDB()
   return  
}


func (dbt * DBThread) NameObject(obj RObject, name string) (err error) {
   dbt.UseDB()
   err = dbt.dbti.NameObject(obj, name)
   dbt.ReleaseDB() 
   return  
}

func (dbt * DBThread) RenameObject(oldName string, newName string) (err error) {
   dbt.UseDB()
   err = dbt.dbti.RenameObject(oldName, newName)
   dbt.ReleaseDB() 
   return  
}


func (dbt * DBThread)  Delete(obj RObject) (err error) {
   dbt.UseDB()
   err = dbt.dbti.Delete(obj)
   dbt.ReleaseDB() 
   return  
}


func (dbt * DBThread) RecordPackageName(name string, shortName string) {
   dbt.UseDB()
   dbt.dbti.RecordPackageName(name, shortName)
   dbt.ReleaseDB()
   return   	
}

func (dbt * DBThread) FetchByName(name string, radius int) (obj RObject, err error) {
   dbt.UseDB()
   obj, err = dbt.dbti.FetchByName(name, radius)   
   dbt.ReleaseDB() 
   return  
}

func (dbt * DBThread) Fetch(id int64, radius int) (obj RObject, err error) {
   dbt.UseDB()
   obj, err = dbt.dbti.Fetch(id, radius)
   dbt.ReleaseDB()  
   return 
}

func (dbt * DBThread) Refresh(obj RObject, radius int) (err error) {
   dbt.UseDB()
   err = dbt.dbti.Refresh(obj, radius)
   dbt.ReleaseDB()  
   return 
}


func (dbt * DBThread) FetchAttribute(th InterpreterThread, objId int64, obj RObject, attr *AttributeSpec, radius int) (val RObject, err error) {
   dbt.UseDB()
   val, err = dbt.dbti.FetchAttribute(th, objId, obj, attr, radius)
   dbt.ReleaseDB()  
   return 
}

/*
Given an object type and an OQL selection criteria clause in a string, set the argument collection to contain 
the matching objects from the the database.

e.g. of first two arguments: vehicles/Car, "speed > 60 order by speed desc"   
*/
	
func (dbt * DBThread) FetchN(typ *RType, oqlSelectionCriteria string, queryArgs []RObject, coll RCollection, radius int, objs *[]RObject) (mayContainProxies bool, err error) {
   dbt.UseDB()
   mayContainProxies, err = dbt.dbti.FetchN(typ, oqlSelectionCriteria, queryArgs, coll, radius, objs)
   dbt.ReleaseDB()	
   return
}

/*
Close the connection to the database.
*/
func (dbt * DBThread) Close() {
	dbt.dbti.Close()
}













/*
   A handle to the sqlite database that holds the open connection	   
   and has lots of specific methods for creating and working with relish
   data and metadata in the database.
*/
type SqliteDB struct {
	dbName         string
	// conn           *sqlite.Conn
	pool           *ConnectionPool
	
	statementQueue chan string
	// preparedStatements map[string]*sqlite.Stmt

	defaultDBThread DBT
}

/*
   Opens a connection to a sqlite database of the specified name.
   Creates the database file if does not already exist.
   Returns a handle to the database that holds the open connection
   and has lots of specific methods for creating and working with relish
   data and metadata in the database.
*/
func NewDB(dbName string) *SqliteDB {
	db := &SqliteDB{dbName: dbName, statementQueue: make(chan string, 1000)}
	// conn, err := sqlite.Open(dbName)
	// if err != nil {
	//	panic(fmt.Sprintf("Unable to open the database '%s': %s", dbName, err))
	// }
	// db.conn = conn

  maxConnections := params.DbMaxConnections
  if params.DbMaxWriteConnections != -1 {
     if params.DbMaxConnections > 1 {
          Logln(ALWAYS_,"Error. Specify either -pool or both -rpool and -wpool")
          os.Exit(1)     
     } 
     if params.DbMaxReadConnections != -1 {
         maxConnections = params.DbMaxReadConnections
     } else {
         Logln(ALWAYS_,"Error. Specify either -pool or both -rpool and -wpool")
         os.Exit(1)
     }
  }
	db.pool = NewConnectionPool(dbName, maxConnections, params.DbMaxWriteConnections, NewSqliteConn)

  db.defaultDBThread = db.NewDBThread(nil)

    // db.preparedStatements = make(map[string]*sqlite.Stmt)
	db.defaultDBThread.EnsureObjectTable()
	db.defaultDBThread.EnsureObjectNameTable()
	db.defaultDBThread.EnsurePackageTable()	
	
	
    // Obsolete I think. Was going to do db statement execution asynchronously from
    // relish code execution for efficiency, but not happening.
	// go db.executeStatements()

	return db
}


func (db *SqliteDB) DefaultDBThread() DBT {
	return db.defaultDBThread
}

/*
Notes on sqlite db use:
TODO Must reset or finalise all open statements before rollback or commit.

 (c *Conn) Exec(cmd string, args ...interface{}) error
does a finalize at end.

(s *Stmt) Exec(args ...interface{}) error
does a reset before other stuff.

func (s *Stmt) Reset() error

func (s *Stmt) Finalize() error

func (c *Conn) Prepare(cmd string) (*Stmt, error) 

*/



/*
Function that creates a sql db connection. 
Accepts the name of the database, and a numeric identifier for the connection
for debugging purposes.
*/
type ConnectionFactory func (string,int) (conn Connection, err error)


/*
A SQLITE3 connection which implements the Connection interface.
*/
type SqliteConn struct {
	conn *sqlite.Conn
	preparedStatements map[string]*sqlite.Stmt
  isReadOnly bool
  id int
}

/*
Set whether this connection is only to be used for read operations on the database.
e.g. selects but not inserts or updates. If never set, the connection defaults to 
read-or-write-enabled.
*/
func (conn *SqliteConn) SetReadOnly(isReadOnly bool) {
  conn.isReadOnly = isReadOnly
}

/*
Returns whether this connection is restricted to reading from the database.
*/
func (conn *SqliteConn) IsReadOnly() bool {
  return conn.isReadOnly
}


/*
Return a prepared statement based on the sql string.
Returns a cached, pre-prepared statement if the sql string has been used before in this connection.
*/
func (conn *SqliteConn) Prepare(sql string) (statement Statement, err error) {
	statement, err = conn.conn.Prepare(sql)
	return
}

func (conn *SqliteConn) PreparedStatement(sql string) (statement Statement, found bool) {
	statement, found = conn.preparedStatements[sql]
	return
}

func (conn *SqliteConn) CachePreparedStatement(sql string, statement Statement) {
	conn.preparedStatements[sql] = statement.(*sqlite.Stmt)
	return
}

func (conn *SqliteConn) Close() error {
	return conn.conn.Close()
}

func (conn *SqliteConn) Id() int {
  return conn.id
}


func NewSqliteConn(dbName string, connectionId int) (conn Connection, err error) {
	s3conn, err := sqlite.Open(dbName)
    if err != nil {
    	return
    }
    sConn := &SqliteConn{conn: s3conn, preparedStatements: make(map[string]*sqlite.Stmt), id: connectionId}    
	conn = sConn
	return
}

func (db *SqliteDB) GrabConnection(doingWrite bool) Connection {
	return db.pool.GrabConnection(doingWrite)
}

func (db *SqliteDB) ReleaseConnection(conn Connection) {
	db.pool.ReleaseConnection(conn)
}





type SqliteDBThread struct {
	db *SqliteDB  // the database implementation
	dbt *DBThread  // Generic version of myself
}

/*
 Should only be used for sql statements that are not dynamically constructed.
 That is, use this only for a small set of completely predefined queries, which may nonetheless have ? parameters
 in them.
 Tests if the sql statement (cmd) is a key in a prepared statements map. If so, returns the already prepared
 statement. If not found in the map, asks the db connection to prepare the statement.
 Returns the prepared statement.

 TODO maybe should do stmt.Reset() if found in hashtable
*/
func (db *SqliteDBThread) Prepare(cmd string) (stmt *sqlite.Stmt, err error) {
   var statement Statement	
   statement,found := db.dbt.conn.PreparedStatement(cmd)
   if ! found {
	  statement,err = db.dbt.conn.Prepare(cmd)   	 
	  if err != nil {
         return
      }
	  db.dbt.conn.CachePreparedStatement(cmd, statement)
   }	
   stmt = statement.(*sqlite.Stmt)   
   return
}

/*
Execute multiple sql statments, each possibly with arguments.
Return a non-nil error and do not continue to process statements, if 
a database error occurs.
*/
func (db *SqliteDBThread) ExecStatements(statementGroup *StatementGroup) (err error) {

	// Replace everything  with the statementQueue insertion below

	Logln(PERSIST_, fmt.Sprintf("Executing statement:\n%s\n", statementGroup))

	for _, sqlStatement := range statementGroup.Statements {
			
		err = db.ExecStatement(sqlStatement.Statement,sqlStatement.Args...)	
        if err != nil {
        	break
        }
	}
	return
}


/*
Execute a single SQL statement.
Return the db error if any.
*/
func (db *SqliteDBThread) ExecStatement(statement string, args ...interface{}) (err error) {

    Logln(PERSIST_, fmt.Sprintf("Executing statement:\n%s\n", statement))
		
    stmt, prepareErr := db.Prepare(statement)
    if prepareErr != nil {
	err = fmt.Errorf("DB ERROR preparing statement:\n%s\nDetail: %s\n\n", statement, prepareErr) 	   
	    return
    }

    defer stmt.Reset()

    err = stmt.Exec(args...)
    if err != nil {
       err = fmt.Errorf("DB ERROR executing statement:\n%s\nDetail: %s\n\n", statement, err) 	
	   return
    }   
    return
}


/*
OBSOLETE
Runs in its own goroutine, taking statements out of the queue and executing them in the sqlite database.
Loops forever.
TODO add good error handling.

func (db *SqliteDBThread) executeStatements() {
	for {
		statementGroup := <-db.statementQueue

		// if strings.HasPrefix(statementGroup,"SELECT") {}

		Logln(PERSIST_, fmt.Sprintf("Executing statements:\n%s\n", statementGroup))

		statements := strings.SplitAfter(statementGroup, ";")
		for _, statement := range statements {
			if statement != "" {
				err := conn.Exec(statement)
				if err != nil {
					fmt.Printf("DB ERROR on statement:\n%s\nDetail: %s\n\n", statement, err)
				}
			}
		}
	}
}
*/

func (db *SqliteDBThread) Close() {
	db.dbt.conn.Close()
}
