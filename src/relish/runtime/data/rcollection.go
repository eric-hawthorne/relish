// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   rcollection.go - relish collection objects
*/

import (
	"fmt"
	"sort"
	"relish/rterr"
	. "relish/dbg"
	"errors"
	"time"
	"sync"
	"strings"
    "net/http"
)

const MAX_CARDINALITY = 999999999999999999 // Replace with highest int64?

///////////////////////////////////////////////////////////////////////////
////////// COLLECTIONS
///////////////////////////////////////////////////////////////////////////

/*
[] Widget             List

[<] Widget            Sorted list using natural order of Widgets (which must be defined)

[<attr] Widget        Sorted list using attribute/unary function of Widget

[<less] Widget        Sorted ist using binary comparison function over widgets

{} Widget             Set

{<} Widget            Sorted set using natural order of Widgets (which must be defined)

{<attr} Widget        Sorted set using attribute/unary function of Widget

{<less} Widget        Sorted set using binary comparison function over widgets (if "less" unary func defined it will be used instead)


{} String > Widget    Map

{} name > Widget      Map using name attribute of Widget (but does not update itself if widget.name is changed) 

{<} String > Widget   Sorted map using natural order of Strings (which must be defined)

{<} name > Widget     Sorted map using name attribute/unary function of Widget (but does not update itself if widget.name is changed) 

*/

type RCollection interface {
	RObject
	ElementType() *RType
	Length() int64
	Cap() int64
	MinCard() int64
	MaxCard() int64
	IsMap() bool
	IsSet() bool
	IsList() bool
	IsOrdered() bool // true if either the collection is maintained in a sorted order, or if the collection
	// at least holds each member in an index-accessible position and returns members
	// when iterated over in order of their sequential indexed position.

	IsSorting() bool // true if the collection has a defined sort order other index-position.
    IsInsertable() bool // true if the ordered collection supports inserting an element at an integer index
         // Note: Presently, only a List which is not sorting is insertable    
    IsIndexSettable() bool // true if the ordered collection supports setting the value of the ith element.
         // Note: Presently, only a List which is not sorting is index settable


	IsCardOk() bool  // Is my current cardinality within my cardinality constraints?
	Owner() RObject  // if non-nil, this collection is the implementation of a multi-valued attribute.
	Attribute() *AttributeSpec  // if non-nil, the attribute that this is the value of for the owner object.
	// Returns an iterator that yields the objects in the collection. A Map iterator returns the keys.
	Iter(th InterpreterThread) <-chan RObject // Usage: for obj := range s.Iter()  or ch := s.Iter(); x, ok = <-ch

    Contains(th InterpreterThread, obj RObject) bool // true if the collection contains the element, false otherwise
                               // uses value equality for primitive element types, reference equality otherwise.
                               // for maps, it is whether the map contains a key equal to the argument object. 


    SetMayContainProxies(status bool)  // Set to true when collection is lazily fetched, false after deproxifying

    MayContainProxies() bool    // Some or all collection elements may be Proxy objects which need to be replaced by a 
                                // db fetch of the real object.

	/*
	  The currently active DB transaction which this collection is dirty in.
	*/
	Transaction() *RTransaction 

	SetTransaction(tx *RTransaction)    
}

/*
A collection which can have a member added. It is added at the end (appended) if this is an non-sorting list.
It is added in the appropriate place in the order, if this is a sorting collection.
It is added in undetermined place if an unordered set.
*/
type AddableMixin interface {
	Add(obj RObject, context MethodEvaluationContext) (added bool, newLen int)

	/*
		This version of the add method does not sort. It assumes that it is being called with element objects
		already known to be simply insertable (at the end of if applicable) the collection.
		Used by the persistence service. Do not use for general use of the collection.
	*/
	AddSimple(obj RObject) (newLen int)
}

type AddableCollection interface {
	RCollection
	AddableMixin
}

/*
A collection which can have a member removed.
*/
type RemovableMixin interface {
	/*
	   removedIndex will be -1 if not applicable or if removed is false
	*/
	Remove(th InterpreterThread, obj RObject) (removed bool, removedIndex int)

	/*
		Removes all members of the in-memory aspect of the collection, setting its len to 0. 
		Does not affect the database-persisted aspect of the collection. 
		Used to refresh the collection from the db.
	*/
	ClearInMemory()

	/*
	Removes all members of the in-memory and local db version of the collection. 
	Sets Length() to 0
	*/
	// Clear()
}

type RemovableCollection interface {
	RCollection
	RemovableMixin
}


type OrderedCollection interface {
	Index(obj RObject, start int) int
	At(th InterpreterThread, i int) RObject	
}

/*
A collection which can return its go list implementation
*/
type List interface {
	RCollection
	AddableMixin
	RemovableMixin
	OrderedCollection
	Vector() *RVector
	AsSlice(th InterpreterThread) []RObject
    ReplaceContents(objs []RObject)	

    /*
    Returns a new list of the same element type, containing elements from the start index (inclusive)
    to the end index (exclusive) of the original list.
    If the end index is < 0, the length of the original list is added to it. So start:0, end:-1 returns
    all but the last element of the list.
    */
    Slice(th InterpreterThread, start int, end int) List
}

/*
A collection which can return its go map implementation
*/
type Set interface {
	RCollection
	AddableMixin
	RemovableMixin
	BoolMap() map[RObject]bool
}

type Map interface {
	RCollection
	RemovableMixin
	KeyType() *RType
	ValType() *RType
	Get(key RObject) (val RObject, found bool)
	Put(key RObject, val RObject, context MethodEvaluationContext) (added bool, newLen int)

	/*
		This version of the put method does not sort. It assumes that it is being called with key and val objects
		already known to be simply insertable (at the end of if applicable) the collection.
		Used by the persistence service. Do not use for general use of the collection.
	*/
	PutSimple(key RObject, val RObject) (newLen int)	
}

type Insertable interface {
	RCollection
	Insert(i int, val RObject) (newLen int)
}

type IndexSettable interface {
	RCollection
	Set(i int, val RObject) 
}














/*
func (c *container) Iter () <-chan item {
    ch := make(chan item);
    go func () {
        for i := 0; i < c.size; i++ {
            ch <- c.items[i]
        }
    } ();
    return ch
}

*/

/*
   Abstract 
*/
type rcollection struct {
	robject
	minCard     int64
	maxCard     int64
	elementType *RType
	owner       RObject
	attr *AttributeSpec

	sortWith *sortOp // Which attribute of a member, or which unary func of member or which less function to sort with. May be nil.

	mayContainProxies bool // If this collection was fetched from the db, could there still be some elements which are proxies?
	                       // set to false when deproxified.
	transaction *RTransaction
}

func (o *rcollection) SetMayContainProxies(status bool) {
   o.mayContainProxies = status
}

func (o *rcollection) MayContainProxies() bool {
   return o.mayContainProxies
}



/*
Only one of the attr or unaryFunction will be non-nil.
If attr or unaryFunction is non-nil, then lessFunction must be the "lt" multiMethod.

collection.sortWith.lessFunction,_ := RT.MultiMethods["lt"]

If attr and unaryFunction are nil, lessFunction may be any binary boolean function which has a method whose
parameter signature is compatible with a pair of values of the elementType of the collection. lessFunction MAY
be the "lt" function in this case but need not be. The function is treated as a "less-than" predicate.

*/
type sortOp struct {
	attr          *AttributeSpec
	unaryFunction *RMultiMethod
	lessFunction  *RMultiMethod
	descending    bool
	sortingDeferred bool  // if true, do not do any sorting of collection for the moment.
	                      // Used while inserting a large number of elements all at once into a sorting collection.
}

func (o *rcollection) String() string {
	return (&(o.robject)).StringTh(nil)  // Not ThreadSafe
}

func (o *rcollection) StringTh(th InterpreterThread) string {
	return (&(o.robject)).StringTh(th)
}

func (o *rcollection) Debug(th InterpreterThread) string {
	return (&(o.robject)).Debug(th) 
}

/*
Only applies to sorting collections
*/
func (o *rcollection) IsSortingDeferred() bool {
	return o.sortWith.sortingDeferred
}

/*
Only applies to sorting collections
*/
func (o *rcollection) SetSortingDeferred(status bool) {
	o.sortWith.sortingDeferred = status
}

func (c rcollection) IsUnit() bool {
	return false
}

func (c rcollection) IsCollection() bool {
	return true
}

func (c rcollection) MinCard() int64 {
	return c.minCard
}

func (c rcollection) MaxCard() int64 {
	return c.maxCard
}

func (c rcollection) ElementType() *RType {
	return c.elementType
}

