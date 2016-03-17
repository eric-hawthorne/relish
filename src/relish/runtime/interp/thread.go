// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the operation of the intermediate-code interpreter
// in the relish language.

package interp

/*
   thread.go - One Thread is created per goroutine. The Thread maintains an execution-stack of RObjects
*/

import (
	"fmt"
	"relish/compiler/ast"
	. "relish/dbg"
	. "relish/runtime/data"
	"errors"
	"net/http"
)


const DEFAULT_STACK_DEPTH = 50 // DEFAULT INITIAL STACK DEPTH PER THREAD - stack grows as needed

/*
If parent is nil, something else must take care of initializing 
the ExecutingMethod and ExecutingPackage attributes of the new thread.
*/
func (i *Interpreter) NewThread(parent *Thread) *Thread {
	defer UnM(parent,TraceM(parent,INTERP_TR, "NewThread"))
	return newThread(DEFAULT_STACK_DEPTH, i, parent)
}

/*
If parent is nil, something else must take care of initializing 
the ExecutingMethod and ExecutingPackage attributes of the new thread.
*/
func newThread(initialStackDepth int, i *Interpreter, parent *Thread) *Thread {

	t := &Thread{Pos: -1, Base: -1, Stack: make([]RObject, initialStackDepth), EvalContext: nil, GCLockCounter: -1}
 
	if i.rt.DB() != nil {
	   t.dbConnectionThread = i.rt.DB().NewDBThread(t)
    }	

	if parent != nil {
		t.ExecutingMethod = parent.ExecutingMethod
		t.ExecutingPackage = parent.ExecutingPackage
	} 
	t.EvalContext = &methodEvaluationContext{i, t}

	i.registerThread(t)
	
	return t
}


func (i *Interpreter) registerThread(t *Thread) {
    // GCMutexRLock(fmt.Sprintf("registerThread %p",t))	
    GCMutexRLock("")	
    t.GCLockCounter = MAX_GC_LOCKED_STACK_OPS

	defer UnM(t, TraceM(t, GC3_, "RegisterThread"))
	
	i.threads[t] = true
}

/*
Remember to call this when thread execution is done!
*/
func (i *Interpreter) DeregisterThread(t *Thread) {
    LoglnM(t, GC3_, "DeregisterThread(")
    // fmt.Println("DeregisterThread")
	delete(i.threads,t)
    RemoveContext(t) 	

    t.GCLockCounter = -1
    // GCMutexRUnlock(fmt.Sprintf("DeregisterThread %p",t))	
    GCMutexRUnlock("")	    
    Logln(GC3_, ")DeregisterThread")
}

/*
Not really a thread. Actually equivalent to a goroutine.

Each routine that is called pushes its context onto the stack, thus has a stack "frame".
The first element in the stack frame is an Int32 which holds the position on the stack of the beginning of the 
previous routine stack frame, or holds -1 if this is the first routine call pushed on the stack.
The value of a local variable or a routine parameter is found by taking the variable's Offset and adding it
to the Base, then dereferencing that stack item.
*/
type Thread struct {
	Pos         int // position of the top of the stack
	Base        int // position on stack of the beginning of the currently executing routine's stack frame
	Stack       []RObject
	EvalContext *methodEvaluationContext // Basically a self-reference, but includes a reference to the interpreter

	ExecutingMethod *RMethod       // Shortcut for dispatch efficiency
	ExecutingPackage *RPackage   // Shortcut for dispatch efficiency

	dbConnectionThread DBT  // Manages serialized and transactional access to the database in
	                              // multi-threaded environment

    // This may be temporary - it is used for generators inside collection constructors, but may
    // be replaced by proper go-routine-and-channel generators or closures.
    Objs       []RObject   // A list of objects that will be built up then become owned by a proper collection object 
                           // and detached from the Thread.
	YieldCardinality int   // How many objects per iteration the current generator is expected to append to Objs
	
	err string  // Will be "" unless we are in a stack-unrolling panic

	GCLockCounter int  // If 0, thread should RUnlock(),RLock() the GCMutex to allow GC to run.
	                   // If -1, means this thread does not have an RLock on the GCMutex.
	                   // If positive, means this thread will keep holding an RLock on GCMutex and decrementing counter

	transaction *RTransaction

	httpClient *http.Client
}

