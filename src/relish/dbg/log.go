// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// This file implements configurable, multi-concern, multi-level logging
// for purposes of debugging relish.

// Usage: 
// relish -log 1 otherargs
// relish -log 2 otherargs
//  
// To change what will be printed out by -log 1, set below, as shown in examples
// const SOME_DEBUG_FLAGS = some flag constants bit-OR'ed together  
// 
// To change what will be printed out by -log 2, set below, as shown in examples
// const FULL_DEBUG_FLAGS = SOME_DEBUG_FLAGS | some more flag constants bit-OR'ed together  
//
// then 
// go install relish/relish 
// to recompile the language with the  new logging topics and levels.


package dbg

import (
	"fmt"
	"sync"
)	
	
var indent uint  // Current call nesting level
var threadNum uint
var indents map[interface{}]uint = make(map[interface{}]uint)
var threadNums map[interface{}]uint = make(map[interface{}]uint)
var currentThreadNum uint

	
// These single-bit-set flags should be used (possibly |'d together explicitly) as the 
// first arg to Log,Logln,Logging,Trace functions.
// Each bit-flag represents a program aspect (or aspect x verbosity-level) to log.
//	
const (
	STACK_      = 1 << iota   // == 1 (iota has been reset)
	AST_        = 1 << iota   // == 2
	PARSE_	    = 1 << iota   // == 4
	LOAD_       = 1 << iota
	LOAD2_      = 1 << iota
	GENERATE_   = 1 << iota
	GENERATE2_  = 1 << iota	
	INTERP_     = 1 << iota   // Log INTERPreter aspect of program when logging at minimum (or above) verbosity level
	INTERP2_    = 1 << iota	  // Log INTERPreter aspect of program only when logging at medium (or above) verbosity level 
	INTERP3_    = 1 << iota	  // Log INTERPreter aspect of program only when logging at max verbosity level	
	INTERP_TR   = 1 << iota	
	INTERP_TR2  = 1 << iota	
	INTERP_TR3  = 1 << iota	
	PERSIST_    = 1 << iota	
	PERSIST2_   = 1 << iota	
	PERSIST_TR  = 1 << iota	
	PERSIST_TR2 = 1 << iota		
	ALWAYS_     = 1 << iota		
	WEB_        = 1 << iota	
	WEB2_       = 1 << iota	
	GC_         = 1 << iota		
	GC2_        = 1 << iota		
	GC3_        = 1 << iota		
	COLL_       = 1 << iota		
	COLL2_      = 1 << iota		
	LOCK_       = 1 << iota		
	INTERP_TRF  = 1 << iota		 		
	ANY_        = 0xFFFFFFFFFFFFFFFF	
)

// These multi-bit flags should be used to set the DEBUG_FLAGS. They are shortcuts
// which represent which other bits should also be set when a particular bit is set. 
// i.e. They implement level-hierarchy such as INTERP_,INTERP2_,INTERP3_ in such a way that
// INTERP3_ implies INTERP2_ and INTERP1_   etc.
//
const (
	STACK__     = STACK_
	AST__       = AST_
	PARSE__	    = PARSE_
	LOAD__      = LOAD_
	LOAD2__     = LOAD2_ | LOAD__
	GENERATE__	= GENERATE_
	GENERATE2__ = GENERATE2_ | GENERATE__	
	INTERP__    = INTERP_                   // Log INTERPreter aspect of program at minimum verbosity level
	INTERP2__   = INTERP2_ | INTERP__	    // Log INTERPreter aspect of program at medium verbosity level
	INTERP3__   = INTERP3_ | INTERP2__    	// Log INTERPreter aspect of program at max verbosity level	
	INTERP__TRF = INTERP_TRF
	INTERP__TR  = INTERP_TR	| INTERP_TRF
	INTERP__TR2 = INTERP_TR2 | INTERP__TR
	INTERP__TR3 = INTERP_TR3 | INTERP__TR2	
	PERSIST__   = PERSIST_
	PERSIST2__  = PERSIST2_ | PERSIST__
	PERSIST__TR = PERSIST_TR
	PERSIST__TR2 = PERSIST_TR2 | PERSIST__TR	
	WEB__       = WEB_
	WEB2__      = WEB2_ | WEB__	
	ALL__       = ANY_
	GC__        = GC_	
	GC2__       = GC2_ | GC__
	GC3__       = GC3_ | GC2__	
	COLL__        = COLL_	
	COLL2__       = COLL2_ | COLL__	
	LOCK__        = LOCK_
)
	
	



	