/*
   If not nil, it means this collection is the implemnentation of a multiple-valued attribute.
*/
func (c rcollection) Owner() RObject {
	return c.owner
}

func (c rcollection) Attribute() *AttributeSpec {
	return c.attr
}

func (c rcollection) IsInsertable() bool {
   return false  // default, override in sub-types
}

func (c rcollection) IsIndexSettable() bool {
   return false  // default, override in sub-types
}

func (c *rcollection) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
	
	//fmt.Println("rcollection.ToMapListTree")
	
	// Record myself in the visited set so we don't infinitely loop.
	visited[c.This()] = true	
	
	var val interface{}
	coll := c.This().(RCollection)
	if coll.IsMap() {
		if c.ElementType() != StringType {
           err = errors.New("Cannot represent Map with non-String key in JSON.")
           return	
        }
        strMap := coll.(*rstringmap) 
        resultMap := make(map[string]interface{})
        for key, value := range strMap.m {
	        if visited[value] {
		       continue
	        }
	        val, err = value.ToMapListTree(th, includePrivate, visited)
	        if err != nil {
		       return
	        }
	        resultMap[key] = val
        }	 			
		tree = resultMap
	} else {
		resultSlice := make([]interface{},0)
		for value := range coll.Iter(nil) {
            if visited[value] {
	           continue
            }			
            // fmt.Println("coll",coll)
            // fmt.Println ("value",value)
	        val, err = value.ToMapListTree(th, includePrivate, visited)
	        if err != nil {
		       return
	        }	
	        resultSlice = append(resultSlice, val)		
		}
		tree = resultSlice
    	// fmt.Println(resultSlice)		
	}
    return
}











/*
A set of relish objects constrained to be of some type.
Implements RCollection
Object address defines element equality. May want to fix that!!! It may not even be true.
*/
type rset struct {
	rcollection
	m map[RObject]bool // use this as set 
}

// DEBUGGING KLUDGE - REMOVE THIS
//
//var DieInUseDB bool = false

///////////////////

func (o *rset) String() string {
   //DieInUseDB = true
   return "ERROR: CANNOT PRINT rset WITHOUT THREAD CONTEXT"
}

func (o *rset) StringTh(th InterpreterThread) string {
   s := ""
   if o.Length() > 4 {
	   sep := "\n   {"
	   for obj := range o.Iter(th) {
	      s += sep + obj.StringTh(th)
	      sep = "\n      "
	   }
	   s += "\n   }"
   	} else { // Horizontal layout
	   s = "{"
	   sep := ""	
	   for obj := range o.Iter(th) {
	      s += sep + obj.StringTh(th)
	      sep = "   "
	   }
	   s += "}"
   }
   return s
}

func (o *rset) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s len:%d\n%s",  (&(o.rcollection)).Debug(th) , o.Length(), o.StringTh(th))
}

func (s *rset) BoolMap() map[RObject]bool {
	return s.m
}

func (s *rset) Add(obj RObject, context MethodEvaluationContext) (added bool, newLen int) {

	if s.m == nil {
		s.m = make(map[RObject]bool)
	} else {
	   th := context.InterpThread()		
       s.deproxify(th)	
    }

	added = !s.m[obj]
	s.m[obj] = true
	newLen = len(s.m)
	return
}

func (s *rset) AddSimple(obj RObject) (newLen int) {

	if s.m == nil {
		s.m = make(map[RObject]bool)
	} 
	s.m[obj] = true
	newLen = len(s.m)
	return
}

func (s *rset) Remove(th InterpreterThread, obj RObject) (removed bool, removedIndex int) {

    if s.m == nil {	
    	return
    } 
    s.deproxify(th)	

	removed = s.m[obj]
	delete(s.m, obj) // delete(s.m,obj)
	removedIndex = -1
	return
}

func (s *rset) ClearInMemory() {

	s.m = nil
}


/*
*/
func (c *rset) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
	var relishVal RObject
	switch tree.(type) {
	case []interface{}:
		slice := tree.([]interface{})
		for _,val := range slice {
			prototypeObj := c.ElementType().Prototype()
			relishVal, err = prototypeObj.FromMapListTree(th, val)
			if err != nil {
				return
			}
			c.AddSimple(relishVal)
		}
	default:
	   err = errors.New("When unmarshalling into a Set, a JSON list is expected.")		
	}
	obj = c
	return 
}



/*
Weird behaviour: Only if the iteration is allowed to complete (i.e. to exhaust the map) will proxies
in the map be replaced by real objects.

TODO !!!! TODO !!!!   use more standard and optimized deproxify with a flag.

TODO MAPS AND PROXIES ARE NOT HAPPY TOGETHER YET!!!!
IT WOULD ONLY WORK IF THE ROBJECT IDENTITY IN THE MAP IS BASED ON THE DBID instead of object address.
TODO

TODO TODO TODO - Uses the wrong db to do Fetch !!! Should be th.DBT().
*/
func (c *rset) Iter(th InterpreterThread) <-chan RObject {
		
    c.deproxify(th)	

	ch := make(chan RObject)
	go func() {
		for obj, _ := range c.m {
			ch <- obj
		}
		close(ch)
	}()
	return ch
}



var mapFetchMutex sync.Mutex  // Why allowing only one map globally to be deproxified at one time?

/*
   Converts all proxies in the collection to real objects.
   TODO: Reconsider the kludge of accepting a nil interpreterThread.
   Currently used in String method to list the elements.
*/
func (c *rset) deproxify(th InterpreterThread) {

    if c.MayContainProxies() {

		var db DBT
		if th == nil {
			db = RT.DBT()
		} else {
			db = th.DBT()
		}

        mapFetchMutex.Lock()

		var fromPersistence map[RObject]RObject
		for robj, _ := range c.m {
			if robj.IsProxy() {
				var err error
				proxy := robj.(Proxy)
				robj, err = db.Fetch(int64(proxy), 0)
				if err != nil {
					panic(fmt.Sprintf("Error fetching set element: %s", err))
				}
				if fromPersistence == nil {
					fromPersistence = make(map[RObject]RObject)
				}
				fromPersistence[proxy] = robj
			}
		}
		// Replace proxies in the set with real objects.

		// TODO Need to mutex lock the map here to guarantee the len of the map is always correct.


		for proxy, robj := range fromPersistence {
			c.m[robj] = true
			delete(c.m, proxy) // delete(c,m)				
		}

		c.SetMayContainProxies(false)

		mapFetchMutex.Unlock()		
    }
}



/*
Creates a fresh new slice.
*/
func (c *rset) AsSlice(th InterpreterThread) []RObject {
    s := make([]RObject,0,c.Length())     
	for obj := range c.Iter(th) {
        s = append(s,obj)
	}
    return s
}

/*
*/
func (s *rset) Iterable() (interface{},error) {
	var fakeThread FakeInterpreterThread	
	return s.AsSlice(fakeThread),nil
}



func (c *rset) Contains(th InterpreterThread, obj RObject) (found bool) {
	if c.m == nil {
		found = false
		return
	} else {	
       c.deproxify(th)	
    }

	_,found = c.m[obj]
	return
}

/*
If the object is not already marked as reachable, flag it as reachable.
Return whether we had to flag it as reachable. false if was already marked reachable.
Also recursively Mark all (in-memory) elements of the collection.
*/
func (o *rset) Mark() bool { 
   if ! (&(o.rcollection.robject)).Mark() {
      return false
   }
   if o.elementType.IsPrimitive {
   	   return true
   }
   Logln(GC3_,"rset.Mark(): Marking",len(o.m),"elements:")
   for obj := range o.m {
   	  obj.Mark()
   }

   return true
}


/*
   Constructor
*/
func (rt *RuntimeEnv) Newrset(elementType *RType, minCardinality, maxCardinality int64, owner RObject, attr *AttributeSpec) (coll RCollection, err error) {
	typ, err := rt.GetSetType(elementType)
	if err != nil {
		return nil, err
	}
	if maxCardinality == -1 {
		maxCardinality = MAX_CARDINALITY
	}
	s := &rset{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, elementType, owner, attr, nil, false, nil}, nil}
	s.rcollection.robject.this = s
	coll = s		
    if ! markSense {
	    coll.SetMarked()
    }	

    if Logging(GC_) {
       n_rsets_ever ++
	   Logln(GC3_,"Newrset: IsMarked",coll.IsMarked())		
    }

	return
}

func (s rset) Length() int64   { return int64(len(s.m)) }
func (s rset) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rset) IsMap() bool     { return false }
func (s rset) IsSet() bool     { return true }
func (s rset) IsList() bool    { return false }
func (s rset) IsOrdered() bool { return false } // This may change!! Actually, need an rsortedset type
func (s rset) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rsortedset
func (s rset) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

