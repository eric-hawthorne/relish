// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements data persistence in the relish language environment.

package persist

/*
   transaction.go 
   
   This file implements transaction management and database access serialization aspects of
   sqlite persistence of relish objects

   SEE ALSO the dbThread type in interp/interpreter.go for an extended implementation of the methods
   below that fully implements safe multi-threaded access to the database.
*/

import (
   sqlite "github.com/mxk/go-sqlite/sqlite3"

	"sync"
   "math/rand"
   "time"
   . "relish/dbg"
)


// A mutual-exclusion lock which ensures single-goroutine-at-a-time serialized access to the database.
var dbMutex sync.Mutex

const N_BEGIN_TRIES = 10
const N_COMMIT_TRIES = 10

var TRY_GAP_WIDENING_MS_INCREMENT = 100  // ms 
 
var dummyQuery *sqlite.Stmt

/*
Begins an immediate-mode database transaction.
*/
func (db *SqliteDBThread) BeginTransaction(transactionType string) (err error) {
   
   db.dbt.th.AllowGC()
   defer db.dbt.th.DisallowGC()
   
   stmt := "BEGIN " + transactionType + " TRANSACTION"
   err = db.ExecStatement(stmt)

   var TRY_GAP_WIDENING_MS_INCREMENT int64 = 100  // ms - longest wait will be 6 seconds + random factor 
   var tryGapWidth int64 = TRY_GAP_WIDENING_MS_INCREMENT
   var r int64
   for i := 0; err != nil && i < N_BEGIN_TRIES ; i ++ {
      Logln(PERSIST2_,"BEGIN", transactionType, "ERR:",err)
      r = rand.Int63n(500 + tryGapWidth)
      if i > 5 {
         TRY_GAP_WIDENING_MS_INCREMENT = 1000
      }
      tryGapWidth += TRY_GAP_WIDENING_MS_INCREMENT
      time.Sleep( time.Duration( (tryGapWidth + r) * 1000000 ) )
      err = db.ExecStatement(stmt)
   }

   if transactionType == "DEFERRED" {
      // We really don't want a deferred transaction. We want to create a SHARED lock right
      // away in the SQLITE database. If we cannot, we need to fail-fast here, so that
      // the code inside the transaction-protected block will not be executed.

      // So do a dummy read query.

      if dummyQuery == nil {
          dummyQuery, err = db.Prepare("select rowid from RPackage where rowid=1")
          if err != nil {
             panic(err)
          }
      }

      err = dummyQuery.Query()  
      dummyQuery.Reset()

      var TRY_GAP_WIDENING_MS_INCREMENT int64 = 100  // ms - longest wait will be 6 seconds + random factor 
      var tryGapWidth int64 = TRY_GAP_WIDENING_MS_INCREMENT
      var r int64
      for j := 0; err != nil && j < N_BEGIN_TRIES  ; j++ {
         Logln(PERSIST2_,"BEGIN DEFERRED: ERR executing dummy select query:",err)
         r = rand.Int63n(500 + tryGapWidth)
         if j > 5 {
            TRY_GAP_WIDENING_MS_INCREMENT = 1000
         }
         tryGapWidth += TRY_GAP_WIDENING_MS_INCREMENT
         time.Sleep( time.Duration( (tryGapWidth + r) * 1000000 ) )
         err = dummyQuery.Query()
         dummyQuery.Reset()
      }
   }



   if err == nil {
      Logln(PERSIST2_,">>>>>>>>>>>>>>>>>>>>>>>>> SUCCESSFULLY BEGAN", transactionType, "TRANSACTION")
   }
   return
}

/*
Commits the in-effect database transaction.
*/
func (db *SqliteDBThread) CommitTransaction() (err error) {
   db.dbt.th.AllowGC()
   defer db.dbt.th.DisallowGC()
   err = db.ExecStatement("COMMIT TRANSACTION")
   var TRY_GAP_WIDENING_MS_INCREMENT int64 = 100  
   var tryGapWidth int64 = TRY_GAP_WIDENING_MS_INCREMENT
   var r int64
   for i := 0;  err != nil && i < N_COMMIT_TRIES ; i ++ {
      Logln(PERSIST2_,"COMMIT ERR:", err)      
      r = rand.Int63n(500 + tryGapWidth)
      if i > 5 {
         TRY_GAP_WIDENING_MS_INCREMENT = 1000
      }      
      tryGapWidth += TRY_GAP_WIDENING_MS_INCREMENT
      time.Sleep(  time.Duration( (tryGapWidth + r) * 1000000 ) )
      err = db.ExecStatement("COMMIT TRANSACTION")
   }
   if err == nil {
      Logln(PERSIST2_,"<<<<<<<<<<<<<<<<<<<<<<<<<< SUCCESSFULLY COMMITTED TRANSACTION")
   }
   return
}

/*
Rolls back the in-effect database transaction.
*/
func (db *SqliteDBThread) RollbackTransaction() (err error) {
   db.dbt.th.AllowGC()
   defer db.dbt.th.DisallowGC()  
   err = db.ExecStatement("ROLLBACK TRANSACTION")
   if err == nil {
      Logln(PERSIST2_,"<<<<<<<<<<<<<<<<<<<<<<<<<< SUCCESSFULLY ROLLED BACK TRANSACTION")
   } else {
      Logln(PERSIST2_,"FAILED !!!!! To ROLLBACK TRANSACTION !!!!!!!! ???")
   }   
   return
}


func (db *SqliteDBThread) UseDB() { 
   panic("should not ever be called. Generic DBThread type calls its method instead.")
}



func (db *SqliteDBThread) ReleaseDB() bool {
   panic("should not ever be called. Generic DBThread type calls its method instead.")
   return false
}

