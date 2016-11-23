// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the operation of the intermediate-code interpreter
// in the relish language.

package interp

/*
   interpreter.go - Interpreter 

   Runs a relish program, after the code objects such as datatypes, methods, constants etc
   have been loaded into the runtime.
   
   Each relish method ** has a code body which is represented by an abstract-syntax-tree (ast).

   The method's ast is executed (interpreted) directly by the interpreter. 
   There is no translation into a sequence of virtual-machine instruction codes.  

   relish compilation (prior to interpretation) then, is simply the parsing of relish syntax to create the
   asts. Relish intermediate-code files are simply GOB-serialized asts which are written as .rlc files by the
   compiler, then loaded from file and de-serialized back into an in-memory tree prior to interpretation. 

   ** Some relish methods (relish built-in methods and some standard-library methods) 
      actually have a Go-native implementation, rather than an ast to be interpreted.


   A note on debugging relish runtime errors which may be caused by an error in the relish implementation,
   assuming you have a source distribution of relish and can recompile relish:

   Search for "defer m" in this interpreter.go file.  
   Comment out the  
   		defer methodCallErrHandle(t,call)
   and go install relish/relish

   This will get you a relish interpreter which will panic with a full Go stack trace on the occurrence of 
   the runtime error. Without commenting out the methodCallErrHandle, you will instead get an error
   message which gives the relish source file line number which induced the runtime error, but no Go stack trace.
*/

import (
	"fmt"
	"relish/compiler/ast"
	"relish/compiler/token"
	. "relish/dbg"
	. "relish/runtime/data"
	"relish/rterr"
	"strconv"
	"net/url"
	"strings"
	. "relish/defs"
	"net/http"
)



type Interpreter struct {
	rt         *RuntimeEnv
	dispatcher *dispatcher
	threads    map[*Thread]bool  // goroutines running in this interpreter 
}

func NewInterpreter(rt *RuntimeEnv) *Interpreter {
	return &Interpreter{rt: rt, dispatcher: newDispatcher(rt), threads:make(map[*Thread]bool)}
}

func (i *Interpreter) Dispatcher() Dispatcher {
	return i.dispatcher
}

/*
Give the runtime knowledge of which artifact was originally run by the relish command.
Used to determine which web package tree to use for handling web requests.
*/
func (i *Interpreter) SetRunningArtifact(originAndArtifact string) {
	i.rt.RunningArtifact = originAndArtifact
}


/* 
Give the runtime access to the package loader.
*/
func (i *Interpreter) SetPackageLoader(loader PackageLoader) {
	i.rt.Loader = loader
}


/*
Runs the main method found in the specified package. 

Detail: Currently, when multimethods/methods are generated, "main" ones are prefixed by full unversioned package name
(as are zero arg methods).
*/
func (i *Interpreter) RunMain(fullUnversionedPackagePath string, quiet bool) {
	if ! quiet {
	   Logln(ANY_, " ")
	   Logln(ANY_, "==============================")
	   Logln(ANY_, "== RELISH Interpreter 0.1.3 ==")
	   Logln(ANY_, "==============================")
    }
	Logln(GENERATE2_, " ")
	Logln(GENERATE2_, "----------")
	Logln(GENERATE2_, "Data Types")
	Logln(GENERATE2_, "----------")
	if Logging(GENERATE2_) {
		i.rt.ListTypes()
	}
	Logln(GENERATE2_, "----------")	
	
    pkg, pkgFound := i.rt.Packages[fullUnversionedPackagePath]
    if !pkgFound {
	   rterr.Stop("Package for main function is not loaded. Is package name in .rel file different than package directory?")
    }
	mm, found := pkg.MultiMethods[fullUnversionedPackagePath + "/main"]
	if !found {
		rterr.Stop("No main function defined.")
	}
	t := i.NewThread(nil)

	args := []RObject{}
	
	defer methodCallErrHandle(t,nil)	

	// TODO Figure out a way to pass command line args (or maybe just keyword ones) to the main program 
	
	method, typeTuple := i.dispatcher.GetMethod(mm, args)

	if method == nil {
		rterr.Stopf("No method '%s' is compatible with %s", mm.Name, typeTuple)
	}
	t.Push(method)
	t.Reserve(1) // For code offset pointer within method bytecode (future)

	t.Reserve(method.NumLocalVars)
	
	t.ExecutingMethod = method
	t.ExecutingPackage = pkg

    go i.GCLoop()

	err := i.apply1(t, method, args)
    if err != nil {
    	rterr.Stopf("Error calling main: %s",err.Error())
    }

	t.PopN(t.Pos + 1) // Pop everything off the stack for good measure.	

	i.DeregisterThread(t)
}


/*
   Return the annotations of the URL-mapped method.
   These determine transaction behaviour etc.
*/
func (i *Interpreter) GetServiceMethodModifiers(mm *RMultiMethod) (modifiers map[string]bool, err error) {
	method := i.dispatcher.GetSingletonMethod(mm)

	if method == nil {
		err = fmt.Errorf("No method '%s' found.", mm.Name)
		return
	}
    modifiers = method.Code.ModifierKeywords
    return
}


/*
Runs a service handler method. 
Assumes there is only one method for the multimethod, so a simpler dispatch that does not consider arg types is used.
Converts the arguments, whose values are all passed in as strings, into the appropriate primitive datatypes to match
the method's parameter type signature, after matching the correct input arguments to method parameters.
The parameter to argument matching process is as follows:
1. Named arguments are used as the values of the matching named method parameter.
2. Additional unmatched method parameters are filled (left to right i.e. top to bottom) from the positionalArgStringValues list.
future:
3. Additional positional and keyword arguments are assigned to variadic and kw parameters of the method, if such exist.
Matching discrepencies such as unmapped leftover arguments, or not enough arguments, cause a non-nil error return value.


Runs the method in a new stack and returns a slice of the method's return values.

TODO WE HAVE SOME SERIOUS DRY VIOLATIONS between this method and a few others in this file.

TODO See EvalMethodCall for updates (to nArgs etc) that have not been applied here yet!!!!
*/
func (i *Interpreter) RunServiceMethod(t *Thread, mm *RMultiMethod, positionalArgStringValues []string, keywordArgStringValues url.Values, request *http.Request) (resultObjects []RObject, err error) {
	// defer UnM(t,TraceM(t,INTERP_TR, "RunServiceMethod", fmt.Sprintf("%s", mm.Name)))	
	
	method := i.dispatcher.GetSingletonMethod(mm)

	if method == nil {
		err = fmt.Errorf("No method '%s' found.", mm.Name)
		return
	}
	
	args, err := i.matchServiceArgsToMethodParameters(method, positionalArgStringValues, keywordArgStringValues, request)
	if err != nil {
		return
	}



	nReturnArgs := mm.NumReturnArgs 
	
    if nReturnArgs > 0 {
	    t.Reserve(nReturnArgs)
    }
    t.Push(Int32(t.Base))  // Useless - pushing -1 here
    t.Base = t.Pos

	t.Push(method)
	t.Reserve(1) // For code offset pointer within method bytecode (future)
	
	


	// NOTE NOTE !!
	// At some point, when leaving this context, we may want to also push just above this the offset into the method's code
	// where we left off. We might wish to leave a space on the stack for that, and make initial variableOffset 3 instead of 2

	for _, arg := range args {
		t.Push(arg)
	}	
	

	t.Reserve(method.NumLocalVars)
	
	t.ExecutingMethod = method
	t.ExecutingPackage = method.Pkg
 
	err = i.apply1(t, method, args)
	if err != nil {
    	rterr.Stopf("Error calling web service handler %s: %s",method, err.Error())
    }
	
	t.PopN(t.Pos - t.Base + 1) 	// Leave only the return values on the stack
	
	t.SetBase(-2)  // Set to invalid - indicating that no function is running.
	
    resultObjects = t.TopN(nReturnArgs)  // Note we are hanging on to the stack array here.
	return 
}

/*
TODO: Why is  this not in thread.go ???

In future, a thread panic would cause t.err to be an error message.

TODO: Make some way of setting t.err when relish-panicking.

TODO TODO: If error on the commit, wait, try again a couple of times, backing off,
then try a rollback.
If error on the rollback, wait, try again a couple of times, backing off, 
then do a releaseDB.


TODO: If we get an exception percolating all the way up to here, should we
have a recover(..) call in here so as to roll back the transaction. Seems like the
best thing to do. Not sure if any exceptions can get up this high.

*/
func (t *Thread) CommitOrRollback() {
	if t.err == "" {
	   err := t.DBT().CommitTransaction()
	   if err != nil {
	      Logln(ALWAYS_,err.Error())

	      err := t.DBT().RollbackTransaction()
	      if err != nil {
	         Logln(ALWAYS_,err.Error())
          }
	      if t.transaction != nil {
	          t.transaction.RollBack()
	      }  	      
       } else {  // successful commit
	      if t.transaction != nil {
	         t.transaction.Commit()
	      }     	
	   }	   
    } else {
	   err := t.DBT().RollbackTransaction()
	   if err != nil {
	      Logln(ALWAYS_,err.Error())
       }
	
       if t.transaction != nil {
          t.transaction.RollBack()
       }
	
	   for ! t.DBT().ReleaseDB() {}  // Loop til we definitely unlock the dbMutex
    }
}


/*
Converts the arguments, whose values are all passed in as strings, into the appropriate primitive datatypes to match
the method's parameter type signature, after matching the correct input arguments to method parameters.
The parameter to argument matching process is as follows:
1. Named arguments are used as the values of the matching named method parameter.
2. Additional unmatched method parameters are filled (left to right i.e. top to bottom) from the positionalArgStringValues list.
future:
3. Additional positional and keyword arguments are assigned to variadic and kw parameters of the method, if such exist.
Matching discrepencies such as unmapped leftover arguments, or not enough arguments, cause a non-nil error return value.

TODO - Not handling multi valued arguments, that need to be mapped to a list of strings or list of ints parameter

RMethod
	parameterNames []string               // names of parameters
	Signature      *RTypeTuple            // types of parameters
*/
func (i *Interpreter) matchServiceArgsToMethodParameters(method *RMethod, positionalArgStringValues []string, keywordArgStringValues url.Values, request *http.Request) (args []RObject, err error) {
   
   arity := len(method.ParameterNames)
   args = make([]RObject,arity)
   paramTypes := method.Signature.Types

   firstNonSpecialArgIndex := 0
   // Special case. If first parameter of the method is of type http.Request, create
   // a wrapped Go-native object representing the request.
   if arity > 0 && paramTypes[0].IsNative && paramTypes[0].Name == "shared.relish.pl2012/relish_lib/pkg/http_srv/Request" {	
	  args[0],err = i.createRequest(request)
	  if err != nil {		   
	  	 return
	  }  			  
	  firstNonSpecialArgIndex = 1
   }
	
   var extraArgKeys []string
   for key := range keywordArgStringValues {
      foundKey := false
      for ix,paramName := range method.ParameterNames {
         if ix < firstNonSpecialArgIndex {
	        continue
	     }
      	 if paramName == key {
            valStr := keywordArgStringValues.Get(key)   
//            if valStr == "" {
//            	panic(fmt.Sprintf("How is it that the value of argument '%s' is the empty string? Shouldn't be able to happen.", key))
//            }

            // Convert string arg to an RObject, checking for type conversion errors
            err = i.setMethodArg(args, paramTypes, ix, key, valStr) 
			if err != nil {
			   return
			}            
            
            foundKey = true
            break	 	
      	 }
      }
      if ! foundKey {
      	extraArgKeys = append(extraArgKeys, key)
      }
   }

   // Finished handling keyword args whose keys are method parameter names.
   // Now check to see if the method declared a wildcardKeywords parameter.
   // If so, append the remaining unmatched key=val arguments to a new Map, and 
   // make that map the value of the wildcardKeywords parameter.
   // If the method does not declare that it expects wildcard keywords, 
   // throw an error if the web request included extra unmatched key=val arguents.

   var keywordsValType *RType
   var keywordsArgMap Map
   if method.WildcardKeywordsParameterName != "" {  // method has a keywords parameter  
	
      keywordsValType = method.WildcardKeywordsParameterType.ValType()

	  keywordsArgMap = method.WildcardKeywordsParameterType.Prototype().(Map)	
	
	  for _,key := range extraArgKeys {
	     	
          valStr := keywordArgStringValues.Get(key)   
//          if valStr == "" {
//          	panic(fmt.Sprintf("How is it that the value of argument '%s' is the empty string? Shouldn't be able to happen.", key))
//          }	
	
		  var obj RObject
          obj, err = i.variadicArg(keywordsValType, valStr)  
	      if err != nil {   // Parameter type incompatibility
		     return
		  }
          keywordsArgMap.PutSimple(String(key), obj)	
	  }
	
	  // Now what? assign the keywordArgMap to something.
   	
	 
   } else if len(extraArgKeys) == 1 {
       err = fmt.Errorf("Web service request has extra argument %s.",extraArgKeys[0])
       return
   } else if len(extraArgKeys) > 1 {
       err = fmt.Errorf("Web service request has extra arguments %v.",extraArgKeys)
       return
   }

   var variadicElementType *RType
   var variadicArgList List
   if method.VariadicParameterName != "" {
      variadicElementType = method.VariadicParameterType.ElementType()
	  variadicArgList = method.VariadicParameterType.Prototype().(List)
   }
   exhaustedPositionalArgs := false

   // Now map URL path components to unfilled method parameters

   for j,valStr := range positionalArgStringValues {
   	  slotFound := false
      if ! exhaustedPositionalArgs {
	   	  for ix,v := range args {
	   	  	 if v == nil {
			      // Convert string arg to an RObject, checking for type conversion errors
			     err = i.setMethodArg(args, paramTypes, ix, valStr, valStr) 
			     if err != nil { // Parameter type incompatibility
				     return
				 } 
	             slotFound = true
	             break
	   	  	 }
	   	  }
      }
   	  if ! slotFound {   // Raise error if too many URL path components in web request
	      exhaustedPositionalArgs = true
   	  	  nExtraArgs := len(positionalArgStringValues) - j

	      if method.VariadicParameterName != "" {
		     var obj RObject
             obj, err = i.variadicArg(variadicElementType, valStr) 
	         if err != nil {   // Parameter type incompatibility
		        return
		     }
             variadicArgList.AddSimple(obj)

          } else {   // no variadic parameter to absorb extra arguments.
	   	  	  if nExtraArgs == 1 {
	             err = fmt.Errorf("Web service request has %d extra URI path component that doesn't map to a handler method parameter.",nExtraArgs)   
	          } else {	  	  
	             err = fmt.Errorf("Web service request has %d extra URI path components that don't map to handler method parameters.",nExtraArgs)
	          }
              return	
          }
   	  }   
   }

   if method.WildcardKeywordsParameterName != "" {
      args = append(args,keywordsArgMap)
   }

   if method.VariadicParameterName != "" {
      args = append(args,variadicArgList)
   }


   // Raise error if too few URL path components and other arguments combined to satisfy the handler method parameter signature.

   for _,v := range args {
   	  if v == nil {
   	  	  err = fmt.Errorf("Too few arguments (or URI path components) supplied to web service request.")
   	  	  return 
   	  }
   }

   return
}


