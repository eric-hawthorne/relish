// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   runtime.go - the runtime environment for a relish program - manages other entities such
   as packages, types, methods, and object instances and their attributes and relations.
*/

import (
	"fmt"
	. "relish/dbg"
	"sync"
	"strings"
	"errors"
	// "runtime/debug"
)

////////////////////////////////////////////////////////////////////////////////////////////////
//////// RUNTIME ENVIRONMENT
//////// 
//////// Used by the interpreter. Used and populated by the parser.
//////// Contains maps from names to types and methods.
//////// Contains a tree index of all unique type tuples that have been 
//////// used in the definition or a call of a method.
////////
//////// Will also contain references to Relish data objects (global variables) I imagine.
////////
////////////////////////////////////////////////////////////////////////////////////////////////

/*
   RuntimeEnv - the nexus for Relish types and methods and objects and related indexes.
   Has methods to create types and methods, which are used by the parser to create a program
   runtime.
   Also has methods to set and get values of attributes of objects.
   Note: Many of the most important methods of RuntimeEnv are defined in the other
   relish source files which deal specifically with other kinds of runtime entities.
*/
type RuntimeEnv struct {
	Types         map[string]*RType        // map from type name to RType object. 
	Typs          map[string]*RType        // map from type.ShortName() to RType object. 
	Packages      map[string]*RPackage     // by package Name
	Pkgs          map[string]*RPackage     // by package ShortName
	PkgShortNameToName map[string]string   
	PkgNameToShortName map[string]string
	InbuiltFunctionsPackage *RPackage      // shortcut to package containing inbuilt functions
	RunningArtifact string                 // Full origin-qualified name of the artifact one of whose package main functions was run in relish command
	TypeTupleTree *TypeTupleTreeNode  // Obsolete
	TypeTupleTrees []*TypeTupleTreeNode  // New

	context map[string]RObject  // "Global variable" hashtable.
	contextMutex sync.RWMutex

	objects       map[int64]RObject // persistable objects by DBID()
	// objects map[string] RObject // persistable objects by uuidstr - do we need this for distributed computing?
	objectIds map[RObject]uint64 // non-persistable object to its local numeric id (for debug printing the object)
	idGen     *IdGenerator
	db        DB

	dbt       DBT  // A database connection thread. 

	DatabaseURI string  // filepath (or eventually some other kind of db connection uri) of the database for persisting relish objects
	Loader PackageLoader  // Loads code packages into the runtime.

	// Note about attribute values. An attribute may have a single value or be multi-valued,
	// where the multi-valued attribute is implemented by an RCollection.
	// But how do we tell if we have a single-valued attribute whose value is an ROject that
	// happens to be a collection? The answer is that a multi-value-attribute-implementing
	// RCollection must have a reference to its owner RObject. A plain-old RCollection object
	// which is a single value always has a nil owner.
	//
	// attributes map[*AttributeSpec]map[RObject]RObject


	// TODO HOW do we deal with object networks that are not fully retrieved 
	// from storage yet? i.e. that need to be lazily loaded into memory?

	// Perhaps the trick is to load a collection valued attribute (when it or a member is requested)
	// in a partial way as follows: create the collection in memory but make it a collection
	// of RObjects which are nothing but the uint64 id of the true object. They have an IsProxy =true
	// Upon retrieving from the collection, we do a coll[i] = EnsureFetched(coll[i]) which 
	// replaces the ith member with the real RObject.
	// 

	// from fully qualified constant name to value.
	constants map[string]RObject

    // From name of private constant to package in which it is defined.
    // TODO this should be removed and replaced with compile-time constant accessibility checking
    privateConstantPackage map[string]*RPackage

    // The following is to make sure that objects in-transit in a channel
    // can be marked by the relish garbage collector.
    // The map contains objects which have been sent into a channel and not yet received.
    // The value is the number of times the object is currently in a channel.
    inTransit map[RObject]uint32

	// from fully qualified type name to reflect.DataType, a *GoWrapper RObject.
	ReflectedDataTypes map[string]RObject

	// from fully qualified attribute name to reflect.Attribute, a *GoWrapper RObject.
	ReflectedAttributes map[string]RObject

	// a map of contexts for the evaluation of methods.

	evalContexts map[RObject]MethodEvaluationContext

	evalContextMutex sync.Mutex


    // Each package also has its own multimethod map.
    // This global one is only used to find implementations for trait abstract methods.
    //
    // This particular multimethods map contains only exported methods which are not trait abstract
    // methods and which have at least one positional argument.
    //
	// MultiMethods  map[string]*RMultiMethod // map from method name to RMultiMethod object.
	//
	MultiMethods  map[string]*RMultiMethod // map from method name to RMultiMethod object.


    // This keeps track of all multimethods which include the abstract method and implementing
    // methods of a trait method. The multimethods in this map are the same ones that are
    // owned by packages. 
    // Whenever a new non-trait package is loaded, its (public) non trait-abstract methods need to be checked
    // against these traitMultimethods, and if name and type-compatible, added to the trait-multimethod
    // here. A multi-method here needs to be cache-cleared if it is added to.
    //
	TraitMultiMethods map[string][]*RMultiMethod
}



func (rt *RuntimeEnv) AddTraitMultiMethod(mm *RMultiMethod) {
    globalMethodName := fmt.Sprintf("%s___%d",mm.Name, mm.NumReturnArgs)
    rt.TraitMultiMethods[globalMethodName] = append(rt.TraitMultiMethods[globalMethodName], mm)
}


var RT *RuntimeEnv

func init() {
	RT = NewRuntimeEnv()
	RT.createPrimitiveTypes()
}


func (rt *RuntimeEnv) DebugAttributesMemory() {
   /*
   fmt.Println("--------attributes---------- valMap lengths ---------")
	for attr,valMap := range rt.attributes {
	   n := len(valMap)
	   if n > 0 {
	      fmt.Printf("%8d %s\n",len(valMap),attr.ShortName())
      }
   }
   */
}


func (rt *RuntimeEnv) DB() DB {
	return rt.db
}

func (rt *RuntimeEnv) DBT() DBT {
	return rt.dbt
}

/*
This and the PopDBT method must be used one after the other without fail in a guaranteed serialized exclusively running section of code.
*/
func (rt *RuntimeEnv) PushDBT(temporaryDbt DBT) (dbt DBT) {
   dbt = rt.dbt
   rt.dbt = temporaryDbt 
   return 
}

func (rt *RuntimeEnv) PopDBT(originalDbt DBT) {
   rt.dbt = originalDbt
}






/*
Creates a new constant.
If 
*/
func (rt *RuntimeEnv) CreateConstant(name string, value RObject, privateConstPackage *RPackage) (err error) {
	if _, found := rt.constants[name]; found {
		err = fmt.Errorf("Redefining constant '%s'", name)
		return 
	}
	rt.constants[name] = value

    if privateConstPackage != nil {
    	rt.privateConstantPackage[name] = privateConstPackage
    }

	return
}

func (rt *RuntimeEnv) GetConstant(name string, fromPackage *RPackage) (val RObject, found bool, hidden bool) {
	val, found = rt.constants[name]
	// TODO The following should have been checked at compile time.
	constPackage, constantIsPrivate := rt.privateConstantPackage[name]
	if constantIsPrivate {
       hidden = (constPackage != fromPackage)
	}

	return
}

/*
   Caches the object in the runtime's in-memory object cache. Assumes an object instance with the same
   dbid is not already in the cache.
   Assumes the object already has a dbid i.e. has been persisted locally.
*/
func (rt *RuntimeEnv) Cache(obj RObject) {
	rt.objects[obj.DBID()] = obj
}

/*
   Return the object with the given dbid from the in-memory cache, or nil if not found.
*/
func (rt *RuntimeEnv) GetObject(id int64) (obj RObject, found bool) {
	obj, found = rt.objects[id]
	return
}

func (rt *RuntimeEnv) Uncache(obj RObject) {
	delete(rt.objects, obj.DBID()) 
}

