// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   robject.go -  relish data objects - this source file defines
                 generic object stuff + unitary (non-collection) objects 


*/

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"io"
	"crypto/rand"
	"strings"
	. "relish/dbg"	
)

///////////////////////////////////////////////////////////////////////////
////////// DATA OBJECTS (Typed)
///////////////////////////////////////////////////////////////////////////

/*
Represents an instance of an object in the Relish data model.
Objects have a type
They can have attributes and participate in relations.
They may be a collection or a unit.

TODO: It would be nice to separate the RObject interface out, creating a sub-interface
called Storable which has all the stuff concerned with persistence.
Because right now, all the relish primitive value types have a whole bunch of spurious
methods with dummy implementations.

Or alternatively, a super-interface RValue could be the core interface, and RObject
would have the persistance-specific methods. RValue would still have flags. All methods
which accept or return RObject or []RObject now would then accept or return RValue or []RValue
*/
type RObject interface {
	Type() *RType
	This() RObject // A reference to the most specific (outer wrapping) struct or the primitive 
	IsUnit() bool
	IsCollection() bool
	
	/*
	rivers1
	2502956568
	
	If this is a RCollection, return a slice or map of the elements, depending on the subtype of collection.
	If not a collection,an error is returned as the second return value.
    Can be used by go's template range action to iterate through collection in a template.	

    Note. This needs to fetch all Proxy objects in the collection before returning the slice.
    Or wait a minute...
	*/
	Iterable() (sliceOrMap interface{}, err error)

    /*
    A basic string representation of the object.
    */
	String() string

    /*
    A basic string representation of the object.
    This variant allows the object or its sub-parts to be fetched from persistence within the current thread and its db lock.
    This prevents a db mutex contention.
    */
	StringTh(th InterpreterThread) string

    /*
    A string containing detailed low-level information about the object.
    */
	Debug(th InterpreterThread) string

	Flags() int8 // version of flags byte suitable for storing in the db RObject row

	/*
	   Whether this object has a uuid yet.
	*/
	HasUUID() bool

	/*
	   Return the UUID. Assumes the object has one. Returns a 16-byte slice.
	*/
	UUID() []byte

	/*
	   Create if needed and return the object's 16-byte UUID. Uses /dev/urandom
	   An error will be returned if object had no uuid and /dev/urandom cannot be read.
	*/
	EnsureUUID() (theUUID []byte, err error)

	/*
	   Return the object's uuid as two uint64s. Assumes object has a UUID.
	*/
	UUIDuint64s() (id uint64, id2 uint64)

	/*
	   Return the two uint64s representation of the UUID. Creates a UUID if object doesn't have one.
	*/
	EnsureUUIDuint64s() (id uint64, id2 uint64, err error)

	/*
	   Return the string representation of the UUID. Assumes object has a UUID.
	*/
	UUIDstr() string

	/*
	   Return the string representation of the UUID. Creates a UUID if object doesn't have one.
	*/
	EnsureUUIDstr() (uuidstr string, err error)

	/*
	   Return an abbreviated string representation of the UUID. Assumes object has a UUID.
	*/
	UUIDabbrev() string

	/*
	   Return an abbreviated string representation of the UUID. Creates a UUID if object doesn't have one.
	*/
	EnsureUUIDabbrev() (uuidstr string, err error)

	/*
	   Sets to nil the UUID value in the in-memory object instance. You should never call this. Only the
	   persistence subsystem should call this.
	*/
	RemoveUUID()

	/*
	   ID to be used as the object's record ID in the local SQLite database.
	*/
	DBID() int64

	/*
	   Whether the object is modified since last committed to db.
	*/
	IsBeingStored() bool
	SetBeingStored()
	ClearBeingStored()

	/*
	   Whether the object's uuid is reversed in the db object (i.e. bytes 8-15 of uuid are the db id).
	*/
	IsIdReversed() bool
	SetIdReversed()
	ClearIdReversed()

	IsLoadNeeded() bool
	SetLoadNeeded()
	ClearLoadNeeded()

	/*
	   The object meets its class invariant constraints and attribute and relation cardinality constraints.
	*/
	IsValid() bool
	SetValid()
	ClearValid()
	
	IsMarked() bool 
	SetMarked()    
	ClearMarked()  
	ToggleMarked()  

	/*
	If the object is not already marked as reachable, flag it as reachable.
	Return whether we had to flag it as reachable. false if was already marked reachable.
	Is a no-op that returns false for value-types that don't need marking.
	*/
	Mark() bool 

	/*
	   The object is stored in the local sqlite database.
	*/
	IsStoredLocally() bool
	SetStoredLocally()
	ClearStoredLocally()

	/*
	   Whether this is a proxy storing only the dbid of the real object. If so, a
	   collection[i] = db.Fetch(collection[i]) must be used to get the real object into memory,
	   and fetch should use a runtime memcache to make sure we only have one copy of the object.
	*/
	IsProxy() bool

	/*
	   Is this object unstorable and un-streamable?	
	   Currently RMethods and RMultiMethods are transient.
	*/
	IsTransient() bool

    /*
    zero is defined more widely than just for numbers. This is similar to languages like
    Go and Python and Lisp.
    Empty lists, sets, and maps are zero.
    nil is zero.
    0 is zero
    0.0 is zero
    false is zero

    zero-ness is what is tested for in ifs, whiles, and by the and, or, not functions.
    */
	IsZero() bool
	
   /*
      Converts the value or object or collection to an value tree suitable for encoding into JSON.
      If this ROBJECT is a primitive value, a value of the corresponding Go primitive type is
      returned as tree.
      If this is a structured object, a map from attribute name to tree-ified value is returned.
      if this is a relish map, then if the key type is String, the map is converted to a 
      Go map from string to tree-ified mapped values.
      If this is a relish map with non-String key type, an error is returned. Such a map cannot
      be JSON encoded.
      If this is any other type of relish collection, a slice of the tree-ified elements of the collection
      is returned.
      
      The includePrivate flag determines whether private attributes of a structured object are included
      in the returned map/tree.
   */ 	
   ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) 

   /*
   Populate a tree of relish objects and attribute values given a map/list tree of Go values.
   Typically, the tree of Go values has come from a JSON unmarshalling.
   Allocates new objects as needed.
   The object that this method is called on serves as a prototype for the root of the object
   tree to be created. This object may be re-used as the returned object, or may be thrown
   away after its type information is used to guide construction of the object tree.
   In any case, you should not rely on modification (or non-modification) of the target
   object of this method, but should use the returned object as the root of the JSON-populated
   relish-object tree.
   */
   FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error)
	
}