/*
Converts the valStr to the type of the ith parameter of a method, and sets args[i] to the resulting RObject.
Returns type conversion errors.
argKey should be set to the valStr if the argument is not a keyword argument.
*/
func (interp *Interpreter) setMethodArg(args []RObject, paramTypes []*RType, i int, argKey string, valStr string) (err error) {
    paramType := paramTypes[i]
    switch paramType {
	case IntType : 
       var v64 int64
       v64, err = strconv.ParseInt(valStr, 0, 64)  
	   if err != nil {
		   return
	   }
	   args[i] = Int(v64)		

	case Int32Type : 
       var v64 int64
       v64, err = strconv.ParseInt(valStr, 0, 32)  
	   if err != nil {
		   return
	   }
	   args[i] = Int32(int32(v64))				

	case Int16Type : rterr.Stop("Int16 Parameter type not yet allowed in a web service handler method.")
	case Int8Type : rterr.Stop("Int8 Parameter type not yet allowed in a web service handler method.")
	case UintType : 
       var v64 uint64
       v64, err = strconv.ParseUint(valStr, 0, 64)  
	   if err != nil {
		   return
	   }
	   args[i] = Uint(v64)	

	case Uint32Type : 
       var v64 uint64
       v64, err = strconv.ParseUint(valStr, 0, 32)  
	   if err != nil {
		   return
	   }
	   args[i] = Uint32(uint32(v64))

	case Uint16Type : rterr.Stop("Uint16 Parameter type not yet allowed in a web service handler method.")
	case ByteType : rterr.Stop("Byte Parameter type not yet allowed in a web service handler method.")
	case BitType : rterr.Stop("Bit Parameter type not yet allowed in a web service handler method.")
	case BoolType : 
       var v bool
       v, err = strconv.ParseBool(valStr)  
	   if err != nil {
		   return
	   }
	   args[i] = Bool(v)

	case FloatType : 
       var v64 float64
       v64, err = strconv.ParseFloat(valStr, 64)  
	   if err != nil {
		   return
	   }
	   args[i] = Float(v64)

	case Float32Type : rterr.Stop("Float32 Parameter type not yet allowed in a web service handler method.")
	case StringType :
       args[i] = String(valStr)

       // TODO !! IMPORTANT !! Should I be checking/filtering for injection attacks here? Do Go net libs provide help?

    default: 
       err = fmt.Errorf("Cannot map web service request argument %s to a parameter of type %s",argKey,paramType.Name)
    }
    return
}

/*
Converts the valStr to the specified type.
Returns type conversion errors.
*/
func (interp *Interpreter) variadicArg(paramType *RType, valStr string) (obj RObject, err error) {
    switch paramType {
	case IntType : 
       var v64 int64
       v64, err = strconv.ParseInt(valStr, 0, 64)  
	   if err != nil {
		   return
	   }
	   obj = Int(v64)		

	case Int32Type : 
       var v64 int64
       v64, err = strconv.ParseInt(valStr, 0, 32)  
	   if err != nil {
		   return
	   }
	   obj = Int32(int32(v64))				

	case Int16Type : rterr.Stop("Int16 Parameter type not yet allowed in a web service handler method.")
	case Int8Type : rterr.Stop("Int8 Parameter type not yet allowed in a web service handler method.")
	case UintType : 
       var v64 uint64
       v64, err = strconv.ParseUint(valStr, 0, 64)  
	   if err != nil {
		   return
	   }
	   obj = Uint(v64)	

	case Uint32Type : 
       var v64 uint64
       v64, err = strconv.ParseUint(valStr, 0, 32)  
	   if err != nil {
		   return
	   }
	   obj = Uint32(uint32(v64))

	case Uint16Type : rterr.Stop("Uint16 Parameter type not yet allowed in a web service handler method.")
	case ByteType : rterr.Stop("Byte Parameter type not yet allowed in a web service handler method.")
	case BitType : rterr.Stop("Bit Parameter type not yet allowed in a web service handler method.")
	case BoolType : 
       var v bool
       v, err = strconv.ParseBool(valStr)  
	   if err != nil {
		   return
	   }
	   obj = Bool(v)

	case FloatType : 
       var v64 float64
       v64, err = strconv.ParseFloat(valStr, 64)  
	   if err != nil {
		   return
	   }
	   obj = Float(v64)

	case Float32Type : rterr.Stop("Float32 Parameter type not yet allowed in a web service handler method.")
	case StringType :
       obj = String(valStr)

       // TODO !! IMPORTANT !! Should I be checking/filtering for injection attacks here? Do Go net libs provide help?

    default: 
       err = fmt.Errorf("Cannot map web service request argument to a parameter of type %s",paramType.Name)
    }
    return
}






/*
Runs the multimethod in a new stack and returns a slice of the method's return values 

TODO Check this against EvalMethodCall to see if the improvements made there have been implemented here.

IS THIS EVEN USED ANYMORE??

*/
func (i *Interpreter) RunMultiMethod(mm *RMultiMethod, args []RObject) (resultObjects []RObject) {
//	defer Un(Trace(INTERP_TR, "RunMultiMethod", fmt.Sprintf("%s", mm.Name)))	


	t := i.NewThread(nil)
	
	
	method, typeTuple := i.dispatcher.GetMethod(mm, args)

	if method == nil {
		rterr.Stopf("No method '%s' is compatible with %s", mm.Name, typeTuple)
	}
	
	
	nReturnArgs := mm.NumReturnArgs 
	
    if nReturnArgs > 0 {
	    t.Reserve(nReturnArgs)
    }
    t.Push(Int32(t.Base))  // Useless - pushing -1 here
    t.Base = t.Pos

	t.Push(method)
	t.Reserve(1) // For code offset pointer within method bytecode (future)
	
	


	// NOTE NOTE !!
	// At some point, when leaving this context, we may want to also push just above this the offset into the method's code
	// where we left off. We might wish to leave a space on the stack for that, and make initial variableOffset 3 instead of 2

	for _, arg := range args {
		t.Push(arg)
	}	
	

	t.Reserve(method.NumLocalVars)
	
	t.ExecutingMethod = method
	t.ExecutingPackage = method.Pkg

	err := i.apply1(t, method, args)
	if err != nil {
    	rterr.Stopf("Error calling %s: %s",mm.Name, err.Error())
    }	
	
	
	t.PopN(t.Pos - t.Base + 1) 	// Leave only the return values on the stack
	
    resultObjects = t.TopN(nReturnArgs)  // Note we are hanging on to the stack array here.
	return 
}









func (i *Interpreter) EvalExpr(t *Thread, expr ast.Expr) {
	switch expr.(type) {
	case *ast.MethodCall:
		i.EvalMethodCall(t, nil, expr.(*ast.MethodCall))
	case *ast.BasicLit:
		i.EvalBasicLit(t, expr.(*ast.BasicLit))
	case *ast.Ident:
		ident := expr.(*ast.Ident)
		switch ident.Kind {
		case token.VAR:
			i.EvalVar(t, ident)
		case token.CONST:
			i.EvalConst(t, ident)
		}
	case *ast.SelectorExpr:
		i.EvalSelectorExpr(t, expr.(*ast.SelectorExpr))

	case *ast.IndexExpr:
		i.EvalIndexExpr(t, expr.(*ast.IndexExpr))
		
	case *ast.SliceExpr:
		i.EvalSliceExpr(t, expr.(*ast.SliceExpr))		

	case *ast.ListConstruction:
		i.EvalListConstruction(t, expr.(*ast.ListConstruction))		
		
	case *ast.SetConstruction:
		i.EvalSetConstruction(t, expr.(*ast.SetConstruction))
		
	case *ast.MapConstruction:
		i.EvalMapConstruction(t, expr.(*ast.MapConstruction))	
		
	case *ast.Closure:
		i.EvalClosure(t, expr.(*ast.Closure))	
					
	}

}


/*
   someExpr[indexExpr]

  Evaluates someExpr to yield a collection or map. Evaluates the index exprssion to yield the index or key.
  Applies the index/key to the collection/map and leaves on the stack the value found at the index/key.

    val = list1[2]

    val = map1[! "Four"]        // return val stored under key or zero-val/nil
    val found = map1["Four"]  // return val stored under key, or zero-val/nil, and also whether key is found in the map
    found = map1[? "Four"]    // query as to whether the key is found in the map.  
*/
func (i *Interpreter) EvalIndexExpr(t *Thread, idxExpr *ast.IndexExpr) {
	// defer UnM(t,TraceM(t,INTERP_TR3, "EvalIndexExpr"))

   var val RObject

   i.EvalExpr(t, idxExpr.X) // Evaluate the left part of the index expression.		      

   i.EvalExpr(t, idxExpr.Index) // Evaluate the inside-square-brackets part of the index expression.		


   obj := t.Stack[t.Pos-1]  // the object to be indexed into
   idx := t.Stack[t.Pos]    	

   collection,isCollection := obj.(RCollection)
   if ! isCollection {
		rterr.Stopf1(t,idxExpr,"[ ] (access by index/key) applies to a collection or map; not to an object of type %v. ", obj.Type())
   }

   if collection.IsOrdered() {
        coll := collection.(OrderedCollection)
        var ix int
        switch idx.(type) {
        case Int:
        	ix = int(int64(idx.(Int)))
        case Int32:
        	ix = int(int32(idx.(Int32)))
        case Uint:
        	ix = int(uint64(idx.(Uint)))                    	                    
        case Uint32:
        	ix = int(uint32(idx.(Uint32)))
        default:
		   rterr.Stop1(t,idxExpr,"Index value must be an Integer")
		}
		defer indexErrHandle(t, idxExpr)
        val = coll.At(t, ix)

        if val == nil {
        	val = collection.ElementType().Zero()
        }

        t.PopN(2) // Pop off the collection and its index
	    t.Push(val)        

   } else if collection.IsMap() {

   	   var found bool
   	   theMap := collection.(Map)
         val,found = theMap.Get(idx) 

        t.PopN(2) // Pop off the collection and its index

        if idxExpr.AskWhether {
           t.Push(Bool(found))
        } else {

           if val == nil {
        	  val = theMap.ValType().Zero()
           }

	       t.Push(val)  
	       if ! idxExpr.AssertExists {
        	  t.Push(Bool(found))
	       }
	    }        

   } else {
		rterr.Stopf1(t,idxExpr,"[ ] (access by index/key) applies to an ordered collection or a map; not to a %v. ", obj.Type())
   }
}

func indexErrHandle(t *Thread, idxExpr *ast.IndexExpr) {
      r := recover()	
      if r != nil {
          rterr.Stopf1(t,idxExpr,"%v",r)
      }	
}	

/*
   someExpr[lowIndexExpr:highIndexExpr]
   someExpr[lowIndexExpr:]     // to the end
   someExpr[:highIndexExpr]      // from the beginning
   someExpr[:]    // a copy of the list

  Evaluates someExpr to yield a list collection. Evaluates the index expressions to yield the low and 
  high indexes. 
  Copies the slice of the list specified by the low and high indexes.
  Leaves on the stack the slice copy. 
*/
func (i *Interpreter) EvalSliceExpr(t *Thread, sliceExpr *ast.SliceExpr) {
	// defer UnM(t,TraceM(t,INTERP_TR3, "EvalSliceExpr"))

   var val RObject

   i.EvalExpr(t, sliceExpr.X) // Evaluate the left part of the index expression.		      

   obj := t.Pop() // the object to be indexed into
   list,isList := obj.(List)
   if ! isList {
		rterr.Stopf1(t, sliceExpr,"[:] slicing applies to a list; not to an object of type %v. ", obj.Type())
   }

   var low, high int
   if sliceExpr.Low != nil {
      i.EvalExpr(t, sliceExpr.Low) // Evaluate the inside-square-brackets part of the index expression.	
      lowIdx := t.Pop()	
	  switch lowIdx.(type) {
	  case Int:
	  	low = int(int64(lowIdx.(Int)))
	  case Int32:
	  	low = int(int32(lowIdx.(Int32)))
	  case Uint:
	  	low = int(uint64(lowIdx.(Uint)))                    	                    
	  case Uint32:
	  	low = int(uint32(lowIdx.(Uint32)))
	  default:
		rterr.Stop1(t,sliceExpr,"Index value must be an Integer")
	  }
   }

   if sliceExpr.High == nil {
	  high = int(list.Length()) 	
   } else {
      i.EvalExpr(t, sliceExpr.High) // Evaluate the inside-square-brackets part of the index expression.	
      highIdx := t.Pop()	
	  switch highIdx.(type) {
	  case Int:
	  	high = int(int64(highIdx.(Int)))
	  case Int32:
	  	high = int(int32(highIdx.(Int32)))
	  case Uint:
	  	high = int(uint64(highIdx.(Uint)))                    	                    
	  case Uint32:
	  	high = int(uint32(highIdx.(Uint32)))
	  default:
		rterr.Stop1(t,sliceExpr,"Index value must be an Integer")
	  }
   }
		
   val = list.Slice(t, low, high)

   t.Push(val)        
}




