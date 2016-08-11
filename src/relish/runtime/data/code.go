// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   code-go - methods and expressions    (methods are data too)
*/

import (
	"fmt"
	"errors"
	"strings"
	"relish/compiler/ast"
)

/*
   Logical method - i.e. a family of methods that all have the same purpose
   and are selected depending on the types of the arguments.

   data:

   methods - for each parameter arity, all method implementations of that arity 
             of this multi-method.
   cachedMethods - for each typetuple representing actual-argument type tuples
   that the multi-method has been called on so far in the process execution,
   the most compatible most specific argument-signature method implementation
   for the type-tuple. Added to as more calls of the multi-method (on different
   types of arguments) are made.

NOTE!!!!!! METHODS should be a map by arity to
a map by pointer to first parameter type to a list of methods.

Don't use this for zero arity of course.

The zero arity method should be findable from a special type tuple with a zero-length list of types.


TODO Variadic last parameter and wildcard keywords parameter, as in python.
*/
type RMultiMethod struct {
	Name          string
	Methods       map[int][]*RMethod       // methods of each arity which implement this multimethod
	CachedMethods map[*RTypeTuple]*RMethod //
	NumReturnArgs int
	MaxArity int	
	Pkg            *RPackage  // the package that this multimethod is owned by
	                          // Note that a method may be referenced by multiple packages' multimethods
	                          // the method itself points to one of these (does not matter which - it just gets its name from the mm)	
    IsExported    bool  // If true, this is a public method exported to packages that import this package.
    DefinesNewMethod bool  // If true, this multimethod defines at least one new method (i.e. the pkg that owns mm defines a method)
                           // Distinguishes the mm from one that has been created solely to merge exported methods from 2 or more
                           // dependency packages.    
    TraitAbstractMethod *RMethod  // If non-nil, this multimethod is the implementation map for a trait abstract method.
}

/*
   Constructor of a multi-method. Sets its name and makes its maps.
*/
func newRMultiMethod(name string, numReturnArgs int, pkg *RPackage, isExported bool) *RMultiMethod {
	return &RMultiMethod{Name: name, Methods: make(map[int][]*RMethod), CachedMethods: make(map[*RTypeTuple]*RMethod), NumReturnArgs: numReturnArgs, Pkg: pkg, IsExported: isExported}
}

/*
   Clone this multimethod to make an initially identical one but owned by the argument package.
*/ 
func (p * RMultiMethod) Clone(pkg *RPackage) *RMultiMethod {
	mm := newRMultiMethod(p.Name,p.NumReturnArgs,pkg, p.IsExported)
	
    for arity,methodList := range p.Methods {
	   var ms []*RMethod
	   mm.Methods[arity] = append(ms,methodList...)
    }	
	
	for tt,method := range p.CachedMethods {
		mm.CachedMethods[tt] = method
	}
	mm.MaxArity = p.MaxArity
	mm.IsExported = p.IsExported
	mm.TraitAbstractMethod = p.TraitAbstractMethod
	return mm
}

/*
For methods which are not found in p but are found in q, add them to p
*/
func (p * RMultiMethod) MergeInNewMethodsFrom(q *RMultiMethod) {
	aritiesMerged := make(map[int]bool)
	for arity,pMethods := range p.Methods {
		aritiesMerged[arity] = true
		qMethods, found := q.Methods[arity]
		if found {
			var methodsToAdd []*RMethod
			for _, mq := range qMethods {
				inP := false
				for _,mp := range pMethods {
					if mq == mp {
						inP = true
						break
					}
				} 
				if ! inP {
					methodsToAdd = append(methodsToAdd,mq)
				}
			}
			p.Methods[arity] = append(pMethods, methodsToAdd...)
		}
    }
    for arity,qMethods := range q.Methods {
       if ! aritiesMerged[arity] {
		  var noMethods []*RMethod       	
       	  p.Methods[arity] = append(noMethods,qMethods...)
       }
    }

    // Now clear the cached methods map of p, since we have new methods to consider.
    // The cache should not have anything in it anyway, since should be calling this
    // at initial code loading time before any relish execution takes place.

    if len(p.CachedMethods) > 0 {
	   p.CachedMethods = make(map[*RTypeTuple]*RMethod)
    }
}