func (o rset) IsZero() bool {
	return o.Length() == 0
}

/*
A sorted set of relish objects constrained to be of some type.
Implements RCollection
Object address defines element equality. May want to fix that!!! It may not even be true.
*/
type rsortedset struct {
	rcollection
	m map[RObject]bool // use this as set 
	v *RVector
}

func (o *rsortedset) String() string {
   //DieInUseDB = true
   return "ERROR: CANNOT PRINT rsortedset WITHOUT THREAD CONTEXT"
}

func (o *rsortedset) StringTh(th InterpreterThread) string {
   s := ""
   if o.Length() > 4 {
	   sep := "\n   {"
	   for obj := range o.Iter(th) {
	      s += sep + obj.StringTh(th)
	      sep = "\n      "
	   }
	   s += "\n   }"
   	} else { // Horizontal layout
	   s = "{"
	   sep := ""	
	   for obj := range o.Iter(th) {
	      s += sep + obj.StringTh(th)
	      sep = "   "
	   }
	   s += "}"
   }
   return s
}

func (o *rsortedset) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s len:%d\n%s",  (&(o.rcollection)).Debug(th) , o.Length(), o.StringTh(th))
}

func (s *rsortedset) BoolMap() map[RObject]bool {
	return s.m
}

func (s *rsortedset) Add(obj RObject, context MethodEvaluationContext) (added bool, newLen int) {
	
	if s.m == nil {
		s.m = make(map[RObject]bool)
		s.v = new(RVector)
	} else {	
	   th := context.InterpThread()				
       s.deproxify(th)	
    }	

	_, found := s.m[obj]
	if !found {
		added = true
		s.m[obj] = true
		s.v.Push(obj)

        if ! s.IsSortingDeferred() {	
		   RT.SetEvalContext(s, context)
		   defer RT.UnsetEvalContext(s)
		   sort.Sort(s)
	    } 
	}
	newLen = len(s.m)
	return
}

func (s *rsortedset) AddSimple(obj RObject) (newLen int) {

	if s.m == nil {
		s.m = make(map[RObject]bool)
		s.v = new(RVector)
	} 

	s.m[obj] = true
	s.v.Push(obj)
	newLen = len(s.m)
	return
}

func (s *rsortedset) At(th InterpreterThread, i int) RObject {

    defer vectorAtErrHandle(i, s.v)
	obj := s.v.At(i).(RObject)
	if obj.IsProxy() {
		var err error
		proxy := obj.(Proxy)
		obj, err = th.DBT().Fetch(int64(proxy), 0)
		if err != nil {
			panic(fmt.Sprintf("Error fetching sorted-set element [%v]: %s", i, err))
		}
		(*(s.v))[i] = obj	

  		s.m[obj] = true
		delete(s.m, proxy) 				
	}
	return obj
}


func vectorAtErrHandle(i int, v *RVector) {
      r := recover()	
      if r != nil {
          s, isString := r.(string)
          if isString { 
             if strings.HasPrefix(s, "Error fetching") {
             	panic(s)
             }
          }
          panic(fmt.Sprintf("Error: index [%d] is out of range. Collection length is %d.",i,v.Len()))
      }	
}	




/*
{<} Widget            Sorted set using natural order of Widgets (which must be defined)

{<attr} Widget        Sorted set using attribute/unary function of Widget

{<less} Widget        Sorted set using binary comparison function over widgets (if "less" unary func defined it will be used instead)


type Interface interface {
    // Len is the number of elements in the collection.
    Len() int
    // Less returns whether the element with index i should sort
    // before the element with index j.
    Less(i, j int) bool
    // Swap swaps the elements with indexes i and j.
    Swap(i, j int)
}
*/

func (s *rsortedset) Len() int {
	return len(s.m)
}

// TODO TODO Optimize this one the same as for rlist !!!!!!

func (s *rsortedset) Less(i, j int) bool {	

	var goLess bool

	var evalContext MethodEvaluationContext = RT.GetEvalContext(s)
	th := evalContext.InterpThread()

	if s.sortWith == nil { 

		val1 := s.At(th, i)

		val2 := s.At(th, j)

	    switch val1.(type) {
	        case RTime:
	        	goLess = time.Time(val1.(RTime)).Before(time.Time(val2.(RTime)))
	        case String:
	        	goLess = string(val1.(String)) < string(val2.(String))
	        case Int:
	        	goLess = int64(val1.(Int)) < int64(val2.(Int))
	        case Float:
	        	goLess = float64(val1.(Float)) < float64(val2.(Float))            	
	        case Bool: 
	        	goLess = (bool(val1.(Bool)) == false) && (bool(val2.(Bool)) == true)     
	        default:
				// Use the "less" multimethod to compare them.

			    lessMethod, _ := th.CallingPackage().MultiMethods["lt"]
			    //lessMethod, _ := RT.InbuiltFunctionsPackage.MultiMethods["lt"]

				isLess := evalContext.EvalMultiMethodCall(lessMethod, []RObject{val1, val2})        
				goLess = ! isLess.IsZero()       
	    }
	    return goLess
	}


	
	//var isLess RObject

	if s.sortWith.attr != nil {

		// Get attr value of both list members

		obj1 := s.At(th, i)
		val1, found := RT.AttrVal(th, obj1, s.sortWith.attr)
		if !found {
			panic(fmt.Sprintf("Object %v has no value for attribute %s", obj1, s.sortWith.attr.Part.Name))
		}

		obj2 := s.At(th, j)
		val2, found := RT.AttrVal(th, obj2, s.sortWith.attr)
		if !found {
			panic(fmt.Sprintf("Object %v has no value for attribute %s", obj2, s.sortWith.attr.Part.Name))
		}




        // depending on the type of the values to compare, use an optimized sorting method.

        // TODO - Does not currently permit mixed-type numbers in a sorted collection

        switch val1.(type) {
            case RTime:
            	goLess = time.Time(val1.(RTime)).Before(time.Time(val2.(RTime)))
            case String:
            	goLess = string(val1.(String)) < string(val2.(String))
            case Int:
            	goLess = int64(val1.(Int)) < int64(val2.(Int))
            case Float:
            	goLess = float64(val1.(Float)) < float64(val2.(Float))            	
            case Bool: 
            	goLess = (bool(val1.(Bool)) == false) && (bool(val2.(Bool)) == true)     
            default:
				// Use the "less" multimethod to compare them.

				// Assumes that the sortWith has been given the "less" multimethod. TODO!
				// 
				isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{val1, val2})        
				goLess = ! isLess.IsZero()       
        }

		if s.sortWith.descending {
			return ! goLess
		}
		return goLess

	} else if s.sortWith.unaryFunction != nil {

		// Evaluate the unary function separately on both list members


		obj1 := s.At(th, i)
		val1 := evalContext.EvalMultiMethodCall(s.sortWith.unaryFunction, []RObject{obj1})

		obj2 := s.At(th, j)
		val2 := evalContext.EvalMultiMethodCall(s.sortWith.unaryFunction, []RObject{obj2})


        // depending on the type of the values to compare, use an optimized sorting method.

        // TODO - Does not currently permit mixed-type numbers in a sorted collection

        switch val1.(type) {
            case RTime:
            	goLess = time.Time(val1.(RTime)).Before(time.Time(val2.(RTime)))
            case String:
            	goLess = string(val1.(String)) < string(val2.(String))
            case Int:
            	goLess = int64(val1.(Int)) < int64(val2.(Int))
            case Float:
            	goLess = float64(val1.(Float)) < float64(val2.(Float))            	
            case Bool: 
            	goLess = (bool(val1.(Bool)) == false) && (bool(val2.(Bool)) == true)     
            default:
				// Use the "less" multimethod to compare them.

				// Assumes that the sortWith has been given the "less" multimethod. TODO!
				// 
				isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{val1, val2})        
				goLess = ! isLess.IsZero()       
        }

		if s.sortWith.descending {
			return ! goLess
		}
		return goLess

	}
	// else ... lessFunction

	// Apply the multi-method to the two list members. It may be just the "less" multimethod.

	// Get attr value of both list members

	val1 := s.At(th, i)

	val2 := s.At(th, j)

    switch val1.(type) {
        case RTime:
        	goLess = time.Time(val1.(RTime)).Before(time.Time(val2.(RTime)))
        case String:
        	goLess = string(val1.(String)) < string(val2.(String))
        case Int:
        	goLess = int64(val1.(Int)) < int64(val2.(Int))
        case Float:
        	goLess = float64(val1.(Float)) < float64(val2.(Float))            	
        case Bool: 
        	goLess = (bool(val1.(Bool)) == false) && (bool(val2.(Bool)) == true)     
        default:
			// Use the "less" multimethod to compare them.

			// Assumes that the sortWith has been given the "less" multimethod. TODO!
			// 
			isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{val1, val2})        
			goLess = ! isLess.IsZero()       
    }

	if s.sortWith.descending {
		return ! goLess
	}
	return goLess
}