/*
   someExpr.ident
*/
func (i *Interpreter) EvalSelectorExpr(t *Thread, selector *ast.SelectorExpr) {
	// defer UnM(t, TraceM(t, INTERP_TR3, "EvalSelectorExpr", fmt.Sprintf(".%s", selector.Sel.Name)))
	i.EvalExpr(t, selector.X) // Evaluate the left part of the selector expression.		      
	obj := t.Pop()            // the robject whose attribute value is going to be fetched.

	// To speed this up at runtime, could, during parse time, have set an attr field (tbd) of the Sel ident.
	//
	// Except... We don't even know if this name is a one-arg method or an attribute, or which setter
	// (from which type) to use. TODO TODO TODO. In this usage, in lhs of an assignment,
	// it has to be an attribute or it's an error.
	//
	
	attr, found := obj.Type().GetAttribute(selector.Sel.Name)
	if ! found {
       rterr.Stopf1(t,selector,"Attribute or relation '%s' not found in type '%v' or supertypes.", selector.Sel.Name, obj.Type())
    }
	
	val, found := RT.AttrVal(t, obj, attr)
	if !found {
		if attr.Part.ArityLow > 0 {
		    // pos = t.ExecutingMethod.File.Position(selector.Pos())
		    if attr.IsRelation() {		
             if Logging(GC_) {  
		          rterr.Stopf1(t,selector,"Object %v has no value for attribute %s. The declared relation with type %s requires the attribute to have a value.\n" +
		          "obj.IsMarked==%v,GCMarkSense()==%v (should be opposite)", obj, selector.Sel.Name, attr.Part.Type.Name,obj.IsMarked(),GCMarkSense())                
             } else {		       
		          rterr.Stopf1(t,selector,"Object %v has no value for attribute %s. The declared relation with type %s requires the attribute to have a value.", obj, selector.Sel.Name, attr.Part.Type.Name)
	          }
		    } else {
		       
             if Logging(GC_) {  
		          rterr.Stopf1(t,selector,"Object %v has no value for attribute %s. The attribute declaration requires the attribute to have a value.\n" +
		          "obj.IsMarked==%v,GCMarkSense()==%v (should be opposite)", obj, selector.Sel.Name, obj.IsMarked(),GCMarkSense())                
             }	else {	       
		          rterr.Stopf1(t,selector,"Object %v has no value for attribute %s. The attribute declaration requires the attribute to have a value.", obj, selector.Sel.Name)
             }
		    }
		}
		val = attr.Part.Type.Zero()

        // Now obsolete, until we start implementing mandatory attribute types. i.e. if not b ?Car then it expects a car,
        // and should complain if there is a nil.
		// rterr.Stopf("Object %v has no value for attribute %s", obj, selector.Sel.Name)
	}		

	t.Push(val)
}

/*
Push the value of the variable onto the top of the stack. This means the object is referred to at least twice on the 
stack. Once at the current stack top, and once at the variable's position in the stack frame.
*/
func (i *Interpreter) EvalVar(t *Thread, ident *ast.Ident) {
	// defer UnM(t,TraceM(t,INTERP_TR, "EvalVar", ident.Name))
	obj, err := t.GetVar(ident.Offset)
	if err != nil {
		rterr.Stopf1(t,ident,"Attempt to access the value of unassigned variable %s.",ident.Name)
	}
	t.Push(obj)
}

/*
Push the value of the constant with the given name onto the thread's stack.
TODO handle fully qualified constant names properly.
*/
func (i *Interpreter) EvalConst(t *Thread, id *ast.Ident) {
	// defer UnM(t,TraceM(t,INTERP_TR, "EvalConst", id.Name))
	val, found, hidden := i.rt.GetConstant(id.Name, t.ExecutingPackage)
	if ! found {
		rterr.Stopf1(t,id,"Constant %s used before it has been assigned a value.",id.Name)
	} else if hidden {
		rterr.Stopf1(t,id,"Constant %s is private to its package and is not visible in %s.",id.Name, t.ExecutingPackage.Name)
	}
	t.Push(val)	
}

/*
TODO
BasicLit struct {
	ValuePos token.Pos   // literal position
	Kind     token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
	Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
}
*/
func (i *Interpreter) EvalBasicLit(t *Thread, lit *ast.BasicLit) {
	// defer UnM(t,TraceM(t,INTERP_TR, "EvalBasicLit", lit))
	switch lit.Kind {
	case token.INT:
		i, err := strconv.ParseInt(lit.Value, 10, 64)
		if err != nil {
			panic(err)
		}
		t.Push(Int(i))

	case token.FLOAT:
		f, err := strconv.ParseFloat(lit.Value, 64)
		if err != nil {
			panic(err)
		}
		t.Push(Float(f))
	case token.STRING:
		t.Push(String(lit.Value))
	case token.BOOL:
		b, err := strconv.ParseBool(lit.Value)
		if err != nil {
			panic(err)
		}
		t.Push(Bool(b))		
	case token.NIL:
		t.Push(NIL)				
	default:
		panic("I don't know how to interpret this kind of literal yet.")
	}
}


/*
Create a closure and bind its free variables. Leave the closure at the top of the stack.
*/
func (i *Interpreter) EvalClosure(t *Thread, clos *ast.Closure) {
	// defer UnM(t,TraceM(t,INTERP_TR, "EvalClosure", clos.MethodName))
	
	method := t.ExecutingPackage.ClosureMethods[clos.MethodName]
	var bindings []RObject
	for _,offset := range clos.Bindings {
		obj, err := t.GetVar(offset)
		// fmt.Printf("closure.Binding offset: %d holds: %s\n",offset,obj.Debug())		
		if err != nil {
			rterr.Stop1(t,clos,"While creating closure, one of its free variables has not yet been assigned a value in enclosing method.")
		}
		bindings = append(bindings, obj)
	}
	obj := &RClosure{Method:method, Bindings:bindings}
	t.Push(obj)
}


/*
   Evaluate a method call or type constructor call, including evaluation of the arguments. 
   Place the result(s) on the stack.

   1. Evaluate the function expression placing a method or multimethod or newly constructed object
      or newly constructed object + init multimethod
   2. If it was a type construction expression (a type spec) with no init function, stop there. The new object is
      on the thread's stack.
   3. Pop the multi-method reference and/or the newly constructed object off the stack. They'll go back on after the
      base for the new method execution is set. That is, after the creation of a new stack frame for executing the method.
   4. Determine the number of return arguments of the multi-method.
   5. PushBase - Reserve space on the stack for 
         a. the return value(s), 
         b. the base pointer which has the index of the previous base pointer. base pointer value will be -1 if stack empty
         c. the method that is to be executed.
      Note: This does not yet set the stack frame base to the new one, because we need to evaluate method call args in the
      old existing stack frame context.
   6. Place the newly constructed object if any back on the stack.
   7. Evaluate each call argument expression, pushing the resulting argument values on the stack. 
   8. If a multi-method, dispatch to find the correct RMethod. 
       (An RMethod must point to the ast.MethodDeclaration which has the code body.) 
       The correct method is determined by multimethod dispatch on the runtime types of the required positional arguments.
   9. Set t.Base to the new base pointer's stack index, thus switching context to the new stack frame.
   10. Apply the method to the arguments. (If a relish method, the method code uses the stack to find the arguments. If is
       a builtin method implemented in go, uses a slice of RObjects which is a slice containing the args at the top of the stack)
   11. Result value(s) are placed just below the current base pointer, during execution in the method relish code of assignments
       to return arg variables, or alternately upon the execution of a relish "=>" expr1 expr2 (return) statement. 
   12. PopBase - set t.Base to the value stored in the current base pointer. Set t.Pos to the index of the last return val. 


   working data for statement, expr evaluation during method execution            <-- t.Pos
   working data for statement, expr evaluation during method execution
   paramOrLocalVar2
   paramOrLocalVar1 
   paramOrLocalVar0
   [reserve for code position in method body bytecode? RESERVED BUT NOT USED]
   method
   base              ---       . . . . . . . . . . . . . . . . . . . . . . . .  . <-- t.Base
   retval2              |
   retval1              |
   working data         | 
   working data         |
   paramOrLocalVar2     |    
   paramOrLocalVar1     |
   paramOrLocalVar0     |
   [reserve for code position in method body bytecode? RESERVED BUT NOT USED]
   method               |
   base      --       <-
   retval1     |
               v


   Note. If Thread t2 is supplied non-nil, it refers to a newly spawned thread of which t is the parent.
   In this case, the method call arguments (and the expression that yields the multimethod) will be evaluated 
   in the parent thread t and its stack, then the stack frame of the call will be copied to t2's stack, 
   and the method application to the arguments will be performed in the new thread (go-routine actually) using t2's stack.


TODO Put nArgs in variants of this method !!!!!!!!!!!!  !!!!!!!!!!!! !!!!!!!!!!! !!!!!!!!!!!!!

*/
func (i *Interpreter) EvalMethodCall(t *Thread, t2 *Thread, call *ast.MethodCall) (nReturnArgs int) {
	// defer UnM(t,TraceM(t,INTERP_TR, "EvalMethodCall"))

    // Evaluate the function expression - function name, lambda (TBD), or type name
    // and put the method or multimethod on the stack,
    // If it is a type constructor, puts the type constructor function and the new object on the
    // stack. 

    // PASS THE WHOLE call , return the remaining args, after slicing the arg list to remove the
    // RClosure object if the call is "func anRClosure remainingArgs..."
    // args returned is remaining args  
	isTypeConstructor, hasInitFunction, isClosureApplication, args := i.EvalFunExpr(t, call) // token.FUN (regular method) or token.TYPE (constructor) or token.CLOSURE (closure application)
	
	

	// This still doesn't handle implicit constructors that have keyword args corresponding to attributes of the
	// type ad supertypes. !!!! TODO
	//
	
	var newObject RObject
	var meth RObject
	var closure *RClosure
	
	if isTypeConstructor {
		if ! hasInitFunction {
			nReturnArgs = 1
			return 
		}
	    meth = t.Pop() // put it back after the base pointer.	
	    newObject = t.Pop()	
	} else {
	    meth = t.Pop() // put it back after the base pointer.			
	}
	
    // TODO Question - Why are we pushing the method and popping it again like this,
    // Why not just have EvalFunExpr return the method or multimethod as an RObject? breaks Eval.. method conventions, but more efficient.


	// LoglnM(t,INTERP_TR, meth)
	switch meth.(type) {
	case *RMultiMethod:
		nReturnArgs = meth.(*RMultiMethod).NumReturnArgs
	case *RMethod:
		nReturnArgs = meth.(*RMethod).NumReturnArgs
	case *RClosure:
		closure = meth.(*RClosure)
		nReturnArgs = closure.Method.NumReturnArgs		
	default:
		panic("Expecting a Method or MultiMethod.")
	}

    // Now we know how many return argument slots to reserve on stack below the base pointer...

	newBase := t.PushBase(nReturnArgs) // begin but don't complete, storing outer routine context. 

	// NOTE NOTE !!
	// When leaving this stack context, we may in future want to also push just above the method the offset into the method's code
	// where we left off. PushBase is leaving a space on the stack for that, and makes initial variableOffset 3 instead of 2

    constructorArg := 0
    if isTypeConstructor {
	   t.Push(newObject)
	   constructorArg = 1
    }  

    p0 := t.Pos
	for _, expr := range args {
		i.EvalExpr(t, expr)
	}
	p1 := t.Pos
	nArgs := p1 - p0
	
    evaluatedArgs := t.TopN(constructorArg + nArgs) 

	// 
	// TODO We are going to have to handle varargs differently here. Basically, eval and push only the non-variable
	// args here, then, below, reserve space for one list (of varargs) then reserve space for the local vars, 
	// and finally, just before apply1, eval and push the extra args onto the stack, then remove them into the list.
	// Or do we just use the stack itself as the list of extra args?s???????? TODO !!!!!
	//

	t.SetBase(newBase) // Now we're in the context of the newly called function.

	// Put this back here!!!
	//defer t.PopBase() // We need to worry about panics leaving the stack state inconsistent. TODO

	var method *RMethod
	var typeTuple *RTypeTuple

	switch meth.(type) {
	case *RMultiMethod:
		mm := meth.(*RMultiMethod)
			
//		// Shouldn't this possibility be caught by the compiler?
//		//	
//		for j,arg := range evaluatedArgs {
//			if arg == nil {
//			    rterr.Stopf1(t, call, "Argument %d in method %s call has no value - used before assigned.",j+1,mm.Name)   
//			}
//		}
			
        // fmt.Println(mm.Debug())

		// Comment out the defer statement to get panics with go stack traces for debugging runtime errors	
		// defer methodCallErrHandle(t,call)	
		method, typeTuple = i.dispatcher.GetMethod(mm, evaluatedArgs) // nArgs is WRONG! Use Type.Param except vararg
		if method == nil {
			if isTypeConstructor && nArgs == 0 {  // There is no other-argless init<TypeName> method.
               // Unsetup the aborted constructor method call.
	           // TODO This is really inefficient!! Do something else for simple no-method constructions.
	           t.PopBase() // We need to worry about relish-level panics leaving the stack state inconsistent. TODO   
	           t.PopN(nReturnArgs) // Get rid of space reserved for return args of aborted constructor method call.     
               t.Push(newObject)  // Again, this time under the base of the alleged but no found constructor.
	           nReturnArgs = 1     
               return                       // Just return the uninitialized object.
            }

            // This is actually a no-compatible method found dynamic-dispatch error (i.e. a runtime-detected type compatibility error).
			//
////			t.ExecutingPackage.ListMethod("dbg")
////			fmt.Println("--------------")
			rterr.Stopf1(t, call, "No method '%s' visible from within %s is compatible with %s", mm.Name, t.ExecutingPackage.Name,typeTuple)
		}
		LoglnM(t,INTERP_, "Multi-method dispatched to ", method)
		// fmt.Println("Multi-method dispatched to ", method)		
	case *RMethod:
		method = meth.(*RMethod)
	case *RClosure:
		method = closure.Method		
	default:
		panic("Expecting a Method or MultiMethod.")
	}

    errorContextMethod := t.ExecutingMethod 

    if t2 == nil { 

		// put currently executing method on stack in reserved parking place
		t.Stack[newBase+1] = method

		t.ExecutingMethod = method       // Shortcut for dispatch efficiency
		t.ExecutingPackage = method.Pkg  // Shortcut for dispatch efficiency

		t.Reserve(method.NumLocalVars) 

        if isClosureApplication {
	        for _,obj := range closure.Bindings {
		       t.Push(obj)
	        }
        }

		// t.Dump()
		// fmt.Println("nArgs",nArgs)

		err := i.apply1(t, method, evaluatedArgs) // Puts results on stack BELOW the current stack frame.	
	    if err != nil {
    	   rterr.Stop1(errorContextMethod, call, err.Error())
        }			

    } else { // This is a go statement execution
    	t2.copyStackFrameFrom(t, nReturnArgs)  // copies the frame for the execution of this method, 
    	                                               // including space reserved for return val(s)
        go i.GoApply(t2, method, errorContextMethod, call)
    }

	t.PopBase() // We need to worry about relish-level panics leaving the stack state inconsistent. TODO

	return
}