/*
   Abstract
*/
type robject struct {
	rtype *RType
	uuid  []byte // will be 16 bytes
	this  RObject
	flags byte
	transaction *RTransaction  // which db transaction this is dirty in, or nil
}

// Could it be that dirty objects are just a list in the RuntimeEnv instead?

const FLAG_BEING_STORED = 0x1       // Begun to store it in the db. Not necessarily finished.
const FLAG_LOAD_NEEDED = 0x2 // object state must be reloaded from db because has uuid but never loaded 
// or because is dirty and transaction was aborted
const FLAG_VALID = 0x4 // Object has all valid attribute values and ok attribute and relationship cardinality
const FLAG_UNUSED_1 = 0x8
const FLAG_UNUSED_2 = 0x10
const FLAG_MARKED = 0x20 // Toggle reachable directly or indirectly from a thread stack or a constant
const FLAG_STORED_LOCALLY = 0x40 // Object is pesisted in local database
const FLAG_ID_REVERSED = 0x80 // The uuid is stored reversed in the db object (db id=bytes 8-15 of uuid).

/*
TODO We will have to see how to handle this. nil should be zero. But perhaps we should be able to 
define type-specific zero functions.
*/
func (o robject) IsZero() bool {
	return false
}

func (o robject) This() RObject {
	return o.this
}