/*
type sortOp {
	attr *AttributeSpec
	unaryFunction *RMultiMethod
	lessFunction *RMultiMethod
	descending bool
}
*/

type SortableMixin interface {
   sort.Interface
   IsSortingDeferred() bool 

   SetSortingDeferred(status bool)

   // This just means the collection is in general supposed to maintain itself in sorted order. 
   // It does not reflect whether sorting is deferred right now or not.
   IsSorting() bool

   // Swap(i, j int)    
}


/*
Not valid to call on indexes >= the length of the collection.
*/
func (s *rsortedset) Swap(i, j int) {
	s.v.Swap(i, j)
}

/*
Returns the index of the first-found occurrence of the argument object with the search beginning at the start index.
TODO Should make this more efficient by doing a binary search. !!!!!!!

TODO !! This DOES NOT HANDLE a collection that has been lazily restored from persistence as a bunch of Proxy objects.

*/
func (s *rsortedset) Index(obj RObject, start int) int {
	if s.m != nil {
		ln := len(*(s.v))
		for i := start; i < ln; i++ {
			if obj == s.v.At(i) {
				return i
			}
		}
	}
	return -1
}


func (c *rsortedset) Contains(th InterpreterThread, obj RObject) (found bool) {
	if c.m == nil {
		found = false
		return
	} else {				
       c.deproxify(th)	
    }	

	_,found = c.m[obj]
	return
}



func (s *rsortedset) Remove(th InterpreterThread, obj RObject) (removed bool, removedIndex int) {

	if s.v == nil {
        Logln(COLL2_,"rsortedset.Remove(): s.v == nil, removed == false, removedIndex == -1")   		
		removedIndex = -1
	} else {			
        s.deproxify(th)	
        len1 := len(s.m)
		delete(s.m, obj) // delete (s.m,obj)
		len2 := len(s.m)			
		weird := (len1 == len2)
		if weird {
		   Logln(COLL2_,"rsortedset.Remove(): len(s.m) same after delete(s.m,obj). len =", len2)  
		   Logln(COLL2_,"rsortedset.Remove(): obj =", obj.Debug(th))   
		   for key,val := range s.m {
		      Logln(COLL2_,"rsortedset.Remove(): m key =", key.Debug(th), "val =", val) 		   	   
		   }	 	
		}
		removedIndex = s.Index(obj, 0)
		if weird {
		   for i,val := range *s.v {
		      Logln(COLL2_,"rsortedset.Remove(): v[",i,"] =", val.Debug(th)) 		   	  
		   }	
		   Logln(COLL2_,"rsortedset.Remove(): removedIndex =", removedIndex)  			
		}
		if removedIndex >= 0 {
			s.v.Delete(removedIndex)
			removed = true
		    if weird {
		       Logln(COLL2_,"rsortedset.Remove(): deleted from vector but not map. len(s.v) =",len(*s.v))  			
		    }			
		}
	}
	return
}

func (s *rsortedset) ClearInMemory() {	

	s.m = nil
	if s.v != nil {
		s.v = s.v.Resize(0, s.v.Cap())
	}
}


/*
Note: Currently, this does not sort the elements.
This could lead to illegal state. Need a MethodEvaluationContext to sort.
*/
func (c *rsortedset) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
	var relishVal RObject
	switch tree.(type) {
	case []interface{}:
		slice := tree.([]interface{})
		for _,val := range slice {
			prototypeObj := c.ElementType().Prototype()
			relishVal, err = prototypeObj.FromMapListTree(th, val)
			if err != nil {
				return
			}
			c.AddSimple(relishVal)
		}
	default:
	   err = errors.New("When unmarshalling into a SortedSet, a JSON list is expected.")		
	}
	obj = c
	return 
}

/*
TODO Use more standard flag-based deproxify
 */
func (c *rsortedset) Iter(th InterpreterThread) <-chan RObject {

    c.deproxify(th)

	ch := make(chan RObject)
	go func() {
		if c.v != nil {
			for _, obj := range *(c.v) {
				ch <- obj
			}
		}
		close(ch)
	}()
	return ch
}


/*
If the object is not already marked as reachable, flag it as reachable.
Return whether we had to flag it as reachable. false if was already marked reachable.
Also recursively Mark all (in-memory) elements of the collection.
*/
func (o *rsortedset) Mark() bool { 
   if ! (&(o.rcollection.robject)).Mark() {
      return false
   }
   if o.elementType.IsPrimitive {
   	   return true
   }
   Logln(GC3_,"rsortedset.Mark(): Marking",len(o.m),"elements:")
   for obj := range o.m {
   	  obj.Mark()
   }
   return true
}



func (c *rsortedset) deproxify(th InterpreterThread) {
    if c.MayContainProxies() {
		var db DBT
		if th == nil {
			db = RT.DBT()
		} else {
			db = th.DBT()
		}	
		if c.v != nil {
			for i, obj := range *(c.v) {
				robj := obj.(RObject)
				if robj.IsProxy() {
					var err error
					proxy := robj.(Proxy)
					robj, err = db.Fetch(int64(proxy), 0)
					if err != nil {
						panic(fmt.Sprintf("Error fetching sorted set element: %s", err))
					}
					(*(c.v))[i] = robj

  			        c.m[robj] = true
			        delete(c.m, proxy) 				
				}
			}
		}

		c.SetMayContainProxies(false)	
    } 
}

/*
Do not modify the returned slice
*/
func (s *rsortedset) AsSlice(th InterpreterThread) []RObject {
	s.deproxify(th)
	if s.v == nil {
		return []RObject{}
	}
	return []RObject(*(s.v))
}

/*
Do not modify the returned slice
*/
func (s *rsortedset) Iterable() (interface{},error) {
	var fakeThread FakeInterpreterThread	
	return s.AsSlice(fakeThread),nil
}


/*
   Constructor
*/
func (rt *RuntimeEnv) Newrsortedset(elementType *RType, minCardinality, maxCardinality int64, owner RObject, attr *AttributeSpec, sortWith *sortOp) (coll RCollection, err error) {
	typ, err := rt.GetSetType(elementType)
	if err != nil {
		return nil, err
	}
	if maxCardinality == -1 {
		maxCardinality = MAX_CARDINALITY
	}
	s := &rsortedset{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, elementType, owner, attr, sortWith, false, nil}, nil, nil}
	s.rcollection.robject.this = s
	coll = s		
	if ! markSense {
	    coll.SetMarked()
	}

    if Logging(GC_) {
       n_rsortedsets_ever ++
	   Logln(GC3_,"Newrsortedset: IsMarked",coll.IsMarked())	
    }

		
	return
}

func (s rsortedset) Length() int64   { return int64(len(s.m)) }
func (s rsortedset) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rsortedset) IsMap() bool     { return false }
func (s rsortedset) IsSet() bool     { return true }
func (s rsortedset) IsList() bool    { return false }
func (s rsortedset) IsOrdered() bool { return true } // This may change!! Actually, need an rsortedset type
func (s rsortedset) IsSorting() bool { return true } // This may change!! Depends on presence of an ordering function in rsortedset
func (s rsortedset) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

func (o rsortedset) IsZero() bool {
	return o.Length() == 0
}

/*
A list of relish objects constrained to be of some type.
Implements RCollection
*/
type rlist struct {
	rcollection
	v *RVector
}

func (o *rlist) String() string {
   //DieInUseDB = true
   return "ERROR: CANNOT PRINT rlist WITHOUT THREAD CONTEXT"
}

func (o *rlist) StringTh(th InterpreterThread) string {
   s := ""
   //if th == nil {
   //   fmt.Println("VVV rlist.StringTh(nil)")
   //} else {
   //   fmt.Println("VVV rlist.StringTh()")
   //}
   if o.Length() > 4 {
	   sep := "\n   ["
	   for obj := range o.Iter(th) {
	      s += sep + obj.StringTh(th)
	      sep = "\n      "
	   }
	   s += "\n   ]"
   	} else { // Horizontal layout
	   s = "["
	   sep := ""
	   for obj := range o.Iter(th) {
	      s += sep + obj.StringTh(th)
	      sep = "   "
	   }
	   s += "]"
   }
   //fmt.Println("^^^ rlist.StringTh()")
   return s
}

func (o *rlist) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s len:%d\n%s",  (&(o.rcollection)).Debug(th) , o.Length(), o.StringTh(th))
}