/*
TODO Should benchmark with and without the deferred call to this.
*/
func methodCallErrHandle(t *Thread, call *ast.MethodCall) {
      r := recover()	
      if r != nil {
          if t.CodeFile() == nil {
             if t.ExecutingMethod == nil {
             	fmt.Println()
                rterr.Stopf("%v\n",r) 
             } else {
             	fmt.Println()             	
                rterr.Stopf("%v (%s):\n\n%v\n",t.ExecutingMethod,t.ExecutingPackage.Name,r)                 
             }                                 
          } else if call == nil {
             if t.ExecutingMethod == nil {
             	fmt.Println()             	
                rterr.Stopf("%v\n",r) 
             } else {
             	fmt.Println()             	
                rterr.Stopf("%v:\n\n%v\n",t.ExecutingMethod.Name(),r)                 
             }                                 
          } else {
             rterr.Stopf1(t,call,"%v",r)
          }
      }	
}	


/*
Assumes that the base pointer for executing the new method has been set in thread t,
but that the method has not been pushed onto t's stack nor has the space for the method's local vars been reserved on t's stack.

Applies the method using thread t to pre-evaluated arguments that are in the current frame of t's stack.

TODO This is NOT updated to handle closure applications yet !!!

*/
func (i *Interpreter) GoApply(t *Thread, method *RMethod, file rterr.CodeFileLocated, pos rterr.Positioned) {
    
	t.Stack[t.Base+1] = method

	t.ExecutingMethod = method       // Shortcut for dispatch efficiency
	t.ExecutingPackage = method.Pkg  // Shortcut for dispatch efficiency

    nArgs := t.Pos - t.Base - 2

    // t.Dump()
    // fmt.Println("len(t.Stack)",len(t.Stack))
    // fmt.Println("nArgs",nArgs)
    // fmt.Println("t.Pos", t.Pos)

    evaluatedArgs := t.TopN(nArgs) 

	t.Reserve(method.NumLocalVars)   
 
	err := i.apply1(t, method, evaluatedArgs) // Puts results on stack BELOW the current stack frame.	
	if err != nil {
       rterr.Stop1(file, pos, err.Error())
    }		

	t.PopFinalBase(method.NumReturnArgs)  // Pop off all of the executing method's context (for GC non-reference)
	                                      // But do not set the thread's ExecutingMethod, because there is none.
	// t.PopN(method.NumReturnArgs) // Also pop any return args off t's stack (for GC non-reference)	

	i.DeregisterThread(t)	
}



/*
If the type name is unqualified (has no package path), prefixes it with the specified package path.
A little bit inefficient.
*/
func (i *Interpreter) QualifyTypeName(packagePath string, typeName string) string {
   if strings.LastIndex(typeName,"/") == -1 && ! BuiltinTypeName[typeName] {
       return packagePath + typeName  	
   }	
   return typeName
}

/*
Used to create List_of etc types on the fly if needed.
Looks up the type by name, with the appropriate "List_of" etc prefixed on if appropriate.
If not found in the runtime types list, creates the type.
Returns the type corresponding to the TypeSpec.
If the base simple type is not found, returns nil and an appropriate error message.
*/
func (i *Interpreter) EnsureTypeOriginal(packagePath string, typeSpec *ast.TypeSpec) (typ *RType, err error) {
	
   baseTypeName := i.QualifyTypeName(packagePath, typeSpec.Name.Name) 
   baseType, baseTypeFound := i.rt.Types[baseTypeName]

   if ! baseTypeFound {
      err = fmt.Errorf("Type %s not found.",baseTypeName)
      return		
   }
  	
   if typeSpec.CollectionSpec == nil {
	  typ = baseType
   } else {	
	    switch typeSpec.CollectionSpec.Kind {
	    case token.LIST:		
			typ, err = i.rt.GetListType(baseType)
			if err != nil {
			   return			
			}
	    case token.SET:
			typ, err = i.rt.GetSetType(baseType)
			if err != nil {
			   return			
			}			
	    case token.MAP:
		    var valTyp *RType
			valTyp, err = i.EnsureType(packagePath, typeSpec.Params[0])
			if err != nil {
			   return			
			}		
            typ, err = i.rt.GetMapType(baseType, valTyp) 		
		
	        // panic("I don't handle Map type specifications in this context yet.")			
	    }
	}	
    return
}


func (i *Interpreter) EnsureType(packagePath string, typeSpec *ast.TypeSpec) (typ *RType, err error) {
	

   if typeSpec.CollectionSpec == nil {	
	   baseTypeName := i.QualifyTypeName(packagePath, typeSpec.Name.Name) 
	   baseType, baseTypeFound := i.rt.Types[baseTypeName]

	   if ! baseTypeFound {
	      err = fmt.Errorf("Type %s not found.",baseTypeName)
	      return		
	   }
	   typ = baseType
    } else {
		var elementTyp *RType
        if typeSpec.Name != nil {  // A relation-end with an implied set-collection type
	       baseTypeName := i.QualifyTypeName(packagePath, typeSpec.Name.Name) 
	       baseType, baseTypeFound := i.rt.Types[baseTypeName]

	       if ! baseTypeFound {
	          err = fmt.Errorf("Type %s not found.",baseTypeName)
	          return		
	       }
	       elementTyp = baseType
        } else {
			elementTyp, err = i.EnsureType(packagePath, typeSpec.Params[0])
			if err != nil {
			   return			
			}	    
		}	
	    switch typeSpec.CollectionSpec.Kind {
	    case token.LIST:		
			typ, err = i.rt.GetListType(elementTyp)
			if err != nil {
			   return			
			}
	    case token.SET:
			typ, err = i.rt.GetSetType(elementTyp)
			if err != nil {
			   return			
			}			
	    case token.MAP:
		    var valTyp *RType
			valTyp, err = i.EnsureType(packagePath, typeSpec.Params[1])
			if err != nil {
			   return			
			}		
            typ, err = i.rt.GetMapType(elementTyp, valTyp) 		
		
	        // panic("I don't handle Map type specifications in this context yet.")			
	    }   	
    }
    return
}



func (i *Interpreter) CreateList(t *Thread, elementType *ast.TypeSpec) (List, error) {
   // Find the element type
   typ, err := i.EnsureType(t.ExecutingPackage.Path, elementType)
   if err != nil {
      rterr.Stopf1(t, elementType, "List Element Type Error: %s", err.Error())	
   }
//   fmt.Printf("CreateList: typ.Name=%s from elementType %s\n",typ.Name,elementType.Name.Name)

   // TODO sorting-lists
   return i.rt.Newrlist(typ, 0, -1, nil, nil, nil)
}

/*
Not handling sorting sets yet.
*/
func (i *Interpreter) CreateSet(t *Thread, elementType *ast.TypeSpec) (RCollection, error) {
   // Find the element type
   typ, err := i.EnsureType(t.ExecutingPackage.Path, elementType)
   if err != nil {
      rterr.Stopf1(t, elementType, "Set Element Type Error: %s", err.Error())	
   }

   // TODO sorting-sets
   return i.rt.Newrset(typ, 0, -1, nil, nil)
}

/*
Note it is not really the key type, since the type of the keyType typeSpec has a collectionTypeSpec
*/
func (i *Interpreter) CreateMap(t *Thread, keyType *ast.TypeSpec, valType *ast.TypeSpec) (Map, error) {

   // Find the key and value types
   keyTyp, err := i.EnsureType(t.ExecutingPackage.Path, keyType)
   if err != nil {
      rterr.Stopf1(t, keyType, "Map Key Type Error: %s", err.Error())	
   }   

   valTyp, err := i.EnsureType(t.ExecutingPackage.Path, valType)
   if err != nil {
      rterr.Stopf1(t, valType, "Map Value Type Error: %s", err.Error())	
   } 

   // TODO sorting-maps
   return i.rt.Newmap(keyTyp, valTyp, 0, -1, nil, nil, nil)
}



/*
Creates a list, and populates it from explicit element expressions or by executing a SQL query in the database.
Leaves the constructed and possibly populated list as the top of the stack.

// EGH A ListConstruction node represents a list constructor invocation, which may be a list literal, a new empty list of a type, or
// a list with a db sql query where clause specified as the source of list members.

ListConstruction struct {
    Type *TypeSpec     // Includes the CollectionTypeSpec which must be a spec of a List.
	Elements  []Expr    // explicitly listed elements; or nil        
	Query     Expr     // must be an expression evaluating to a String containing a SQL WHERE clause (without the "WHERE"), or nil
	                   // Note eventually it should be more like OQL where you can say e.g. engine.horsePower > 120 when fetching []Car
}
*/
func (i *Interpreter) EvalListConstruction(t *Thread, listConstruction *ast.ListConstruction) {
	// defer UnM(t,TraceM(t,INTERP_TR, "EvalListConstruction"))

    list, err := i.CreateList(t, listConstruction.Type)
    if err != nil {
	   panic(err)
    }
    t.Push(list)

   nElem := len(listConstruction.Elements)
   if nElem > 0 {
		for _, expr := range listConstruction.Elements {
			i.EvalExpr(t, expr)
		}	
		
       err = i.rt.ExtendCollection(list, t.TopN(nElem), true, t.EvalContext) 	
       if err != nil {
	      rterr.Stop1(t, listConstruction, err)
       }		

       t.PopN(nElem)
	
   } else if listConstruction.Query != nil { // Database select query to fill the list
	  // TODO
	
	
	  // TODO Why can't we do this query syntax transformation at generation time, so we only do it
	  // once per occurrence in the program text, as long as it is a literal string.
	  // !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	  //
			
	  i.EvalExpr(t, listConstruction.Query)
	
	  qExpr := t.Pop()	
	
	  query := ""
	  queryArgs := []RObject{}
	  isList := false
	  qS,isString := qExpr.(String)
	  if isString {
	     query = string(qS)		
	  } else {
		 if qExpr.IsCollection() {
			coll := qExpr.(RCollection)
			if coll.IsList() {
				isList = true
				list := coll.(List)
				v := list.Vector()
				objList := []RObject(*v)
				if len(objList) < 1 {
	               isList = false
	            } else {				
				   qS,isString = objList[0].(String)
			   	   if ! isString {
				  	  isList = false
			  	   } else {
	                  query = string(qS)					      
				      queryArgs = objList[1:]
				   }
				}	
			} 
		}
		if ! (isString || isList) {   
	      rterr.Stop1(t, listConstruction, "Query expression used in list construction must be a String or a list starting with a String.")	
	    }
	  }
	  radius := 1
	  if strings.HasPrefix(query, "lazy: ") {
		 query = query[6:]
	     radius = 0	 
	  }

	
	  objs := []RObject{} // TODO Use the existing List's RVector somehow
	
	  // objs := list.Vector().(*[]RObject)
	
//	  rv := list.Vector()
//      objs := (*[]RObject)(rv)


	

      mayContainProxies, err := t.DBT().FetchN(list.ElementType(), query, queryArgs, nil, radius, &objs)		
      if err != nil {
	      rterr.Stop1(t, listConstruction, err)
      }	

      list.ReplaceContents(objs)
      list.SetMayContainProxies(mayContainProxies)

      // fmt.Println(len(*objs))

   } else if listConstruction.Generator != nil { // Generator expression to yield elements  
      i.iterateGenerator(t, listConstruction.Generator, 1)
      list.ReplaceContents(t.Objs)
      t.Objs = nil
   }
   return
}

/*
   Given a for-range statement which is of the special type that constitutes a generator expression,
   causes the iteration of the range statement to produce all of the generated results.
   nResultsPerIteration is the expected number of result values expected to be yielded per iteration
   of the for loop. A runtime error will occur if the for-range statement does not yield this number
   of results per iteration.
   Results are collected in a slice of RObjects. If there is more than one result per iteration,
   the result values are interleaved in the returned slice as follows iteration1result1 i1r2 i2r1 i2r2 i3r1 i3r2 
*/
func (i *Interpreter) iterateGenerator(t *Thread, rangeStmt *ast.RangeStatement, nResultsPerIteration int) {
	t.YieldCardinality = nResultsPerIteration
	i.ExecForRangeStatement(t, rangeStmt)
}


/*
Creates a set, and populates it from explicit element expressions or by executing a SQL query in the database.
Leaves the constructed and possibly populated set as the top of the stack.

// EGH A SetConstruction node represents a set constructor invocation, which may be a set literal, a new empty set of a type, or
// a set with a db sql query where clause specified as the source of set members.

SetConstruction struct {
    Type *TypeSpec     // Includes the CollectionTypeSpec which must be a spec of a Set.
	Elements  []Expr    // explicitly listed elements; or nil        
	Query     Expr     // must be an expression evaluating to a String containing a SQL WHERE clause (without the "WHERE"), or nil
	                   // Note eventually it should be more like OQL where you can say e.g. engine.horsePower > 120 when fetching []Car
}
*/
func (i *Interpreter) EvalSetConstruction(t *Thread, setConstruction *ast.SetConstruction) {
	// defer UnM(t,TraceM(t,INTERP_TR, "EvalSetConstruction"))

    set, err := i.CreateSet(t, setConstruction.Type)
    if err != nil {
	   panic(err)
    }
    t.Push(set)

   nElem := len(setConstruction.Elements)
   if nElem > 0 {
		for _, expr := range setConstruction.Elements {
			i.EvalExpr(t, expr)
		}	
		
       err = i.rt.ExtendCollection(set.(AddableCollection), t.TopN(nElem), true, t.EvalContext) 	
       if err != nil {
	      rterr.Stop1(t, setConstruction, err)
       }		

       t.PopN(nElem)
	
   } else if setConstruction.Query != nil { // Database select query to fill the list
	  // TODO
	
	
	  // TODO Why can't we do this query syntax transformation at generation time, so we only do it
	  // once per occurrence in the program text, as long as it is a literal string.
	  // !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	  //
			
	  i.EvalExpr(t, setConstruction.Query)
	
	  qExpr := t.Pop()	
	
	  queryArgs := []RObject{}	
	
	  qS,isString := qExpr.(String)
	  if ! isString {
	     rterr.Stop1(t, setConstruction, "Query expression used in set construction must evaluate to a String.")	
	  }
	  query := string(qS)
	  radius := 1
	  if strings.HasPrefix(query, "lazy: ") {
		 query = query[6:]
	     radius = 0	 
	  }

	
	  objs := []RObject{} // TODO Use the existing List's RVector somehow
	
	  // objs := list.Vector().(*[]RObject)
	
//	  rv := list.Vector()
//      objs := (*[]RObject)(rv)


	

      mayContainProxies, err := t.DBT().FetchN(set.ElementType(), query, queryArgs, nil, radius, &objs)		
      if err != nil {
	      rterr.Stop1(t, setConstruction, err)
      }	

      aSet := set.(AddableMixin)
      for _,obj := range objs {
		 aSet.Add(obj, t.EvalContext) 
      }

      set.SetMayContainProxies(mayContainProxies)


      // fmt.Println(len(*objs))

   } else if setConstruction.Generator != nil { // Generator expression to yield elements  
      i.iterateGenerator(t, setConstruction.Generator, 1)
      aSet := set.(AddableMixin)
      for _,obj := range t.Objs {
		 aSet.Add(obj, t.EvalContext) 
      }
      t.Objs = nil
   }
	return
}