const MAX_GC_LOCKED_STACK_OPS = 100  // Do this many pops and pushes before relinquishing RLock on GCMutex.
                                     // So GC has opportunity to run and block this thread, after this many pops/pushes.




func (t *Thread) Push(obj RObject) {		
	defer UnM(t, TraceM(t, INTERP_TR3, "Push", obj))
	t.Pos++
	if len(t.Stack) <= t.Pos {
		oldStack := t.Stack
		t.Stack = make([]RObject, len(t.Stack)*2)
		copy(t.Stack, oldStack)
	}
	t.Stack[t.Pos] = obj
	if Logging(STACK_) && Logging(INTERP_TR3) {
		t.Dump()
	}

    // Manage locking of the garbage collector RWMutex
	if t.GCLockCounter == 0 {
		t.GCLockCounter = -1
        // GCMutexRUnlock("Push")		
        GCMutexRUnlock("")
        // Garbage Collector goroutine can block me and run in here
        // GCMutexRLock("Push")        
        GCMutexRLock("")
        t.GCLockCounter = MAX_GC_LOCKED_STACK_OPS
	} else {
		t.GCLockCounter--
	}
}

func (t *Thread) Reserve(n int) {		
	defer UnM(t, TraceM(t, INTERP_TR3, "Reserve", n))
	for len(t.Stack) <= t.Pos+n {
		oldStack := t.Stack
		t.Stack = make([]RObject, len(t.Stack)*2)
		copy(t.Stack, oldStack)
	}
	t.Pos += n
}


/*
Call this at the beginning of a method call, to push the previous (outer) method's stack frame base onto the stack.
When finished evaluating arguments to the call, then complete the context store by calling setBase with the
value that has been returned by PushBase().
The numReturnArgs argument says how many stack positions to leave free below the base-pointer position for
results (return values) of the method call.

Returns the position of the base pointer which holds the stack-pointer to the previous stack-frame base position.
*/
func (t *Thread) PushBase(numReturnArgs int) int {		
	defer UnM(t, TraceM(t, INTERP_TR3, "PushBase"))

	if numReturnArgs > 0 {
		t.Reserve(numReturnArgs)
	}
	t.Push(Int32(t.Base))
	t.Reserve(2) // Reserve space for the currently-executing-method reference and code offset in current method	
	return t.Pos - 2  // Return the stack position where the previous base-index is now stored.
	                  // This position will become the new "base" position, when SetBase is soon called
	                  // to complete the switch into the new method's stack context.
	
}

func (t *Thread) SetBase(newBase int) {
	defer UnM(t, TraceM(t, INTERP_TR3, "SetBase", newBase))
	t.Base = newBase
}

/*
Call this to return from a method call.
Pops the stack to just before the beginning of the current routine's stack frame, and
sets the thread's Base pointer to the beginning of the previous (outer) routine's stack frame.


                            B       P
0   1   2   3   4   5   6   7   8   9
-1  v1  v2  0   v1  v2  v3  3   v1  v2

*/
func (t *Thread) PopBase() {
	defer UnM(t, TraceM(t, INTERP_TR3, "PopBase"))
	obj := t.PopN(t.Pos - t.Base + 1) // 9 - 7 + 1 = 3
	t.Base = int(obj.(Int32))
	if t.Base == -2 {
		t.ExecutingMethod = nil
		t.ExecutingPackage = nil
	} else {		
	    t.ExecutingMethod = t.Stack[t.Base+1].(*RMethod)
	    t.ExecutingPackage = t.ExecutingMethod.Pkg
    }
	LogM(t, INTERP3_, "---Base = %d\n", t.Base)
}

/*
Call this to return from the initial method call in a new Thread used for GoRoutine execution.
Pops all objects off the stack, and
sets the Pos to -1 thread's Base pointer to -2, indicating a de-activated goroutine thread.
*/
func (t *Thread) PopFinalBase(numReturnArgs int) {
	defer UnM(t, TraceM(t, INTERP_TR3, "PopBase"))
	obj := t.PopN(t.Pos - t.Base + numReturnArgs + 1) // 9 - 7 + 1 = 3
	t.Base = int(obj.(Int32))  // Sets to -2 meaning A de-activated thread.
	t.ExecutingMethod = nil
	t.ExecutingPackage = nil
	LogM(t, INTERP3_, "---Final Base = %d\n", t.Base)
}


// Note: These should be supplemented by a full stack backtrace capability.