func (o robject) Flags() int8 {
	return int8(o.flags)
}

func (o robject) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}



func (o robject) IsBeingStored() bool { return o.flags&FLAG_BEING_STORED != 0 }
func (o *robject) SetBeingStored()    { o.flags |= FLAG_BEING_STORED }
func (o *robject) ClearBeingStored()  { o.flags &^= FLAG_BEING_STORED }

func (o robject) IsLoadNeeded() bool { return o.flags&FLAG_LOAD_NEEDED != 0 }
func (o *robject) SetLoadNeeded()    { o.flags |= FLAG_LOAD_NEEDED }
func (o *robject) ClearLoadNeeded()  { o.flags &^= FLAG_LOAD_NEEDED }

func (o robject) IsValid() bool { return o.flags&FLAG_VALID != 0 }
func (o *robject) SetValid()    { o.flags |= FLAG_VALID }
func (o *robject) ClearValid()  { o.flags &^= FLAG_VALID }

func (o robject) IsMarked() bool { return o.flags&FLAG_MARKED != 0 }
func (o *robject) SetMarked()    { o.flags |= FLAG_MARKED }
func (o *robject) ClearMarked()  { o.flags &^= FLAG_MARKED }
func (o *robject) ToggleMarked()  { o.flags ^= FLAG_MARKED }

/*
If the object is not already marked as reachable, flag it as reachable.
Return whether we had to flag it as reachable. false if was already marked reachable.
*/
func (o *robject) Mark() bool { 
   if o.IsMarked() == markSense {
       Logln(GC3_,"Mark(): Already marked with",markSense)	
   	   return false
   } 
   o.ToggleMarked()
   Logln(GC3_,"Mark(): Marked with",o.IsMarked())

   // COMMENT OUT NEXT 3 LINES FOR SPEED INCREASE IF DO NOT NEED MEMORY LEAK DEBUGGING SUPPORT ENABLED
   if mappingMemory {
      instanceMap[o.rtype]++
   }
   return true
}

func (o robject) IsIdReversed() bool { return o.flags&FLAG_ID_REVERSED != 0 }
func (o *robject) SetIdReversed()    { o.flags |= FLAG_ID_REVERSED }
func (o *robject) ClearIdReversed()  { o.flags &^= FLAG_ID_REVERSED }

func (o robject) IsStoredLocally() bool { return o.flags&FLAG_STORED_LOCALLY != 0 }
func (o *robject) SetStoredLocally()    { o.flags |= FLAG_STORED_LOCALLY }
func (o *robject) ClearStoredLocally()  { o.flags &^= FLAG_STORED_LOCALLY }

func (o *robject) IsProxy() bool { return false }

func (o *robject) IsTransient() bool { return false }

func (o *robject) String() string {
   return o.StringTh(nil)
}

func (o *robject) StringTh(th InterpreterThread) string {
	var id string
	if o.HasUUID() {
		//id = o.UUIDabbrev()
		id = strconv.FormatInt(o.DBID(), 10)
	} else {
		id = strconv.FormatUint(o.LocalID(), 10)
	}
	return fmt.Sprintf("%v:%v", o.Type(), id)
}


func (o *robject) Debug(th InterpreterThread) string {
	return fmt.Sprintf("%s@%p",o.StringTh(th),o)
}


/*
   Returns the localId. Creates the local id if it does not yet exist.
   An object should not have both a local id and a UUID.
*/
func (o robject) LocalID() uint64 {
	localId, found := RT.objectIds[o.this]
	if !found {
		localId = RT.NextLocalID()
		RT.objectIds[o.this] = localId
	}
	return localId
}

/*
   Whether this object has a uuid yet.
*/
func (o robject) HasUUID() bool {
	return o.uuid != nil
}

/*
   Return the UUID. Assumes the object has one. Returns a 16-byte slice.
*/
func (o robject) UUID() []byte {
	return o.uuid
}