/*
Creates a map, and populates it from explicit entry expressions. 
Leaves the constructed and possibly populated map as the top of the stack.

MapConstruction struct {
    Type *TypeSpec     // Includes the CollectionTypeSpec which must be a spec of a Map.
    ValType *TypeSpec     // Type of the values
    Keys []Expr         // explicitly listed keys; or nil
	Elements  []Expr    // explicitly listed elements; or nil        
}

*/
func (i *Interpreter) EvalMapConstruction(t *Thread, mapConstruction *ast.MapConstruction) {
	// defer UnM(t,TraceM(t,INTERP_TR, "EvalMapConstruction"))

    theMap, err := i.CreateMap(t, mapConstruction.Type, mapConstruction.ValType)
    if err != nil {
	   panic(err)
    }
    t.Push(theMap)

   nElem := len(mapConstruction.Elements)
   if nElem > 0 {
		for j, valExpr := range mapConstruction.Elements {
			i.EvalExpr(t, mapConstruction.Keys[j])
			i.EvalExpr(t, valExpr)
		}	
		
       err = i.rt.ExtendMapTypeChecked(theMap, t.TopN(nElem *2), t.EvalContext) 	
       if err != nil {
	      rterr.Stop1(t, mapConstruction, err)
       }		

       t.PopN(nElem * 2)
	} else if mapConstruction.Generator != nil { // Generator expression to yield elements  
      i.iterateGenerator(t, mapConstruction.Generator, 2)
      n := len(t.Objs)
      for k := 0; k < n; k+=2 {
         key := t.Objs[k]
         val := t.Objs[k+1] 
		   //theMap.Put(key, val, t.EvalContext) 
         err = i.rt.PutInMapTypeChecked(theMap, key, val, t.EvalContext) 	 
         if err != nil {
   	      rterr.Stop1(t, mapConstruction, err)
         }		 
      }
      t.Objs = nil
   }
	
 
	return
}






///// From here to next ///// is special purpose code
/*
Implements interface MethodEvaluationContext defined in package relish/runtime/data
*/
type methodEvaluationContext struct {
	interpreter *Interpreter
	thread      *Thread
}

/*
Evaluates a single-valued function after dispatching to find the correct function implementation.
*/
func (context *methodEvaluationContext) EvalMultiMethodCall(mm *RMultiMethod, args []RObject) RObject {
	return context.interpreter.evalMultiMethodCall1ReturnVal(context.thread, mm, args)
}

func (context *methodEvaluationContext) InterpThread() InterpreterThread {
	return context.thread
}

func (context *methodEvaluationContext) GC() { 
   context.interpreter.GC() 
}


/*
Special-purpose variant of method evaluation, used as part of implementation of MethodEvaluationContext interface.
Evaluates a single-valued multi-method on some pre-evaluated arguments, and returns the result. 

Used for example to evaluate collection-sort comparison functions on collection members.
*/
func (i *Interpreter) evalMultiMethodCall1ReturnVal(t *Thread, mm *RMultiMethod, args []RObject) RObject {
	// defer UnM(t,TraceM(t,INTERP_TR, "evalMultiMethodCall"))
	// LoglnM(t,INTERP_TR, mm)

	//   t.Push(multiMethod)

	//   meth := t.Pop() // put it back after the base pointer.

	nReturnArgs := mm.NumReturnArgs // Assuming this is 1   !!!

	newBase := t.PushBase(nReturnArgs) // begin but don't complete, storing outer routine context. 

	// NOTE NOTE !!
	// At some point, when leaving this context, we may want to also push just above this the offset into the method's code
	// where we left off. We might wish to leave a space on the stack for that, and make initial variableOffset 3 instead of 2

	for _, arg := range args {
		t.Push(arg)
	}
	// 
	// TODO We are going to have to handle varargs differently here. Basically, eval and push only the non-variable
	// args here, then, below, reserve space for one list (of varargs) then reserve space for the local vars, 
	// and finally, just before apply1, eval and push the extra args onto the stack, then remove them into the list.
	// Or do we jusy use the list on the stack???????? TODO !!!!!
	//

	t.SetBase(newBase) // Now we're in the context of the newly called function.

	// Put this back here!!!
	//defer t.PopBase() // We need to worry about panics leaving the stack state inconsistent. TODO

	var method *RMethod
	var typeTuple *RTypeTuple

	method, typeTuple = i.dispatcher.GetMethod(mm, args) // len call.Args is WRONG! Use Type.Param except vararg
	if method == nil {
		if t.ExecutingMethod == nil {
		   rterr.Stopf("Function call in template: No method '%s' is compatible with %s", mm.Name, typeTuple)			
		} else {
		   rterr.Stopf1(t, t.ExecutingMethod.File, "No method '%s' is compatible with %s", mm.Name, typeTuple)
	   }
	}
	LoglnM(t,INTERP_, "Multi-method dispatched to ", method)

	// put currently executing method on stack in reserved parking place
	t.Stack[newBase+1] = method

	t.ExecutingMethod = method       // Shortcut for dispatch efficiency
	t.ExecutingPackage = method.Pkg  // Shortcut for dispatch efficiency	

	t.Reserve(method.NumLocalVars)

	err := i.apply1(t, method, args) // Puts results on stack BELOW the current stack frame.	
	if err != nil {
		rterr.Stopf1(t, t.ExecutingMethod.File, "Error calling sorting method %s: %s", mm.Name, err.Error())		
	}

	t.PopBase() // We need to worry about panics leaving the stack state inconsistent. TODO

	return t.Pop() // Assuming single valued function!
}

///// From here up to previous ///// is special purpose code

/*
Evaluate the expression which must end up as either a RMultiMethod or an RMethod or a Type. 
If a RMultiMethod or RMethod, put that on the stack.
If a Type, then if TODO!!!!!!!!!!!!

If this is a type constructor call, then return isTypeConstructor = true
and place the newly allocated but uninitialized object on the stack.
If additionally an init<TypeName> function was found, set the hasInitFunc = true
and place the init<TypeName> multimethod on the stack.
If it is not a constructor call, place the found multimethod on the stack. 
*/
func (i *Interpreter) EvalFunExpr(t *Thread, call *ast.MethodCall) (isTypeConstructor bool, hasInitFunc bool, isClosureApplication bool, args []ast.Expr) {
	// defer UnM(t,TraceM(t,INTERP_TR2, "EvalFunExpr"))
	var methodKind token.Token
	fun := call.Fun
	switch fun.(type) {
	case *ast.Ident:
		id := fun.(*ast.Ident)
		methodKind = id.Kind
		switch methodKind {
		case token.FUNC:
			multiMethod, found := t.ExecutingPackage.MultiMethods[id.Name]
			if !found {
			    // t.ExecutingPackage.ListMethods()				
				rterr.Stopf1(t, fun, "'%s' is not a method visible from within package %s, nor an assigned local variable, nor a method-parameter name.", id.Name, t.ExecutingPackage.Name)

			}
			t.Push(multiMethod)
			args = call.Args
			
		case token.TYPE:
			var obj RObject
			var err error
			obj, err = i.rt.NewObject(id.Name)
			if err != nil {
				rterr.Stop1(t, fun, err)
			}
			if obj.Type().IsPrivate && obj.Type().Package != t.ExecutingPackage {
				rterr.Stopf1(t, fun, "'%s' is a package-private Type not visible from within package %s", id.Name, t.ExecutingPackage.Name)				
			}
			t.Push(obj)
			
			isTypeConstructor = true			
			
			slashPos := strings.LastIndex(id.Name,"/")
			var typeName string
			var initMethodName string
			if slashPos >= 0 {
				typeName = id.Name[slashPos + 1:]
				initMethodName = id.Name[:slashPos+1] + "init" + typeName
			} else {
				typeName = id.Name
				initMethodName = "init" + typeName
			}
			
			multiMethod, found := t.ExecutingPackage.MultiMethods[initMethodName]
			if found {
			   t.Push(multiMethod)	
			   hasInitFunc = true	
			} 
			args = call.Args			
				
		case token.CLOSURE:	
		    isClosureApplication = true
		    
		    i.EvalExpr(t, call.Args[0])	 // Leave the RClosure on the stack	
		
			args = call.Args[1:]
		default:
			panic("Wrong type of ident for function name.")
		}
	default:
		panic("I don't handle lambda expressions yet!")
	}
	return 
}



/*
Apply the method implementation to the arguments.
Puts the results on the stack, in reserved slots BELOW the current method-call's stack frame.
Does not pop the m method's stack frame from the stack i.e. does not pop (move down) the base pointer.

TODO TODO We cannot have the return values on the stack in reverse order like this.
It will not work for using the values as args to the next outer method.
*/
func (i *Interpreter) apply1(t *Thread, m *RMethod, args []RObject) (err error) {
//	defer UnM(t, TraceM(t,INTERP_TR, "apply1", m, "to", args))	
//	defer UnM(t, TraceM(t,INTERP_TR, m.String()))		
//	if strings.Contains(m.String(),"spew") {
//		fmt.Println(args)
//	} 

	if Logging(STACK_) {
		t.Dump()
	}
	if m.PrimitiveCode == nil {
		if m.ReturnArgsNamed {
			n := m.NumReturnArgs
            for j,typ := range m.ReturnSignature.Types {
            	t.Stack[t.Base+j-n] = typ.Zero()
            }
		}

        // Experimental
        for j, arg := range args {
        	if arg == nil {
               err = fmt.Errorf("Argument %d in method call has no value - used before assigned.",j+1)
               return
           } else if arg == NIL {
               if ! m.NilArgAllowed[j] {
                  err = fmt.Errorf("nil value not permitted for argument %d in method call.",j+1)      
                  return         	
               } 
           } 
        }
        if m.Code.Body == nil {
           err = fmt.Errorf("Method body statements not defined for method %v.",m)  
           return      	
        }
		i.ExecBlock(t, m.Code.Body)

		// Now maybe type-check the return values !!!!!!!! This is really expensive !!!!

        // At least make sure, if we were supposed to return something, we did.
        if m.NumReturnArgs > 0 && ! m.ReturnArgsNamed && t.Stack[t.Base-m.NumReturnArgs] == nil {
           // We've missed encountering a return statement.
           if m.NumReturnArgs == 1 {	
              err = fmt.Errorf("Method %v is specified to return one value but returned none because no => statement was executed.",m)  
           } else {
              err = fmt.Errorf("Method %v is specified to return %d values but returned none because no => statement was executed.",m,m.NumReturnArgs)  
           }
           return        	
        }

	} else {

		objs := m.PrimitiveCode(t, args)
		
		n := len(objs)		
		for j, obj := range objs {
			t.Stack[t.Base+j-n] = obj   
		}	
	}
	return
}

/*
Evaluate a block statement. Any return values will have been placed into the appropriate return-value stack
slots. Returns whether the next outermost loop should be broken or continued, or whether the containing methods should be
returned from.
*/
func (i *Interpreter) ExecBlock(t *Thread, b *ast.BlockStatement) (breakLoop, continueLoop, returnFrom bool) {
	// defer UnM(t,TraceM(t,INTERP_TR2, "ExecBlock"))
	for _, statement := range b.List {
		breakLoop, continueLoop, returnFrom = i.ExecStatement(t, statement)
		if breakLoop || continueLoop || returnFrom {
			break
		}
	}
	return
}

func (i *Interpreter) ExecStatement(t *Thread, stmt ast.Stmt) (breakLoop, continueLoop, returnFrom bool) {
	// defer UnM(t,TraceM(t,INTERP_TR3, "ExecStatement"))
	switch stmt.(type) {
	case *ast.IfStatement:
		breakLoop, continueLoop, returnFrom = i.ExecIfStatement(t, stmt.(*ast.IfStatement))
	case *ast.WhileStatement:
		breakLoop, continueLoop, returnFrom = i.ExecWhileStatement(t, stmt.(*ast.WhileStatement))
	case *ast.RangeStatement:
		breakLoop, continueLoop, returnFrom = i.ExecForRangeStatement(t, stmt.(*ast.RangeStatement))
	case *ast.ForStatement:
		breakLoop, continueLoop, returnFrom = i.ExecForStatement(t, stmt.(*ast.ForStatement))		
	case *ast.BreakStatement:
		breakLoop = true
	case *ast.ContinueStatement:
		continueLoop = true
	case *ast.MethodCall:
		i.ExecMethodCall(t, nil, stmt.(*ast.MethodCall))
	case *ast.AssignmentStatement:
		i.ExecAssignmentStatement(t, stmt.(*ast.AssignmentStatement))
	case *ast.ReturnStatement:		
		returnFrom = i.ExecReturnStatement(t, stmt.(*ast.ReturnStatement)) 
	case *ast.BlockStatement:
		breakLoop, continueLoop, returnFrom = i.ExecBlock(t, stmt.(*ast.BlockStatement))
	case *ast.GoStatement:
		i.ExecGoStatement(t, stmt.(*ast.GoStatement))		
//	case *ast.DeferStatement:
//		breakLoop, continueLoop, returnFrom = i.ExecDeferStatement(t, stmt.(*ast.DeferStatement))		
		
	default:
		panic("I don't know how to handle this kind of statement.")
	}
	return
}

/*
 */
func (i *Interpreter) ExecIfStatement(t *Thread, stmt *ast.IfStatement) (breakLoop, continueLoop, returnFrom bool) {
	// defer UnM(t, TraceM(t,INTERP_TR2, "ExecIfStatement"))
	i.EvalExpr(t, stmt.Cond)
	if t.Pop().IsZero() {
		if stmt.Else != nil {
			breakLoop, continueLoop, returnFrom = i.ExecStatement(t, stmt.Else)
		}
	} else {
		breakLoop, continueLoop, returnFrom = i.ExecBlock(t, stmt.Body)
	}
	return
}