type MemCache interface {
	/*
	   Caches the object in the in-memory object cache. Assumes an object instance with the same
	   dbid is not already in the cache.
	   Assumes the object already has a dbid i.e. has been persisted locally.
	*/	
	Cache(obj RObject)
	
    GetObject(id int64) (obj RObject, found bool) 
	
    Uncache(obj RObject)	
}

/*
   A generator of unsigned 64-bit integer ids starting at 1 and incrementing by 1 each time a new one
   is requested. Call the NextID() method to get a runtime-unique id.
*/
type IdGenerator struct {
	ch chan uint64
}

func NewIdGenerator() *IdGenerator {
	gen := &IdGenerator{ch: make(chan uint64)}
	go gen.sendID()
	return gen
}

/*
   Blocks until a next id is available.
*/
func (gen *IdGenerator) NextID() uint64 {
	id := <-gen.ch
	return id
}

func (gen *IdGenerator) sendID() {
	var id uint64
	for id = 1; ; id++ {
		gen.ch <- id
	}
}

/*
Still needs SetDB(db) before is valid.
*/
func NewRuntimeEnv() *RuntimeEnv {
	rt := &RuntimeEnv{Types: make(map[string]*RType),
		Typs: make(map[string]*RType),
		MultiMethods:  make(map[string]*RMultiMethod),
		Packages:      make(map[string]*RPackage),
		Pkgs:          make(map[string]*RPackage),
		PkgShortNameToName: make(map[string]string),   
		PkgNameToShortName: make(map[string]string),		
		TypeTupleTree: &TypeTupleTreeNode{},
		TypeTupleTrees: make([]*TypeTupleTreeNode,100),
		objects:       make(map[int64]RObject),
		objectIds:     make(map[RObject]uint64),
		idGen:         NewIdGenerator(),
		// attributes:    make(map[*AttributeSpec]map[RObject]RObject),
		context:       make(map[string]RObject),
		constants:     make(map[string]RObject),
		privateConstantPackage:     make(map[string]*RPackage),		
		inTransit:     make(map[RObject]uint32),
	    ReflectedDataTypes: make(map[string]RObject),
	    ReflectedAttributes: make(map[string]RObject),
		evalContexts:  make(map[RObject]MethodEvaluationContext),
		TraitMultiMethods: make(map[string][]*RMultiMethod),
	}
	for i := 0; i < 20; i++ {
	   tttn := &TypeTupleTreeNode{}
	   tttn.LastTypeTuple = make(map[*RType]*RTypeTuple)
	   rt.TypeTupleTrees[i] = tttn
   }
	rt.PkgNameToShortName["relish.pl2012/core/inbuilt"] = "inbuilt"
	rt.PkgShortNameToName["inbuilt"] = "relish.pl2012/core/inbuilt"	
	return rt
}

/*
   Return the next available identifier for a not-yet-persistable object which must be
   identified.
*/
func (rt *RuntimeEnv) NextLocalID() uint64 {
	return rt.idGen.NextID()
}

func (rt *RuntimeEnv) SetDB(db DB) {
	rt.db = db

	rt.dbt = db.DefaultDBThread()
}





/*
   Return the value of the specified attribute for the specified object.
   Does not currently distinguish between multi-value attributes and single-valued.
   If it is a multi-valued attribute, returns the RCollection which implements
   the multi-value attribute.
   What does it return if no value has been defined? How about a found boolean
*/
func (rt *RuntimeEnv) AttrVal(th InterpreterThread, obj RObject, attr *AttributeSpec) (val RObject, found bool) {
	return rt.AttrValue(th, obj, attr, true, true, true)
}

var txOpsMutex sync.Mutex

/*
Tries to make sure that if we are in a thread other than one participating in the transaction,
the attribute value we will get will wait til a transaction that the
object is dirty in commits or rolls back.
Also makes sure that if the object state is invalid, due to a rolled back transaction it was dirty in,
that the object state is first restored from database before getting the attribute value.
*/
func ensureMemoryTransactionConsistency1(th InterpreterThread, unit *runit) {
	defer Un(Trace(PERSIST_TR,"ensureMemoryTransactionConsistency1"))
	th.AllowGC()
	// fmt.Println("txOpsMutex.Lock() ing etc1")
    txOpsMutex.Lock()
	// fmt.Println("txOpsMutex.Lock() ed etc1")    
    th.DisallowGC()
    defer txOpsMutex.Unlock()


//   fmt.Println("unit.transaction=",unit.transaction)
   if unit.transaction == RolledBackTransaction || unit.IsLoadNeeded() {  // The object state has been rolled back.
//       fmt.Println("Yeah! Refreshing")
       err := unit.Refresh(th)    // TODO Should we really refresh if not supposed to check persistence??
                         // What does refresh mean if the object was only persisted in the db
                         // in the rolled back transaction. Need to abort the refresh before updating
                         // the attribute value.
                         // NB. Should set the unit's transaction to nil
       if err != nil {
       	  panic(err)
       }
    }

    if unit.transaction != nil {  // Is definitely stored locally. TODO Note the race condition here.

    	if th != nil && th.Transaction() != unit.transaction {
           tx := unit.transaction	
           txOpsMutex.Unlock()
	       // fmt.Println("txOpsMutex.Unlock() etc1 diff xactions")

           th.AllowGC()
	       // fmt.Println("tx.RLock() ing etc1")           
           tx.RLock()  // Block til the transaction that the object is dirty in commits or rolls back.
	       // fmt.Println("tx.Lock() ed etc1")            
           th.DisallowGC()
           tx.RUnlock()
	       // fmt.Println("tx.Unlock() etc1")            
           th.AllowGC()
	       // fmt.Println("txOpsMutex.Lock() ing etc1 #2")           
           txOpsMutex.Lock()
	       // fmt.Println("txOpsMutex.Lock() ed etc1 #2")            
           th.DisallowGC()
        }
        if unit.transaction == RolledBackTransaction || unit.IsLoadNeeded() {  // The object state has been rolled back.

	       err := unit.Refresh(th)    // TODO Should we really refresh if not supposed to check persistence??
	                         // What does refresh mean if the object was only persisted in the db
	                         // in the rolled back transaction. Need to abort the refresh before updating
	                         // the attribute value.
	                         // NB. Should set the unit's transaction to nil
	       if err != nil {
	       	  panic(err)
	       }
        }
    } 

	// fmt.Println("txOpsMutex.Unlock() etc1 end")    
}

func ensureMemoryTransactionConsistency2(th InterpreterThread, unit *runit) (err error) {
	defer Un(Trace(PERSIST_TR,"ensureMemoryTransactionConsistency2"))	
    th.AllowGC()	
    txOpsMutex.Lock()
    th.DisallowGC()    
    defer txOpsMutex.Unlock()
    if unit.transaction == RolledBackTransaction || unit.IsLoadNeeded() {
// LOCK    	
    	err = unit.Refresh(th)  // Should set the unit's transaction to nil
    	if err != nil {
    		return
    	}
// UNLOCK    	
    }
    if th.Transaction() != nil {

// LOCK

    	if unit.transaction == nil {  // Marking this object as dirty in the thread's transaction.
    		unit.transaction = th.Transaction()

     		unit.SetTransaction(th.Transaction())
   	

    	} else if unit.transaction != th.Transaction() {
            err = errors.New("object transaction is different than goroutine's transaction.")
            return
    	}
// UNLOCK

    } else if unit.transaction != nil {  // This is a thread that is not participating in the transaction.
        tx := unit.transaction
        txOpsMutex.Unlock() 
        th.AllowGC()               
        tx.RLock()   // Wait for the transaction to commit or rollback.
        th.DisallowGC()
        tx.RUnlock()
        th.AllowGC()         
        txOpsMutex.Lock()
        th.DisallowGC()        
        if unit.transaction == RolledBackTransaction || unit.IsLoadNeeded() {  // The object state has been rolled back.

	       err := unit.Refresh(th)    // TODO Should we really refresh if not supposed to check persistence??
	                         // What does refresh mean if the object was only persisted in the db
	                         // in the rolled back transaction. Need to abort the refresh before updating
	                         // the attribute value.
	       if err != nil {
	       	  return err
	       }
        }
    }
    return
}