/*
   Create if needed and return the object's 16-byte UUID. Uses /dev/urandom
   An error will be returned if object had no uuid and /dev/urandom cannot be read.
*/
func (o *robject) EnsureUUID() (theUUID []byte, err error) {
	if o.uuid == nil {
//		var f *os.File
//		f, err = os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
//		defer f.Close()
//		if err != nil {
//			return
//		}
		aUUID := make([]byte, 16)
//		var n int
		
		_,err = io.ReadFull(rand.Reader, aUUID)
		//n, err = f.Read(theUUID)
		if err != nil {
			return
		}
//		if n != 16 {
//			err = errors.New(fmt.Sprintf("When creating UUID could not read 16 random bytes from /dev/urandom. Read %v bytes.",n))
//			return
//		}
		o.uuid = aUUID
		delete(RT.objectIds, o.this) // Remove the local id from the runtime.
		return
	}
	theUUID = o.uuid
	return
}


/*
   Create if needed and return the object's 16-byte UUID. Uses /dev/urandom
   An error will be returned if object had no uuid and /dev/urandom cannot be read.

func (o *robject) EnsureUUID() (theUUID []byte, err error) {
	if o.uuid == nil {
		var f *os.File
		f, err = os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
		defer f.Close()
		if err != nil {
			return
		}
		aUUID := make([]byte, 16)
		var n int
		n, err = f.Read(aUUID)
		if err != nil {
			return
		}
		if n != 16 {
			err = errors.New(fmt.Sprintf("When creating UUID could not read 16 random bytes from /dev/urandom. Read %v bytes.",n))
			return
		}
		o.uuid = aUUID
		delete(RT.objectIds, o.this) // Remove the local id from the runtime.
		return
	}
	theUUID = o.uuid
	return
}
*/

/*
   Return the object's uuid as two uint64s. Assumes object has a UUID.
*/
func (o robject) UUIDuint64s() (id uint64, id2 uint64) {
    if o.uuid == nil {
	    panic(fmt.Sprintf("Object %s uuid is nil !",&o))
    }
	buf := bytes.NewBuffer(o.UUID())
	err := binary.Read(buf, binary.BigEndian, &id)
	if err != nil {
		panic(err)
	}
	err = binary.Read(buf, binary.BigEndian, &id2)
	if err != nil {
		panic(err)
	}
	return
}

/*
   Return the two uint64s representation of the UUID. Creates a UUID if object doesn't have one.
*/
func (o *robject) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	_, err = o.EnsureUUID()
	if err != nil {
		return
	}
	id, id2 = o.UUIDuint64s()
	return
}

/*
   Return the string representation of the UUID. Assumes object has a UUID.
*/
func (o robject) UUIDstr() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", o.uuid[0:4], o.uuid[4:6], o.uuid[6:8], o.uuid[8:10], o.uuid[10:])
}

/*
   Return an abbreviation of the string representation of the UUID. Assumes object has a UUID.
*/
func (o robject) UUIDabbrev() string {
	return fmt.Sprintf("%x..%x", o.uuid[0:4], o.uuid[14:])
}

/*
   Return the string representation of the UUID. Creates a UUID if object doesn't have one.
*/
func (o *robject) EnsureUUIDstr() (uuidstr string, err error) {
	_, err = o.EnsureUUID()
	if err != nil {
		return
	}
	uuidstr = o.UUIDstr()
	return
}

/*
   Return the string representation of the UUID. Creates a UUID if object doesn't have one.
*/
func (o *robject) EnsureUUIDabbrev() (uuidstr string, err error) {
	_, err = o.EnsureUUID()
	if err != nil {
		return
	}
	uuidstr = o.UUIDabbrev()
	return
}

func (o *robject) RemoveUUID() {
	o.uuid = nil
}

func (o robject) DBID() int64 {
	id, id2 := o.UUIDuint64s()
	if o.IsIdReversed() {
		return int64(id2)
	}
	return int64(id)
}