/*
   Execute the MethodCall in a new go thread, with a new Relish stack. (A "Thread" is an object representing a relish stack)
*/
func (i *Interpreter) ExecGoStatement(parent *Thread, stmt *ast.GoStatement) {
	t := i.NewThread(parent)
	i.ExecMethodCall(parent, t, stmt.Call)	
}

func (i *Interpreter) ExecWhileStatement(t *Thread, stmt *ast.WhileStatement) (breakLoop, continueLoop, returnFrom bool) {
	// defer UnM(t,TraceM(t,INTERP_TR2, "ExecWhileStatement"))
	i.EvalExpr(t, stmt.Cond)
	if t.Pop().IsZero() {
		if stmt.Else != nil {
			breakLoop, continueLoop, returnFrom = i.ExecStatement(t, stmt.Else)
		}
	} else {
		breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)
		for (!breakLoop) && (!returnFrom) {
			i.EvalExpr(t, stmt.Cond)
			if t.Pop().IsZero() {
				break
			}
			breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)
		}
		breakLoop = false
		continueLoop = false
	}
	return
}



func (i *Interpreter) ExecForStatement(t *Thread, stmt *ast.ForStatement) (breakLoop, continueLoop, returnFrom bool) {
	// defer UnM(t,TraceM(t,INTERP_TR2, "ExecForStatement"))

    i.ExecAssignmentStatement(t, stmt.Init)

	for (!breakLoop) && (!returnFrom) {
		i.EvalExpr(t, stmt.Cond)
		if t.Pop().IsZero() {
			break
		}
		breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)
		if ! breakLoop && ! returnFrom {
           i.ExecAssignmentStatement(t, stmt.Post)
        }		
	}
	breakLoop = false
	continueLoop = false
	return
}



/*
If there is only a single collection to be ranged over, this if possible uses an optimized iteration that does not
use the Iter() method of the collection.
If there is more than one collection, though, or the collection is a map, we have to use the Iter() method of the
collection, which uses a Go channel in a closure to do the iterating, and could be a tad slow. 
*/
func (i *Interpreter) ExecForRangeStatement(t *Thread, stmt *ast.RangeStatement) (breakLoop, continueLoop, returnFrom bool) {
	// defer UnM(t,TraceM(t,INTERP_TR2, "ExecForRangeStatement"))

	kvLen := len(stmt.KeyAndValues)

	stackPosBefore := t.Pos

	// evaluate the collection expression(s) and push it (them) on the stack.

	// Collection 0 is first one pushed on stack. Others up from there.
	for _, expr := range stmt.X {
		i.EvalExpr(t, expr)
	}
	lastCollectionPos := t.Pos
	nCollections := lastCollectionPos - stackPosBefore // the number of collections pushed onto the stack.	

   var haveNilCollection bool = false  // check for a special case

	var iters []<-chan RObject
	for collPos := stackPosBefore + 1; collPos <= lastCollectionPos; collPos++ {
		var iter <-chan RObject
		coll := t.Stack[collPos]
		switch coll.(type) {
		   case Nil:
		   	  haveNilCollection = true
		   	  break
		   	  // iter = coll.(Nil).Iter(t)  // unnecessary
		   default:
			 collection, isCollection := coll.(RCollection)
			 if ! isCollection {
				rterr.Stopf1(t, stmt, "Attempt to iterate over an object which is not a list, set, or map.")	
			 }	
			 if nCollections == 1 && collection.IsOrdered() && ! collection.IsMap() {
			    iter = nil
			 } else {	 
		      iter = collection.Iter(t)	
		    }   	
		}
		iters = append(iters, iter)
	}

	if haveNilCollection {
	    t.PopN(nCollections) // Pop the collections off the stack. 
	   return
	}

	var idx int64 = 0 // value of index integer in each loop iteration

	keyOffset := kvLen - len(iters) // number of index or key vars before the first value var in for statement

	// TODO Move the decision on what kind this is to here!
	// Based on num and type of collections and lvars

	// Here are the different varieties this could be:
	// 1. for i key val in orderedMap  // keyOffset == 2, 1 coll, coll[0] is map
	// 2. for key val in mapOrOrderedMap // keyOffset == 1, 1 coll, coll[0] is map
	// 3. for i val in listOrOrderedSet // keyOffset == 1, 1 coll, coll[0] is listOrOrderedSet
	// 4. for i val1 val2 in listOrOrderedSetOrOrderedMap listOrOrderedSetOrOrderedMap  // if map is keys
	// 5. for val in anyCollection  // if map then is keys	
	// 6. for val1 val2 in listOfOrderedSetOrOrderedMap listOrOrderedSetOrOrderedMap  // if map is keys

	collPos := stackPosBefore + 1
	collection := t.Stack[collPos].(RCollection)

	switch keyOffset {
	case 2:

		// 1. for i key val in orderedMap  // keyOffset == 2, 1 coll, coll[0] is map		

		if nCollections != 1 {
			rterr.Stop1(t,stmt,"Expecting only one collection, (an ordered map), when there are two more vars than collections.")
		}
		if !collection.IsMap() {
			rterr.Stop1(t,stmt,"Expecting an ordered map, when construct is 'for i key val in orderedMap'.")
		}
		if !collection.IsOrdered() {
			rterr.Stop1(t,stmt,"Expecting an ordered map, when construct is 'for i key val in orderedMap'.")
		}

		// now do the looping

		for {
			moreMembers := false

			key, nextMemberFound := <-iters[0]

			if nextMemberFound {
				moreMembers = true
			}

			if !moreMembers {
				break
			}

			// Assign to the index variable

			idxVar := stmt.KeyAndValues[0].(*ast.Ident)
			LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, idxVar.Name, idxVar.Offset)
			t.Stack[t.Base+idxVar.Offset] = Int(idx)

			// Assign to the key variable

			keyVar := stmt.KeyAndValues[1].(*ast.Ident)
			LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, keyVar.Name, keyVar.Offset)
			t.Stack[t.Base+keyVar.Offset] = key

			// Fetch the map value for the key 	

			mapColl := collection.(Map)
			obj, _ := mapColl.Get(key) // TODO Implement Get!!!!!!!!!!!!	

			// Assign to the value variable

			valVar := stmt.KeyAndValues[2].(*ast.Ident)
			LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, valVar.Name, valVar.Offset)
			t.Stack[t.Base+valVar.Offset] = obj

			// Execute loop body	

			breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)

			if breakLoop || returnFrom {
				breakLoop = false
				continueLoop = false
				break
			}

			// increment the loop iteration index
			idx += 1

		}
	case 1:

		if nCollections == 1 {
			if collection.IsMap() {

				// 2. for key val in mapOrOrderedMap 
				
				theMap := collection.(Map)
				
				for {
					moreMembers := false

					key, nextMemberFound := <-iters[0]

					if nextMemberFound {
						moreMembers = true
					}

					if !moreMembers {
						break
					}

					// Assign to the key variable

					keyVar := stmt.KeyAndValues[0].(*ast.Ident)
					LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, keyVar.Name, keyVar.Offset)
					t.Stack[t.Base+keyVar.Offset] = key

					// Assign to the value variable

					valVar := stmt.KeyAndValues[1].(*ast.Ident)
					LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, valVar.Name, valVar.Offset)
					t.Stack[t.Base+valVar.Offset],_ = theMap.Get(key)

					// Execute loop body	

					breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)

					if breakLoop || returnFrom {
						breakLoop = false
						continueLoop = false
						break
					}
				}				

			} else {

				// 3. for i val in listOrOrderedSet 

				if !collection.IsOrdered() {
					rterr.Stop1(t,stmt,"Expecting a list or ordered set when construct is 'for i val in listOrOrderedSet'.")
				}
				ordColl := collection.(OrderedCollection)
				length := collection.Length()

				// now do the looping

				for ; idx < length; idx++ {
					// moreMembers := false

               obj := ordColl.At(t, int(idx)) 
               
					// obj, nextMemberFound := <-iters[0]

					// if nextMemberFound {
					// 	moreMembers = true
					// }

					// if !moreMembers {
					// 	break
					// }

					// Assign to the index variable

					idxVar := stmt.KeyAndValues[0].(*ast.Ident)
					LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, idxVar.Name, idxVar.Offset)
					t.Stack[t.Base+idxVar.Offset] = Int(idx)

					// Assign to the value variable

					valVar := stmt.KeyAndValues[1].(*ast.Ident)
					LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, valVar.Name, valVar.Offset)
					t.Stack[t.Base+valVar.Offset] = obj

					// Execute loop body	

					breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)

					if breakLoop || returnFrom {
						breakLoop = false
						continueLoop = false
						break
					}

					// increment the loop iteration index
					// idx += 1
				}
			}
		} else { // more than one collection

			// 4. for i val1 val2 in listOrOrderedSetOrOrderedMap listOrOrderedSetOrOrderedMap  // if map, is keys

			for collPos = stackPosBefore + 1; collPos <= lastCollectionPos; collPos++ {
				collection := t.Stack[collPos].(RCollection)
				if !collection.IsOrdered() {
					rterr.Stop1(t,stmt,"Expecting lists or other ordered collections when construct is 'for i val1 val2 ... in coll1 coll2 ...'.")
				}
			}

			// now do the looping

			for {
				moreMembers := false

				for k := 0; k < len(iters); k++ {
					obj, nextMemberFound := <-iters[k]

					if nextMemberFound {
						moreMembers = true
					} else if (k == len(iters)-1) && (!moreMembers) { // we are done. All iterators are exhausted.
						break
					}

					// Assign to the value variable

					valVar := stmt.KeyAndValues[k+1].(*ast.Ident)
					LogM(t,INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, valVar.Name, valVar.Offset)
					t.Stack[t.Base+valVar.Offset] = obj
				}

				if !moreMembers {
					break
				}

				// Assign to the index variable

				idxVar := stmt.KeyAndValues[0].(*ast.Ident)
				LogM(t,INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, idxVar.Name, idxVar.Offset)
				t.Stack[t.Base+idxVar.Offset] = Int(idx)

				// Execute loop body	

				breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)

				if breakLoop || returnFrom {
					breakLoop = false
					continueLoop = false
					break
				}

				// increment the loop iteration index
				idx += 1
			}

		}

	case 0:
		if nCollections == 1 {

			// 5. for val in anyCollection  // if map, is keys	

			// now do the looping

         if collection.IsMap() {
   			for {
   				moreMembers := false

   				obj, nextMemberFound := <-iters[0]

   				if nextMemberFound {
   					moreMembers = true
   				}

   				if !moreMembers {
   					break
   				}

   				// Assign to the value variable

   				valVar := stmt.KeyAndValues[0].(*ast.Ident)
   				LogM(t,INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, valVar.Name, valVar.Offset)
   				t.Stack[t.Base+valVar.Offset] = obj

   				// Execute loop body	

   				breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)

   				if breakLoop || returnFrom {
   					breakLoop = false
   					continueLoop = false
   					break
   				}

   				// increment the loop iteration index
   				idx += 1
   			}
		 } else if collection.IsOrdered() {
					
				
			ordColl := collection.(OrderedCollection)
		  	length := collection.Length()			
			
			// now do the looping

			for ; idx < length; idx++ {
					// moreMembers := false

               obj := ordColl.At(t, int(idx)) 
               
					// obj, nextMemberFound := <-iters[0]

					// if nextMemberFound {
					// 	moreMembers = true
					// }

					// if !moreMembers {
					// 	break
					// }

					// Assign to the index variable

					// idxVar := stmt.KeyAndValues[0].(*ast.Ident)
					// LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, idxVar.Name, idxVar.Offset)
					// t.Stack[t.Base+idxVar.Offset] = Int(idx)

					// Assign to the value variable

					valVar := stmt.KeyAndValues[0].(*ast.Ident)
					LogM(t, INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, valVar.Name, valVar.Offset)
					t.Stack[t.Base+valVar.Offset] = obj

					// Execute loop body	

					breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)

					if breakLoop || returnFrom {
						breakLoop = false
						continueLoop = false
						break
					}
					// increment the loop iteration index
					// idx += 1
				}			
			} else {  // Unordered
				for {
					obj, nextMemberFound := <-iters[0]

					if ! nextMemberFound {
						break
					}

					// Assign to the value variable

					valVar := stmt.KeyAndValues[0].(*ast.Ident)
					LogM(t,INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, valVar.Name, valVar.Offset)
					t.Stack[t.Base+valVar.Offset] = obj

					// Execute loop body	

					breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)

					if breakLoop || returnFrom {
						breakLoop = false
						continueLoop = false
						break
					}

					// increment the loop iteration index
					idx += 1
				}
			}			
		} else { // more than one collection - they must be ordered 	

			// 6. for val1 val2 in listOfOrderedSetOrOrderedMap listOrOrderedSetOrOrderedMap  // if map, is keys

			for collPos = stackPosBefore + 1; collPos <= lastCollectionPos; collPos++ {
				collection := t.Stack[collPos].(RCollection)
				if !collection.IsOrdered() {
					rterr.Stop1(t,stmt,"Expecting lists or other ordered collections when construct is 'for val1 val2 ... in coll1 coll2 ...'.")
				}
			}

			// now do the looping

			for {
				moreMembers := false

				for k := 0; k < len(iters); k++ {
					obj, nextMemberFound := <-iters[k]

					if nextMemberFound {
						moreMembers = true
					} else if (k == len(iters)-1) && (!moreMembers) { // we are done. All iterators are exhausted.
						break
					}

					// Assign to the value variable

					valVar := stmt.KeyAndValues[k].(*ast.Ident)
					LogM(t,INTERP2_, "for range assignment base %d varname %s offset %d\n", t.Base, valVar.Name, valVar.Offset)
					t.Stack[t.Base+valVar.Offset] = obj
				}

				if !moreMembers {
					break
				}

				// Execute loop body	

				breakLoop, _, returnFrom = i.ExecBlock(t, stmt.Body)

				if breakLoop || returnFrom {
					breakLoop = false
					continueLoop = false
					break
				}

				// increment the loop iteration index
				idx += 1
			}

		}

	default:
		rterr.Stop1(t,stmt,"too many or too few variables in for statement.")
	}

	/*	




			for  {
				moreMembers := false

			    var key RObject 

				handled := false

				switch keyOffset {
				   case 2: // there must be a single ordered map collection

		              idxVar := stmt.KeyAndValues[0].(*ast.Ident)
				      Log(INTERP2_,"for range assignment base %d varname %s offset %d\n",t.Base,idxVar.Name,idxVar.Offset)
					  t.Stack[t.Base + idxVar.Offset] = Int(idx)

					  key = <-iters[0] 

		              keyVar := stmt.KeyAndValues[1].(*ast.Ident)	
				      Log(INTERP2_,"for range assignment base %d varname %s offset %d\n",t.Base,keyVar.Name,keyVar.Offset)
					  t.Stack[t.Base + keyVar.Offset] = key			

					  collPos := stackPosBefore + 1
					  mapColl := t.Stack[collPos].(Map)
				      obj,_ = mapColl.Get(key)		       // TODO Implement Get!!!!!!!!!!!!	

		              valVar := stmt.KeyAndValues[2].(*ast.Ident)	
				      Log(INTERP2_,"for range assignment base %d varname %s offset %d\n",t.Base,valVar.Name,valVar.Offset)
					  t.Stack[t.Base + valVar.Offset] = obj	

				 case 1:	 
		              idxVar := stmt.KeyAndValues[0].(*ast.Ident)

		              // TODO TODO !!!!
		              // Have to check if the first and only collection is a Map
		              // - if so, idxVar gets the key - if not, it gets the index

					  collPos := stackPosBefore + 1
					  coll := t.Stack[collPos].(RCollection)
			          if coll.IsMap() && nCollections == 1 {
				      }

				      Log(INTERP2_,"for range assignment base %d varname %s offset %d\n",t.Base,idxVar.Name,idxVar.Offset)
					  t.Stack[t.Base + idxVar.Offset] = Int(idx)

				}	

				for j:= 0; j < len(iters); j++ {
					obj, nextMemberFound := <-iters[j]
					if nextMemberFound {
						moreMembers = true
					} 

					switch keyOffset {
				       case 2: // Collection must be an OrderedMap
					      key = obj
						  mapColl := t.Stack[collPos].(Map)
					      obj,_ = mapColl.Get(key)

					      stmt.KeyAndValues[0]   // assign it Int(idx)
					      stmt.KeyAndValues[1]   // assign it key
					      stmt.KeyAndValues[2]   // assign it obj
					   case 1:
						  // Have

					   case 0:
					   default: 
					      rterr.Stop("More l-values than allowed in for range statement.")	
					}

					// Assign to the right variable



				}
				if ! moreMembers {
					break
				}

				// set index variable if there is one to idx

				// Execute loop body



				idx += 1
			}






			collection := t.Pop().(RCollection)
			if collection.IsList() {
				list := collection.(List)
				v := list.Vector()
		        len := v.Length()

		        for i := 0; i < len; i++ {
			         idx := Int(i)
			         obj := v.At(i)
		        }		
			} else if collection.IsSet() {
				set := collection.(Set)
			}	

			   case *rlist:
		          list := collection.(*rlist)
		          len := list.Length()		
		          for i := 0; i < len; i++ {
			         idx := Int(i)

		          }

			   case *rset:
			   // TODO: maps
			   default:
				  rterr.Stop("Argument to 'for var(s) in collection(s)' is not a collection.")

			}
			for i,val := range 


				i.EvalExpr(t, stmt.Cond) 
				if t.Pop().IsZero() {

				} else {
				   breakLoop,_,returnFrom = i.ExecBlock(t, stmt.Body)
				   for (! breakLoop) && (! returnFrom) {
				      i.EvalExpr(t, stmt.Cond) 		
				      if t.Pop().IsZero() {
					     break
					  }	   	
			          breakLoop,_,returnFrom = i.ExecBlock(t, stmt.Body)		    
				   }
				   breakLoop = false
				   continueLoop = false
				}
		    }



	*/
	t.PopN(nCollections) // Pop the collections off the stack. 

	return
}