// RObject interface methods

func (p *RMultiMethod) IsZero() bool {
	return false
}

func (p *RMultiMethod) Type() *RType {
	return MultiMethodType
}

func (p *RMultiMethod) This() RObject {
	return p
}

func (p *RMultiMethod) IsUnit() bool {
	return true
}

func (p *RMultiMethod) IsCollection() bool {
	return false
}

func (p *RMultiMethod) String() string {
	return p.Name
}

func (p *RMultiMethod) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p *RMultiMethod) Debug(th InterpreterThread) string {
	s := fmt.Sprintf("%s.%s Exported:%v MaxArity:%v #RetArgs: %v\n", p.Pkg.ShortName, p.String(), p.IsExported, p.MaxArity, p.NumReturnArgs)
    s += "--- Methods: ---\n"  
    for arity,methods := range p.Methods {
       s += fmt.Sprintf("%d =>\n",arity)
       for _,meth := range methods {
       	  s += "      " + meth.String() + "\n" 
       }
    }
    s += "--- CachedMethods: ---\n"   
    for tt, meth := range p.CachedMethods {
    	s += fmt.Sprintf("%v => %v",tt, meth)
    }

	return s
}

func (p *RMultiMethod) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p *RMultiMethod) UUID() []byte {
	panic("A multimethod cannot have a UUID.")
	return nil
}

func (p *RMultiMethod) DBID() int64 {
	panic("A multimethod cannot have a DBID.")
	return 0
}

func (p *RMultiMethod) EnsureUUID() (theUUID []byte, err error) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) UUIDstr() string {
	panic("A multimethod cannot have a UUID.")
	return ""
}

func (p *RMultiMethod) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) UUIDabbrev() string {
	panic("A multimethod cannot have a UUID.")
	return ""
}

func (p *RMultiMethod) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A multimethod cannot have a UUID.")
	return
}

func (p *RMultiMethod) RemoveUUID() {
	panic("A multimethod does not have a UUID.")
	return
}

func (p *RMultiMethod) Flags() int8 {
	panic("A multimethod has no Flags.")
	return 0
}

func (p *RMultiMethod) IsBeingStored() bool {
	return false
}
func (p *RMultiMethod) SetBeingStored() {
}
func (p *RMultiMethod) ClearBeingStored() {
}

func (p *RMultiMethod) IsIdReversed() bool {
	return false
}

func (p *RMultiMethod) SetIdReversed() {}

func (p *RMultiMethod) ClearIdReversed() {}

func (p *RMultiMethod) IsLoadNeeded() bool {
	return false
}

func (p *RMultiMethod) SetLoadNeeded()   {}
func (p *RMultiMethod) ClearLoadNeeded() {}

func (p *RMultiMethod) IsValid() bool { return true }
func (p *RMultiMethod) SetValid()     {}
func (p *RMultiMethod) ClearValid()   {}

func (p *RMultiMethod) IsMarked() bool { return false }
func (p *RMultiMethod) SetMarked()    {}
func (p *RMultiMethod) ClearMarked()  {}
func (p *RMultiMethod) ToggleMarked()  {}

func (p *RMultiMethod) Mark() bool { return false }

func (p *RMultiMethod) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p *RMultiMethod) SetStoredLocally()     {}
func (p *RMultiMethod) ClearStoredLocally()   {}

func (p *RMultiMethod) IsProxy() bool { return false }

func (o *RMultiMethod) IsTransient() bool { return true }

func (o *RMultiMethod) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (o *RMultiMethod) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a MultiMethod in JSON.")
   return
}

func (o *RMultiMethod) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a MultiMethod.")
   return
}