/*
TODO: Reconsider the kludge of accepting a nil interpreterThread.
Currently used in String method to list the elements.
*/
func (c *rlist) Iter(th InterpreterThread) <-chan RObject {

	c.deproxify(th)

	ch := make(chan RObject)
	go func() {
		if c.v != nil {
			for _, obj := range *(c.v) {
				ch <- obj
			}
		}
		close(ch)
	}()
	return ch
}




func (c *rlist) Slice(th InterpreterThread, start int, end int) (slice List) {
	slice = c.Type().Prototype().(List)
	length := int(c.Length())	
	if end < 0 {
		end = length + end
	}
	vec := c.v
	if vec == nil {
		return
	}
	defer sliceErrHandle(start,end, length)		
	sliceVec := vec.Slice(start,end)
	sl := slice.(*rlist)
	sl.v = sliceVec

	sl.SetMayContainProxies(c.MayContainProxies())

	return	
}	

func sliceErrHandle(start int, end int, length int) {
      r := recover()	
      if r != nil {
          rterr.Stopf("Error in list slice [%d:%d]: index out of range. List length is %d.",start,end,length)
      }	
}		
	




	    // substring or slice
	/*
		slice s String start Int end Int > String
		
		slice s String start Int > String		

Counts in bytes.

func builtinStringSlice(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	start := int(objects[1].(Int))
	length := len(s)
	end := length
	if len(objects) == 3 {
		end = int(objects[2].(Int))
		if end < 0 {
			end = length + end
		}
	}
	substr := s[start:end]
    return []RObject{String(substr)}
}
*/


/*
If the object is not already marked as reachable, flag it as reachable.
Return whether we had to flag it as reachable. false if was already marked reachable.
Also recursively Mark all (in-memory) elements of the collection.
*/
func (o *rlist) Mark() bool { 
   if ! (&(o.rcollection.robject)).Mark() {
      return false
   }
   if o.elementType.IsPrimitive {
   	   return true
   }   
   if o.v != nil {
       Logln(GC3_,"rlist.Mark(): Marking",len(*(o.v)),"elements:")	
	   for _,obj := range *(o.v) {
	   	  obj.Mark()
	   }
   }
   return true
}

/*
   Converts all proxies in the collection to real objects.
   TODO: Reconsider the kludge of accepting a nil interpreterThread.
   Currently used in String method to list the elements.
*/
func (c *rlist) deproxify(th InterpreterThread) {
    if c.MayContainProxies() {
		var db DBT
		if th == nil {
			db = RT.DBT()
		} else {
			db = th.DBT()
		}
		if c.v != nil {
			for i, obj := range *(c.v) {
				robj := obj.(RObject)
				if robj.IsProxy() {
					var err error
					proxy := robj.(Proxy)
					robj, err = db.Fetch(int64(proxy), 0)
					if err != nil {
						panic(fmt.Sprintf("Error fetching list element: %s", err))
					}
					(*(c.v))[i] = robj				
				}
			}
		}
		c.SetMayContainProxies(false)
    }
}

/*
Return the underlying collection.
*/
func (s *rlist) Vector() *RVector {
	return s.v
}


/*
Insert the element at the specified index. Shift elements from that index on to have
the next higher index.
TODO: DOES NOT HANDLE PERSISTENCE YET !!!!
*/
func (s *rlist)	Insert(i int, val RObject) (newLen int) {
	if s.IsSorting() {
        rterr.Stop("Cannot insert at specified index into a sorting list.")				
	}
	if s.v == nil {
		if i != 0 {
           rterr.Stopf("Error in list-element insert: index %d is out of range.",i)			
		}
		newLen = s.AddSimple(val)
	} else {

       s.v.Insert(i, val)
       newLen = len(*(s.v))
	}
	return
}

/*
Set the element at index i to be the specified value.
Note: It is illegal to call this on a IsSorting() == true list. Not enforced. Watch out!

TODO: Note that the current use of this e.g. Interpreter.go 2662 is not persisting the set operation.
*/
func (s *rlist) Set(i int, val RObject) {	
   if s.v == nil {
      rterr.Stopf("Error in list-element set: index %d is out of range.",i)
   }   
   s.v.Set(i, val)
}


func (s *rlist) ReplaceContents(objs []RObject) {

	var rv RVector = RVector(objs)
	s.v = &rv
}

func (s *rlist) AsSlice(th InterpreterThread) []RObject {
	s.deproxify(th)
	if s.v == nil {
		return []RObject{}
	}
	return []RObject(*(s.v))
}

func (s *rlist) Contains(th InterpreterThread, obj RObject) bool {
	if s.v == nil {
		return false
	} 
	for _,element := range s.AsSlice(th) {
       if obj == element {
       	   return true
       }
	}
    return false
}




func (s *rlist) Iterable() (interface{},error) {
	var fakeThread FakeInterpreterThread
	return s.AsSlice(fakeThread),nil
}

// RT.SetEvalContext(obj, context)
// defer RT.UnsetEvalContext(obj)
// context := RT.GetEvalContext(obj)	

func (s *rlist) Add(obj RObject, context MethodEvaluationContext) (added bool, newLen int) {	

	if s.v == nil {
		s.v = new(RVector)
	}
	s.v.Push(obj)

	if s.IsSorting() && ! s.IsSortingDeferred() {
		RT.SetEvalContext(s, context)
		defer RT.UnsetEvalContext(s)
		s.deproxify(context.InterpThread())
		sort.Sort(s)
	}
	added = true
	newLen = s.v.Len()
	return
}

func (s *rlist) AddSimple(obj RObject) (newLen int) {

	if s.v == nil {
		s.v = new(RVector)
	}
	s.v.Push(obj)
	newLen = s.v.Len()
	return
}

func (s *rlist) At(th InterpreterThread, i int) RObject {

    defer vectorAtErrHandle(i, s.v)
	obj := s.v.At(i).(RObject)
	if obj.IsProxy() {
		var err error
		proxy := obj.(Proxy)
		obj, err = th.DBT().Fetch(int64(proxy), 0)
		if err != nil {
			panic(fmt.Sprintf("Error fetching list element [%v]: %s", i, err))
		}
		(*(s.v))[i] = obj		
	}
	return obj
}

func (s *rlist) IsInsertable() bool {
   return ! s.IsSorting()
}

func (s *rlist) IsIndexSettable() bool {
    return ! s.IsSorting()
}

func (s *rlist) Len() int {
	return s.v.Len()
}