/*
 */
func (i *Interpreter) ExecMethodCall(t *Thread, t2 *Thread, call *ast.MethodCall) {
	// defer UnM(t,TraceM(t,INTERP_TR, "ExecMethodCall"))
	nResults := i.EvalMethodCall(t, t2, call)
	t.PopN(nResults) // Discard the results of the method call. No one wants them.
}

/*
TODO
*/
func (i *Interpreter) ExecAssignmentStatement(t *Thread, stmt *ast.AssignmentStatement) {
	// defer UnM(t,TraceM(t,INTERP_TR2, "ExecAssignmentStatement"))
	
	startPos := t.Pos  // top of stack at beginning of assignment statement execution

    for _,rhsExpr := range stmt.Rhs {
 		i.EvalExpr(t, rhsExpr)   	
    }
    numRhsValues := t.Pos - startPos
    numLhsLocations := len(stmt.Lhs)


    if stmt.Tok == token.ARROW { // send to channel
       if numLhsLocations != 1 || numRhsValues != 1 {   // Channel send operator only accepts one lhs channel and one rhs expr
 		   rterr.Stop1(t, stmt, "Can only send a single value to a single channel in each invocation of '<-' operator.")  
       }

	   lhsExpr := stmt.Lhs[0]
			
	   var c *Channel
	   switch lhsExpr.(type) {
	   case *ast.Ident: // A local variable or parameter or result parameter
		   LogM(t,INTERP2_, "send to channel varname %s\n", lhsExpr.(*ast.Ident).Name)
			obj, err := t.GetVar(lhsExpr.(*ast.Ident).Offset)
			if err != nil {
				rterr.Stopf1(t, lhsExpr, "Attempt to access the value of unassigned variable %s.",lhsExpr.(*ast.Ident).Name)
			}		
            c= obj.(*Channel)
		case *ast.SelectorExpr:
		   selector := lhsExpr.(*ast.SelectorExpr)			
		   LogM(t,INTERP2_, "send to channel attr name %s\n", selector.Sel.Name)			
	  	   i.EvalSelectorExpr(t, selector)	      
		   c = t.Pop().(*Channel)      


		// TODO case index expression []
		   		
       }

       val := t.Pop()
       // TODO do a runtime type-compatibility check of val's type with c.ElementType

       if val.IsUnit() || val.IsCollection() || val.Type() == ClosureType {
          i.rt.IncrementInTransitCount(val)
       }
   
       // t.AllowGC(" Channel <-")      
       t.AllowGC()
	   c.Ch <- val
	   // t.DisallowGC(" Channel <-")	  
	   t.DisallowGC()	      


    } else { // assignment

	    if numLhsLocations != numRhsValues  {
	    	var errMessage string
	    	if numRhsValues == 1 { 
	    		errMessage = fmt.Sprintf("Cannot assign 1 right-hand-side value to %d left-hand-side variables/attributes.",numLhsLocations)
	    	} else if numLhsLocations == 1 {
	    		errMessage = fmt.Sprintf("Cannot assign %d right-hand-side values to 1 left-hand-side variable/attribute.",numRhsValues)    		
	    	} else {
	            errMessage = fmt.Sprintf("Cannot assign %d right-hand-side values to %d left-hand-side variables/attributes.",numRhsValues, numLhsLocations)  
	        }    	
	 		rterr.Stop1(t,stmt,errMessage)   	
	    }

	    // Pop rhs values of the stack one by one, assigning each of them to a successive lhs location, starting with the last lhs expr and going
	    // from right to left.
	    //
	    for j := numLhsLocations - 1; j >= 0; j-- {
	       lhsExpr := stmt.Lhs[j]

			switch lhsExpr.(type) {
			case *ast.Ident: // A local variable or parameter or result parameter
				LogM(t,INTERP2_, "assignment base %d varname %s offset %d\n", t.Base, lhsExpr.(*ast.Ident).Name, lhsExpr.(*ast.Ident).Offset)
                             


				switch stmt.Tok {
				case token.ASSIGN:

				   t.Stack[t.Base+lhsExpr.(*ast.Ident).Offset] = t.Pop()

	              
				case token.ADD_ASSIGN:

					assignee, err := t.GetVar(lhsExpr.(*ast.Ident).Offset)
					if err != nil {
						rterr.Stopf1(t, lhsExpr, "Attempt to access the value of unassigned variable %s.",lhsExpr.(*ast.Ident).Name)
					}	
					coll,isAddColl := assignee.(AddableCollection) 
					if ! isAddColl {
						rterr.Stopf1(t, lhsExpr, "%s is not an appendable list or set.",lhsExpr.(*ast.Ident).Name)						
					}
                    val := t.Pop()
					if !val.Type().LessEq(coll.ElementType()) {
						rterr.Stopf1(t,lhsExpr,"Cannot append a '%v' to %s. Elements must be of type '%v'.",  val.Type(), lhsExpr.(*ast.Ident).Name, coll.ElementType())
						return
					}					
					// TODO The persistence aspect of appending to a collection
                    //coll.Add(val, t.EvalContext)	

	               	// Replaced because now also handle the persistence aspect of appending to a collection
	                // coll.Add(val, t.EvalContext)					   
						  
	                err = RT.AddToCollection(coll, val, false, t.EvalContext) 
		            if err != nil {					
		               rterr.Stop1(t,lhsExpr,err)       				                               
		            } 



				case token.SUB_ASSIGN:

					assignee, err := t.GetVar(lhsExpr.(*ast.Ident).Offset)
					if err != nil {
						rterr.Stopf1(t, lhsExpr, "Attempt to access the value of unassigned variable %s.",lhsExpr.(*ast.Ident).Name)
					}	
					coll,isRemColl := assignee.(RemovableCollection) 
					if ! isRemColl {
						rterr.Stopf1(t, lhsExpr, "%s is not a list or set allowing element removal.",lhsExpr.(*ast.Ident).Name)						
					}
                    val := t.Pop()					

           	        // Replaced because now also handle he persistence aspect of removing from a collection
					// coll.Remove(val)					   

	                err = RT.RemoveFromCollection(t, coll, val, true) 
		            if err != nil {					
		               rterr.Stop1(t,lhsExpr,err)       				                               
		            } 	

				default:
					panic("Unrecognized assignment operator")
				}


			// Handle dot (selector) expressions. First lhsEvaluate lhsExpr to put a
			// cell reference onto the stack, then pop it and assign to it.
			case *ast.SelectorExpr:
			   //
			   // a.b =
			   // a.b +=
			   // a.b -=
			   //
				selector := lhsExpr.(*ast.SelectorExpr)
				i.EvalExpr(t, selector.X) // Evaluate the left part of the selector expression.		          

				assignee := t.Pop()       // the robject whose attribute is being assigned to.

				// To speed this up at runtime, could, during parse time, have set an attr field (tbd) of the Sel ident.
				//
				// Except... We don't even know if this name is a one-arg method or an attribute, or which setter
				// (from which type) to use. TODO TODO TODO. In this usage, in lhs of an assignment,
				// it has to be an attribute or it's an error.
				//
				if assignee == nil {					
					rterr.Stop1(t,selector,fmt.Sprintf("%s is unassigned. Can't evaluate %s.%s", selector.X, selector.X, selector.Sel.Name))									
				}
				attr, found := assignee.Type().GetAttribute(selector.Sel.Name)
				if !found {
					rterr.Stop1(t,selector,fmt.Sprintf("Attribute %s not found in type %v or supertypes.", selector.Sel.Name, assignee.Type()))
				}

				switch stmt.Tok {
				case token.ASSIGN:
	                if attr.Part.CollectionType != "" && attr.Part.ArityHigh != 1 {
	                   // A multi-valued collection

		                val := t.Pop()
		                if val.IsCollection() {
			                coll := val.(RCollection)
			                for v := range coll.Iter(t) {
								err := RT.AddToAttr(t, assignee, attr, v, true, t.EvalContext, false)
								if err != nil {
										rterr.Stop1(t, selector, err)
									} 					
			                }
		                } else if val == NIL {
			                err := i.rt.ClearAttr(t, assignee, attr)
			                if err != nil {
							    rterr.Stop1(t, selector, err)	
							    }	
					       } else {
			                rterr.Stop1(t, selector,"Only nil or a collection can be assigned to a multi-valued attribute.")						
						    }
	                } else { // Not a multi-valued attribute.		

						   err := RT.SetAttr(t, assignee, attr, t.Pop(), true, t.EvalContext, false)
						   if err != nil {
								   rterr.Stop1(t,selector, err)
							   } 
				    }
				    		    
				case token.ADD_ASSIGN:
				   if attr.Part.ArityHigh != 1 {
					    // A multi-valued attribute
	   					err := RT.AddToAttr(t, assignee, attr, t.Pop(), true, t.EvalContext, false)
	   					if err != nil {
	   							rterr.Stop1(t,selector,err)
	   						} 					
				   } else {
				      // A single-valued attribute whose value is a collection

                        attrVal, err := RT.EnsureCollectionAttributeVal(t, assignee, attr) 
                        if err != nil {
	   						rterr.Stop1(t,selector,err)                        	
                        }

						// attrVal, found := RT.AttrVal(assignee, attr)
						// if ! found {
                              // TODO Create collection
							  // TODO assign it to the attribute, which persists it if assignee is persistent.
							  // 
                              // Obsolete
						      // rterr.Stopf1(t,selector,"Can't add to collection. %s.%s has not been assigned a collection yet.", selector.X, selector.Sel.Name)               	   
						// }									
																		
						coll,isAddColl := attrVal.(AddableCollection) 
						if ! isAddColl {
						      rterr.Stopf1(t,selector,"%s.%s is not an appendable list or set.", selector.X, selector.Sel.Name)               								
						}
						val := t.Pop()
						if !val.Type().LessEq(coll.ElementType()) {
						      rterr.Stopf1(t,selector,"Cannot append a '%v' to %s.%s. Elements must be of type '%v'.",  val.Type(), selector.X, selector.Sel.Name, coll.ElementType())                	   
							return
						}					

		               	// Replaced because now also handle the persistence aspect of appending to a collection
		                // coll.Add(val, t.EvalContext)					   
							  
		                err = RT.AddToCollection(coll, val, false, t.EvalContext) 
			            if err != nil {					
			               rterr.Stop1(t,lhsExpr,err)       				                               
			            } 
				   }
					
				case token.SUB_ASSIGN:
					if attr.Part.ArityHigh != 1 {	
				      // A multi-valued attribute				      			   
						// TODO TODO	
						err := RT.RemoveFromAttr(t, assignee, attr, t.Pop(), false, true)
						if err != nil {
                           rterr.Stop1(t,selector,err)
						}
				    } else {
				   	   // A single-valued attribute whose value is a collection
						attrVal, found := RT.AttrVal(t, assignee, attr)
						if ! found {
							  return
							  // Who cares.
						      // rterr.Stopf1(t,selector,"Can't remove from collection. %s.%s has not been assigned a collection yet.", selector.X, selector.Sel.Name)               	   
						}	
						coll,isRemColl := attrVal.(RemovableCollection) 
						if ! isRemColl {
							rterr.Stopf1(t, selector, "%s.%s is not a list or set allowing element removal.", selector.X, selector.Sel.Name)						
						}
						val := t.Pop()					

						// Replaced because now also handle he persistence aspect of removing from a collection
						// coll.Remove(val)					   

						err := RT.RemoveFromCollection(t, coll, val, true) 
						if err != nil {					
						   rterr.Stop1(t,lhsExpr,err)  
						}

			   		}
				default:
					panic("Unrecognized assignment operator")
				}

			// Handle [ ] (index) expressions. First lhsEvaluate lhsExpr to put a
			// cell reference onto the stack, then pop it and assign to it.
			case *ast.IndexExpr:
				indexExpr := lhsExpr.(*ast.IndexExpr)
				i.EvalExpr(t, indexExpr.X) // Evaluate the left part of the index expression.		
				i.EvalExpr(t, indexExpr.Index) // Evaluate the index of the index expression.		  
				 
				idx := t.Pop()       // the index or map key			     
				assignee := t.Pop()       // the robject whose attribute is being assigned to. OrderedCollection or Map


			    if assignee == nil {					
				   rterr.Stop1(t,indexExpr,fmt.Sprintf("%s is unassigned. Can't evaluate %s[%s]", indexExpr.X, indexExpr.X, indexExpr.Index))									
			    }
			    if idx == nil {					
				   rterr.Stop1(t,indexExpr,fmt.Sprintf("%s is unassigned. Can't evaluate %s[%s]", indexExpr.Index, indexExpr.X, indexExpr.Index))									
			    }			

				collection,isCollection := assignee.(RCollection)
				if ! isCollection {
					rterr.Stopf1(t, indexExpr,"Cannot [index] into a non-collection of type %v.", assignee.Type())
				}

				if collection.IsOrdered() { // Should also check to make sure is not ordered map.
	               // coll := collection.(IndexSettable)
	               
	               var ix int
	               switch idx.(type) {
	               case Int:
	               	ix = int(int64(idx.(Int)))
	               case Int32:
	               	ix = int(int32(idx.(Int32)))
	               case Uint:
	               	ix = int(uint64(idx.(Uint)))                    	                    
	               case Uint32:
	               	ix = int(uint32(idx.(Uint32)))
	               default:
	                 rterr.Stop1(t,indexExpr,"Index value must be an Integer")
	               }

				   switch stmt.Tok {
				   case token.ASSIGN:
	   				   if collection.IsIndexSettable() {
							coll := collection.(IndexSettable)		
							val := t.Pop()
							typeCheck := true 
							if typeCheck && !val.Type().LessEq(coll.ElementType()) {
							  err := fmt.Errorf("Cannot put a value of type '%v' in collection with element type '%v'.", val.Type(), coll.ElementType())				
							  rterr.Stop1(t,indexExpr,err)       				                               
							}
						    owner := coll.Owner()                 
		                    if owner == nil {
		                        if coll.IsBeingStored() {
		                            err := t.DBT().PersistSetCollectionElement(t, coll, val, ix)                         		
		                            if err != nil {					
		   							       rterr.Stop1(t,indexExpr,err)       				                               
		                            }    
		                         }                     
		                    } else {
		                         if owner.IsBeingStored() {
		                            attr := coll.Attribute()
		                            err := t.DBT().PersistSetAttrElement(t, owner, attr , val, ix)  
		                            if err != nil {					
		         							 rterr.Stop1(t,indexExpr,err)       				                               
		                            }   
		                         }                 
		                   	}
	                   	 
	               		    coll.Set(ix,val)     	 
                     			
					   } else {
   					      if collection.IsList() { // Must be a sorting list
   					         rterr.Stop1(t, indexExpr,"Cannot set element at [index] of a sorting list.")	
   					      }					      
      					  rterr.Stopf1(t, indexExpr, "Can only set [index] of an index-settable ordered collection or a map; not a %v.", assignee.Type())					
				       }
				      
					case token.ADD_ASSIGN:
						defer indexErrHandle(t, indexExpr)

						coll := collection.(OrderedCollection)		            
						attrVal := coll.At(t, ix)

						if attrVal == nil {               	
						     rterr.Stopf1(t,indexExpr,"Can't add to collection. %s[%s] has not been assigned a collection yet.", indexExpr.X, indexExpr.Index)	
						}               	        								
																	
		               	innerColl,isAddColl := attrVal.(AddableCollection) 
		               	if ! isAddColl {
						         rterr.Stopf1(t,indexExpr,"%s[%s] is not an appendable list or set.", indexExpr.X, indexExpr.Index)					                   								
		               	}
		                val := t.Pop()
		               	if !val.Type().LessEq(innerColl.ElementType()) {                	   
		 				         rterr.Stopf1(t,indexExpr,"Cannot append a '%v' to %s[%s]. Elements must be of type '%v'.", val.Type(), indexExpr.X, indexExpr.Index, innerColl.ElementType())	
		               		return
		               	}		

		               	// Replaced because now also handle the persistence aspect of appending to a collection
		                // innerColl.Add(val, t.EvalContext)					   
							  
		                err := RT.AddToCollection(innerColl, val, false, t.EvalContext) 
			            if err != nil {					
			               rterr.Stop1(t,indexExpr,err)       				                               
			            } 


					   

						
						
					case token.SUB_ASSIGN:
		            	defer indexErrHandle(t, indexExpr)
		            
						coll := collection.(OrderedCollection)		            	            
						attrVal := coll.At(t, ix)

						if attrVal == nil {               	
						 rterr.Stopf1(t,indexExpr,"Can't remove from collection. %s[%s] has not been assigned a collection yet.", indexExpr.X, indexExpr.Index)	
						}					   
					   	
		               	innerColl,isRemColl := attrVal.(RemovableCollection) 
		               	if ! isRemColl {
		               		rterr.Stopf1(t,indexExpr, "%s[%s] is not a list or set allowing element removal.", indexExpr.X, indexExpr.Index)						
		               	}
						val := t.Pop()	

               	        // Replaced because now also handle he persistence aspect of removing from a collection
						// innerColl.Remove(val)					   
	
		                err := RT.RemoveFromCollection(t, innerColl, val, true) 
			            if err != nil {					
			               rterr.Stop1(t,indexExpr,err)       				                               
			            } 					   
					   
						// rterr.Stop1(t, indexExpr,"[index] -= val  is not supported yet.")
						
					default:
						panic("Unrecognized assignment operator")
					}	


				} else if collection.IsMap() { //  a["key"] = val   a["key"] += val   a["key"] -= val

               theMap := collection.(Map)
               
					switch stmt.Tok {
					case token.ASSIGN:
	                    //theMap.Put(idx, t.Pop(), t.EvalContext)	
	               
						err := i.rt.PutInMapTypeChecked(theMap, idx, t.Pop(), t.EvalContext) 	 
						if err != nil {
						  rterr.Stop1(t, indexExpr, err)
						}	               
	               				   
					case token.ADD_ASSIGN:
					   
						attrVal,found := theMap.Get(idx)					   
						 
						if ! found {               	
						 rterr.Stopf1(t,indexExpr,"Can't add to collection. %s[%s] has not been assigned a collection yet.", indexExpr.X, indexExpr.Index)	
						}               	        								
																
		               	innerColl,isAddColl := attrVal.(AddableCollection) 
		               	if ! isAddColl {
		   			         rterr.Stopf1(t,indexExpr,"%s[%s] is not an appendable list or set.", indexExpr.X, indexExpr.Index)					                   								
		               	}
		                  val := t.Pop()
		               	if !val.Type().LessEq(innerColl.ElementType()) {                	   
		   			         rterr.Stopf1(t,indexExpr,"Cannot append a '%v' to %s[%s]. Elements must be of type '%v'.", val.Type(), indexExpr.X, indexExpr.Index, innerColl.ElementType())	
		               		return
		               	}					
			
		               	// Replaced because now also handle the persistence aspect of appending to a collection
		                // innerColl.Add(val, t.EvalContext)	                  			   
					   
		                err := RT.AddToCollection(innerColl, val, false, t.EvalContext) 
			            if err != nil {					
			               rterr.Stop1(t,indexExpr,err)       				                               
			            } 


						
			
					case token.SUB_ASSIGN:
					   
						attrVal,found := theMap.Get(idx)					   
						 
						if ! found {					                	
						 rterr.Stopf1(t,indexExpr,"Can't remove from collection. %s[%s] has not been assigned a collection yet.", indexExpr.X, indexExpr.Index)	
						}					   
				   	
		               	innerColl,isRemColl := attrVal.(RemovableCollection) 
		               	if ! isRemColl {
		               		rterr.Stopf1(t,indexExpr, "%s[%s] is not a list or set allowing element removal.", indexExpr.X, indexExpr.Index)						
		               	}
						val := t.Pop()		

               	        // Replaced because now also handle he persistence aspect of removing from a collection
						// innerColl.Remove(val)					   
	
		                err := RT.RemoveFromCollection(t, innerColl, val, true) 
			            if err != nil {					
			               rterr.Stop1(t,indexExpr,err)       				                               
			            } 	
						
					default:
						panic("Unrecognized assignment operator")
					}			

				} else {
					rterr.Stopf1(t, indexExpr, "Can only set or get [index] of an index-settable ordered collection or a map; not a %v.", assignee.Type())					
				}			

			default:
				rterr.Stop1(t, lhsExpr, "Left-hand side expr must be variable, attribute, or indexed position in map/collection.")

			}       
	    }
    }

/*
	for j, lhsExpr := range stmt.Lhs {
		rhsExpr := stmt.Rhs[j]
		i.EvalExpr(t, rhsExpr)

		//i.EvalExpr(lhsExpr) // This is not right. What I do here depends on the type of lhsExpr and type of cell it results in.
		// It is not necessarily full evaluation that is wanted.

		switch lhsExpr.(type) {
		case *ast.Ident: // A local variable or parameter or result parameter
			Log(INTERP2_, "assignment base %d varname %s offset %d\n", t.Base, lhsExpr.(*ast.Ident).Name, lhsExpr.(*ast.Ident).Offset)

			t.Stack[t.Base+lhsExpr.(*ast.Ident).Offset] = t.Pop()

			// TODO TODO TODO
			// Will have to have reserved space for the local variables here when calling the method!!!     
			// So have to know how many locals there are in the body!!!! Store it in the RMethod!!!
			// Why is this comment here? We already do that.

		// TODO handle dot expressions and [ ] selector expressions. First lhsEvaluate lhsExpr to put a
		// cell reference onto the stack, then pop it and assign to it.
		case *ast.SelectorExpr:
			selector := lhsExpr.(*ast.SelectorExpr)
			i.EvalExpr(t, selector.X) // Evaluate the left part of the selector expression.		      
			assignee := t.Pop()       // the robject whose attribute is being assigned to.

			// To speed this up at runtime, could, during parse time, have set an attr field (tbd) of the Sel ident.
			//
			// Except... We don't even know if this name is a one-arg method or an attribute, or which setter
			// (from which type) to use. TODO TODO TODO. In this usage, in lhs of an assignment,
			// it has to be an attribute or it's an error.
			//
			attr, found := assignee.Type().GetAttribute(selector.Sel.Name)
			if !found {
				panic(fmt.Sprintf("Attribute %s not found in type %v or supertypes.", selector.Sel.Name, assignee.Type()))
			}

			switch stmt.Tok {
			case token.ASSIGN:
				err := RT.SetAttr(assignee, attr, t.Pop(), true, t.EvalContext, false)
				if err != nil {
					panic(err)
				}
			case token.ADD_ASSIGN:
				err := RT.AddToAttr(assignee, attr, t.Pop(), true, t.EvalContext, false)
				if err != nil {
					panic(err)
				}
			case token.SUB_ASSIGN:
				// TODO TODO	
				err := RT.RemoveFromAttr(assignee, attr, t.Pop(), false)
				if err != nil {
					panic(err)
				}
			default:
				panic("Unrecognized assignment operator")
			}

		default:
			rterr.Stop("I only handle simple variable or attribute assignments currently. Not indexed ones.")

		}
	}
	*/
}