func DBID(id1 int64, id2 int64, flags int) int64 {
	if flags&FLAG_ID_REVERSED != 0 {
		return id2
	}
	return id1
}

/*
   Used when summoning an object from the database. The object has already been
   created in memory. Now its contents have to be restored from the DB.
   This method accepts the object's id,id2,and flags from the database and
   restores this object's UUID and flags.
*/
func (o *robject) RestoreIdsAndFlags(id, id2 int64, flags int) {
	theUUID := make([]byte, 0, 16)

	wasMarked := o.IsMarked()  // Make sure that marked flag stored in db is ignored.
	o.flags = byte(flags)
	if wasMarked {
		o.SetMarked()
	} else {
		o.ClearMarked()
	}

    o.SetBeingStored() // Kludge so that old db data without this flag set is ok.

	if o.IsIdReversed() {
		id, id2 = id2, id
	}
	uid := uint64(id)
	uid2 := uint64(id2)

	// write id, then id2 into the byte array.
	buf := bytes.NewBuffer(theUUID)

	err := binary.Write(buf, binary.BigEndian, uid)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buf, binary.BigEndian, uid2)
	if err != nil {
		panic(err)
	}

	o.uuid = buf.Bytes()
	delete(RT.objectIds, o.this) // Remove the local id from the runtime. OOPS! Probably should never be a local id in this case!!!
}

/*
   Used when summoning an object from the database. The object has already been
   created in memory. Now its contents have to be restored from the DB.
   This method accepts the object's id,id2,and flags from the database and
   restores this object's UUID and flags.

   Also has methods for setting and reading which db transaction the object is currently dirty in.
*/
type Persistable interface {
	RestoreIdsAndFlags(id, id2 int64, flags int)
	SetTransaction(tx *RTransaction)  // Set to nil to end association with a transaction
	Transaction() *RTransaction  // Transaction in which this is dirty, or nil, or RolledBackTransaction
	RollBack()  // set the transaction to the RolledBackTransaction.
	IsRolledBack() bool
	This() RObject
	Refresh(th InterpreterThread) error  // Re-fetch the object's state and associations from the database. Clear IsRolledBack() status.
}

func (o robject) Type() *RType { return o.rtype }



func (o *robject) SetTransaction(tx *RTransaction) {
   o.transaction = tx
   if tx != nil && tx != RolledBackTransaction {
   	   tx.DirtyObjects[o.This().(Persistable)] = true
   }
}

func (o *robject) Transaction() *RTransaction {
   return o.transaction
}

func (o *robject) RollBack() {
	if o.IsBeingStored() && ! o.IsStoredLocally() {
//		fmt.Println("Was not stored locally flag.")
		o.ClearBeingStored()
		o.SetTransaction(nil)
	} else {
//	   fmt.Println("Was stored locally flag.")	
//	   fmt.Println(o.IsBeingStored(),o.IsStoredLocally())	
	   o.transaction = RolledBackTransaction
    }
}

func (o *robject) IsRolledBack() bool {
	return o.transaction == RolledBackTransaction
}

/* Can only be called while txOpsMutex is locked.
*/
func (o *robject) Refresh(th InterpreterThread) (err error) {
	// TODO refresh attributes from database.


	o.SetTransaction(nil)
    o.ClearLoadNeeded()
	return
}


/*
A unitary object. i.e. Not a collection.
*/
type runit struct {
	robject
	attrs []RObject
}

func (o *runit) Mark() bool {
	if (&(o.robject)).Mark() {
		for _,val := range o.attrs {
			if val != nil {
			   val.Mark()
		    }
		}
		return true
	}
	return false
}

/*
Create the array of references to attribute values.
*/
func (r *runit) initialize(n int) {
   r.attrs = make([]RObject,n)
}

/* Can only be called while txOpsMutex is locked.
   Refresh attributes from database.
*/
func (u *runit) Refresh(th InterpreterThread) (err error) {
	n := len(u.attrs)
	for i := 0; i < n; i++ {
		u.attrs[i] = nil
	}

    th.DBT().Refresh(u,1)   
	

	
	if err == nil {
		err = u.robject.Refresh(th)
	}
	return err
}