/*
   A method implementation that applies to a particular tuple of argument types.
*/
type RMethod struct {
	multiMethod    *RMultiMethod
	ParameterNames []string               // names of parameters

    // TODO We're going to need a slice of bools as to whether the method parameters
    // accept nil or not.
    //
    NilArgAllowed []bool    

	Signature      *RTypeTuple            // types of parameters
	WildcardKeywordsParameterName string  // "" or name of the special parameter which takes any number of keyword=val args
	WildcardKeywordsParameterType *RType  // nil or type of Map used for wildcard keyword=val args
	VariadicParameterName string          // "" or name of the special variadic parameter which accepts remainder args
	VariadicParameterType *RType          // nil or type of List used for remainder (variadic) args
	ReturnSignature *RTypeTuple           // types of return values
	ReturnArgsNamed bool                  // Whether the return arguments are named and have to be initialized on method call.
	Code           *ast.MethodDeclaration // abstract syntax tree
	NumReturnArgs  int
	NumLocalVars   int
	NumFreeVars    int  // Usually 0 but may be more in the case of an anonymous Closure-method
	PrimitiveCode  func(InterpreterThread, []RObject) []RObject
	Pkg            *RPackage  // the package that this method is defined in
	File           *ast.File
}


func (p *RMethod) IsAbstract() bool {
   return p.Code != nil && p.Code.Body == nil
}

/*
Number of named positional parameters.
Does not include a variadic (list of extra last args) parameter. Does not include a wildcard keywords parameter. 
*/
func (p *RMethod) Arity() int {
   return len(p.Signature.Types)
}

/*
Obtain the ast.File node for the source code file this method was found in.
Used in printing context info for runtime error messages.
*/
func (p *RMethod) CodeFile() *ast.File {
	return p.File
}

func (p *RMethod) IsZero() bool {
	return false
}

func (m RMethod) Name() string {
	return m.multiMethod.Name
}

func (m RMethod) String() string {
	return fmt.Sprintf("%s %v %v", m.multiMethod.Name, m.ParameterNames, m.Signature)
}

func (m RMethod) StringTh(th InterpreterThread) string {
	return m.String() 
}

func (p *RMethod) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s.%s (Multimethod: %s)", p.Pkg.ShortName, p.StringTh(th), p.multiMethod.Debug(th))
}

func (p *RMethod) Type() *RType {
	return MethodType
}

func (p *RMethod) This() RObject {
	return p
}

func (p *RMethod) IsUnit() bool {
	return true
}

func (p *RMethod) IsCollection() bool {
	return false
}

func (p *RMethod) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p *RMethod) UUID() []byte {
	panic("A Method cannot have a UUID.")
	return nil
}

func (p *RMethod) DBID() int64 {
	panic("A Method cannot have a DBID.")
	return 0
}

func (p *RMethod) EnsureUUID() (theUUID []byte, err error) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) UUIDstr() string {
	panic("A Method cannot have a UUID.")
	return ""
}

func (p *RMethod) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) UUIDabbrev() string {
	panic("A Method cannot have a UUID.")
	return ""
}

func (p *RMethod) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Method cannot have a UUID.")
	return
}

func (p *RMethod) RemoveUUID() {
	panic("A Method does not have a UUID.")
	return
}

func (p *RMethod) Flags() int8 {
	panic("A Method has no Flags.")
	return 0
}

func (p *RMethod) IsBeingStored() bool {
	return false
}
func (p *RMethod) SetBeingStored() {
}
func (p *RMethod) ClearBeingStored() {
}

func (p *RMethod) IsIdReversed() bool {
	return false
}

func (p *RMethod) SetIdReversed() {}

func (p *RMethod) ClearIdReversed() {}

func (p *RMethod) IsLoadNeeded() bool {
	return false
}

func (p *RMethod) SetLoadNeeded()   {}
func (p *RMethod) ClearLoadNeeded() {}

func (p *RMethod) IsValid() bool { return true }
func (p *RMethod) SetValid()     {}
func (p *RMethod) ClearValid()   {}

func (p *RMethod) IsMarked() bool { return false }
func (p *RMethod) SetMarked()    {}
func (p *RMethod) ClearMarked()  {}
func (p *RMethod) ToggleMarked()  {}

func (p *RMethod) Mark() bool { return false }

func (p *RMethod) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p *RMethod) SetStoredLocally()     {}
func (p *RMethod) ClearStoredLocally()   {}

func (p *RMethod) IsProxy() bool { return false }

func (o *RMethod) IsTransient() bool { return true }

func (o *RMethod) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (o *RMethod) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a Method in JSON.")
   return
}

func (o *RMethod) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Method.")
   return
}