//const DEBUG_FLAGS = INTERP__TR3 | INTERP3__ | STACK__ | AST__ | PARSE__
//const DEBUG_FLAGS = INTERP__TR | INTERP__ | STACK__ | AST__ | PARSE__	
//const DEBUG_FLAGS = INTERP__TR | INTERP__ | AST__ | PARSE__	

// Was using this one for a long time.
//const SOME_DEBUG_FLAGS =  AST__ | PARSE__ | PERSIST2__ | PERSIST__TR2| INTERP__TR2 | INTERP2__	
//const SOME_DEBUG_FLAGS =   PERSIST2__ | PERSIST__TR2 | WEB__ | ALWAYS_

const SOME_DEBUG_FLAGS =   INTERP__TRF | ALWAYS_

// THIS WAS MOST RECENT
// const SOME_DEBUG_FLAGS =   PERSIST2__ | PERSIST__TR2 | ALWAYS_

// THIS WAS NEXT MOST RECENT
// const SOME_DEBUG_FLAGS =   PERSIST2__ | PERSIST__TR2 | GC__ | ALWAYS_

// const SOME_DEBUG_FLAGS =   INTERP_TR | GC__ | ALWAYS_

// const SOME_DEBUG_FLAGS =   COLL2__ | ALWAYS_

//const SOME_DEBUG_FLAGS =   LOAD2__ | ALWAYS_
//const SOME_DEBUG_FLAGS =   INTERP__TR2 | ALWAYS_
//const SOME_DEBUG_FLAGS =   PERSIST__ | PERSIST__TR | INTERP__TR | INTERP__ | WEB__ | ALWAYS_

// const SOME_DEBUG_FLAGS =  GC__ | WEB2__ | PERSIST__TR2 | PERSIST2__ | ALWAYS_

// const SOME_DEBUG_FLAGS =  INTERP__TR3 | INTERP3__ | ALWAYS_

// const SOME_DEBUG_FLAGS =  INTERP__TR2 | INTERP2__ | ALWAYS_

// const SOME_DEBUG_FLAGS =  PERSIST__TR2 | WEB2__ | ALWAYS_

// Note: To get a trace of generated AST, also need to 
// uncomment 
// var parserDebugMode uint = parser.DeclarationErrors | parser.Trace
// on line 218 of global_loader/code_load.go
// const SOME_DEBUG_FLAGS =  AST__ | PARSE__ | ALWAYS_

// const SOME_DEBUG_FLAGS =  GENERATE2__ | ALWAYS_

// const SOME_DEBUG_FLAGS =  INTERP__TR3 | INTERP3__ | STACK__ | ALWAYS_

// const SOME_DEBUG_FLAGS =  INTERP__TR3 | INTERP3__ | STACK__ | WEB2__ | ALWAYS_

// const FULL_DEBUG_FLAGS = SOME_DEBUG_FLAGS | GENERATE_

const FULL_DEBUG_FLAGS = SOME_DEBUG_FLAGS | GC2__
// const FULL_DEBUG_FLAGS = SOME_DEBUG_FLAGS | GENERATE_ | PARSE__ | AST__
	
const NO_DEBUG_FLAGS = ALWAYS_

var DEBUG_FLAGS uint64 = NO_DEBUG_FLAGS

/*
   Set the overall debug level to 0 (none), 1 (some), or 2 (full). Debug level defaults to 0 if this method is not called.
   This is called during initialization of the relish compiler+interpreter.
*/
func InitLogging(masterLevel int32) {
	switch masterLevel {
	case 0: // No debugging
	   DEBUG_FLAGS = NO_DEBUG_FLAGS
	case 1: // Some
	   DEBUG_FLAGS = SOME_DEBUG_FLAGS	
    case 2: // Full 
	   DEBUG_FLAGS = FULL_DEBUG_FLAGS
	}
}

var LogMutex sync.Mutex
	
/*
Equivalent of fmt.Printf
*/	
func Log(flags uint64,s string,args ...interface{}) {
   if flags & DEBUG_FLAGS != 0 {
      printDots(threadNum, indent)		
      fmt.Printf(s,args...)
   }
}