func (o *runit) String() string {
	return (&(o.robject)).String()
}

func (o *runit) StringTh(th InterpreterThread) string {
	return (&(o.robject)).StringTh(th)
}

func (o *runit) Debug(th InterpreterThread) string {
	return (&(o.robject)).Debug(th)
}

func (u runit) IsUnit() bool {
	return true
}

func (u runit) IsCollection() bool {
	return false
}


/*
Convert object to a map of strings (attribute names) to attribute values,
with attribute values themselves converted to maps and slices.
*/
func (o *runit) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
	
   // fmt.Println("runit.ToMapListTree",o.String())
	
	// Record myself in the visited set so we don't infinitely loop.
	visited[o] = true
	
	// Go through all the attributes and map-tree-ify their values
		
	m := make(map[string]interface{})
	
	var key string
	var val interface{}
	
	for _, attr := range o.rtype.Attributes {

        if includePrivate || attr.IsPublicReadable() {
            key = attr.Part.Name

			  value, found := RT.AttrVal(th, o, attr)
			  // fmt.Println(attr.Part.Name,":",value)			
			  if !found {
			     //fmt.Println("attr val not found")
				  value = NIL // TODO Decide how to really handle these !
			  } else if visited[value] {
	           continue
	        }			
			
			 val, err = value.ToMapListTree(th, includePrivate, visited)	
			 if err != nil {
			 	return
			 }
			 m[key] = val
		 }
	}

	for _, typ := range o.rtype.Up {
		for _, attr := range typ.Attributes {
	        if includePrivate || attr.IsPublicReadable() {
	            key = attr.Part.Name

				value, found := RT.AttrVal(th, o, attr)
				//fmt.Println(attr.Part.Name,":",value)
				if !found {
			      //fmt.Println("attr val not found")				   
				   value = NIL // TODO Decide how to really handle these !
 			   } else if visited[value] {
 	            continue
 	         }
					
				val, err = value.ToMapListTree(th, includePrivate, visited)	
				if err != nil {
					return
				}
				m[key] = val
			}
		}
	}
	
	tree = m
	
	return	
}


func (o *runit) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
	myType := o.Type()
	
	obj = o
	
	var relishVal RObject	
	
	switch tree.(type) {
	case map[string]interface{}:
		theMap := tree.(map[string]interface{})
		
        bestType, nMatchingAttrs := myType.BestMatchingSubtype(theMap) 			
		
		if nMatchingAttrs == 0 {
		   err = errors.New("Unmarshalling target object's type is not compatible with JSON map-keys.")	
		}
		
		// re-use the target object, unless the best type is a subtype, in which case,
		// create a new instance of the subtype.
		if bestType != myType {
			obj, err = RT.NewObject(bestType.Name) 
			if err != nil {
				return
			}
		}
		
		for key, value := range theMap {
			
			// Find the attribute spec given the rtype of o and the key.
			
			attr, found := bestType.GetAttribute(key)
			
			if ! found {
				continue
			}
			
			// check if the object already has a value of that attribute.
			
            val, valFound := RT.AttrVal(th, obj, attr) 
            if valFound {
	           relishVal, err = val.FromMapListTree(th, value) 
  	     	   if err != nil {
				  return
			   }	
	           RT.RestoreAttr(obj, attr, relishVal)
            } else {	
			
			   // if not, ... instantiate an instance of the attribute value type 
			  
			   var prototypeVal RObject
			   var relishVal RObject
			 
			   attrValType := attr.Part.Type
		
		      
			   // Is it a collection type? Only handling lists so far. TODO others

			   if attrValType.LessEq(ListType) {
                  prototypeVal, err = RT.Newrlist(attrValType.ElementType(),0,-1,nil,nil,nil)
				  if err != nil {
					 return
				  }	                  
               } else if attrValType.IsPrimitive {    // Is it a primitive type?
	              prototypeVal = attrValType.Zero()
               } else {			
				  prototypeVal, err = RT.NewObject(attrValType.Name) 
				  if err != nil {
					 return
				  }			
		       }

	           relishVal, err = prototypeVal.FromMapListTree(th, value) 
  	     	   if err != nil {
				  return
			   }	
	           RT.RestoreAttr(obj, attr, relishVal)		
			}
		}
	default:
		err = errors.New("Cannot restore a structured object except from a map.")
		return
	}
	return 
}