/*
   Create a new relish method.
   Methods are collected by name into multi-methods.
   A multi-method maintains a list of method implementations under each
   parameter arity. At runtime, multi-method dispatch selects for application
   the method (of the correct arity), whose parameter type signature is 
   compatible with (same or a generalization of) the types of the actual arguments
   and whose parameter type signature is closest in type to the actual argument types,
   and whose parameter type signature, being arg-type-compatible and closest to the
   arg types, consists of the most specific types.

   This function determines whether a multi-method of the given name already
   exists in the runtime environment. If so, it adds the method implementation
   with the given argument type tuple to the appropriate list of method implementations
   of the multimethod. If the multi-method does not yet exist, it is first
   created and added to the runtime environment.

   If packageName is the empty string, it defaults to "relish.pl2012/core/inbuilt"

   Returns an error if a method implementation with the name and parameter type signature
   already exists. (What scope?)
   TODO How do we handle incremental compilation that includes redefinitions of method
   implementations?

   This is the variant that is used I think to create built-in methods.   
*/
func (rt *RuntimeEnv) CreateMethod(packageName string, file *ast.File, methodName string, parameterNames []string, parameterTypes []string, 
	returnValTypes []string,
	                  returnValNamed bool, numLocalVars int, allowRedefinition bool) (*RMethod, error) {
	    nilArgAllowed := make([]bool,len(parameterTypes))
	    isExported := true
		return rt.CreateMethodGeneral(packageName, file, methodName, parameterNames, nilArgAllowed, parameterTypes, 
						                  nil,
						                  nil,
						                  "",
						                  "",
						                  "",
						                  "",
						                  returnValTypes,
						                  returnValNamed, 
						                  numLocalVars, 
						                  false,
						                  allowRedefinition,
						                  isExported)
}


/*
This is the variant that is called by the generator. It is starting to handle variadic and wildcard keyword args.
*/
func (rt *RuntimeEnv) CreateMethodV(packageName string, file *ast.File, methodName string, parameterNames []string, nilArgAllowed []bool, parameterTypes []string, 
    wildcardKeywordsParameterName string,
    wildcardKeywordsParameterType string,	
	variadicParameterName string,
	variadicParameterType string,
	returnValTypes []string,
	returnValNamed bool, 
	numLocalVars int, 
	isTraitAbstractMethod bool,
	allowRedefinition bool,
	isExported bool) (*RMethod, error) {
		return rt.CreateMethodGeneral(packageName, file, methodName, parameterNames, nilArgAllowed, parameterTypes, 
						                  nil,
						                  nil,
						                  wildcardKeywordsParameterName,
						                  wildcardKeywordsParameterType,						
						                  variadicParameterName,
						                  variadicParameterType,						
						                  returnValTypes,
						                  returnValNamed, 
						                  numLocalVars, 
						                  isTraitAbstractMethod,
						                  allowRedefinition,
						                  isExported)
}