/*
Executes expressions in left to right order then places them under the Base pointer on the stack, ready to be
the results of the evaluation of the method.
*/
func (i *Interpreter) ExecReturnStatement(t *Thread, stmt *ast.ReturnStatement) (returnFrom bool) {
	// defer UnM(t,TraceM(t,INTERP_TR, "ExecReturnStatement", "stack top index ==>", t.Base-1))

	p0 := t.Pos
    for _, resultExpr := range stmt.Results {
	   i.EvalExpr(t, resultExpr)
    }	
	p1 := t.Pos
	n := p1 -p0
		
	if stmt.IsYield {
		
		if n != t.YieldCardinality {
			rterr.Stopf1(t, stmt, "Generator expression should yield %d values but yields %d instead.",t.YieldCardinality,n)
		}
		
		// TODO This may be a temporary implementation
        for p := p0+1; p <= p1; p++ {
			t.Objs = append(t.Objs, t.Stack[p])   
		}
		t.PopN(n)
				
	} else {
		returnFrom = true

		if  t.ExecutingMethod.NumReturnArgs != n && ! t.ExecutingMethod.ReturnArgsNamed {          
		   rterr.Stopf1(t,stmt,"Method is declared to return %d values. Returning %d values.", t.ExecutingMethod.NumReturnArgs,n)
		}

        types := t.ExecutingMethod.ReturnSignature.Types 

		for j := n-1; j >=0; j-- {	
			val := t.Pop()   

			// TODO IMPORTANT This type checking should first be sped up by a typetuple compatibility checking cache,
			// Then eventually replaced by static type inference within the method body.

            if val == nil {
           	   rterr.Stopf1(t,stmt,"Returned value #%d is a local variable that has not been assigned a value.",j+1)
            }
            if ! val.Type().LessEq(types[j]) {
            	rterr.Stopf1(t,stmt,"Returned value #%d has type '%v' - not compatible with method's declared return type '%v'.",j+1,val.Type(),types[j])
            }
			t.Stack[t.Base+j-n] = val
			// t.Stack[t.Base+j-n] = t.Pop()  

		}
    }
    return
}



/*
Create an http.Request object. Special-case constructor used in webv service method evaluation.
*/
func (i *Interpreter) createRequest(r *http.Request) (req RObject, err error) {
   
    req, err = i.rt.NewObject("shared.relish.pl2012/relish_lib/pkg/http_srv/Request")
    if err != nil {
    	return 
    }

    reqWrapper := req.(*GoWrapper)

    reqWrapper.GoObj = r

	return 
}