func (s *rlist) Less(i, j int) bool {

	var evalContext MethodEvaluationContext = RT.GetEvalContext(s)
	th := evalContext.InterpThread()

	var goLess bool

	if s.sortWith == nil { // Not a sorted list. 

		val1 := s.At(th, i)

		val2 := s.At(th, j)

	    switch val1.(type) {
	        case RTime:
	        	goLess = time.Time(val1.(RTime)).Before(time.Time(val2.(RTime)))
	        case String:
	        	goLess = string(val1.(String)) < string(val2.(String))
	        case Int:
	        	goLess = int64(val1.(Int)) < int64(val2.(Int))
	        case Float:
	        	goLess = float64(val1.(Float)) < float64(val2.(Float))            	
	        case Bool: 
	        	goLess = (bool(val1.(Bool)) == false) && (bool(val2.(Bool)) == true)     
	        default:
				// Use the "less" multimethod to compare them.

			    lessMethod, _ := th.CallingPackage().MultiMethods["lt"]
			    //lessMethod, _ := RT.InbuiltFunctionsPackage.MultiMethods["lt"]

				isLess := evalContext.EvalMultiMethodCall(lessMethod, []RObject{val1, val2})        
				goLess = ! isLess.IsZero()       
	    }
	    return goLess
	}

	//var isLess RObject

	if s.sortWith.attr != nil {

		// Get attr value of both list members

		obj1 := s.At(th, i)
		val1, found := RT.AttrVal(th, obj1, s.sortWith.attr)
		if !found {
			panic(fmt.Sprintf("Object %v has no value for attribute %s", obj1, s.sortWith.attr.Part.Name))
		}

		obj2 := s.At(th, j)
		val2, found := RT.AttrVal(th, obj2, s.sortWith.attr)
		if !found {
			panic(fmt.Sprintf("Object %v has no value for attribute %s", obj2, s.sortWith.attr.Part.Name))
		}

        // depending on the type of the values to compare, use an optimized sorting method.

        // TODO - Does not currently permit mixed-type numbers in a sorted collection

        switch val1.(type) {
            case RTime:
            	goLess = time.Time(val1.(RTime)).Before(time.Time(val2.(RTime)))
            case String:
            	goLess = string(val1.(String)) < string(val2.(String))
            case Int:
            	goLess = int64(val1.(Int)) < int64(val2.(Int))
            case Float:
            	goLess = float64(val1.(Float)) < float64(val2.(Float))            	
            case Bool: 
            	goLess = (bool(val1.(Bool)) == false) && (bool(val2.(Bool)) == true)     
            default:
				// Use the "less" multimethod to compare them.

				// Assumes that the sortWith has been given the "less" multimethod. TODO!
				// 
				isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{val1, val2})        
				goLess = ! isLess.IsZero()       
        }

		if s.sortWith.descending {
			return ! goLess
		}
		return goLess

	} else if s.sortWith.unaryFunction != nil {

		// Evaluate the unary function separately on both list members

		obj1 := s.At(th, i)
		val1 := evalContext.EvalMultiMethodCall(s.sortWith.unaryFunction, []RObject{obj1})

		obj2 := s.At(th, j)
		val2 := evalContext.EvalMultiMethodCall(s.sortWith.unaryFunction, []RObject{obj2})


        switch val1.(type) {
            case RTime:
            	goLess = time.Time(val1.(RTime)).Before(time.Time(val2.(RTime)))
            case String:
            	goLess = string(val1.(String)) < string(val2.(String))
            case Int:
            	goLess = int64(val1.(Int)) < int64(val2.(Int))
            case Float:
            	goLess = float64(val1.(Float)) < float64(val2.(Float))            	
            case Bool: 
            	goLess = (bool(val1.(Bool)) == false) && (bool(val2.(Bool)) == true)     
            default:
				// Use the "less" multimethod to compare them.

				// Assumes that the sortWith has been given the "less" multimethod. TODO!
				// 
				isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{val1, val2})        
				goLess = ! isLess.IsZero()       
        }

		if s.sortWith.descending {
			return ! goLess
		}
		return goLess
	}
	// else ... lessFunction

	// Apply the multi-method to the two list members. It may be just the "less" multimethod.

	// Get attr value of both list members

	val1 := s.At(th, i)

	val2 := s.At(th, j)

    switch val1.(type) {
        case RTime:
        	goLess = time.Time(val1.(RTime)).Before(time.Time(val2.(RTime)))
        case String:
        	goLess = string(val1.(String)) < string(val2.(String))
        case Int:
        	goLess = int64(val1.(Int)) < int64(val2.(Int))
        case Float:
        	goLess = float64(val1.(Float)) < float64(val2.(Float))            	
        case Bool: 
        	goLess = (bool(val1.(Bool)) == false) && (bool(val2.(Bool)) == true)     
        default:
			// Use the "less" multimethod to compare them.

			// Assumes that the sortWith has been given the "less" multimethod. TODO!
			// 
			isLess := evalContext.EvalMultiMethodCall(s.sortWith.lessFunction, []RObject{val1, val2})        
			goLess = ! isLess.IsZero()       
    }

	if s.sortWith.descending {
		return ! goLess
	}
	return goLess
}

/*
type sortOp {
	attr *AttributeSpec
	unaryFunction *RMultiMethod
	lessFunction *RMultiMethod
	descending bool
}
*/

/*
Not valid to call on indexes >= the length of the collection.
*/
func (s *rlist) Swap(i, j int) {
	s.v.Swap(i, j)
}

/*
Returns the index of the first-found occurrence of the argument object with the search beginning at the start index.
*/
func (s *rlist) Index(obj RObject, start int) int {
	if s.v != nil {
		ln := len(*(s.v))
		for i := start; i < ln; i++ {
			if obj == s.v.At(i) {
				return i
			}
		}
	}
	return -1
}

func (s *rlist) Remove(th InterpreterThread, obj RObject) (removed bool, removedIndex int) {
	
	if s.v == nil {
		removedIndex = -1
	} else {
		s.deproxify(th)
		removedIndex = s.Index(obj, 0)
		if removedIndex >= 0 {
			s.v.Delete(removedIndex)
			removed = true
		}
	}
	return
}

func (s *rlist) ClearInMemory() {

	if s.v != nil {
		s.v = s.v.Resize(0, s.v.Cap())
	}
}


/*
*/
func (c *rlist) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
	var relishVal RObject
	switch tree.(type) {
	case []interface{}:
		slice := tree.([]interface{})
		for _,val := range slice {
			prototypeObj := c.ElementType().Prototype()
			relishVal, err = prototypeObj.FromMapListTree(th, val)
			if err != nil {
				return
			}
			c.AddSimple(relishVal)
		}
	default:
	   err = errors.New("When unmarshalling into a List, a JSON list is expected.")		
	}
	obj = c
	return 
}


/*
   Constructor
*/
func (rt *RuntimeEnv) Newrlist(elementType *RType, minCardinality, maxCardinality int64, owner RObject, attr *AttributeSpec, sortWith *sortOp) (coll List, err error) {
	typ, err := rt.GetListType(elementType)
	if err != nil {
		return nil, err
	}
	if maxCardinality == -1 {
		maxCardinality = MAX_CARDINALITY
	}
	lst := &rlist{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, elementType, owner, attr, sortWith, false, nil}, nil}
	lst.rcollection.robject.this = lst
	coll = lst	
	if ! markSense {
	    coll.SetMarked()
	}	

    if Logging(GC_) {
       n_rlists_ever ++
	   Logln(GC3_,"Newrlist: IsMarked",coll.IsMarked())	
    }	
	
	return
}

func (s rlist) Length() int64   { if s.v == nil { return 0};  return int64(s.v.Len()) }
func (s rlist) Cap() int64      { if s.v == nil { return 0}; return int64(s.v.Cap()) }
func (s rlist) IsMap() bool     { return false }
func (s rlist) IsSet() bool     { return false }
func (s rlist) IsList() bool    { return true }
func (s rlist) IsOrdered() bool { return true } // This may change!! Depends on presence of an ordering function
func (s rlist) IsSorting() bool { return s.sortWith != nil }
func (s rlist) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }

func (o rlist) IsZero() bool {
	return o.Length() == 0
}




/*
   Constructor - decides which kind of map to use depending on the keyType
*/
func (rt *RuntimeEnv) Newmap(keyType *RType, valType *RType, minCardinality, maxCardinality int64, owner RObject, attr *AttributeSpec, sortWith *sortOp) (coll Map, err error) {
	typ, err := rt.GetMapType(keyType,valType)
	if err != nil {
		return nil, err
	}
	if maxCardinality == -1 {
		maxCardinality = MAX_CARDINALITY
	}
	//fmt.Println(keyType)
	//fmt.Println(valType)
	//fmt.Println(keyType == StringType)	
	switch keyType {
	case StringType:
		//fmt.Println("Making a stringmap")
		m := &rstringmap{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, keyType, owner, attr, sortWith, false, nil}, valType, make(map[string]RObject)}	
	    m.rcollection.robject.this = m
	    coll = m
	    //fmt.Println(m)
	    //fmt.Println(m.m)
   	 //fmt.Println(m.m == nil)	    
	case IntType, Int32Type:
		m := &rint64map{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, keyType, owner, attr, sortWith, false, nil}, valType, make(map[int64]RObject)}
	    m.rcollection.robject.this = m
	    coll = m	
	case UintType, Uint32Type:
		m := &ruint64map{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, keyType, owner, attr, sortWith, false, nil},valType, make(map[uint64]RObject)}
	    m.rcollection.robject.this = m				
	    coll = m	
	default:
		m := &rpointermap{rcollection{robject{rtype: typ}, minCardinality, maxCardinality, keyType, owner, attr, sortWith, false, nil},valType, make(map[RObject]RObject)}		
	    m.rcollection.robject.this = m			
	    coll = m				
	}
	
	if ! markSense {
	    coll.SetMarked()
	}
	
    if Logging(GC_) {
       n_maps_ever ++
	   Logln(GC3_,"Newmap: IsMarked",coll.IsMarked())	
    }	

	return
}


type rstringmap struct {
	rcollection
	valType *RType
	m map[string]RObject
}

func (o *rstringmap) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s len:%d\n%s",  (&(o.rcollection)).Debug(th) , o.Length(),o.StringTh(th))
}

func (o *rstringmap) String() string {
        return o.StringTh(nil)
}

func (o *rstringmap) StringTh(th InterpreterThread) string {
	encoded, err := JsonMarshal(nil, o, false)
	if err != nil {
		return "({}String > T with unserializable elements)"
	}
	return encoded
}

func (c *rstringmap) Iter(th InterpreterThread) <-chan RObject {
	ch := make(chan RObject)
	go func() {
		for key, _ := range c.m {
			ch <- String(key)
		}
		close(ch)
	}()
	return ch
}