/*
Create a new unitary object.
*/

func (rt *RuntimeEnv) NewObject(typeName string) (RObject, error) {
	
	// fmt.Println("NewObject:",typeName)
	
	/*
    switch typeName {	
    case "Time": get rid od this switch at this level
    case "Channel":
    case "String":
    case "Int":
    case "Int32":
    case "Uint":
    case "Uint32": 
    default:
    	*/

	// TODO Need to handle parameterized types here
        TypesMutex.RLock()
	typ, found := rt.Types[typeName]
        TypesMutex.RUnlock()
	if !found {
		return nil, fmt.Errorf("Type '%s' not found.", typeName)
	}
	
	// Detect primitive types
    switch typ {
    case TimeType:
        var t RTime  // this is a zero Time value
        return t, nil    	
    case StringType:
    	var s String
    	return s, nil
    case IntType:
    	var i Int
    	return i, nil
    case Int32Type:
    	var i32 Int32
    	return i32, nil
    case UintType:
    	var u Uint
    	return u, nil
    case Uint32Type:
    	var u32 Uint32
    	return u32,nil
    case FloatType:
    	var f Float
    	return f, nil
    case BoolType:
    	var b Bool
    	return b, nil
    case MutexType:
    	return &Mutex{}, nil
    case OwnedMutexType:
    	return &OwnedMutex{}, nil    	
    case RWMutexType:
    	return &RWMutex{}, nil    
	
   case BitType:
	   var b Bit
	   return b, nil
	
   case ByteType:
	   var b Byte
	   return b, nil	
	
   case CodePointType:
	   var c CodePoint
	   return c, nil
	
   case BitsType:
	  fallthrough
   case BytesType:
      fallthrough	
   case CodePointsType:
	   return typ.Zero(), nil		
	    	
   default:
	   if typ.IsNative {
		   // w := &GoWrapper{nil,typ,0}

	       w := &GoWrapper{runit{robject{rtype: typ},nil},nil}

           w.initialize(typ.TotalAttributeCount)

  	      if ! markSense {
  		      w.SetMarked()
  	      }
  	      Logln(GC3_,"NewObject: IsMarked",w.IsMarked())		  
	      return w, nil		
	   }
      if typ.IsParameterized {
    	   if typ == ChannelType || strings.HasPrefix(typeName, "Channel of ") {  // TODO May need to change to just == "Channel"
           c := &Channel{} // this is an uninitialized Channel value
           return c, nil   
    	   }
    	   // TODO Not handling other parameterized types yet.       	   
        return nil, fmt.Errorf("Sorry. Parameterized data types are not handled yet.")
    	   // TODO Not handling other parameterized types yet.
       } 
    }
	// It's not a primitive type nor a parameterized type

	unit := &runit{robject{rtype: typ},nil}

// NEED TO FIND OUT HOW BIG TO MAKE THE INSTANCE HERE ! HOW MANY ATTRIBUTES
///////////////////////////
//////////////////////////
/////////////////////////	
////////////////////////
///////////////////////
//////////////////////
/////////////////////
////////////////////
///////////////////	
    unit.initialize(typ.TotalAttributeCount)
    

	unit.robject.this = unit
	if ! markSense {
		unit.SetMarked()
	}
    if Logging(GC_) {
       n_runits_ever ++
	   Logln(GC3_,"NewObject: IsMarked",unit.IsMarked())	
    }
	return unit, nil
}