/*
Returns the method that called the currently executing method, 
or nil if at stack bottom.
*/
func (t *Thread) CallingMethod() *RMethod {
	defer UnM(t, TraceM(t, INTERP_TR3, "CallingMethod"))
	if t.Base >= 0 {
		previousBase := int(t.Stack[t.Base].(Int32))

		if previousBase == -2 {
			return nil
		} 		
		return t.Stack[previousBase+1].(*RMethod)
    }
    return nil
}

/*
Returns the package from which the currently executing method was called, 
or nil if at stack bottom.
*/
func (t *Thread) CallingPackage() *RPackage {
	defer UnM(t, TraceM(t, INTERP_TR3, "CallingPackage"))
	method := t.CallingMethod()
    if method == nil {
    	return nil
    }
	return method.Pkg
}


/*
Return the value of the local variable (or parameter) with the given offset from the current routine's 
stack base.
*/
func (t *Thread) GetVar(offset int) (obj RObject, err error) {
	defer UnM(t, TraceM(t, INTERP_TR3, "GetVar", "offset", offset, "stack index", t.Base+offset))
	if offset == -99 {
		err = errors.New("Unassigned variable.")
		return
	}
	obj = t.Stack[t.Base+offset]
	return
}

func (t *Thread) Pop() RObject {

    // Manage locking of the garbage collector RWMutex
	if t.GCLockCounter == 0 {
		t.GCLockCounter = -1
        // GCMutexRUnlock("pop")
        GCMutexRUnlock("")        
        // Garbage Collector goroutine can block me and run in here
        // GCMutexRLock("pop")
        GCMutexRLock("")        
        t.GCLockCounter = MAX_GC_LOCKED_STACK_OPS
	} else {
		t.GCLockCounter--
	}	

	obj := t.Stack[t.Pos]
	defer UnM(t, TraceM(t, INTERP_TR3, "Pop", "==>", obj))
	t.Stack[t.Pos] = nil // ensure var/param value is garbage-collectable if not otherwise referred to.
	t.Pos--
	if Logging(STACK_) && Logging(INTERP_TR3) {
		t.Dump()
	}
	return obj
}



/*
Efficiently pop n items off the stack.
*/
func (t *Thread) PopN(n int) RObject {

    // Manage locking of the garbage collector RWMutex
	if t.GCLockCounter == 0 {
		t.GCLockCounter = -1
        // GCMutexRUnlock("popN")
        GCMutexRUnlock("")        
        // Garbage Collector goroutine can block me and run in here
        // GCMutexRLock("popN")
        GCMutexRLock("")        
        t.GCLockCounter = MAX_GC_LOCKED_STACK_OPS
	} else {
		t.GCLockCounter--
	}
	
	lastPopped := t.Pos - n + 1
	obj := t.Stack[lastPopped]
	defer UnM(t, TraceM(t, INTERP_TR3, "PopN(", n, ") ==> val:", obj," | Pos: ", t.Pos - n))
	for i := t.Pos; i >= lastPopped; i-- {
		t.Stack[i] = nil // ensure var/param value is garbage-collectable if not otherwise referred to.
	}
	t.Pos -= n
	if Logging(STACK_) && Logging(INTERP_TR3) {
		t.Dump()
	}
	return obj
}

/*
Return a slice which represents the top n objects on the thread's stack.
In order from bottom most (oldest pushed) to top most (most recently pushed).
*/
func (t *Thread) TopN(n int) []RObject {
	return t.Stack[t.Pos-n+1 : t.Pos+1]
}

/*
Preparatory to executing the first-called method in a new goroutine, on t the new Stack,
we copy 
*/
func (t *Thread) copyStackFrameFrom(parent *Thread, numReturnVals int) {
   n := parent.Pos - parent.Base + numReturnVals + 1
   // defer UnM(t, TraceM(t, INTERP_TR3, "copyStackFrameFrom: copying ",n))	
   src := parent.Stack[parent.Base - numReturnVals:parent.Pos+1]
   if copy(t.Stack, src) != n {
   	   panic("stack copy range exception during go-routine spawn.")
   }
   t.Base = numReturnVals
   t.Stack[t.Base] = Int32(-2)  // Indicates that thread will be de-activated if this base is ever popped.
   t.Pos = n - 1	   
}