/*
Tries to make sure that if we are in a thread other than one participating in the transaction,
the attribute value we will get will wait til a transaction that the
object is dirty in commits or rolls back.
Also makes sure that if the object state is invalid, due to a rolled back transaction it was dirty in,
that the object state is first restored from database before getting the attribute value.
*/
func ensureMemoryTransactionConsistency3(th InterpreterThread, coll RCollection) {
	defer Un(Trace(PERSIST_TR,"ensureMemoryTransactionConsistency3"))		
    th.AllowGC()   	
    txOpsMutex.Lock()
    th.DisallowGC()    
    defer txOpsMutex.Unlock()

   if coll.Transaction() == RolledBackTransaction || coll.IsLoadNeeded() {  // The object state has been rolled back.

       err := coll.(Persistable).Refresh(th)    // TODO Should we really refresh if not supposed to check persistence??
                         // What does refresh mean if the object was only persisted in the db
                         // in the rolled back transaction. Need to abort the refresh before updating
                         // the attribute value.
                         // NB. Should set the unit's transaction to nil
       if err != nil {
       	  panic(err)
       }
    }

    if coll.Transaction() != nil {  // Is definitely stored locally. TODO Note the race condition here.

    	if th != nil && th.Transaction() != coll.Transaction() {
           tx := coll.Transaction()	
           txOpsMutex.Unlock()
           th.AllowGC()            
           tx.RLock()  // Block til the transaction that the object is dirty in commits or rolls back.
           th.DisallowGC()            
           tx.RUnlock()
           th.AllowGC()             
           txOpsMutex.Lock()
           th.DisallowGC()             
        }
        if coll.Transaction() == RolledBackTransaction || coll.IsLoadNeeded() {  // The object state has been rolled back.

	       err := coll.(Persistable).Refresh(th)    // TODO Should we really refresh if not supposed to check persistence??
	                         // What does refresh mean if the object was only persisted in the db
	                         // in the rolled back transaction. Need to abort the refresh before updating
	                         // the attribute value.
	                         // NB. Should set the unit's transaction to nil
	       if err != nil {
	       	  panic(err)
	       }
        }
    } 
}

func ensureMemoryTransactionConsistency4(th InterpreterThread, coll RCollection) (err error) {
	defer Un(Trace(PERSIST_TR,"ensureMemoryTransactionConsistency4"))		
    th.AllowGC()  	
    txOpsMutex.Lock()
    th.DisallowGC()        
    defer txOpsMutex.Unlock()
    if coll.Transaction() == RolledBackTransaction || coll.IsLoadNeeded() {   	
    	err = coll.(Persistable).Refresh(th)  // Should set the unit's transaction to nil
    	if err != nil {
    		return
    	}   	
    }
    if th.Transaction() != nil {
    	if coll.Transaction() == nil {  // Marking this object as dirty in the thread's transaction.
    		coll.SetTransaction(th.Transaction())
    	} else if coll.Transaction() != th.Transaction() {
            err = errors.New("object transaction is different than goroutine's transaction.")
            return
    	}
    } else if coll.Transaction() != nil {  // This is a thread that is not participating in the transaction.
        tx := coll.Transaction()
        txOpsMutex.Unlock()  
        th.AllowGC()              
        tx.RLock()   // Wait for the transaction to commit or rollback.
        th.DisallowGC()         
        tx.RUnlock()
        th.AllowGC()           
        txOpsMutex.Lock()
        th.DisallowGC()          
        if coll.Transaction() == RolledBackTransaction || coll.IsLoadNeeded() {  // The object state has been rolled back.

	       err := coll.(Persistable).Refresh(th)    // TODO Should we really refresh if not supposed to check persistence??
	                         // What does refresh mean if the object was only persisted in the db
	                         // in the rolled back transaction. Need to abort the refresh before updating
	                         // the attribute value.
	       if err != nil {
	       	  return err
	       }
        }
    }
    return
}









/*
   Return the value of the specified attribute for the specified object.
   Does not currently distinguish between multi-value attributes and single-valued.
   If it is a multi-valued attribute, returns the RCollection which implements
   the multi-value attribute.
   What does it return if no value has been defined? val = nil (Go nil) and found=false.
*/
func (rt *RuntimeEnv) AttrValue(th InterpreterThread, obj RObject, attr *AttributeSpec, checkPersistence bool, allowNoValue bool, lock bool) (val RObject, found bool) {
	

    t := obj.Type()

    if ! attr.PublicReadable {
    	if t.Package != th.Package() {
	       panic(fmt.Sprintf("Attribute %s.%s is private; not readable outside of the package in which the attribute is declared.", obj, attr.Part.Name))		    		
    	}
    }



    i := attr.Index[t]
    var unit *runit
    var isUnit bool
    unit,isUnit = obj.(*runit)
    
    if ! isUnit {
        unit = &(obj.(*GoWrapper).runit)    	
    }

    if obj.IsBeingStored() {
       // fmt.Println("AttrValue - ensureMemoryTransactionConsistency1") 
       ensureMemoryTransactionConsistency1(th, unit)
       // fmt.Println("done AttrValue - ensureMemoryTransactionConsistency1") 
    }

    val = unit.attrs[i]

    if val != nil {
    	found = true
    	return
    }

    if (! checkPersistence) && (! allowNoValue) {
	   panic(fmt.Sprintf("Attribute %s.%s has no value.", obj, attr.Part.Name))			
	}    	

	//Logln(PERSIST_,"AttrVal ! found in mem and strdlocally=",obj.IsStoredLocally())
	//Logln(PERSIST_,"AttrVal ! found in mem and attr.Part.CollectionType=",attr.Part.CollectionType)
	//Logln(PERSIST_,"AttrVal ! found in mem and attr.Part.Type.IsPrimitive=",attr.Part.Type.IsPrimitive)
	if checkPersistence && obj.IsStoredLocally() && (attr.Part.CollectionType != "" || !attr.Part.Type.IsPrimitive) {
		var err error

		val, err = th.DBT().FetchAttribute(th, obj.DBID(), obj, attr, 0)		
	        //Logln(PERSIST2_, "Finished FetchAttribute!")
	        //Logln(PERSIST2_, "Finished FetchAttribute:", err)
	        //if val != nil { Logln(PERSIST2_, val.String()) }
                
		// val, err = rt.db.FetchAttribute(th, obj.DBID(), obj, attr, 0)
		if err != nil {
			// TODO  - NOT BEING PRINCIPLED ABOUT WHAT TO DO IF NO VALUE! Should sometimes allow, sometimes not!
			
            if strings.Contains(err.Error(), "has no value for attribute") {
               if ! allowNoValue {
          	      panic(fmt.Sprintf("Error fetching attribute %s.%s from database: %s", obj, attr.Part.Name, err))     	
               }
		    } else {
		       panic(fmt.Sprintf("Error fetching attribute %s.%s from database: %s", obj, attr.Part.Name, err))
			}
		}
        if val != nil {
                        // THE LOGGING STATEMENT IS NOT THREAD-SAFE!!! 
			Logln(ALWAYS_, "Fetched Attribute", attr.ShortName())

			Logln(PERSIST2_, "AttrVal (fetched) =", val.StringTh(th))
            found = true
		} 
	}

	if val == nil && attr.Part.ArityHigh != 1 && attr.Part.CollectionType != ""  {
		var err error
	   	val, err = rt.EnsureMultiValuedAttributeCollection(obj, attr)
		if err != nil {
          	panic(fmt.Sprintf("Error ensure multivalued attribute collection %s.%s: %s", obj, attr.Part.Name, err))  
		}   
		found = true          	
	}

	return
}


/*
Version to be used in template execution.

Note: Now checks relations as well as one-way attributes.
*/
func (rt *RuntimeEnv) AttrValByName(th InterpreterThread, obj RObject, attrName string) (val RObject, err error) {

	attr, found := obj.Type().GetAttribute(attrName)
	if ! found {
       err = fmt.Errorf("Attribute or relation '%s' not found in type '%v' or supertypes.", attrName, obj.Type())		
	   return	
	}	
	val, _ = RT.AttrVal(th, obj, attr)
	return
}