/*
   Create a new relish method.
   Methods are collected by name into multi-methods.
   A multi-method maintains a list of method implementations under each
   parameter arity. At runtime, multi-method dispatch selects for application
   the method (of the correct arity), whose parameter type signature is 
   compatible with (same or a generalization of) the types of the actual arguments
   and whose parameter type signature is closest in type to the actual argument types,
   and whose parameter type signature, being arg-type-compatible and closest to the
   arg types, consists of the most specific types.

   This function determines whether a multi-method of the given name already
   exists in the runtime environment. If so, it adds the method implementation
   with the given argument type tuple to the appropriate list of method implementations
   of the multimethod. If the multi-method does not yet exist, it is first
   created and added to the runtime environment.

   Returns an error if a method implementation with the name and parameter type signature
   already exists. (What scope?)
   TODO How do we handle incremental compilation that includes redefinitions of method
   implementations?
*/
func (rt *RuntimeEnv) CreateMethodGeneral(packageName string, file *ast.File, methodName string,  parameterNames []string, nilArgAllowed []bool, parameterTypes []string, 
	                  // FROM HERE DOWN IS NEW
	                  keyWordParameterDefaults map[string] RObject,
	                  keyWordParameterTypes map[string]string,
	                  wildcardKeywordsParameterName string,
	                  wildcardKeywordsParameterType string,
	                  variadicParameterName string,
	                  variadicParameterType string,
	                  // FROM HERE UP IS NEW
	                  returnValTypes []string,
	                  returnArgsNamed bool,
	                  numLocalVars int, 
	                  isTraitAbstractMethod bool, 
	                  allowRedefinition bool,
	                  isExported bool) (*RMethod, error) {

	if packageName == "" { 
		packageName = "relish.pl2012/core/inbuilt"
	}

    // Set the package into which the method is to be added.
    // If it is the default inbuilt functions package and does not exist, create it.
    // Note, the default inbuilt functions package, "relish.pl2012/core/inbuilt", is
    // the package which inbuilt methods, which do not need to have any package imported for them
    // to be visible in code, are added to.
    // packages whose full path starts with "relish/pkg" are packages which hold
    // only native methods and have no real relish artifact or package loaded. So these packages
    // need to be created here.
    // 
	pkg := rt.Packages[packageName]
	if pkg == nil {
		if packageName == "relish.pl2012/core/inbuilt"  {
		    pkg = rt.CreatePackage(packageName, false)
	    } else if strings.HasPrefix(packageName, "relish/pkg/") {
		   pkg = rt.CreatePackage(packageName, true)
	    }
	}


	arity := len(parameterTypes)
    numReturnArgs := len(returnValTypes)

    var multiMethod *RMultiMethod
    var found bool

    if isTraitAbstractMethod {
    	if ! isExported {
		   return nil, fmt.Errorf("Trait abstract method '%v' cannot be defined in a _private.rel file.", methodName)
        }
		multiMethod = newRMultiMethod(methodName, numReturnArgs, pkg, isExported)
	    pkg.MultiMethods[methodName] = multiMethod	

    } else {
		//multiMethod, found := rt.MultiMethods[methodName]
		multiMethod, found = pkg.MultiMethods[methodName]

		if !found {
			multiMethod = newRMultiMethod(methodName, numReturnArgs, pkg, isExported)
			pkg.MultiMethods[methodName] = multiMethod
		} else {
			if multiMethod.IsExported != isExported {

               if (! multiMethod.IsExported) && (multiMethod.Pkg == pkg) && (! multiMethod.DefinesNewMethod) {
               	  // This is just a merge mm, but we are now adding a new exported method to it.
               	  multiMethod.IsExported = true
               } else {
   			      fmt.Println(multiMethod.Debug(nil))
			      fmt.Println(isExported)
		          return nil, fmt.Errorf("Method '%v' cannot be defined as private in one source code file and as exported in another.", methodName)				
			   }
			}
			if multiMethod.NumReturnArgs != numReturnArgs {
			   return nil, fmt.Errorf("Method '%v' is already defined to have %v return arguments and cannot have other than that.", methodName, multiMethod.NumReturnArgs)
	        }
	        if multiMethod.Pkg != pkg { // Make sure we have a multimethod that this package is allowed to modify
		
		                                // TODO ---- WARNING ----- We modify a multimethod when caching! is it ok to 
		                                // do caching on a dependency package's multimethod? probably not!! 
		                                // Needs more thought. Ok for a package's multimethod to cache typetuples
		                                // where the types are not even known within that package??? Maybe it's ok
		                                // because we only cache based on a lookup done with that multimethod in first place!!!
		
		        multiMethod = multiMethod.Clone(pkg)
		        pkg.MultiMethods[methodName] = multiMethod
	        }
		}
	}
		
	if arity > multiMethod.MaxArity {
		multiMethod.MaxArity = arity
	}
    

	typeTuple, err := rt.getTypeTupleFromTypes(parameterTypes)
	if err != nil {
		return nil, err
	}
	resultTypeTuple, err := rt.getTypeTupleFromTypes(returnValTypes)
	if err != nil {
		return nil, err
	}	
	
	
    var wildcardKeywordsParamType *RType = nil
	var variadicParamType *RType = nil
	var typFound bool


	if wildcardKeywordsParameterName != "" {
	   wildcardKeywordsParamType, typFound = rt.Types[wildcardKeywordsParameterType]
	   if ! typFound {
	      return nil, fmt.Errorf("Method '%v' keyword parameter type %v not found.", methodName, wildcardKeywordsParameterType)
	   }  	
	}

	if variadicParameterName != "" {
	   variadicParamType, typFound = rt.Types[variadicParameterType]
	   if ! typFound {
	      return nil, fmt.Errorf("Method '%v' variadic parameter type %v not found.", methodName, variadicParameterType)
	   } 		
	}
/* TODO Not sure the status of this suggestion. Is it older code or a suggested change. I forget.
    if wildcardKeywordsParameterName != "" {
       typ, typFound := rt.Types[wildcardKeywordsParameterType]
       if ! typFound {
	      return nil, fmt.Errorf("Method '%v' keyword parameter type %v not found.", methodName, wildcardKeywordsParameterType)
       }  	
       wildcardKeywordsParamType, err = rt.GetMapType(StringType, typ)	 
       if err != nil {
	      return nil, err
       }
    }

	if variadicParameterName != "" {
	   typ, typFound := rt.Types[variadicParameterType]
	   if ! typFound {
	      return nil, fmt.Errorf("Method '%v' variadic parameter type %v not found.", methodName, variadicParameterType)
	   } 	
       variadicParamType, err = rt.GetListType(typ)	 
       if err != nil {
	      return nil, err
       }	
	}
*/	
	
	method := &RMethod{
		multiMethod:    multiMethod,
		ParameterNames: parameterNames,
		NilArgAllowed: nilArgAllowed,
		Signature:      typeTuple,
		WildcardKeywordsParameterName: wildcardKeywordsParameterName,
		WildcardKeywordsParameterType: wildcardKeywordsParamType,
		VariadicParameterName: variadicParameterName,
		VariadicParameterType: variadicParamType,
		ReturnSignature: resultTypeTuple,
		ReturnArgsNamed: returnArgsNamed,		
		NumReturnArgs:  numReturnArgs,
		NumLocalVars:   numLocalVars,
		Pkg:            pkg,
		File:           file,
	}

	methodsOfRightArity, multiMethodHasArity := multiMethod.Methods[arity]

    // TODO This check does not seem correct. For one, how do we know the other methods have been put in the cache?
	_, found = multiMethod.CachedMethods[typeTuple]
	if found {
		if !allowRedefinition {
			return nil, fmt.Errorf("Method '%v' is already defined for types %v.", methodName, typeTuple)
		}

		for i, oldMethod := range methodsOfRightArity {
			if oldMethod.Signature == typeTuple {
				methodsOfRightArity[i] = method
			}
		}

	} else { // new method

		if multiMethodHasArity {
			multiMethod.Methods[arity] = append(methodsOfRightArity, method)
		} else {
			multiMethod.Methods[arity] = []*RMethod{method}
		}
	}


	multiMethod.DefinesNewMethod = true  // This multi-method defines at least one method that has not been defined in merged-in mms.



	if isTraitAbstractMethod {
		multiMethod.TraitAbstractMethod = method

		// Search rt.MultiMethods for compatible implementing methods.
		// Add each of them into this multiMethod

        globalMethodName := fmt.Sprintf("%s___%d",methodName, numReturnArgs)

        implementorsMultiMethod := rt.MultiMethods[globalMethodName]
        if implementorsMultiMethod != nil {
            possibleImplementorMethods := implementorsMultiMethod.Methods[arity]
            for _,possibleImplementorMethod := range possibleImplementorMethods {
            	if possibleImplementorMethod.Signature.LessEq(method.Signature) {
            	   // Yes. This method is an implementor of the trait abstract method (interface).
			       if possibleImplementorMethod.Signature == method.Signature {
			       	  // Remove the abstract method from the multimethod.Methods
			       	  meths := multiMethod.Methods[arity]
			       	  meths2 := []*RMethod{}
			       	  for _,meth := range meths {
			       	  	  if meth != method {
			       	  	     meths2 = append(meths2, meth)
			       	  	  }
			       	  }
			       	  multiMethod.Methods[arity] = meths2 
			       } 

			       multiMethod.Methods[arity] = append(multiMethod.Methods[arity], possibleImplementorMethod)  

			    }   

            }
        }

        // Keep a global record of the multimethods of all trait abstract methods.
		rt.AddTraitMultiMethod(multiMethod)

	} else if arity > 0 && isExported {  // Add to the global runtime multimethod map.

        globalMethodName := fmt.Sprintf("%s___%d",methodName, numReturnArgs)

		globalMultiMethod, found := rt.MultiMethods[globalMethodName]


		if !found {
			globalMultiMethod = newRMultiMethod(globalMethodName, numReturnArgs, nil, isExported)
			rt.MultiMethods[globalMethodName] = globalMultiMethod
		}
		
		
		if arity > globalMultiMethod.MaxArity {
			globalMultiMethod.MaxArity = arity
		}


        traitMultiMethods := rt.TraitMultiMethods[globalMethodName]
        for _,traitMultiMethod := range traitMultiMethods {
             traitMethod := traitMultiMethod.TraitAbstractMethod        	
        	if traitMethod.Arity() == arity {

               if method.Signature.LessEq(traitMethod.Signature) {
        	   
               	   // This new method is type compatible with the trait abstract method.


			       if method.Signature == traitMethod.Signature {

			       	  // Remove the abstract method from the multimethod.Methods
			       	  meths := traitMultiMethod.Methods[arity]
			       	  meths2 := []*RMethod{}
			       	  for _,meth := range meths {
			       	  	  if meth != traitMethod {
			       	  	     meths2 = append(meths2, meth)
			       	  	  }
			       	  }
			       	  traitMultiMethod.Methods[arity] = meths2 
			       } 

                   // So add this new method to the trait multimethod.

			       traitMultiMethod.Methods[arity] = append(traitMultiMethod.Methods[arity], method)  

                   // Uncache the multimethod methods. A new method means the best method per
                   // argument typetuple needs to be recomputed by specialization distance comparison.
                   if len(traitMultiMethod.CachedMethods) > 0 {
                      traitMultiMethod.CachedMethods = make(map[*RTypeTuple]*RMethod)			
                   }                      	
               }	
        	}
        }
	}

	return method, nil
}


	
	