func (o *rstringmap) Mark() bool { 
   if ! (&(o.rcollection.robject)).Mark() {
      return false
   }
   if o.valType.IsPrimitive {
   	   return true
   }
   Logln(GC3_,"rstringmap.Mark(): Marking",len(o.m),"elements:")
   for _,obj := range o.m {
   	  obj.Mark()
   }
   return true
}

func (s *rstringmap) Iterable() (interface{},error) {
	return s.m,nil
}




func (s rstringmap) Length() int64   { return int64(len(s.m)) }
func (s rstringmap) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rstringmap) IsMap() bool     { return true }
func (s rstringmap) IsSet() bool     { return false }
func (s rstringmap) IsList() bool    { return false }
func (s rstringmap) IsOrdered() bool { return false } // This may change!! Need an rorderedstringmap
func (s rstringmap) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rorderedstringmap
func (s rstringmap) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }
func (o rstringmap) IsZero() bool {
	return o.Length() == 0
}

func (s rstringmap) KeyType() *RType {
	return s.ElementType()
}

func (s rstringmap) ValType() *RType {
	return s.valType
}

func (s *rstringmap) Get(key RObject) (val RObject, found bool) {
	k := string(key.(String))
	val, found = s.m[k]
	return
}

func (s *rstringmap) Put(key RObject, val RObject, context MethodEvaluationContext) (added bool, newLen int) {
	
	k := string(key.(String))	
	_,found := s.m[k] 
	added = ! found
   s.m[k] = val
   newLen = len(s.m)
   return
}

func (s *rstringmap) Contains(th InterpreterThread, key RObject) (found bool) {
	k := string(key.(String))
	_, found = s.m[k]
	return
}

/*
	This version of the put method does not sort. It assumes that it is being called with key and val objects
	already known to be simply insertable (at the end of if applicable) the collection.
	Used by the persistence service. Do not use for general use of the collection.
*/
func (s *rstringmap) PutSimple(key RObject, val RObject) (newLen int) {	

	k := string(key.(String))
	 //fmt.Println(s.m)
	 //fmt.Println(s.m == nil)	
    s.m[k] = val
    newLen = len(s.m)
    return
}

func (s *rstringmap) Remove(th InterpreterThread, key RObject) (removed bool, removedIndex int) {

	k := string(key.(String))		
	_,removed = s.m[k] 
	delete(s.m, k) 
	removedIndex = -1
	return
}

func (s *rstringmap) ClearInMemory() {	

	s.m = make(map[string]RObject)
}


/*
*/
func (c *rstringmap) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
	var relishVal RObject
	switch tree.(type) {
	case []interface{}:

	case map[string]interface{}:
		theMap := tree.(map[string]interface{})		

		for key,val := range theMap {
			prototypeObj := c.ValType().Prototype()
			relishVal, err = prototypeObj.FromMapListTree(th, val)
			if err != nil {
				return
			}
			c.PutSimple(String(key), relishVal)
		}
	default:
	   err = errors.New("When unmarshalling into a Map, a JSON map is expected.")		
	}
	obj = c
	return 
}








type ruint64map struct {
	rcollection
	valType *RType	
	m map[uint64]RObject
}

func (o *ruint64map) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s len:%d",  (&(o.rcollection)).Debug(th) , o.Length())
}

func (c *ruint64map) Iter(th InterpreterThread) <-chan RObject {
	ch := make(chan RObject)
	go func() {
		for key, _ := range c.m {
			ch <- Uint(key)
		}
		close(ch)
	}()
	return ch
}


func (o *ruint64map) Mark() bool { 
   if ! (&(o.rcollection.robject)).Mark() {
      return false
   }
   if o.valType.IsPrimitive {
   	   return true
   }
   Logln(GC3_,"ruint64map.Mark(): Marking",len(o.m),"elements:")
   for _,obj := range o.m {
   	  obj.Mark()
   }
   return true
}

func (s *ruint64map) Iterable() (interface{},error) {
	return s.m,nil
}

func (s ruint64map) Length() int64   { return int64(len(s.m)) }
func (s ruint64map) Cap() int64      { return int64(len(s.m)) * 2 }
func (s ruint64map) IsMap() bool     { return true }
func (s ruint64map) IsSet() bool     { return false }
func (s ruint64map) IsList() bool    { return false }
func (s ruint64map) IsOrdered() bool { return false } // This may change!! Need an rordereduintmap
func (s ruint64map) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rorderedstringmap
func (s ruint64map) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }
func (o ruint64map) IsZero() bool {
	return o.Length() == 0
}

func (s ruint64map) KeyType() *RType {
	return s.ElementType()
}

func (s ruint64map) ValType() *RType {
	return s.valType
}

func (s *ruint64map) Get(key RObject) (val RObject, found bool) {
    var k uint64
	switch key.(type) {
	   case Uint:	
	      k = uint64(key.(Uint))
       case Uint32:
	      k = uint64(uint32(key.(Uint32)))
	   case Int:	
	      k = uint64(int64(key.(Int)))
       case Int32:
	      k = uint64(int32(key.(Int32)))	
	   default:
	     rterr.Stop("Invalid type for map key.")     
	} 
	val, found = s.m[k]
	return
}

func (s *ruint64map) Put(key RObject, val RObject, context MethodEvaluationContext) (added bool, newLen int) {	

    var k uint64
	switch key.(type) {
	   case Uint:	
	      k = uint64(key.(Uint))
       case Uint32:
	      k = uint64(uint32(key.(Uint32)))
	   case Int:	
	      k = uint64(int64(key.(Int)))
       case Int32:
	      k = uint64(int32(key.(Int32)))	
	   default:
	     rterr.Stop("Invalid type for map key.")     
	} 
	_,found := s.m[k] 
	added = ! found
    s.m[k] = val
    newLen = len(s.m)
    return
}

func (s *ruint64map) Contains(th InterpreterThread, key RObject) (found bool) {
	_,found = s.Get(key)
	return
}

/*
	This version of the put method does not sort. It assumes that it is being called with key and val objects
	already known to be simply insertable (at the end of if applicable) the collection.
	Used by the persistence service. Do not use for general use of the collection.
*/
func (s *ruint64map) PutSimple(key RObject, val RObject) (newLen int) {

    var k uint64
	switch key.(type) {
	   case Uint:	
	      k = uint64(key.(Uint))
       case Uint32:
	      k = uint64(uint32(key.(Uint32)))
	   case Int:	
	      k = uint64(int64(key.(Int)))
       case Int32:
	      k = uint64(int32(key.(Int32)))	
	   default:
	     rterr.Stop("Invalid type for map key.")     
	} 
    s.m[k] = val
    newLen = len(s.m)
    return
}


func (s *ruint64map) Remove(th InterpreterThread, key RObject) (removed bool, removedIndex int) {	

    var k uint64
	switch key.(type) {
	   case Uint:	
	      k = uint64(key.(Uint))
       case Uint32:
	      k = uint64(uint32(key.(Uint32)))
	   case Int:	
	      k = uint64(int64(key.(Int)))
       case Int32:
	      k = uint64(int32(key.(Int32)))	
	   default:
	     rterr.Stop("Invalid type for map key.")     
	} 

	_,removed = s.m[k] 
	delete(s.m, k) 
	removedIndex = -1
	return
}

func (s *ruint64map) ClearInMemory() {

	s.m = make(map[uint64]RObject)
}




func (c *ruint64map) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
	err = errors.New("Cannot unmarshal JSON into a Map unless the key-type is String.")		
	return 
}









/*
Can I just use a ruint64map and cast the int64 to uint64? probably
*/
type rint64map struct {
	rcollection
	valType *RType	
	m map[int64]RObject
}

func (o *rint64map) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s len:%d",  (&(o.rcollection)).Debug(th) , o.Length())
}

func (c *rint64map) Iter(th InterpreterThread) <-chan RObject {
	ch := make(chan RObject)
	go func() {
		for key, _ := range c.m {
			ch <- Int(key)
		}
		close(ch)
	}()
	return ch
}

func (o *rint64map) Mark() bool { 
   if ! (&(o.rcollection.robject)).Mark() {
      return false
   }
   if o.valType.IsPrimitive {
   	   return true
   }
   Logln(GC3_,"rint64map.Mark(): Marking",len(o.m),"elements:")
   for _,obj := range o.m {
   	  obj.Mark()
   }
   return true
}

func (s *rint64map) Iterable() (interface{},error) {
	return s.m,nil
}

func (s rint64map) Length() int64   { return int64(len(s.m)) }
func (s rint64map) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rint64map) IsMap() bool     { return true }
func (s rint64map) IsSet() bool     { return false }
func (s rint64map) IsList() bool    { return false }
func (s rint64map) IsOrdered() bool { return false } // This may change!! Depends on presence of an ordering function
func (s rint64map) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rorderedstringmap
func (s rint64map) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }
func (o rint64map) IsZero() bool {
	return o.Length() == 0
}