/*
Untypechecked assignment. Used in restoration (summoning) of an object from persistent storage.
*/
func (rt *RuntimeEnv) RestoreAttr(obj RObject,  attr *AttributeSpec, val RObject) {
	
	defer Un(Trace(PERSIST_TR2, "RestoreAttr", obj, attr, val))

    t := obj.Type()
    i := attr.Index[t]

    var unit *runit
    var isUnit bool
    unit,isUnit = obj.(*runit)
    
    if ! isUnit {
        unit = &(obj.(*GoWrapper).runit)    	
    }

    unit.attrs[i] = val
}

/*
Untypechecked assignment. Used in restoration (summoning) of an object from persistent storage.
Not mutex locked.
*/
func (rt *RuntimeEnv) RestoreAttrNonLocking(obj RObject,  attr *AttributeSpec, val RObject) {
	
	defer Un(Trace(PERSIST_TR2, "RestoreAttrNonLocking", obj, attr, val))
	
    t := obj.Type()
    i := attr.Index[t]
    var unit *runit
    var isUnit bool
    unit,isUnit = obj.(*runit)
    
    if ! isUnit {
        unit = &(obj.(*GoWrapper).runit)    	
    }
    unit.attrs[i] = val
}


/*
Optionally typechecked assignment. Never used in inverse.

!!!!!!!!!!!!!!!!!
TODO !!!!!!!!!!!!
!!!!!!!!!!!!!!!!!
AttributeSpec should have its own .attributes and also its own Mutex / RWMutex
So attr setting/getting is only one hashtable lookup not two, and
mutex locking has less scope (only applies to the particular attributespec.)

*/
func (rt *RuntimeEnv) SetAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, typeCheck bool, context MethodEvaluationContext, isInverse bool) (err error) {

    if obj == NIL {
		err = fmt.Errorf("nil cannot have a value assigned to '%v' attribute.", attr.Part.Name)
		return		
    }

    if val == NIL {
       err = rt.UnsetAttr(th, obj, attr, isInverse, true)
       return
    }

    if val == nil {
		err = fmt.Errorf("Attempt to assign value of uninitialized variable to '%v' attribute.", attr.Part.Name)
		return    	
    }

	if typeCheck {

       // No. This should have been deproxified already!!
       //if val.IsProxy() {
       //
       //}

        // This is a kludge
        if attr.Part.CollectionType != "" { // "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap" ""
           
           valType := val.Type()

           if valType.ElementType() != attr.Part.Type {
 		      err = fmt.Errorf("Cannot assign  '%v.%v %s of %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.CollectionType, attr.Part.Type, val.Type())
		      return          	
           } 
           
            if strings.HasPrefix(valType.Name,"List_of_") {
               if attr.Part.CollectionType != "list" {
 		          err = fmt.Errorf("Cannot assign  '%v.%v %s of %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.CollectionType, attr.Part.Type, val.Type())
		          return  
               }
           	} else if strings.HasPrefix(valType.Name,"Set_of_") {
                if attr.Part.CollectionType != "set" {
  		           err = fmt.Errorf("Cannot assign  '%v.%v %s of %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.CollectionType, attr.Part.Type, val.Type())
		           return                	
                }
           	} else if strings.HasPrefix(valType.Name,"Map_of_") {
                if attr.Part.CollectionType != "map" {
  		           err = fmt.Errorf("Cannot assign  '%v.%v %s of %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.CollectionType, attr.Part.Type, val.Type())
		           return                	
                }           	
           	} else {
 		      err = fmt.Errorf("Cannot assign  '%v.%v %s of %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.CollectionType, attr.Part.Type, val.Type())
		      return  
           	}
       	} else if !val.Type().LessEq(attr.Part.Type) {
		   err = fmt.Errorf("Cannot assign  '%v.%v %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.Type, val.Type())
		   // debug.PrintStack()
		   return
	   }
	}
	

    t := obj.Type()

    if ! attr.PublicWriteable {
    	if t.Package != th.Package() {
    	   fmt.Println(t.Package)
    	   fmt.Println(th.Package)
	       err = fmt.Errorf("Attribute %s.%s is private; not settable outside of the package in which the attribute is declared.", obj, attr.Part.Name)
	       return	    		
    	}
    }


    i := attr.Index[t]

    var unit *runit
    var isUnit bool
    unit,isUnit = obj.(*runit)
    
    if ! isUnit {
        unit = &(obj.(*GoWrapper).runit)    	
    }    

    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if obj.IsBeingStored() {
       ensureMemoryTransactionConsistency2(th, unit)    	
    }


    oldVal := unit.attrs[i]
    found := (oldVal != nil)

    // Also have to do this in UnsetAttr !!!
	if ! found {
		if obj.IsStoredLocally() && attr.IsRelation() {
			oldVal, found = rt.AttrValue(th, obj, attr,true,true, true)
		}
	}

   unit.attrs[i] = val


	if obj.IsBeingStored() {
		th.DBT().PersistSetAttr(th, obj, attr, val, found)
	}

	if ! isInverse && attr.Inverse != nil {
	   if oldVal != nil && oldVal != NIL {
		   // Remove (one entry for) obj from oldVal's inverse attr. 
		   err = rt.RemoveAttrGeneral(th,oldVal,attr.Inverse,obj,true,true)
		   if err != nil {
			  return
		   }
	   }	
	   if val != NIL {
	   	  err = rt.SetOrAddToAttr(th, val, attr.Inverse, obj, context, true)
	   }
	}	
	return
}

// attributes map[*AttributeSpec] map[RObject] RObject

/*
Typechecked adding a member to a multi-valued attribute.
TODO TODO TODO

Create the collection on demand.

context is a context for evaluating a sorting comparison operator on the collection.

*/
func (rt *RuntimeEnv) AddToAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, typeCheck bool, context MethodEvaluationContext, isInverse bool) (err error) {

	if typeCheck && !val.Type().LessEq(attr.Part.Type) {
		err = fmt.Errorf("Cannot assign  '%v.%v %v' a value of type '%v'.", obj.Type(), attr.Part.Name, attr.Part.Type, val.Type())
		return
	}
	
	if obj == NIL {
		err = fmt.Errorf("nil cannot have a value added to '%v' attribute.", attr.Part.Name)
		return		
	}
	

    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if obj.IsBeingStored() {
       unit := obj.(*runit)    	
       ensureMemoryTransactionConsistency2(th, unit)    	
    }	

	// Note: Need to put in a check here as to whether the collection accepts NIL elements, and
	// if val == NIL, reject the addition.	

    // !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!1
    // NOTE 2014 07 24 ! This does another ensureMemoryTransactionConsistency1 inside AttrVal! !!!
    // Two transaction state checks? Why? Inefficient?
    //
    objColl, collectionFound := rt.AttrVal(th, obj, attr) 
    if ! collectionFound {
    	panic("There was supposed to be a collection value ensured as the value of the multi-valued attribute.")
	}



	// objColl, err := rt.EnsureMultiValuedAttributeCollection(obj, attr)
	// if err != nil {
	//	return
	// }
	// NOTE THE rt.Ensure... ABOVE CREATES A COLLECTION WHETHER THE ATTRIBUTE WAS collectionType = "" or not
    // BUT IT DOES NOT RETRIEVE THE ATTRIBUTE FROM PERSISTENCE IF NOT FETCHED YET.!!!!!!
    // And we need to do that.



	addColl := objColl.(AddableMixin)     // Will throw an exception if collection type does not implement Add(..)
	added, newLen := addColl.Add(val, context) // returns false if is a set and val is already a member.

	// fmt.Println("AddToAttr (added, newLen)",added,newLen)

	/* TODO figure out efficient persistence of collection updates
	 */
	//fmt.Printf("added=%v\n",added)
	//fmt.Printf("IsStoredLocally=%v\n",obj.IsStoredLocally())

	if added {
	    if obj.IsBeingStored() {
			var insertIndex int
			if objColl.(RCollection).IsSorting() {
				orderedColl := objColl.(OrderedCollection)
				insertIndex = orderedColl.Index(val, 0)
			} else {
				insertIndex = newLen - 1
			}
			th.DBT().PersistAddToAttr(th, obj, attr, val, insertIndex)
		}
		if ! isInverse && attr.Inverse != nil {
			err = rt.SetOrAddToAttr(th, val, attr.Inverse, obj, context, true)
		}
	}

	return
}

/*
   Used to make val the value or a value of the attribute of obj. 
   If the attribute is multi-valued, val will be added to the collection of values.
   If the attribute is single-valued, val will be set as the value of the attribute. 
*/
func (rt *RuntimeEnv) SetOrAddToAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, context MethodEvaluationContext, isInverse bool) (err error) {
	
   if attr.Part.CollectionType == "" {	
      err = rt.SetAttr(th, obj, attr, val, false, context, isInverse) 
   } else {	
      err = rt.AddToAttr(th, obj, attr, val, false, context, isInverse)
   }	
   return
}



func (rt *RuntimeEnv) AddToCollection(coll AddableCollection, val RObject, typeCheck bool, context MethodEvaluationContext) (err error) {

	if typeCheck && !val.Type().LessEq(coll.ElementType()) {
		err = fmt.Errorf("Cannot add a value of type '%v' to a collection with element-type constraint '%v'.", val.Type(),coll.ElementType())	
		return
	}
	
    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if coll.IsBeingStored() {
       ensureMemoryTransactionConsistency4(context.InterpThread(), coll)    	
    }	

	// Note: Need to put in a check here as to whether the collection accepts NIL elements, and
	// if val == NIL, reject the addition.	

	added, newLen := coll.Add(val, context) // returns false if is a set and val is already a member.

	/* TODO figure out efficient persistence of collection updates
	 */
	//fmt.Printf("added=%v\n",added)
	//fmt.Printf("IsStoredLocally=%v\n",obj.IsStoredLocally())

	if added {
	    if coll.IsBeingStored() {
			var insertIndex int
			if coll.IsSorting() {
				orderedColl := coll.(OrderedCollection)
				insertIndex = orderedColl.Index(val, 0)
			} else {
				insertIndex = newLen - 1
			}
			err = context.InterpThread().DBT().PersistAddToCollection(context.InterpThread(),coll, val, insertIndex)
		}
	}

	return
}

/*
Removes val from the multi-valued attribute if val is in the collection. Does nothing and does not complain if val is not in the collection.
If removePersistent is true, also removes the value from the persistent version of the attribute association.
*/
func (rt *RuntimeEnv) RemoveFromCollection(th InterpreterThread, collection RemovableCollection, val RObject, removePersistent bool) (err error) { 	
	
    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if collection.IsBeingStored() {  	
       ensureMemoryTransactionConsistency4(th, collection)    	
    }	

	removed, removedIndex := collection.Remove(th, val)
	
	if removed  {
	   if removePersistent && collection.IsBeingStored() {
	    	th.DBT().PersistRemoveFromCollection(collection, val, removedIndex)
	   }
	}

	return
}





/*
 Remove all elements of the multivalued attribute, in memory and in the db.
 If the attribute has an inverse, also removes the inverse attribute values.
*/
func (rt *RuntimeEnv) ClearAttr(th InterpreterThread, obj RObject, attr *AttributeSpec) (err error) {

    objColl, foundCollection := rt.AttrVal(th, obj, attr)

 	if !foundCollection { // this object does not have the collection implementation of this multi-valued attribute	
                         // Must be already empty or unassigned?	
	  return
    }
		

    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if obj.IsBeingStored() {
       unit := obj.(*runit)   	
       ensureMemoryTransactionConsistency2(th, unit)    	
    }	


	if attr.IsRelation() {
	   inverseAttr := attr.Inverse

       collection := objColl.(RCollection)
	   for val := range collection.Iter(th) {
           rt.RemoveAttrGeneral(th, val, inverseAttr, obj, true, false)		
	   }
    }

	collection := objColl.(RemovableMixin) // Will throw an exception if collection type does not implement ClearInMemory()
	collection.ClearInMemory()	
	
	if obj.IsBeingStored() {
	   err = th.DBT().PersistClearAttr(obj, attr)
    }
	return
}


/*
 Remove all elements of the collection, in memory and in the db.
*/
func (rt *RuntimeEnv) ClearCollection(th InterpreterThread, collection RemovableCollection) (err error) {

    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if collection.IsBeingStored() {   	
       ensureMemoryTransactionConsistency4(th, collection)    	
    }	

	collection.ClearInMemory()	
	
	if collection.IsBeingStored() {
	   err = th.DBT().PersistClearCollection(collection)
    }
	return
}




/*
TODO Optimize this  to add all at once with a slice copy or similar, then persist in fewer
separate DB calls.
*/
func (rt *RuntimeEnv) ExtendCollection(coll AddableCollection, vals []RObject, typeCheck bool, context MethodEvaluationContext) (err error) {

    for _,val := range vals {
       err = rt.AddToCollection(coll, val, typeCheck, context)	
       if err != nil {
	      return
       }
    }
    return
}

/*
func (rt *RuntimeEnv) AddToCollectionTypeChecked(coll RCollection, val RObject, context MethodEvaluationContext) (err error) {

	if !val.Type().LessEq(coll.ElementType()) {
		err = fmt.Errorf("Cannot add a value of type '%v' to a collection with element-type constraint '%v'.", val.Type(),coll.ElementType())
		return
	}
	
	addColl := coll.(AddableMixin)     // Will throw an exception if collection type does not implement Add(..)
	
	// Re-enable when we implement persisting of independent collections.
	// added, newLen := addColl.Add(val, context) // returns false if is a set and val is already a member.

	addColl.Add(val, context) // returns false if is a set and val is already a member.	
*/
/*
Need to decide how to persist collections and check if persisted and handle persisting add

	TODO figure out efficient persistence of collection updates
	
	//fmt.Printf("added=%v\n",added)
	//fmt.Printf("IsStoredLocally=%v\n",obj.IsStoredLocally())

    // This part is COPYITIS from AddToAttrTypeChecked method.
	if added && obj.IsStoredLocally() {
		var insertIndex int
		if objColl.(RCollection).IsSorting() {
			orderedColl := objColl.(OrderedCollection)
			insertIndex = orderedColl.Index(val, 0)
		} else {
			insertIndex = newLen - 1
		}
		rt.db.PersistAddToAttr(obj, attr, val, insertIndex)
	}

	*/
/*
	return
}
*/

/*
TODO Optimize this  to add all at once with a slice copy or similar, then persist in fewer
separate DB calls.
*/
func (rt *RuntimeEnv) ExtendMapTypeChecked(theMap Map, keysVals []RObject, context MethodEvaluationContext) (err error) {

    n := len(keysVals)
    for i := 0; i < n; i+=2 {
       key := keysVals[i]
       val := keysVals[i+1]
       err = rt.PutInMapTypeChecked(theMap, key, val, context)	
       if err != nil {
	      return
       }
    }
    return
}

func (rt *RuntimeEnv) PutInMapTypeChecked(theMap Map, key RObject, val RObject, context MethodEvaluationContext) (err error) {

	if !key.Type().LessEq(theMap.KeyType()) {
		err = fmt.Errorf("Cannot use a key of type '%v' in a map with key-type constraint '%v'.", key.Type(),theMap.KeyType())
		return
	}
	
	if !val.Type().LessEq(theMap.ValType()) {
		err = fmt.Errorf("Cannot put a value of type '%v' in a map with value-type constraint '%v'.", val.Type(),theMap.ValType())
		return
	}	

    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if theMap.IsBeingStored() {   	
       ensureMemoryTransactionConsistency4(context.InterpThread(), theMap)    	
    }	

    isNewKey,_ := theMap.Put(key, val, context)
	
	if theMap.IsBeingStored() {
	   err = context.InterpThread().DBT().PersistMapPut(context.InterpThread(), theMap, key, val, isNewKey)  
    }
	return
}



/*
Helper method. Ensures that a collection exists in memory to manage the values of a multi-valued attribute of an object.
Assumes the attribute is multi-valued or collection-valued.
*/
func (rt *RuntimeEnv) EnsureMultiValuedAttributeCollection(obj RObject, attr *AttributeSpec) (collection RCollection, err error) {

    //defer Un(Trace(ALWAYS_,"EnsureMultiValuedAttributeCollection", attr.Part.Name))


    t := obj.Type()
    i := attr.Index[t]
    var unit *runit
    var isUnit bool
    unit,isUnit = obj.(*runit)
    
    if ! isUnit {
        unit = &(obj.(*GoWrapper).runit)    	
    }


    val := unit.attrs[i]

    if val != nil {
	   collection = val.(RCollection)
       return
    }

	var owner RObject
	var attribute *AttributeSpec
	var minCardinality, maxCardinality int64
	if attr.Part.ArityHigh == 1 { // This is a collection-valued attribute of arity 1. (1 collection)
		minCardinality = 0
		maxCardinality = -1 // largest possible collection is allowed - the attribute is not constraining it.	
		// panic("Should not rt.EnsureMultiValuedAttributeCollection on a collection-valued attribute.")
		// We will allow this. but we should not allow persisting of such collections really.
		// This all has to be checked.
	} else { // This is a multi-valued attribute. The collection is a hidden implementation detail. 
		minCardinality = int64(attr.Part.ArityLow)
		maxCardinality = int64(attr.Part.ArityHigh)
		owner = obj //  	Collection is owned by the "whole" object.
		attribute = attr

	}
	// Create the list or set collection

	var unaryMethod *RMultiMethod
	var binaryMethod *RMultiMethod
	var orderAttr *AttributeSpec
	var isAscending bool

	// fmt.Println(attr.Part.CollectionType)

	if attr.Part.CollectionType == "sortedlist" || attr.Part.CollectionType == "sortedset" || attr.Part.CollectionType == "sortedmap" || attr.Part.CollectionType == "sortedstringmap" {
		if attr.Part.OrderMethodArity == 1 {
			unaryMethod = attr.Part.OrderMethod
		} else if attr.Part.OrderMethodArity == 2 {
			binaryMethod = attr.Part.OrderMethod
		} else { // must be an attribute
			orderAttr = attr.Part.OrderAttr
		}			
		isAscending = attr.Part.IsAscending	
/*
		if binaryMethod == nil {
			binaryMethod, _ = rt.InbuiltFunctionsPackage.MultiMethods["lt"]
		}

		sortWith = &sortOp{
			attr:          orderAttr,
			unaryFunction: unaryMethod,
			lessFunction:  binaryMethod,
			descending:    !attr.Part.IsAscending,
		}
*/			
	}

	/*
		@@@@@@@@@@@@@@@@@@@@@@@

		NEED TO FIND OUT IF THERE IS A UNARY METHOD OF THE TYP2, otherwise it must be a binary method.

		Only one of the attr or unaryFunction will be non-nil.
		If attr or unaryFunction is non-nil, then lessFunction must be the "lt" multiMethod.

		collection.sortWith.lessFunction,_ := RT.MultiMethods["lt"]

		If attr and unaryFunction are nil, lessFunction may be any binary boolean function which has a method whose
		parameter signature is compatible with a pair of values of the elementType of the collection. lessFunction MAY
		be the "lt" function in this case but need not be. The function is treated as a "less-than" predicate.

		type sortOp struct {
			attr *AttributeSpec
			unaryFunction *RMultiMethod
			lessFunction *RMultiMethod
			descending bool
		}

		$$$$$$$$$$$$$$$$$$$$$$$$


		  attr = &AttributeSpec{typ1,
		                        RelEnd{
		 									    Name:endName2,
		                                       Type:typ2,
		                                       ArityLow:arityLow2,
		                                       ArityHigh:arityHigh2,
		                                       CollectionType:collectionType2,
		                                       OrderAttr:orderAttrName,
		                                       OrderMethod: orderMethod,
		 									   OrderMethodArity: int32,
		                                       IsAscending:isAscending,
		                                      },

		@@@@@@@@@@@@@@
	*/


/////////


   collection,err = rt.NewCollection(minCardinality,
     maxCardinality,
     owner,   
     attribute,
     attr.Part.CollectionType, 
     isAscending,
     unaryMethod,
     binaryMethod,
     orderAttr,  
     nil, // keyType *RType, 
     attr.Part.Type) 

	unit.attrs[i] = collection

   return
}










/*
Helper method. Ensures that a collection exists in memory (and in db if obj is persistent) if someone
wants to do a += to a collection-valued attribute.
Ensure the collection is assigned as value of the attribute.
*/
func (rt *RuntimeEnv) EnsureCollectionAttributeVal(th InterpreterThread, obj RObject, attr *AttributeSpec) (collection RCollection, err error) {

   // A single-valued attribute whose value is a collection
	attrVal, found := RT.AttrVal(th, obj, attr)
	var isCollection bool
	if found {
		collection, isCollection = attrVal.(RCollection)
        if ! isCollection {
        	err = fmt.Errorf("Value of %s.%s is not a collection but must be.", obj, attr)
        }
		return
	}

	var minCardinality int64 = 0
	var maxCardinality int64 = -1 // largest possible collection is allowed - the attribute is not constraining it.	


	// Create the list or set collection

	var unaryMethod *RMultiMethod
	var binaryMethod *RMultiMethod
	var orderAttr *AttributeSpec
	var isAscending bool

	// fmt.Println(attr.Part.CollectionType)

    collectionType := attr.Part.Type.CollectionImplementationType()
	if collectionType == "sortedlist" || collectionType == "sortedset" || collectionType == "sortedmap" || collectionType == "sortedstringmap" {
		if attr.Part.OrderMethodArity == 1 {
			unaryMethod = attr.Part.OrderMethod
		} else if attr.Part.OrderMethodArity == 2 {
			binaryMethod = attr.Part.OrderMethod
		} else { // must be an attribute
			orderAttr = attr.Part.OrderAttr
		}			
		isAscending = attr.Part.IsAscending			
	}


    collection,err = rt.NewCollection(minCardinality,
         maxCardinality,
         nil, // owner,   
         nil, // attribute,
         collectionType, 
         isAscending,
         unaryMethod,
         binaryMethod,
         orderAttr,  
         nil, // keyType *RType, 
         attr.Part.Type.ElementType())  
    if err != nil {
    	return
    }
	

    err = RT.SetAttr(th, obj, attr, collection, true, th.EvaluationContext(), false)

    return
}




/*
Creates a new RCollection object of the appropriate type, for purposes of restoring a collection
from persistent storage.
The typeDescriptor is from the type field of the collection's instance entry in the RObject table in the
local database. 
Will have to get fancier here about sorting specifications, by enhancing the type descriptor info.
This method is only for independent collections. Not for multivalued attribute collections.

A collection type descriptor is something like:

"["[ordered_][map|stringmap|list|set]"_of_"<someshorttypename>"]"
*/
func (rt *RuntimeEnv) NewCollectionFromDB(collectionTypeDescriptor string) (collection RCollection, err error) {
   // , keyType *RType, elementType *RType

   // Extract the collectionType part and the key type and element type from the collection type descriptor
   
   var keyTypeShortName string
   var elementTypeShortName string
   
   ofPos := strings.Index(collectionTypeDescriptor, "_of_")
   collectionType := collectionTypeDescriptor[1:ofPos]
   switch collectionType {   
   case "map","stringmap","sortedmap","sortedstringmap","int64map","uint64map":
      keyStartPos := ofPos + 5
      keyEndPos := strings.Index(collectionTypeDescriptor, ")=>(")
      elementStartPos := keyEndPos + 4
      elementEndPos := len(collectionTypeDescriptor)-2
      keyTypeShortName = collectionTypeDescriptor[keyStartPos:keyEndPos]  // the key type   
      elementTypeShortName = collectionTypeDescriptor[elementStartPos:elementEndPos]  // the element type         
   default:
      elementTypeShortName = collectionTypeDescriptor[ofPos+4:len(collectionTypeDescriptor)-1]  // the element type   
   }
   // load the key type and element type if not in the runtime yet.
   
   var keyType *RType
   
   if keyTypeShortName != "" {
      keyType = rt.Typs[keyTypeShortName]
      if keyType == nil {

         pkgShortName := PackageShortName(keyTypeShortName)  
   //      localTypeName := LocalTypeName(typeName)   
         pkgFullName := RT.PkgShortNameToName[pkgShortName]
         originAndArtifact := OriginAndArtifact(pkgFullName) 
         packagePath := LocalPackagePath(pkgFullName)      
   
         // TODO Dubious values of version and mustBeFromShared here!!!
         err = rt.Loader.LoadRelishCodePackage(originAndArtifact,"",packagePath,false)
         if err != nil {
            return
         }
      
         keyType = rt.Typs[keyTypeShortName]    
       
         // Alternate strategy!!    
   	   // rterr.Stop("Can't summon object. The package which defines its type, '%s', has not been loaded into the runtime.",localTypeName) 
       }      
   }
   
   
   var elementType *RType
   elementType = rt.Typs[elementTypeShortName]
   if elementType == nil {

      pkgShortName := PackageShortName(elementTypeShortName)  
//      localTypeName := LocalTypeName(typeName)   
      pkgFullName := RT.PkgShortNameToName[pkgShortName]
      originAndArtifact := OriginAndArtifact(pkgFullName) 
      packagePath := LocalPackagePath(pkgFullName)      
   
      // TODO Dubious values of version and mustBeFromShared here!!!
      err = rt.Loader.LoadRelishCodePackage(originAndArtifact,"",packagePath,false)
      if err != nil {
         return
      }
      
      elementType = rt.Typs[elementTypeShortName]    
       
      // Alternate strategy!!    
	   // rterr.Stop("Can't summon object. The package which defines its type, '%s', has not been loaded into the runtime.",localTypeName) 
    }   


	var minCardinality int64 = 0
	var maxCardinality int64 = -1


   var isAscending bool = true
	var unaryMethod *RMultiMethod
	var binaryMethod *RMultiMethod
	var orderAttr *AttributeSpec
	
	
	
   collection,err = rt.NewCollection(minCardinality,
      maxCardinality,
      nil, // owner   
      nil, // attribute
      collectionType, 
      isAscending,
      unaryMethod,
      binaryMethod,
      orderAttr,  
      keyType, 
      elementType) 	
	
	return
}



/*
minCardinality int64,
maxCardinality int64,
owner RObject,  // can be nil   
isList bool,   // "list","sortedlist"
isSet bool,    // "set","sortedset"
isMap bool, // "map" "stringmap" "sortedmap" "sortedstringmap"
isStringMap bool,  // "stringmap" "sortedstringmap"
isSorted bool, // "sortedlist" "sortedset" "sortedmap" "sortedstringmap"

keyType *RType,  // can be nil   
elementType *RType) (collection RCollection, err error) {
*/   
func (rt *RuntimeEnv) NewCollection(
   minCardinality int64,
   maxCardinality int64,
   owner RObject,  // can be nil   
   attribute *AttributeSpec,  // can be nil
   collectionType string, 
   isAscending bool,
   unaryMethod *RMultiMethod,
   binaryMethod *RMultiMethod,
   orderAttr *AttributeSpec,  
   keyType *RType, 
   elementType *RType) (collection RCollection, err error) {


	var objColl RObject

	var sortWith *sortOp

   // Why not use
   // "list" "set" "map" "stringmap" "intmap" "sortedlist" "sortedset" "sortedmap" "sortedstringmap" "sortedintmap"
   // in the type descriptor !!!!!!!
   //
   //
   
   // Also, should be unpacking the shorttypename of element (and possibly map key) here
   // and loading their packages if not loaded.

	if collectionType == "sortedlist" || collectionType == "sortedset" || collectionType == "sortedmap" || collectionType == "sortedstringmap" {			

		if binaryMethod == nil {
			binaryMethod, _ = rt.InbuiltFunctionsPackage.MultiMethods["lt"]
		}

		sortWith = &sortOp{
			attr:          orderAttr,
			unaryFunction: unaryMethod,
			lessFunction:  binaryMethod,
			descending:    !isAscending,
		}
	}

	/*
		@@@@@@@@@@@@@@@@@@@@@@@

		NEED TO FIND OUT IF THERE IS A UNARY METHOD OF THE TYP2, otherwise it must be a binary method.

		Only one of the attr or unaryFunction will be non-nil.
		If attr or unaryFunction is non-nil, then lessFunction must be the "lt" multiMethod.

		collection.sortWith.lessFunction,_ := RT.MultiMethods["lt"]

		If attr and unaryFunction are nil, lessFunction may be any binary boolean function which has a method whose
		parameter signature is compatible with a pair of values of the elementType of the collection. lessFunction MAY
		be the "lt" function in this case but need not be. The function is treated as a "less-than" predicate.

		type sortOp struct {
			attr *AttributeSpec
			unaryFunction *RMultiMethod
			lessFunction *RMultiMethod
			descending bool
		}

		$$$$$$$$$$$$$$$$$$$$$$$$


		  attr = &AttributeSpec{typ1,
		                        RelEnd{
		 									    Name:endName2,
		                                       Type:typ2,
		                                       ArityLow:arityLow2,
		                                       ArityHigh:arityHigh2,
		                                       CollectionType:collectionType2,
		                                       OrderAttr:orderAttrName,
		                                       OrderMethod: orderMethod,
		 									   OrderMethodArity: int32,
		                                       IsAscending:isAscending,
		                                      },

		@@@@@@@@@@@@@@
	*/

	switch collectionType {
	case "list", "sortedlist":
		objColl, err = rt.Newrlist(elementType, minCardinality, maxCardinality, owner, attribute, sortWith)
	case "set":
		objColl, err = rt.Newrset(elementType, minCardinality, maxCardinality, owner, attribute)
	case "sortedset":
		objColl, err = rt.Newrsortedset(elementType, minCardinality, maxCardinality, owner, attribute, sortWith)
	case "map","sortedmap","stringmap","sortedstringmap","uint64map","int64map":		
      objColl, err = rt.Newmap(keyType, elementType, minCardinality, maxCardinality, owner, attribute, sortWith)		
	default:
		panic(fmt.Sprintf("I don't handle %s attributes yet.",collectionType))		
	}

	collection = objColl.(RCollection)

	return
}














/*
Creates a new RCollection object of the appropriate type, for purposes of restoring a collection
from persistent storage.
The typeDescriptor is from the type field of the collection's instance entry in the RObject table in the
local database. 
Will have to get fancier here about sorting specifications, by enhancing the type descriptor info.
This method is only for independent collections. Not for multivalued attribute collections.

A collection type descriptor is something like:

[ordered_][map|stringmap|list|set]

func (rt *RuntimeEnv) NewCollection(
   minCardinality int64,
   maxCardinality int64,
   owner RObject,  // can be nil   
   isList bool,   // "list","sortedlist"
   isSet bool,    // "set","sortedset"
   isMap bool, // "map" "stringmap" "sortedmap" "sortedstringmap"
   isStringMap bool,  // "stringmap" "sortedstringmap"
   isSorted bool, // "sortedlist" "sortedset" "sortedmap" "sortedstringmap"
   isAscending bool,
	unaryMethod *RMultiMethod,
	binaryMethod *RMultiMethod,
	orderAttr *AttributeSpec,
	keyType *RType,  // can be nil   
   elementType *RType) (collection RCollection, err error) {


	var objColl RObject
	var sortWith *sortOp,

	// fmt.Println(attr.Part.CollectionType)

   if isSorted {			

		if binaryMethod == nil {
			binaryMethod, _ = rt.InbuiltFunctionsPackage.MultiMethods["lt"]
		}

		sortWith = &sortOp{
			attr:          orderAttr,
			unaryFunction: unaryMethod,
			lessFunction:  binaryMethod,
			descending:    !isAscending,
		}
	}


	if isList {
		objColl, err = rt.Newrlist(attr.Part.Type, minCardinality, maxCardinality, owner, sortWith)
	} else if isSet {
	   if isSorted {
		   objColl, err = rt.Newrsortedset(attr.Part.Type, minCardinality, maxCardinality, owner, sortWith)	      
      } else {
         objColl, err = rt.Newrset(attr.Part.Type, minCardinality, maxCardinality, owner)         
      }
   } else if isStringMap {
      if isSorted {
         panic("I don't handle sortedstringmap creation yet.")         
      } else {
         panic("I don't handle stringmap creation yet.")
      }      
   } else if isMap {
      if isSorted {
         panic("I don't handle sortedmap creation yet.")         
      } else {
         panic("I don't handle map creation yet.")
      }
   }
	
	collection = objColl.(RCollection)
	return
}

*/










/*
Removes val from the multi-valued attribute if val is in the collection. Does nothing and does not complain if val is not in the collection.
If removePersistent is true, also removes the value from the persistent version of the attribute association.
*/
func (rt *RuntimeEnv) RemoveFromAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, isInverse bool, removePersistent bool) (err error) { 	
	
	objColl, foundCollection := rt.AttrVal(th, obj, attr)

	if !foundCollection { // this object does not have the collection implementation of this multi-valued attribute	


	    // TODO WHOA WHOA WHOA  just because the collection wasn't there in mem doesn't mean the value shouldnt be removed
	    // from the database representation of the multi-valued attribute association table does it????
	    // Can this situation arise?
        // OK OK, it looks like now that rt.AttrVal(...) will pull the collection into memory.


		return
	}

    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if obj.IsBeingStored() {
       unit := obj.(*runit)
       ensureMemoryTransactionConsistency2(th, unit)    	
    }	



	collection := objColl.(RemovableMixin) // Will throw an exception if collection type does not implement Remove(..)
	removed, removedIndex := collection.Remove(th, val)



    // fmt.Println("collection.Remove(val)", removed, removedIndex)
	
	if removed  {
	   if removePersistent && obj.IsBeingStored() {
            // fmt.Println("calling th.DB().PersistRemoveFromAttr(obj, attr, val, removedIndex)")	   	
	    	err = th.DBT().PersistRemoveFromAttr(obj, attr, val, removedIndex)
	    	if err != nil {
	    		return
	    	}
	   }
	
	   if ! isInverse && attr.Inverse != nil {
	   	  err = rt.RemoveAttrGeneral(th, val, attr.Inverse, obj, true, removePersistent)
	   }
	}

	return
}




/*

Unsets the attribute. Used when setting to nil.
Deletes the corresponding row from the attribute database table if it exists.
Ok to call this even if attribute had no value.
If removePersistent, removes the attribute association from the database if it was persisted.
*/
func (rt *RuntimeEnv) UnsetAttr(th InterpreterThread, obj RObject, attr *AttributeSpec, isInverse bool, removePersistent bool) (err error) {

    if attr.Part.Type.IsPrimitive {
       err = fmt.Errorf("Cannot assign primitive-valued attribute '%v.%v %v' a value of nil", obj.Type(), attr.Part.Name, attr.Part.Type)
       return
    }

    t := obj.Type()
    i := attr.Index[t]

    var unit *runit
    var isUnit bool
    unit,isUnit = obj.(*runit)
    
    if ! isUnit {
        unit = &(obj.(*GoWrapper).runit)    	
    }

    // Note. This needs to be locked, so that the attribute setting only gets associated with one
    // transaction, and there is no race.

    if obj.IsBeingStored() {
       ensureMemoryTransactionConsistency2(th, unit)    	
    }	

    val := unit.attrs[i] 
    found := (val != nil)

	if ! found {
		if obj.IsStoredLocally() {

			val, found = rt.AttrVal(th, obj, attr)
		}
	}

	if found {

      unit.attrs[i] = nil

       if removePersistent && obj.IsBeingStored() {
	          err = th.DBT().PersistRemoveAttr(obj, attr) 	 
	          if err != nil {
	          	 return
	          } 
       }
       

       if ! isInverse && attr.Inverse != nil && val != NIL {
           err = rt.RemoveAttrGeneral(th, val, attr.Inverse, obj, true, removePersistent)
       }
    } else {

    }	
	
    return
}

/*
If the attribute is multi-valued, removes the val from it, otherwise
if single valued, unsets the attribute.
*/
func (rt *RuntimeEnv) RemoveAttrGeneral(th InterpreterThread, obj RObject, attr *AttributeSpec, val RObject, isInverse bool, removePersistent bool) (err error) {
   if attr.Part.CollectionType == "" {	
	  err = rt.UnsetAttr(th, obj, attr, isInverse, removePersistent)
   } else {
      err = rt.RemoveFromAttr(th, obj, attr, val, isInverse, removePersistent)	
   }
   return
}

/*
Returns the object that has been given the specified name in the global context map.
Returns *nil* if no object found in the global context map under the name.
*/
func (rt *RuntimeEnv) ContextGet(name string) RObject {
	rt.contextMutex.RLock()
	defer rt.contextMutex.RUnlock()
	val,found := rt.context[name]
	if ! found {
		val = NIL
	}
	return val
}

func (rt *RuntimeEnv) ContextExists(name string) (found bool) {
	rt.contextMutex.RLock()
	defer rt.contextMutex.RUnlock()
	_,found = rt.context[name]	
	return
}

func (rt *RuntimeEnv) ContextPut(obj RObject, name string) {
	rt.contextMutex.Lock()
	defer rt.contextMutex.Unlock()
	rt.context[name] = obj
}

func (rt *RuntimeEnv) ContextRemove(name string) {
	rt.contextMutex.Lock()
	defer rt.contextMutex.Unlock()
	delete(rt.context,name)
}




















/*
type AttributeSpec struct {
   WholeType *RType
   Part RelEnd
   IsTransient bool
}


One end of a relation - specifies arity and type constraints and a few other details.

type RelEnd struct {
   Name string
   Type *RType
   ArityLow int32
   ArityHigh int32
   CollectionType string // "list", "sortedlist", "set", "sortedset", "map", "sortedmap", ""
   OrderAttrName string   // which primitive attribute of other is it ordered by when retrieving? "" if none

   OrderMethod *RMultiMethod

   DependentPart bool // delete of parent results in delete of attribute value
}
*/

/*
Can only be used in between a SetEvalContext and UnsetEvalContext call.
*/
func (rt *RuntimeEnv) GetEvalContext(obj RObject) MethodEvaluationContext {
	return rt.evalContexts[obj]
}

/*
You MUST call UnsetEvalContext after calling this!!
*/
func (rt *RuntimeEnv) SetEvalContext(obj RObject, context MethodEvaluationContext) {
	rt.evalContextMutex.Lock()
	rt.evalContexts[obj] = context
}

func (rt *RuntimeEnv) UnsetEvalContext(obj RObject) {
	delete(rt.evalContexts, obj)
	rt.evalContextMutex.Unlock()
}

// Usage
// RT.SetEvalContext(obj, context)
// defer RT.UnsetEvalContext(obj)
// context := RT.GetEvalContext(obj) 

type MethodEvaluationContext interface {

	/*
	   Evaluates a call of the (single-valued) multimethod on the argument objects.
	   Returns the result of the method call.
	*/
	EvalMultiMethodCall(mm *RMultiMethod, args []RObject) RObject
	
	/*
	   Return the interpreter thread aspect of the evaluation context.
	*/
    InterpThread() InterpreterThread 	
}

type Dispatcher interface {

	/*
	   The main dispatch function.
	   First looks up a cache (map) of method implementations keyed by typetuples.
	   a method will be found in this cache if the type tuple of the arguments
	   has had the multimethod called on it before.
	   If there is a cache miss, uses a multi-argument dynamic dispatch algorithm
	   to find the best-matching method implementation (then caches the find under
	   the type-tuple of the arguments for next time.)
	   Returns the best method implementation for the types of the argument objects,
	   or nil if the multimethod has no method signature which is compatible with
	   the types of the argument objects.
	   Also returns the type-tuple of the argument objects, which can be used to
	   report the lack of a compatible method.
	*/
	GetMethod(mm *RMultiMethod, args []RObject) (*RMethod, *RTypeTuple)

	/*
	   Same as GetMethod but for types instead of object instances.
	*/
	GetMethodForTypes(mm *RMultiMethod, types ...*RType) (*RMethod, *RTypeTuple)
}