func (rt *RuntimeEnv) CreateClosureMethod(packageName string, file *ast.File, methodName string, parameterNames []string, nilArgAllowed []bool, parameterTypes []string, 
	returnValTypes []string,
	                  returnValNamed bool, numLocalVars int, numFreeVars int) (*RMethod, error) {
		return rt.CreateClosureMethodGeneral(packageName, file, methodName, parameterNames, nilArgAllowed, parameterTypes, 
						                  nil,
						                  nil,
						                  "",
						                  "",
						                  "",
						                  "",
						                  returnValTypes,
						                  returnValNamed, 
						                  numLocalVars, 
						                  numFreeVars)
}	


func (rt *RuntimeEnv) CreateClosureMethodGeneral(packageName string, file *ast.File, methodName string, parameterNames []string, nilArgAllowed []bool, parameterTypes []string, 
	                  // FROM HERE DOWN IS NEW
	                  keyWordParameterDefaults map[string] RObject,
	                  keyWordParameterTypes map[string]string,
	                  wildcardKeywordsParameterName string,
	                  wildcardKeywordsParameterType string,
	                  variadicParameterName string,
	                  variadicParameterType string,
	                  // FROM HERE UP IS NEW
	                  returnValTypes []string,
	                  returnArgsNamed bool,
	                  numLocalVars int, 
	                  numFreeVars int) (*RMethod, error) {

    // Set the package into which the method is to be added.
    // If it is the default inbuilt functions package and does not exist, create it.
	pkg := rt.Packages[packageName]

    numReturnArgs := len(returnValTypes)

    // TODO - Have to find and/or create this in the pkg object !!!!!!!!!!!!!!!!!
    //
    //
    //
    //
	

	typeTuple, err := rt.getTypeTupleFromTypes(parameterTypes)
	if err != nil {
		return nil, err
	}
	resultTypeTuple, err := rt.getTypeTupleFromTypes(returnValTypes)
	if err != nil {
		return nil, err
	}	
	method := &RMethod{
		multiMethod:    nil,
		ParameterNames: parameterNames,
		NilArgAllowed: nilArgAllowed,
		Signature:      typeTuple,
		ReturnSignature: resultTypeTuple,
		ReturnArgsNamed: returnArgsNamed,		
		NumReturnArgs:  numReturnArgs,
		NumLocalVars:   numLocalVars,
		NumFreeVars:    numFreeVars,
		Pkg:            pkg,
		File:           file,
	}

    pkg.ClosureMethods[methodName] = method

	return method, nil
}