func (s rint64map) KeyType() *RType {
	return s.ElementType()
}

func (s rint64map) ValType() *RType {
	return s.valType
}

func (s *rint64map) Get(key RObject) (val RObject, found bool) {
    var k int64
	switch key.(type) {
	   case Int:	
	      k = int64(key.(Int))
       case Int32:
	      k = int64(int32(key.(Int32)))
	   case Uint:	
	      k = int64(uint64(key.(Uint)))
       case Uint32:
	      k = int64(uint32(key.(Uint32)))	
	   default:
	     rterr.Stop("Invalid type for map key.")     
	} 
	val, found = s.m[k]
	return
}

func (s *rint64map) Put(key RObject, val RObject, context MethodEvaluationContext) (added bool, newLen int) {

    var k int64
	switch key.(type) {
	   case Int:	
	      k = int64(key.(Int))
       case Int32:
	      k = int64(int32(key.(Int32)))
	   case Uint:	
	      k = int64(uint64(key.(Uint)))
       case Uint32:
	      k = int64(uint32(key.(Uint32)))	
	   default:
	     rterr.Stop("Invalid type for map key.")     
	} 
	_,found := s.m[k] 
	added = ! found
    s.m[k] = val
    newLen = len(s.m)
    return
}

func (s *rint64map) Contains(th InterpreterThread, key RObject) (found bool) {
	_,found = s.Get(key)
	return
}

/*
	This version of the put method does not sort. It assumes that it is being called with key and val objects
	already known to be simply insertable (at the end of if applicable) the collection.
	Used by the persistence service. Do not use for general use of the collection.
*/
func (s *rint64map) PutSimple(key RObject, val RObject) (newLen int) {

    var k int64
	switch key.(type) {
	   case Int:	
	      k = int64(key.(Int))
       case Int32:
	      k = int64(int32(key.(Int32)))
	   case Uint:	
	      k = int64(uint64(key.(Uint)))
       case Uint32:
	      k = int64(uint32(key.(Uint32)))	
	   default:
	     rterr.Stop("Invalid type for map key.")     
	} 
    s.m[k] = val
    newLen = len(s.m)
    return
}

func (s *rint64map) Remove(th InterpreterThread, key RObject) (removed bool, removedIndex int) {	

    var k int64
	switch key.(type) {
	   case Int:	
	      k = int64(key.(Int))
       case Int32:
	      k = int64(int32(key.(Int32)))
	   case Uint:	
	      k = int64(uint64(key.(Uint)))
       case Uint32:
	      k = int64(uint32(key.(Uint32)))	
	   default:
	     rterr.Stop("Invalid type for map key.")     
	}

	_,removed = s.m[k] 
	delete(s.m, k) 
	removedIndex = -1
	return
}

func (s *rint64map) ClearInMemory() {	

	s.m = make(map[int64]RObject)
}

func (c *rint64map) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
    err = errors.New("Cannot unmarshal JSON into a Map unless the key-type is String.")		
	return 
}



type rpointermap struct {
	rcollection
	valType *RType	
	m map[RObject]RObject
}

func (o *rpointermap) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s len:%d",  (&(o.rcollection)).Debug(th) , o.Length())
}

func (c *rpointermap) Iter(th InterpreterThread) <-chan RObject {
	ch := make(chan RObject)
	go func() {
		for key, _ := range c.m {
			ch <- key //.This()
		}
		close(ch)
	}()
	return ch
}

func (o *rpointermap) Mark() bool { 
   if ! (&(o.rcollection.robject)).Mark() {
      return false
   }
   if o.elementType.IsPrimitive && o.valType.IsPrimitive {
      return true
   } 
   if o.elementType.IsPrimitive {
       Logln(GC3_,"rpointermap.Mark(): Marking",len(o.m),"map values:")	
	   for _,obj := range o.m {
	   	  obj.Mark()
	   }
   } else if o.valType.IsPrimitive {
       Logln(GC3_,"rpointermap.Mark(): Marking",len(o.m),"map keys:")	
	   for obj := range o.m {
	   	  obj.Mark()
	   }
   } else {
       Logln(GC3_,"rpointermap.Mark(): Marking",len(o.m),"map keys and same number of map values.")	
	   for key,obj := range o.m {
	   	  key.Mark()
	   	  obj.Mark()
	   }
   }
   return true
}


func (s *rpointermap) Iterable() (interface{},error) {
	return s.m,nil
}

func (s rpointermap) Length() int64   { return int64(len(s.m)) }
func (s rpointermap) Cap() int64      { return int64(len(s.m)) * 2 }
func (s rpointermap) IsMap() bool     { return true }
func (s rpointermap) IsSet() bool     { return false }
func (s rpointermap) IsList() bool    { return false }
func (s rpointermap) IsOrdered() bool { return false } // This may change!! Depends on presence of an ordering function
func (s rpointermap) IsSorting() bool { return false } // This may change!! Depends on presence of an ordering function in rorderedstringmap
func (s rpointermap) IsCardOk() bool  { return s.Length() >= s.MinCard() && s.Length() <= s.MaxCard() }
func (o rpointermap) IsZero() bool {
	return o.Length() == 0
}

func (s rpointermap) KeyType() *RType {
	return s.ElementType()
}

func (s rpointermap) ValType() *RType {
	return s.valType
}

func (s *rpointermap) Get(key RObject) (val RObject, found bool) {
	val, found = s.m[key]
	return
}


func (s *rpointermap) Put(key RObject, val RObject, context MethodEvaluationContext) (added bool, newLen int) {	

	_,found := s.m[key] 
	added = ! found
    s.m[key] = val
    newLen = len(s.m)
    return
}

func (s *rpointermap) Contains(th InterpreterThread, key RObject) (found bool) {
	_,found = s.m[key] 
	return
}

/*
	This version of the put method does not sort. It assumes that it is being called with key and val objects
	already known to be simply insertable (at the end of if applicable) the collection.
	Used by the persistence service. Do not use for general use of the collection.
*/
func (s *rpointermap) PutSimple(key RObject, val RObject) (newLen int) {	

    s.m[key] = val
    newLen = len(s.m)
    return
}

func (s *rpointermap) Remove(th InterpreterThread, key RObject) (removed bool, removedIndex int) {

	_,removed = s.m[key] 
	delete(s.m, key) 
	removedIndex = -1
	return
}

func (s *rpointermap) ClearInMemory() {

	s.m =  make(map[RObject]RObject)
}

func (c *rpointermap) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Map unless the key-type is String.")			
	return 
}

/*
func (s *rpointermap) Get(key RObject) (val RObject, found bool) {
	unit := key.(*runit)
	val, found = s.m[unit]
	return
}
*/


/*
==========================================================================================

Used to get a reference to the database for the Iterable method.
Relies on the fact that the whole web request should be handled inside a single
database transaction, so no need to manage transactions or db access for the iteration.
*/
type FakeInterpreterThread int 

func (f FakeInterpreterThread) Package() *RPackage {
	return nil
}

/*
The executing method.
*/
func (f FakeInterpreterThread) Method() *RMethod {
	return nil
}

/*
Returns the package from which the currently executing method was called, 
or nil if at stack bottom.
*/
func (f FakeInterpreterThread) CallingPackage() *RPackage {
	return nil
}

/*
Returns the method that called the currently executing method, 
or nil if at stack bottom.
*/
func (f FakeInterpreterThread) CallingMethod() *RMethod {
	return nil
}



/*
A db connection thread. Used to serialize access to the database in a multi-threaded environment,
and to manage database transactions.
*/
func (f FakeInterpreterThread) DBT() DBT {
   return RT.DBT()	
}

func (f FakeInterpreterThread) Err() string {
	return ""
}

/*
deadlock debug version
func (f FakeInterpreterThread) AllowGC(msg string)  {
}
*/
/*
deadlock debug version
func (f FakeInterpreterThread) DisallowGC(msg string)  {
}
*/
/*
*/
func (f FakeInterpreterThread) AllowGC()  {
}

/*
*/
func (f FakeInterpreterThread) DisallowGC()  {
}

func (f FakeInterpreterThread) GC() {
}

func (f FakeInterpreterThread) EvaluationContext() MethodEvaluationContext {
	return nil
}

/*
  The currently active DB transaction which this thread started or is participating in.
*/
func (f FakeInterpreterThread) Transaction() *RTransaction {
   return nil
}

func (f FakeInterpreterThread) SetTransaction(tx *RTransaction) {
}

func (f FakeInterpreterThread) HttpClient() *http.Client {
	return nil
}