/*
Equivalent of fmt.Println
*/
func Logln(flags uint64,s ...interface{}) {
   if flags & DEBUG_FLAGS != 0 {
      printDots(threadNum, indent)		
      fmt.Println(s...)
   }
}

/*
Whether we are currently logging aspects/levels identified by the specified flags.
Usage pattern: if Logging(INTERP2_) { t.Dump() }
*/
func Logging(flags uint64) bool {
   return flags & DEBUG_FLAGS != 0 	
}

// Nested printing of the message and arguments. Use single-threaded only!!
// The message might be a Go method name and args some of the arguments passed
// to it.
// Used to trace execution of the relish parser. 
// See Un() function below.
func Trace(flags uint64, msg string, args ...interface{}) string {
	if flags & DEBUG_FLAGS != 0 {
		
	   args2 := []interface{}{msg,"("}
	   args2 = append(args2,args...)
	   printTrace(threadNum, indent,args2...)		
	   //printTrace(msg, "(")
 	   indent++
	   return msg
    }
    return ""
}

// Usage pattern: defer Un(Trace("..."))
func Un(msg string) {
	if msg != "" {
	   indent--
	   printTrace(threadNum,indent,")",msg)
    }
}


/*
Multi-threaded version of Log (like fmt.Printf) used to log inside the multi-threaded relish interpreter.
*/
func LogM(context interface{}, flags uint64,s string,args ...interface{}) {

   if flags & DEBUG_FLAGS != 0 {
      LogMutex.Lock()		
      indnt,found := indents[context]
      if ! found {
            indents[context] = indnt	
            currentThreadNum++
            threadNums[context] = currentThreadNum
      }	
      printDots(threadNums[context],indents[context])		
      fmt.Printf(s,args...)
      LogMutex.Unlock()
   }

}

/*
*/
func LoglnM(context interface{}, flags uint64,s ...interface{}) {

   if flags & DEBUG_FLAGS != 0 {
      LogMutex.Lock()		

      indnt,found := indents[context]
      if ! found {
            indents[context] = indnt	
            currentThreadNum++
            threadNums[context] = currentThreadNum
      }	

      printDots(threadNums[context],indents[context])		
      fmt.Println(s...)
      LogMutex.Unlock()
   }
}

// Nested printing of the message and arguments. Multi-threaded version.
// The message might be a Go method name and args some of the arguments passed
// to it.
// Used to trace execution of the multi-threaded relish interpreter. 
// See UnM() function below.
func TraceM(context interface{}, flags uint64, msg string, args ...interface{}) string {
	if flags & DEBUG_FLAGS != 0 {
       LogMutex.Lock()		
	   args2 := []interface{}{msg,"("}
	   args2 = append(args2,args...)
	
	   indnt,found := indents[context]
	   if ! found {
	      indents[context] = indnt	
	      currentThreadNum++
	      threadNums[context] = currentThreadNum
	   }	
	   printTrace(threadNums[context],indnt,args2...)		
	
 	   indents[context]++
       LogMutex.Unlock()
	   return msg
    }
    return ""
}

/*
Lose track of a thread which is finished. (So it can be garbage collected.)
*/
func RemoveContext(context interface{}) {
	delete(indents, context)
	delete(threadNums,context)
}

// Multi-threaded version of Un() for use to trace execution of the multi-threaded relish interpreter
// Usage pattern: defer UnM(TraceM("..."))
func UnM(context interface{}, msg string) {
	if msg != "" {
       LogMutex.Lock()		
	   indents[context]--	
	   
	   printTrace(threadNums[context],indents[context],")",msg)
       LogMutex.Unlock()	
    }
}



/*
Nested (indented) printing of  message.
Prints a number of dots to the left, corresponding to logical nesting level of executing code.
Then prints the message, indented after the dots.
*/
func printTrace(tNum uint, indnt uint, a ...interface{}) {
	printDots(tNum, indnt)
	fmt.Println(a...)
}

func printDots(tNum uint, indnt uint) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . " +
		". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = uint(len(dots))
	
	if tNum > 0 {
		fmt.Printf("%02d ",tNum)
	}
	i := 2 * indnt
	for ; i > n; i -= n {
		fmt.Print(dots)
	}
	fmt.Print(dots[0:i])
}