/*
Represents a closure with bound free variables.
*/
type RClosure struct {
	Method *RMethod
	Bindings []RObject
	flags byte	
}

func (p *RClosure) IsZero() bool {
	return false
}

func (m RClosure) String() string {
	return fmt.Sprintf("%v %v", m.Method.ParameterNames, m.Method.Signature)
}

func (m RClosure) StringTh(th InterpreterThread) string {
	return m.String()
}

func (p *RClosure) Debug(th InterpreterThread) string {
	s := p.StringTh(th) + "\n"
	for _,obj := range p.Bindings {
		s += "   " + obj.Debug(th) + "\n"
	}
	return s
}

func (p *RClosure) Type() *RType {
	return ClosureType
}

func (p *RClosure) This() RObject {
	return p
}

func (p *RClosure) IsUnit() bool {
	return true
}

func (p *RClosure) IsCollection() bool {
	return false
}

func (p *RClosure) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p *RClosure) UUID() []byte {
	panic("A Closure cannot have a UUID.")
	return nil
}

func (p *RClosure) DBID() int64 {
	panic("A Closure cannot have a DBID.")
	return 0
}

func (p *RClosure) EnsureUUID() (theUUID []byte, err error) {
	panic("A Closure cannot have a UUID.")
	return
}

func (p *RClosure) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Closure cannot have a UUID.")
	return
}