/*
Manage locking of the garbage collector RWMutex. RUnlock the mutex.
Call this just before this goroutine (relish thread) is going to be potentially blocked.
Allow garbage collection to proceed while this goroutine (relish thread) is blocked.
You must call DisallowGC() as soon as the potentially-blocking operation is complete.
*/
//func (t *Thread) AllowGC(msg string) {
func (t *Thread) AllowGC() {
//	fmt.Printf("%p.AllowGC\n",t)
    if t.GCLockCounter == -1 {
    	panic("thread is attempting to doubly runlock the GCMutex.")
    }
	t.GCLockCounter = -1
    //GCMutexRUnlock(msg)	
    GCMutexRUnlock("")
    // Garbage Collector goroutine can block me and run in here
}

/*
Manage locking of the garbage collector RWMutex. RLock the mutex.
Call this just after this goroutine (relish thread) completes an operation that is going to 
be potentially blocked.
*/
//func (t *Thread) DisallowGC(msg string) {
func (t *Thread) DisallowGC() {
//	fmt.Printf("%p.DisallowGC\n",t)
    if t.GCLockCounter != -1 {
    	panic("thread is attempting to doubly rlock the GCMutex.")
    }
    //GCMutexRLock(msg)
    GCMutexRLock("")    
    t.GCLockCounter = MAX_GC_LOCKED_STACK_OPS
}


func (t *Thread) GC() {
   t.AllowGC()
   t.EvalContext.GC()
   t.DisallowGC()
}


/*
Mark all structured objects on the stack as reachable and safe from garbage collection.
*/
func (t *Thread) Mark() {
	defer UnM(t, TraceM(t, GC2_, "Mark"))
	

    
	for i := t.Pos; i >= 0; i-- {
		if t.Stack[i] != nil {
			t.Stack[i].Mark()
		}
	}
	
    for _,obj := range t.Objs {
	   obj.Mark()
    }	

    LoglnM(t,GC2_,"  Marked",t.Pos + 1 + len(t.Objs),"objects on stack and their associates.")
   
	// When finished, signal somehow.
	// The philosophy should be each stack has to signal when it is finished marking,
	// then it has to wait on an RLock() of an RWMutex for the RT.GC() to finish and unlock the RWMutex.

	// So interpreter.GC() should 
	// a) Know how many threads are active
	// b) lightweight flag each thread (set a bool in the thread) to begin marking
	//     - It's ok if the flag is missed for a while
	// - Flags actually have to wait for all of them to rendezvous before
	//   commencing to mark, so that it is safe to traverse a collection and mark its members
	// c) wait for all threads to finish marking
	// d) proceed with the sweep 
}

/*
DEBUG printout of stack
*/
func (t *Thread) Dump() {
	LogMutex.Lock()
	fmt.Println("------STACK----------")
	for i := t.Pos; i >= 0; i-- {
		fmt.Printf("%3d: %v\n", i, t.Stack[i])
	}
	fmt.Printf("Pos : %d\n", t.Pos)
	fmt.Printf("Base : %d\n", t.Base)
	fmt.Println("---------------------")
	LogMutex.Unlock()
}



func (t *Thread) CodeFile() (file *ast.File) {
   file = t.ExecutingMethod.CodeFile()
   if file == nil {
      caller := t.CallingMethod()   
      if caller != nil {
         file = caller.CodeFile()
      }
   }
   return
}

/*
The package context of the executing method.
*/
func (t *Thread) Package() *RPackage {
	return t.ExecutingPackage
}
	
/*
The executing method.
*/
func (t *Thread) Method() *RMethod {
	return t.ExecutingMethod
}



/*
The DBThread which can execute db queries in a serialized fashion in a multi-threaded environment.
*/
func (t *Thread) DBT() DBT {
   return t.dbConnectionThread
}

func (t *Thread) Err() string {
	return t.err
}

func (t *Thread) SetErr(msg string) {
	t.err = msg
}

func (t *Thread) EvaluationContext() MethodEvaluationContext {
	return t.EvalContext
}

/*
  The currently active DB transaction which this thread started or is participating in.
*/
func (t *Thread) Transaction() *RTransaction {
   return t.transaction
}

func (t *Thread) SetTransaction(tx *RTransaction) {
	t.transaction = tx
}

func (t *Thread) HttpClient() *http.Client {
	if t.httpClient == nil {
		t.httpClient = &http.Client{}
	}
	return t.httpClient
}