func (p *RClosure) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Closure cannot have a UUID.")
	return
}

func (p *RClosure) UUIDstr() string {
	panic("A Closure cannot have a UUID.")
	return ""
}

func (p *RClosure) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Closure cannot have a UUID.")
	return
}

func (p *RClosure) UUIDabbrev() string {
	panic("A Closure cannot have a UUID.")
	return ""
}

func (p *RClosure) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Closure cannot have a UUID.")
	return
}

func (p *RClosure) RemoveUUID() {
	panic("A Closure does not have a UUID.")
	return
}

func (o *RClosure) Flags() int8 {
	return int8(o.flags)
}

func (p *RClosure) IsBeingStored() bool {
	return false
}
func (p *RClosure) SetBeingStored() {
}
func (p *RClosure) ClearBeingStored() {
}

func (p *RClosure) IsIdReversed() bool {
	return false
}

func (p *RClosure) SetIdReversed() {}

func (p *RClosure) ClearIdReversed() {}

func (p *RClosure) IsLoadNeeded() bool {
	return false
}

func (p *RClosure) SetLoadNeeded()   {}
func (p *RClosure) ClearLoadNeeded() {}

func (p *RClosure) IsValid() bool { return true }
func (p *RClosure) SetValid()     {}
func (p *RClosure) ClearValid()   {}

func (o *RClosure) IsMarked() bool { return o.flags&FLAG_MARKED != 0 }
func (o *RClosure) SetMarked()    { o.flags |= FLAG_MARKED }
func (o *RClosure) ClearMarked()  { o.flags &^= FLAG_MARKED }
func (o *RClosure) ToggleMarked()  { o.flags ^= FLAG_MARKED }

/*
If the object is not already marked as reachable, flag it as reachable.
Return whether we had to flag it as reachable. false if was already marked reachable.
*/
func (o *RClosure) Mark() bool { 
   if o.IsMarked() == markSense {
   	   return false
   } 
   o.ToggleMarked()

   // Now mark the bound objects 
   for _,obj := range o.Bindings {
   	  obj.Mark()
   }

   return true
}


func (p *RClosure) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p *RClosure) SetStoredLocally()     {}
func (p *RClosure) ClearStoredLocally()   {}

func (p *RClosure) IsProxy() bool { return false }

func (o *RClosure) IsTransient() bool { return true }

func (o *RClosure) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (o *RClosure) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a Closure in JSON.")
   return
}

func (o *RClosure) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Closure.")
   return
}
