// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   rprimitive.go - Primitive data types in relish. These all implement the RObject interface.
                   Most of them are simple value types, like numbers, strings, boolean etc.
                   Others are included too, like Channel, GoWrapper which represent relish object
                   instances, like File instances, that are wrappers around a Go-native object instance.
*/

import ( 
	     "strconv"
         "errors"
         . "time"
         "fmt"
         "sync"
         "strings"
	      . "relish/dbg"	         
)


/*
STRATEGY FOR PRIMITIVES

This is a rubbishy rambling comment

See if the following is possible:

Create an RUnit interface, also an RTyped interface??

Now start defining primitive types in go. Make sure they implement RObject, RTyped, RUnit

NOTE!!!! Primitives cannot be persisted by themselves.
So they cannot have a UUID.
They must be part of something else.

Is this true?

What about global (universal) URN-named constant values.

Or global  (universal) URN-named variables of primitive type:
e.g. GlobalStore.Get("Canada Debt Clock")          (a uint64 i.e. Uint value)

I think in this special case we may need a boxing wrapper type for each primitive type?
ONLY for the purpose of persisting it independently. NEED TO THINK FURTHER ABOUT THIS!!!!
*/

var PrimitiveType *RType
var IntOrStringType *RType
var NumericType *RType
var IntegerType *RType
var IntType *RType
var Int32Type *RType
var Int16Type *RType
var Int8Type *RType
var UintType *RType
var Uint32Type *RType
var Uint16Type *RType
var ByteType *RType
var BitType *RType
var BoolType *RType
var CodePointType *RType
var RealType *RType
var FloatType *RType
var Float32Type *RType
var ComplexNumberType *RType
var ComplexType *RType
var Complex32Type *RType
var TextType *RType
var StringType *RType
var TimeType *RType
var ProxyType *RType

var CallableType *RType
var MultiMethodType *RType
var MethodType *RType
var ClosureType *RType

var AnyType *RType
var NothingType *RType
var NonPrimitiveType *RType
var StructType *RType
var CollectionType *RType
var SetType *RType
var ListType *RType
var MapType *RType

var InChannelType *RType
var OutChannelType *RType
var ChannelType *RType
var MutexType *RType
var OwnedMutexType *RType
var RWMutexType *RType


var ListOfAnyType *RType
var SetOfAnyType *RType

var ListOfStringType *RType
var SetOfStringType *RType

var ListOfNumericType *RType
var SetOfNumericType *RType

var ListOfFloatType *RType
var SetOfFloatType *RType

var ListOfIntType *RType
var SetOfIntType *RType

var ListOfUintType *RType
var SetOfUintType *RType

var ListOfInt32Type *RType
var SetOfInt32Type *RType

var ListOfUint32Type *RType
var SetOfUint32Type *RType


// Slice Types
var SliceType *RType
var BytesType *RType
var CodePointsType *RType
// var IntsType *RType
var BitsType *RType

// IO library Types
var WriterType *RType
var ReaderType *RType
var WriterAtType *RType
var ReaderAtType *RType
var SeekerType *RType
var CloserType *RType
var WriteCloserType *RType
var WriteSeekerType *RType
var ReadCloserType *RType
var ReadSeekerType *RType
var ReadWriterType *RType
var ReadWriteCloserType *RType
var ReadWriteSeekerType *RType
var FileType *RType

func (rt *RuntimeEnv) createPrimitiveTypes() {

	PrimitiveType, _ = rt.CreateType("RelishPrimitive", "", []string{})
	IntOrStringType,_ = rt.CreateType("IntOrString", "", []string{"RelishPrimitive"})
	NumericType, _ = rt.CreateType("Numeric", "", []string{"RelishPrimitive"})
	IntegerType, _ = rt.CreateType("Integer", "", []string{"Numeric"})
	IntType, _ = rt.CreateType("Int", "", []string{"Integer","IntOrString"})
	Int32Type, _ = rt.CreateType("Int32", "", []string{"Integer"})
	Int16Type, _ = rt.CreateType("Int16", "", []string{"Integer"})
	Int8Type, _ = rt.CreateType("Int8", "", []string{"Integer"})
	UintType, _ = rt.CreateType("Uint", "", []string{"Integer"})
	Uint32Type, _ = rt.CreateType("Uint32", "", []string{"Integer"})
	Uint16Type, _ = rt.CreateType("Uint16", "", []string{"Integer"})
	ByteType, _ = rt.CreateType("Byte", "", []string{"Integer"})
	BitType, _ = rt.CreateType("Bit", "", []string{"Integer"})
	BoolType, _ = rt.CreateType("Bool", "", []string{"RelishPrimitive"})
	TextType, _ = rt.CreateType("Text", "", []string{"RelishPrimitive"})
	CodePointType, _ = rt.CreateType("CodePoint", "", []string{"Text"})
	RealType, _ = rt.CreateType("Real", "", []string{"Numeric"})
	FloatType, _ = rt.CreateType("Float", "", []string{"Real"})
	Float32Type, _ = rt.CreateType("Float32", "", []string{"Real"})
	ComplexNumberType, _ = rt.CreateType("ComplexNumber", "", []string{"Numeric"})
	ComplexType, _ = rt.CreateType("Complex", "", []string{"ComplexNumber"})
	Complex32Type, _ = rt.CreateType("Complex32", "", []string{"ComplexNumber"})
	StringType, _ = rt.CreateType("String", "", []string{"Text","IntOrString"})
	TimeType, _ = rt.CreateType("Time", "", []string{"RelishPrimitive"})
	ProxyType, _ = rt.CreateType("Proxy", "", []string{})

	CallableType, _ = rt.CreateType("Callable", "", []string{"Text"})
	MultiMethodType, _ = rt.CreateType("MultiMethod", "", []string{"Callable"})
	MethodType, _ = rt.CreateType("Method", "", []string{"Callable"})
	ClosureType, _ = rt.CreateType("Closure", "", []string{"Callable"})	

	AnyType, _ = rt.CreateType("Any", "", []string{})
	NothingType, _ = rt.CreateType("Nothing", "", []string{})	
	NonPrimitiveType, _ = rt.CreateType("NonPrimitive", "", []string{})
	StructType, _ = rt.CreateType("Struct", "", []string{})
	CollectionType, _ = rt.CreateType("Collection", "", []string{})
	SetType, _ = rt.CreateType("Set", "", []string{"Collection"})	
	ListType, _ = rt.CreateType("List", "", []string{"Collection"})	
	MapType, _ = rt.CreateType("Map", "", []string{"Collection"})	

	// Do a need Iterable and InChannel <: Iterable, Collection <: Iterable
	
	InChannelType, _ = rt.CreateType("InChannel", "", []string{})
	OutChannelType, _ = rt.CreateType("OutChannel", "", []string{})
	ChannelType, _ = rt.CreateType("Channel", "", []string{"InChannel","OutChannel"})	
	ChannelType.IsParameterized = true	

	MutexType, _ = rt.CreateType("Mutex", "", []string{})

	OwnedMutexType, _ = rt.CreateType("OwnedMutex", "", []string{})

	RWMutexType, _ = rt.CreateType("RwMutex", "", []string{})	



//	FileType,_ = rt.CreateType("shared.relish.pl2012/lib/core/pkg/io/File", "", []string{"Any"})

    // Primitive collection types

	ListOfAnyType, _ = rt.GetListType(AnyType)
	SetOfAnyType, _  = rt.GetSetType(AnyType)	

	ListOfStringType, _ = rt.GetListType(StringType)
	SetOfStringType, _  = rt.GetSetType(StringType)	

	ListOfNumericType, _ = rt.GetListType(NumericType)
	SetOfNumericType, _  = rt.GetSetType(NumericType)	

	ListOfFloatType, _ = rt.GetListType(FloatType)
	SetOfFloatType, _  = rt.GetSetType(FloatType)			

	ListOfIntType, _ = rt.GetListType(IntType)
	SetOfIntType, _  = rt.GetSetType(IntType)	

	ListOfInt32Type, _ = rt.GetListType(Int32Type)
	SetOfInt32Type, _  = rt.GetSetType(Int32Type)		

	ListOfUintType, _ = rt.GetListType(UintType)
	SetOfUintType, _  = rt.GetSetType(UintType)		

	ListOfUint32Type, _ = rt.GetListType(Uint32Type)
	SetOfUint32Type, _  = rt.GetSetType(Uint32Type)		

    // Slice types

	SliceType, _ = rt.CreateType("Slice", "", []string{"RelishPrimitive"})    
	BytesType, _ = rt.CreateType("Bytes", "", []string{"Slice"})    
	CodePointsType, _ = rt.CreateType("CodePoints", "", []string{"Slice"})    
// 	IntsType, _ = rt.CreateType("Ints", "", []string{"Slice"})    
	BitsType, _ = rt.CreateType("Bits", "", []string{"Slice"})    
	
}


/*
Return a zero-value of the type.
For structured object types, this is currently defined to be NIL which is a Go constant in this data package.
NIL is of go-type Nil, of relish RType NothingType. 
NIL's name in relish source code is nil, and its string representation when printed is *nil* 
*/
func (t *RType) Zero() RObject {
	var z RObject
    switch t {
	case PrimitiveType:
		z = Int(0)	
	case NumericType:
		z = Int(0)	
	case IntegerType:
		z = Int(0)	
	case IntType:
		z = Int(0)
    case Int32Type:
    	z = Int32(0)
/*
    case Int16Type:
    	z = Int16(0)
    case Int8Type:
    	z = Int8(0)
*/
    case UintType:
    	z = Uint(0)
    case Uint32Type:
        z = Uint32(0)	
/*
    case Uint16Type:
    	z = Uint16(0)
*/
    case ByteType:
    	z = Byte(0)    	
/*
    case BitType:
    	z = Bit(0)
*/
    case BoolType:
    	z = Bool(false)
    case TextType:
    	z = String("")
/*
    case CodePointType:
    	z = CodePoint(0)
*/
    case RealType:
        z = Float(0.)
    case FloatType:
        z = Float(0.)    	
/*
    case Float32Type:
        z = Float32(0.)    	
*/
    case ComplexNumberType:
    case ComplexType:
        z = Complex(0+0i)    	
    case Complex32Type:
        z = Complex32(0+0i)       	
    case StringType:
    	z = String("")
    case TimeType:
    	var t Time
    	z = RTime(t)
    case AnyType:
		z = Int(0)	 
	case IntOrStringType:
		z = Int(0)		
	case BytesType:
		z = Bytes(make([]byte,0))
	case CodePointsType:
		z = CodePoints(make([]rune,0))
//	case IntsType:
//		z = Ints(make([]int64,0))		
	case BitsType:
		
		z = &Bits{   make([]byte,0)  ,    0}	
			
    default:
	    if t.IsNative {
	       // z = &GoWrapper{nil,t,0}		

	       gw := &GoWrapper{runit{robject{rtype: t},nil},nil}

           gw.initialize(t.TotalAttributeCount)

           z = gw


	       if ! markSense {
		       z.SetMarked()
	       }
	       Logln(GC3_,"Zero: IsMarked",z.IsMarked())	       
		} else {
    	   z = NIL   // Hmmm. Do I need one Nil per RType???? With a KnownType attribute?
        }
    }   
    return z
}

/*
If the type is a primitive type, return the zero-value of the type.
If it is a structured-object type, return an instance with no attribute values or relations.
If it is a collection type, return an empty instance of the correct type of collection
(with the specified elment type and/or key and value types)

*/
func (t *RType) Prototype() RObject {
	p := t.Zero()
	if p == NIL {
        var err error

	    if t.Less(ListType) {
	
	       elementTypeName := t.Name[8:]

		   typ, typFound := RT.Types[elementTypeName]
		   if ! typFound {
		      panic(fmt.Sprintf("List Element Type '%s' not found.",elementTypeName))	
		   }	
	
           p, err = RT.Newrlist(typ, 0, -1, nil, nil, nil)		
		   if err != nil {
			   panic(err)
		   }

	    } else if t.Less(MapType) {
		
		   sepPos := strings.Index(t.Name,")=>(")
	       keyTypeName := t.Name[8:sepPos]
	       valTypeName := t.Name[sepPos+4:strings.LastIndex(t.Name,")")]	
		
		   keyTyp, typFound := RT.Types[keyTypeName]
		   if ! typFound {
		      panic(fmt.Sprintf("Map key type '%s' not found.",keyTypeName)	)
		   }
		
		   valTyp, typFound := RT.Types[valTypeName]
		   if ! typFound {
		      panic(fmt.Sprintf("Map value type '%s' not found.",valTypeName))
		   }	
		
           p, err = RT.Newmap(keyTyp, valTyp, 0, -1, nil, nil, nil)		
		   if err != nil {
			   panic(err)
		   }					
	
	    } else if t.Less(SetType) {
	       elementTypeName := t.Name[7:]

		   typ, typFound := RT.Types[elementTypeName]
		   if ! typFound {
		      panic(fmt.Sprintf("Set Element Type '%s' not found.",elementTypeName))
		   }	

           p, err = RT.Newrset(typ, 0, -1, nil, nil)		
		   if err != nil {
			   panic(err)
		   }		
	
        } else { // Must be a struct type
           p, err = RT.NewObject(t.Name) 	
	       if err != nil {
		      panic(err)
	       }
    	}		
	}
	return p
}







type RInChannel interface {
	RObject

	From() RObject
}



type Channel struct {
	Ch chan RObject
	ElementType *RType
}

func (c *Channel) From() RObject {	
   return <- c.Ch
}

func (c *Channel) To(val RObject) {
   c.Ch <- val
}

// TODO

func (p Channel) IsZero() bool {
	return p.Ch == nil
}

func (p Channel) Type() *RType {
	return ChannelType
}

func (p Channel) This() RObject {
	return p
}

/*
Hmmm. TODO
*/
func (p Channel) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a Channel should be considered a collection???
*/
func (p Channel) IsCollection() bool {
	return false
}

func (p Channel) String() string {
   return p.StringTh(nil)
} 

func (p Channel) StringTh(th InterpreterThread) string {
	var descriptor string
	if p.Ch == nil {
		descriptor = "uninitialized"
	} else if cap(p.Ch) > 0 {
		descriptor = fmt.Sprintf("cap: %d len: %d",cap(p.Ch),len(p.Ch))
	} else {
		descriptor = "synchronous"
	}
	return fmt.Sprintf("Channel (%s) of %v", descriptor, p.ElementType)
}

func (p Channel) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Channel) Length() int64 {
	if p.Ch == nil {
		return 0
	}
	return int64(len(p.Ch))
}

func (p Channel) Cap() int64 {
	if p.Ch == nil {
		return 0
	}
	return int64(cap(p.Ch))
}


func (p Channel) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Channel) UUID() []byte {
	panic("A Channel cannot have a UUID.")
	return nil
}

func (p Channel) DBID() int64 {
	panic("A Channel cannot have a DBID.")
	return 0
}

func (p Channel) EnsureUUID() (theUUID []byte, err error) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) UUIDstr() string {
	panic("A Channel cannot have a UUID.")
	return ""
}

func (p Channel) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) UUIDabbrev() string {
	panic("A Channel cannot have a UUID.")
	return ""
}

func (p Channel) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Channel cannot have a UUID.")
	return
}

func (p Channel) RemoveUUID() {
	panic("A Channel does not have a UUID.")
	return
}

func (p Channel) Flags() int8 {
	panic("A Channel has no Flags.")
	return 0
}

func (p Channel) IsBeingStored() bool {
	return false
}
func (p Channel) SetBeingStored() {
}
func (p Channel) ClearBeingStored() {
}

func (p Channel) IsIdReversed() bool {
	return false
}

func (p Channel) SetIdReversed() {}

func (p Channel) ClearIdReversed() {}

func (p Channel) IsLoadNeeded() bool {
	return false
}

func (p Channel) SetLoadNeeded()   {}
func (p Channel) ClearLoadNeeded() {}

func (p Channel) IsValid() bool { return true }
func (p Channel) SetValid()     {}
func (p Channel) ClearValid()   {}

func (p Channel) IsMarked() bool { return false }
func (p Channel) SetMarked()    {}
func (p Channel) ClearMarked()  {}
func (p Channel) ToggleMarked()  {}

/*
TODO TODO TODO !!! What do we do about objects that are sitting in buffered-channel queues but are nowhere else referred to?

Those should not be removed from the objects map nor the attributes maps !!!!!!
Is this going to require a separate flag? Quite possibly, unless we, upon taking an object out of a channel,
immediately mark it as reachable!
*/
func (p Channel) Mark() bool { return false }


func (p Channel) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Channel) SetStoredLocally()     {}
func (p Channel) ClearStoredLocally()   {}

func (p Channel) IsProxy() bool { return false }

func (p Channel) IsTransient() bool { return true }


func (p Channel) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}


func (p Channel) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a Channel in JSON.")
   return
}

func (p Channel) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Channel.")
   return
}


type TimeChannel struct {
	Ch <-chan Time
	ElementType *RType
}

func (c *TimeChannel) From() RObject {
   return RTime(<- c.Ch)
}


// TODO

func (p TimeChannel) IsZero() bool {
	return p.Ch == nil
}

func (p TimeChannel) Type() *RType {
	return ChannelType
}

func (p TimeChannel) This() RObject {
	return p
}

/*
Hmmm. TODO
*/
func (p TimeChannel) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a TimeChannel should be considered a collection???
*/
func (p TimeChannel) IsCollection() bool {
	return false
}

func (p TimeChannel) String() string {
   return p.StringTh(nil)
}

func (p TimeChannel) StringTh(th InterpreterThread) string {
	var descriptor string
	if p.Ch == nil {
		descriptor = "uninitialized"
	} else if cap(p.Ch) > 0 {
		descriptor = fmt.Sprintf("cap: %d len: %d",cap(p.Ch),len(p.Ch))
	} else {
		descriptor = "synchronous"
	}
	return fmt.Sprintf(" Channel (%s) of %v", descriptor, p.ElementType)
}

func (p TimeChannel) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p TimeChannel) Length() int64 {
	if p.Ch == nil {
		return 0
	}
	return int64(len(p.Ch))
}

func (p TimeChannel) Cap() int64 {
	if p.Ch == nil {
		return 0
	}
	return int64(cap(p.Ch))
}


func (p TimeChannel) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p TimeChannel) UUID() []byte {
	panic("A TimeChannel cannot have a UUID.")
	return nil
}

func (p TimeChannel) DBID() int64 {
	panic("A TimeChannel cannot have a DBID.")
	return 0
}

func (p TimeChannel) EnsureUUID() (theUUID []byte, err error) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) UUIDstr() string {
	panic("A TimeChannel cannot have a UUID.")
	return ""
}

func (p TimeChannel) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) UUIDabbrev() string {
	panic("A TimeChannel cannot have a UUID.")
	return ""
}

func (p TimeChannel) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A TimeChannel cannot have a UUID.")
	return
}

func (p TimeChannel) RemoveUUID() {
	panic("A TimeChannel does not have a UUID.")
	return
}

func (p TimeChannel) Flags() int8 {
	panic("A TimeChannel has no Flags.")
	return 0
}

func (p TimeChannel) IsBeingStored() bool {
	return false
}
func (p TimeChannel) SetBeingStored() {
}
func (p TimeChannel) ClearBeingStored() {
}

func (p TimeChannel) IsIdReversed() bool {
	return false
}

func (p TimeChannel) SetIdReversed() {}

func (p TimeChannel) ClearIdReversed() {}

func (p TimeChannel) IsLoadNeeded() bool {
	return false
}

func (p TimeChannel) SetLoadNeeded()   {}
func (p TimeChannel) ClearLoadNeeded() {}

func (p TimeChannel) IsValid() bool { return true }
func (p TimeChannel) SetValid()     {}
func (p TimeChannel) ClearValid()   {}

func (p TimeChannel) IsMarked() bool { return false }
func (p TimeChannel) SetMarked()    {}
func (p TimeChannel) ClearMarked()  {}
func (p TimeChannel) ToggleMarked()  {}

func (p TimeChannel) Mark() bool { return false }

func (p TimeChannel) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p TimeChannel) SetStoredLocally()     {}
func (p TimeChannel) ClearStoredLocally()   {}

func (p TimeChannel) IsProxy() bool { return false }

func (p TimeChannel) IsTransient() bool { return true }


func (p TimeChannel) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p TimeChannel) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a Channel in JSON.")
   return
}

func (p TimeChannel) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Channel.")
   return
}


type Mutex sync.Mutex

// TODO

func (p Mutex) IsZero() bool {
	return sync.Mutex(p) == sync.Mutex{}  // unlocked mutex
}

func (p Mutex) Type() *RType {
	return MutexType
}

func (p Mutex) This() RObject {
	return p
}

/*
Hmmm. TODO
*/
func (p Mutex) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a Mutex should be considered a collection???
*/
func (p Mutex) IsCollection() bool {
	return false
}

func (p Mutex) String() string {
   return p.StringTh(nil)
}

func (p Mutex) StringTh(th InterpreterThread) string {
   return fmt.Sprintf("%v",sync.Mutex(p))
}

func (p Mutex) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Mutex) HasUUID() bool {
	return false
}


/*
   TODO We have to figure out what to do with this.
*/
func (p Mutex) UUID() []byte {
	panic("A Mutex cannot have a UUID.")
	return nil
}

func (p Mutex) DBID() int64 {
	panic("A Mutex cannot have a DBID.")
	return 0
}

func (p Mutex) EnsureUUID() (theUUID []byte, err error) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) UUIDstr() string {
	panic("A Mutex cannot have a UUID.")
	return ""
}

func (p Mutex) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) UUIDabbrev() string {
	panic("A Mutex cannot have a UUID.")
	return ""
}

func (p Mutex) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Mutex cannot have a UUID.")
	return
}

func (p Mutex) RemoveUUID() {
	panic("A Mutex does not have a UUID.")
	return
}

func (p Mutex) Flags() int8 {
	panic("A Mutex has no Flags.")
	return 0
}

func (p Mutex) IsBeingStored() bool {
	return false
}
func (p Mutex) SetBeingStored() {
}
func (p Mutex) ClearBeingStored() {
}

func (p Mutex) IsIdReversed() bool {
	return false
}

func (p Mutex) SetIdReversed() {}

func (p Mutex) ClearIdReversed() {}

func (p Mutex) IsLoadNeeded() bool {
	return false
}

func (p Mutex) SetLoadNeeded()   {}
func (p Mutex) ClearLoadNeeded() {}

func (p Mutex) IsValid() bool { return true }
func (p Mutex) SetValid()     {}
func (p Mutex) ClearValid()   {}

func (p Mutex) IsMarked() bool { return false }
func (p Mutex) SetMarked()    {}
func (p Mutex) ClearMarked()  {}
func (p Mutex) ToggleMarked()  {}

func (p Mutex) Mark() bool { return false }

func (p Mutex) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Mutex) SetStoredLocally()     {}
func (p Mutex) ClearStoredLocally()   {}

func (p Mutex) IsProxy() bool { return false }

func (p Mutex) IsTransient() bool { return true }


func (p Mutex) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Mutex) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a Mutex in JSON.")
   return
}

func (p Mutex) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Mutex.")
   return
}

func (p *Mutex) Lock() {
	(*sync.Mutex)(p).Lock() 
}

func (p *Mutex) Unlock() {
	(*sync.Mutex)(p).Unlock() 
}















type OwnedMutex struct {
	mtx sync.Mutex
	count int64
	threadWithLock InterpreterThread
	mtx2 sync.Mutex  // Mutex to ensure that all variables in the OwnedMutex
	                 // are updated consistently and atomically each time.
}


func (p *OwnedMutex) Lock(th InterpreterThread) {
	p.mtx2.Lock()
    if p.threadWithLock == th {
       p.count++
    } else {  // someone else coming to lock it
	   p.mtx2.Unlock()    
	   p.mtx.Lock()
	   p.mtx2.Lock()	
       p.threadWithLock = th
       p.count++	       
    }
	p.mtx2.Unlock()    
}

func (p *OwnedMutex) Unlock(th InterpreterThread) {
	p.mtx2.Lock()	
	if p.threadWithLock == th {
		p.count--
		if p.count == 0 {
		   p.threadWithLock = nil	
	       p.mtx.Unlock() 			
		}
	} else {
	   p.mtx2.Unlock()			
       panic("A goroutine that did not lock the OwnedMutex is trying to unlock it.")
	}
	p.mtx2.Unlock()	
}

/*
Returns true iff some thread has the OwnedMutex locked presently.
Note that it is only necessarily true of some moment that is already past by the time this
method is returned from.
*/
func (p *OwnedMutex) IsLocked() bool {
	p.mtx2.Lock()
	defer p.mtx2.Unlock()	
	return p.threadWithLock != nil 		
}

/*
Returns true iff the argument thread has the OwnedMutex locked presently.
NOTE that it is only necessarily true of some moment that is already past by the time this
method is returned from, unless it is asked from the same thread as is passed in.
*/
func (p *OwnedMutex) HaveLock(th InterpreterThread) bool {
	p.mtx2.Lock()	
	defer p.mtx2.Unlock()	
	return p.threadWithLock == th 
}



// TODO

func (p OwnedMutex) IsZero() bool {
	return p.mtx == sync.Mutex{}  // unlocked mutex
}

func (p OwnedMutex) Type() *RType {
	return OwnedMutexType
}

func (p OwnedMutex) This() RObject {
	return p
}

/*
Hmmm. TODO
*/
func (p OwnedMutex) IsUnit() bool {
	return true
}

/*
*/
func (p OwnedMutex) IsCollection() bool {
	return false
}

func (p OwnedMutex) String() string {
   return p.StringTh(nil)
}

func (p OwnedMutex) StringTh(th InterpreterThread) string {
   return fmt.Sprintf("%v:%d",p.mtx,p.count)
}

func (p OwnedMutex) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p OwnedMutex) HasUUID() bool {
	return false
}


/*
   TODO We have to figure out what to do with this.
*/
func (p OwnedMutex) UUID() []byte {
	panic("A OwnedMutex cannot have a UUID.")
	return nil
}

func (p OwnedMutex) DBID() int64 {
	panic("A OwnedMutex cannot have a DBID.")
	return 0
}

func (p OwnedMutex) EnsureUUID() (theUUID []byte, err error) {
	panic("A OwnedMutex cannot have a UUID.")
	return
}

func (p OwnedMutex) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A OwnedMutex cannot have a UUID.")
	return
}

func (p OwnedMutex) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A OwnedMutex cannot have a UUID.")
	return
}

func (p OwnedMutex) UUIDstr() string {
	panic("A OwnedMutex cannot have a UUID.")
	return ""
}

func (p OwnedMutex) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A OwnedMutex cannot have a UUID.")
	return
}

func (p OwnedMutex) UUIDabbrev() string {
	panic("A OwnedMutex cannot have a UUID.")
	return ""
}

func (p OwnedMutex) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A OwnedMutex cannot have a UUID.")
	return
}

func (p OwnedMutex) RemoveUUID() {
	panic("A OwnedMutex does not have a UUID.")
	return
}

func (p OwnedMutex) Flags() int8 {
	panic("A OwnedMutex has no Flags.")
	return 0
}

func (p OwnedMutex) IsBeingStored() bool {
	return false
}
func (p OwnedMutex) SetBeingStored() {
}
func (p OwnedMutex) ClearBeingStored() {
}

func (p OwnedMutex) IsIdReversed() bool {
	return false
}

func (p OwnedMutex) SetIdReversed() {}

func (p OwnedMutex) ClearIdReversed() {}

func (p OwnedMutex) IsLoadNeeded() bool {
	return false
}

func (p OwnedMutex) SetLoadNeeded()   {}
func (p OwnedMutex) ClearLoadNeeded() {}

func (p OwnedMutex) IsValid() bool { return true }
func (p OwnedMutex) SetValid()     {}
func (p OwnedMutex) ClearValid()   {}

func (p OwnedMutex) IsMarked() bool { return false }
func (p OwnedMutex) SetMarked()    {}
func (p OwnedMutex) ClearMarked()  {}
func (p OwnedMutex) ToggleMarked()  {}

func (p OwnedMutex) Mark() bool { return false }

func (p OwnedMutex) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p OwnedMutex) SetStoredLocally()     {}
func (p OwnedMutex) ClearStoredLocally()   {}

func (p OwnedMutex) IsProxy() bool { return false }

func (p OwnedMutex) IsTransient() bool { return true }


func (p OwnedMutex) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p OwnedMutex) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a OwnedMutex in JSON.")
   return
}

func (p OwnedMutex) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a OwnedMutex.")
   return
}





















type RWMutex sync.RWMutex

// TODO

func (p RWMutex) IsZero() bool {
	return sync.RWMutex(p) == sync.RWMutex{}  // unlocked mutex
}

func (p RWMutex) Type() *RType {
	return RWMutexType
}

func (p RWMutex) This() RObject {
	return p
}

func (p RWMutex) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a RWMutex should be considered a collection???
*/
func (p RWMutex) IsCollection() bool {
	return false
}

func (p RWMutex) String() string {
   return p.StringTh(nil) 
}
func (p RWMutex) StringTh(th InterpreterThread) string {
   return fmt.Sprintf("%v",sync.RWMutex(p))
}

func (p RWMutex) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p RWMutex) HasUUID() bool {
	return false
}


/*
   TODO We have to figure out what to do with this.
*/
func (p RWMutex) UUID() []byte {
	panic("A RWMutex cannot have a UUID.")
	return nil
}

func (p RWMutex) DBID() int64 {
	panic("A RWMutex cannot have a DBID.")
	return 0
}

func (p RWMutex) EnsureUUID() (theUUID []byte, err error) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) UUIDstr() string {
	panic("A RWMutex cannot have a UUID.")
	return ""
}

func (p RWMutex) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) UUIDabbrev() string {
	panic("A RWMutex cannot have a UUID.")
	return ""
}

func (p RWMutex) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A RWMutex cannot have a UUID.")
	return
}

func (p RWMutex) RemoveUUID() {
	panic("A RWMutex does not have a UUID.")
	return
}

func (p RWMutex) Flags() int8 {
	panic("A RWMutex has no Flags.")
	return 0
}

func (p RWMutex) IsBeingStored() bool {
	return false
}
func (p RWMutex) SetBeingStored() {
}
func (p RWMutex) ClearBeingStored() {
}

func (p RWMutex) IsIdReversed() bool {
	return false
}

func (p RWMutex) SetIdReversed() {}

func (p RWMutex) ClearIdReversed() {}

func (p RWMutex) IsLoadNeeded() bool {
	return false
}

func (p RWMutex) SetLoadNeeded()   {}
func (p RWMutex) ClearLoadNeeded() {}

func (p RWMutex) IsValid() bool { return true }
func (p RWMutex) SetValid()     {}
func (p RWMutex) ClearValid()   {}

func (p RWMutex) IsMarked() bool { return false }
func (p RWMutex) SetMarked()    {}
func (p RWMutex) ClearMarked()  {}
func (p RWMutex) ToggleMarked()  {}

func (p RWMutex) Mark() bool { return false }

func (p RWMutex) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p RWMutex) SetStoredLocally()     {}
func (p RWMutex) ClearStoredLocally()   {}

func (p RWMutex) IsProxy() bool { return false }

func (p RWMutex) IsTransient() bool { return true }


func (p RWMutex) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p RWMutex) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent an RWMutex in JSON.")
   return
}

func (p RWMutex) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Mutex.")
   return
}

func (p *RWMutex) Lock() {
	(*sync.RWMutex)(p).Lock() 
}

func (p *RWMutex) Unlock() {
	(*sync.RWMutex)(p).Unlock() 
}

func (p *RWMutex) RLock() {
	(*sync.RWMutex)(p).Lock() 
}

func (p *RWMutex) RUnlock() {
	(*sync.RWMutex)(p).Unlock() 
}









type RTime Time

func (p RTime) IsZero() bool {
	return Time(p).IsZero()
}

func (p RTime) Type() *RType {
	return TimeType
}

func (p RTime) This() RObject {
	return p
}

func (p RTime) IsUnit() bool {
	return true
}

func (p RTime) IsCollection() bool {
	return false
}

func (p RTime) String() string {
	return p.StringTh(nil) 
}

func (p RTime) StringTh(th InterpreterThread) string {
	return Time(p).String() // TODO May want to change this: Formats as: "2006-01-02 15:04:05.999999999 -0700 MST"
}

func (p RTime) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p RTime) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p RTime) UUID() []byte {
	panic("An RTime cannot have a UUID.")
	return nil
}

func (p RTime) DBID() int64 {
	panic("An RTime cannot have a DBID.")
	return 0
}

func (p RTime) EnsureUUID() (theUUID []byte, err error) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) UUIDuint64s() (id uint64, id2 uint64) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) UUIDstr() string {
	panic("An RTime cannot have a UUID.")
	return ""
}

func (p RTime) EnsureUUIDstr() (uuidstr string, err error) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) UUIDabbrev() string {
	panic("An RTime cannot have a UUID.")
	return ""
}

func (p RTime) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("An RTime cannot have a UUID.")
	return
}

func (p RTime) RemoveUUID() {
	panic("An RTime does not have a UUID.")
	return
}

func (p RTime) Flags() int8 {
	panic("An RTime has no Flags.")
	return 0
}

func (p RTime) IsBeingStored() bool {
	return false
}
func (p RTime) SetBeingStored() {
}
func (p RTime) ClearBeingStored() {
}

func (p RTime) IsIdReversed() bool {
	return false
}

func (p RTime) SetIdReversed() {}

func (p RTime) ClearIdReversed() {}

func (p RTime) IsLoadNeeded() bool {
	return false
}

func (p RTime) SetLoadNeeded()   {}
func (p RTime) ClearLoadNeeded() {}

func (p RTime) IsValid() bool { return true }
func (p RTime) SetValid()     {}
func (p RTime) ClearValid()   {}

func (p RTime) IsMarked() bool { return false }
func (p RTime) SetMarked()    {}
func (p RTime) ClearMarked()  {}
func (p RTime) ToggleMarked()  {}

func (p RTime) Mark() bool { return false }

func (p RTime) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p RTime) SetStoredLocally()     {}
func (p RTime) ClearStoredLocally()   {}

func (p RTime) IsProxy() bool { return false }

func (p RTime) IsTransient() bool { return false }

func (p RTime) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p RTime) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = Time(p)
   return
}

func (p RTime) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Time.")
   return
}






type Int int64

func (p Int) IsZero() bool {
	return p == 0
}

func (p Int) Type() *RType {
	return IntType
}

func (p Int) This() RObject {
	return p
}

func (p Int) IsUnit() bool {
	return true
}

func (p Int) IsCollection() bool {
	return false
}

func (p Int) String() string {
	return strconv.FormatInt(int64(p),10)
}

func (p Int) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p Int) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Int) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Int) UUID() []byte {
	panic("An Int cannot have a UUID.")
	return nil
}

func (p Int) DBID() int64 {
	panic("An Int cannot have a DBID.")
	return 0
}

func (p Int) EnsureUUID() (theUUID []byte, err error) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) UUIDuint64s() (id uint64, id2 uint64) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) UUIDstr() string {
	panic("An Int cannot have a UUID.")
	return ""
}

func (p Int) EnsureUUIDstr() (uuidstr string, err error) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) UUIDabbrev() string {
	panic("An Int cannot have a UUID.")
	return ""
}

func (p Int) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("An Int cannot have a UUID.")
	return
}

func (p Int) RemoveUUID() {
	panic("An Int does not have a UUID.")
	return
}

func (p Int) Flags() int8 {
	panic("An Int has no Flags.")
	return 0
}

func (p Int) IsBeingStored() bool {
	return false
}
func (p Int) SetBeingStored() {
}
func (p Int) ClearBeingStored() {
}

func (p Int) IsIdReversed() bool {
	return false
}

func (p Int) SetIdReversed() {}

func (p Int) ClearIdReversed() {}

func (p Int) IsLoadNeeded() bool {
	return false
}

func (p Int) SetLoadNeeded()   {}
func (p Int) ClearLoadNeeded() {}

func (p Int) IsValid() bool { return true }
func (p Int) SetValid()     {}
func (p Int) ClearValid()   {}

func (p Int) IsMarked() bool { return false }
func (p Int) SetMarked()    {}
func (p Int) ClearMarked()  {}
func (p Int) ToggleMarked()  {}

func (p Int) Mark() bool { return false }

func (p Int) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Int) SetStoredLocally()     {}
func (p Int) ClearStoredLocally()   {}

func (p Int) IsProxy() bool { return false }

func (p Int) IsTransient() bool { return false }

func (p Int) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Int) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = int64(p)
   return
}

func (p Int) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Int(int64(tree.(float64)))
   return
}


type Int32 int32

func (p Int32) IsZero() bool {
	return p == 0
}

func (p Int32) Type() *RType {
	return Int32Type
}

func (p Int32) This() RObject {
	return p
}

func (p Int32) IsUnit() bool {
	return true
}

func (p Int32) IsCollection() bool {
	return false
}

func (p Int32) String() string {
	return strconv.FormatInt(int64(p),10)
}

func (p Int32) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p Int32) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Int32) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Int32) UUID() []byte {
	panic("An Int32 cannot have a UUID.")
	return nil
}

func (p Int32) DBID() int64 {
	panic("An Int32 cannot have a DBID.")
	return 0
}

func (p Int32) EnsureUUID() (theUUID []byte, err error) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) UUIDuint64s() (id uint64, id2 uint64) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) UUIDstr() string {
	panic("An Int32 cannot have a UUID.")
	return ""
}

func (p Int32) EnsureUUIDstr() (uuidstr string, err error) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) UUIDabbrev() string {
	panic("An Int32 cannot have a UUID.")
	return ""
}

func (p Int32) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("An Int32 cannot have a UUID.")
	return
}

func (p Int32) RemoveUUID() {
	panic("An Int32 does not have a UUID.")
	return
}

func (p Int32) Flags() int8 {
	panic("An Int32 has no Flags.")
	return 0
}

func (p Int32) IsBeingStored() bool {
	return false
}
func (p Int32) SetBeingStored() {
}
func (p Int32) ClearBeingStored() {
}

func (p Int32) IsIdReversed() bool {
	return false
}

func (p Int32) SetIdReversed() {}

func (p Int32) ClearIdReversed() {}

func (p Int32) IsLoadNeeded() bool {
	return false
}

func (p Int32) SetLoadNeeded()   {}
func (p Int32) ClearLoadNeeded() {}

func (p Int32) IsValid() bool { return true }
func (p Int32) SetValid()     {}
func (p Int32) ClearValid()   {}

func (p Int32) IsMarked() bool { return false }
func (p Int32) SetMarked()    {}
func (p Int32) ClearMarked()  {}
func (p Int32) ToggleMarked()  {}

func (p Int32) Mark() bool { return false }

func (p Int32) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Int32) SetStoredLocally()     {}
func (p Int32) ClearStoredLocally()   {}

func (p Int32) IsProxy() bool { return false }

func (p Int32) IsTransient() bool { return false }

func (p Int32) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Int32) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = int32(p)
   return
}

func (p Int32) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Int32(int32(tree.(float64)))
   return
}


type Uint uint64

func (p Uint) IsZero() bool {
	return p == 0
}

func (p Uint) Type() *RType {
	return UintType
}

func (p Uint) This() RObject {
	return p
}

func (p Uint) IsUnit() bool {
	return true
}

func (p Uint) IsCollection() bool {
	return false
}

func (p Uint) String() string {
	return strconv.FormatUint(uint64(p),10)
}

func (p Uint) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p Uint) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Uint) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Uint) UUID() []byte {
	panic("A Uint cannot have a UUID.")
	return nil
}

func (p Uint) DBID() int64 {
	panic("A Uint cannot have a DBID.")
	return 0
}

func (p Uint) EnsureUUID() (theUUID []byte, err error) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) UUIDstr() string {
	panic("A Uint cannot have a UUID.")
	return ""
}

func (p Uint) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) UUIDabbrev() string {
	panic("A Uint cannot have a UUID.")
	return ""
}

func (p Uint) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Uint cannot have a UUID.")
	return
}

func (p Uint) RemoveUUID() {
	panic("A Uint does not have a UUID.")
	return
}

func (p Uint) Flags() int8 {
	panic("A Uint has no Flags.")
	return 0
}

func (p Uint) IsBeingStored() bool {
	return false
}
func (p Uint) SetBeingStored() {
}
func (p Uint) ClearBeingStored() {
}

func (p Uint) IsIdReversed() bool {
	return false
}

func (p Uint) SetIdReversed() {}

func (p Uint) ClearIdReversed() {}

func (p Uint) IsLoadNeeded() bool {
	return false
}

func (p Uint) SetLoadNeeded()   {}
func (p Uint) ClearLoadNeeded() {}

func (p Uint) IsValid() bool { return true }
func (p Uint) SetValid()     {}
func (p Uint) ClearValid()   {}

func (p Uint) IsMarked() bool { return false }
func (p Uint) SetMarked()    {}
func (p Uint) ClearMarked()  {}
func (p Uint) ToggleMarked()  {}

func (p Uint) Mark() bool { return false }

func (p Uint) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Uint) SetStoredLocally()     {}
func (p Uint) ClearStoredLocally()   {}

func (p Uint) IsProxy() bool { return false }

func (p Uint) IsTransient() bool { return false }

func (p Uint) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Uint) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = uint64(p)
   return
}

func (p Uint) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Uint(uint64(tree.(float64)))
   return
}


type Uint32 uint32

func (p Uint32) IsZero() bool {
	return p == 0
}

func (p Uint32) Type() *RType {
	return Uint32Type
}

func (p Uint32) This() RObject {
	return p
}

func (p Uint32) IsUnit() bool {
	return true
}

func (p Uint32) IsCollection() bool {
	return false
}

func (p Uint32) String() string {
	return strconv.FormatUint(uint64(p),10)
}

func (p Uint32) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p Uint32) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Uint32) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Uint32) UUID() []byte {
	panic("A Uint32 cannot have a UUID.")
	return nil
}

func (p Uint32) DBID() int64 {
	panic("A Uint32 cannot have a DBID.")
	return 0
}

func (p Uint32) EnsureUUID() (theUUID []byte, err error) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) UUIDstr() string {
	panic("A Uint32 cannot have a UUID.")
	return ""
}

func (p Uint32) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) UUIDabbrev() string {
	panic("A Uint32 cannot have a UUID.")
	return ""
}

func (p Uint32) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Uint32 cannot have a UUID.")
	return
}

func (p Uint32) RemoveUUID() {
	panic("A Uint32 does not have a UUID.")
	return
}

func (p Uint32) Flags() int8 {
	panic("A Uint32 has no Flags.")
	return 0
}

func (p Uint32) IsBeingStored() bool {
	return false
}
func (p Uint32) SetBeingStored() {
}
func (p Uint32) ClearBeingStored() {
}

func (p Uint32) IsIdReversed() bool {
	return false
}

func (p Uint32) SetIdReversed() {}

func (p Uint32) ClearIdReversed() {}

func (p Uint32) IsLoadNeeded() bool {
	return false
}

func (p Uint32) SetLoadNeeded()   {}
func (p Uint32) ClearLoadNeeded() {}

func (p Uint32) IsValid() bool { return true }
func (p Uint32) SetValid()     {}
func (p Uint32) ClearValid()   {}

func (p Uint32) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Uint32) SetStoredLocally()     {}
func (p Uint32) ClearStoredLocally()   {}

func (p Uint32) IsMarked() bool { return false }
func (p Uint32) SetMarked()    {}
func (p Uint32) ClearMarked()  {}
func (p Uint32) ToggleMarked()  {}

func (p Uint32) Mark() bool { return false }

func (p Uint32) IsProxy() bool { return false }

func (p Uint32) IsTransient() bool { return false }

func (p Uint32) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Uint32) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = uint32(p)
   return
}

func (p Uint32) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Uint32(uint32(tree.(float64)))
   return
}


type Float float64

func (p Float) IsZero() bool {
	return p == 0
}

func (p Float) Type() *RType {
	return FloatType
}

func (p Float) This() RObject {
	return p
}

func (p Float) IsUnit() bool {
	return true
}

func (p Float) IsCollection() bool {
	return false
}

func (p Float) String() string {
	return strconv.FormatFloat(float64(p), 'G', -1, 64)
}

func (p Float) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p Float) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Float) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Float) UUID() []byte {
	panic("A Float cannot have a UUID.")
	return nil
}

func (p Float) DBID() int64 {
	panic("A Float cannot have a DBID.")
	return 0
}

func (p Float) EnsureUUID() (theUUID []byte, err error) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) UUIDstr() string {
	panic("A Float cannot have a UUID.")
	return ""
}

func (p Float) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) UUIDabbrev() string {
	panic("A Float cannot have a UUID.")
	return ""
}

func (p Float) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Float cannot have a UUID.")
	return
}

func (p Float) RemoveUUID() {
	panic("A Float does not have a UUID.")
	return
}

func (p Float) Flags() int8 {
	panic("A Float has no Flags.")
	return 0
}

func (p Float) IsBeingStored() bool {
	return false
}
func (p Float) SetBeingStored() {
}
func (p Float) ClearBeingStored() {
}

func (p Float) IsIdReversed() bool {
	return false
}

func (p Float) SetIdReversed() {}

func (p Float) ClearIdReversed() {}

func (p Float) IsLoadNeeded() bool {
	return false
}

func (p Float) SetLoadNeeded()   {}
func (p Float) ClearLoadNeeded() {}

func (p Float) IsValid() bool { return true }
func (p Float) SetValid()     {}
func (p Float) ClearValid()   {}

func (p Float) IsMarked() bool { return false }
func (p Float) SetMarked()    {}
func (p Float) ClearMarked()  {}
func (p Float) ToggleMarked()  {}

func (p Float) Mark() bool { return false }

func (p Float) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Float) SetStoredLocally()     {}
func (p Float) ClearStoredLocally()   {}

func (p Float) IsProxy() bool { return false }

func (p Float) IsTransient() bool { return false }

func (p Float) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Float) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = float64(p)
   return
}

func (p Float) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Float(tree.(float64))
   return
}


type Bool bool

func (p Bool) IsZero() bool {
	return !bool(p)
}

func (p Bool) Type() *RType {
	return BoolType
}

func (p Bool) This() RObject {
	return p
}

func (p Bool) IsUnit() bool {
	return true
}

func (p Bool) IsCollection() bool {
	return false
}

func (p Bool) String() string {
	return strconv.FormatBool(bool(p))
}

func (p Bool) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p Bool) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Bool) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Bool) UUID() []byte {
	panic("A Bool cannot have a UUID.")
	return nil
}

func (p Bool) DBID() int64 {
	panic("A Bool cannot have a DBID.")
	return 0
}

func (p Bool) EnsureUUID() (theUUID []byte, err error) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) UUIDstr() string {
	panic("A Bool cannot have a UUID.")
	return ""
}

func (p Bool) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) UUIDabbrev() string {
	panic("A Bool cannot have a UUID.")
	return ""
}

func (p Bool) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Bool cannot have a UUID.")
	return
}

func (p Bool) RemoveUUID() {
	panic("A Bool does not have a UUID.")
	return
}

func (p Bool) Flags() int8 {
	panic("A Bool has no Flags.")
	return 0
}

func (p Bool) IsBeingStored() bool {
	return false
}
func (p Bool) SetBeingStored() {
}
func (p Bool) ClearBeingStored() {
}

func (p Bool) IsIdReversed() bool {
	return false
}

func (p Bool) SetIdReversed() {}

func (p Bool) ClearIdReversed() {}

func (p Bool) IsLoadNeeded() bool {
	return false
}

func (p Bool) SetLoadNeeded()   {}
func (p Bool) ClearLoadNeeded() {}

func (p Bool) IsValid() bool { return true }
func (p Bool) SetValid()     {}
func (p Bool) ClearValid()   {}

func (p Bool) IsMarked() bool { return false }
func (p Bool) SetMarked()    {}
func (p Bool) ClearMarked()  {}
func (p Bool) ToggleMarked()  {}

func (p Bool) Mark() bool { return false }

func (p Bool) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Bool) SetStoredLocally()     {}
func (p Bool) ClearStoredLocally()   {}

func (p Bool) IsProxy() bool { return false }

func (p Bool) IsTransient() bool { return false }

func (p Bool) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Bool) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = bool(p)
   return
}

func (p Bool) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Bool(tree.(bool))
   return
}



type Nil byte

func (p Nil) IsZero() bool {
	return true
}

func (p Nil) Type() *RType {
	return NothingType
}

func (p Nil) This() RObject {
	return p
}

func (p Nil) IsUnit() bool {
	return true
}

func (p Nil) IsCollection() bool {
	return false
}

/*
Return a closed channel that can be tested to ascertain it will return no collection elements.
*/
func (p Nil) Iter(th InterpreterThread) <-chan RObject {
	ch := make(chan RObject)
	close(ch)
	return ch
}

func (p Nil) Length() int64 {
	return 0
}

func (p Nil) Cap() int64 {
	return 0
}



func (p Nil) String() string {
	return "*nil*"
}

func (p Nil) StringTh(th InterpreterThread) string {
	return "*nil*"
}

func (p Nil) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Nil) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Nil) UUID() []byte {
	panic("A Nil cannot have a UUID.")
	return nil
}

func (p Nil) DBID() int64 {
	panic("A Nil cannot have a DBID.")
	return 0
}

func (p Nil) EnsureUUID() (theUUID []byte, err error) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) UUIDstr() string {
	panic("A Nil cannot have a UUID.")
	return ""
}

func (p Nil) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) UUIDabbrev() string {
	panic("A Nil cannot have a UUID.")
	return ""
}

func (p Nil) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Nil cannot have a UUID.")
	return
}

func (p Nil) RemoveUUID() {
	panic("A Nil does not have a UUID.")
	return
}

func (p Nil) Flags() int8 {
	panic("A Nil has no Flags.")
	return 0
}

func (p Nil) IsBeingStored() bool {
	return false
}
func (p Nil) SetBeingStored() {
}
func (p Nil) ClearBeingStored() {
}

func (p Nil) IsIdReversed() bool {
	return false
}

func (p Nil) SetIdReversed() {}

func (p Nil) ClearIdReversed() {}

func (p Nil) IsLoadNeeded() bool {
	return false
}

func (p Nil) SetLoadNeeded()   {}
func (p Nil) ClearLoadNeeded() {}

func (p Nil) IsValid() bool { return true }
func (p Nil) SetValid()     {}
func (p Nil) ClearValid()   {}

func (p Nil) IsMarked() bool { return false }
func (p Nil) SetMarked()    {}
func (p Nil) ClearMarked()  {}
func (p Nil) ToggleMarked()  {}

func (p Nil) Mark() bool { return false }

func (p Nil) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Nil) SetStoredLocally()     {}
func (p Nil) ClearStoredLocally()   {}

func (p Nil) IsProxy() bool { return false }

func (p Nil) IsTransient() bool { return false }

func (p Nil) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Nil) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = nil
   return
}

func (p Nil) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = NIL
   return
}

const NIL Nil = 0



type Float32 float32



type Complex complex128

func (p Complex) IsZero() bool {
	return p == 0
}

func (p Complex) Type() *RType {
	return ComplexType
}

func (p Complex) This() RObject {
	return p
}

func (p Complex) IsUnit() bool {
	return true
}

func (p Complex) IsCollection() bool {
	return false
}

func (p Complex) String() string {
    return fmt.Sprintf("%G",complex128(p))	
}

func (p Complex) StringTh(th InterpreterThread) string {
    return p.String()
}

func (p Complex) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Complex) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Complex) UUID() []byte {
	panic("A Complex cannot have a UUID.")
	return nil
}

func (p Complex) DBID() int64 {
	panic("A Complex cannot have a DBID.")
	return 0
}

func (p Complex) EnsureUUID() (theUUID []byte, err error) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) UUIDstr() string {
	panic("A Complex cannot have a UUID.")
	return ""
}

func (p Complex) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) UUIDabbrev() string {
	panic("A Complex cannot have a UUID.")
	return ""
}

func (p Complex) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Complex cannot have a UUID.")
	return
}

func (p Complex) RemoveUUID() {
	panic("A Complex does not have a UUID.")
	return
}

func (p Complex) Flags() int8 {
	panic("A Complex has no Flags.")
	return 0
}

func (p Complex) IsBeingStored() bool {
	return false
}
func (p Complex) SetBeingStored() {
}
func (p Complex) ClearBeingStored() {
}

func (p Complex) IsIdReversed() bool {
	return false
}

func (p Complex) SetIdReversed() {}

func (p Complex) ClearIdReversed() {}

func (p Complex) IsLoadNeeded() bool {
	return false
}

func (p Complex) SetLoadNeeded()   {}
func (p Complex) ClearLoadNeeded() {}

func (p Complex) IsValid() bool { return true }
func (p Complex) SetValid()     {}
func (p Complex) ClearValid()   {}

func (p Complex) IsMarked() bool { return false }
func (p Complex) SetMarked()    {}
func (p Complex) ClearMarked()  {}
func (p Complex) ToggleMarked()  {}

func (p Complex) Mark() bool { return false }

func (p Complex) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Complex) SetStoredLocally()     {}
func (p Complex) ClearStoredLocally()   {}

func (p Complex) IsProxy() bool { return false }

func (p Complex) IsTransient() bool { return false }

func (p Complex) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Complex) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a Complex number in JSON.")
   return
}

func (p Complex) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Complex number.")
   return
}


type Complex32 complex64

func (p Complex32) IsZero() bool {
	return p == 0
}

func (p Complex32) Type() *RType {
	return Complex32Type
}

func (p Complex32) This() RObject {
	return p
}

func (p Complex32) IsUnit() bool {
	return true
}

func (p Complex32) IsCollection() bool {
	return false
}

func (p Complex32) String() string {
    return fmt.Sprintf("%G",complex64(p))	
}

func (p Complex32) StringTh(th InterpreterThread) string {
    return p.String()
}

func (p Complex32) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Complex32) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Complex32) UUID() []byte {
	panic("A Complex32 cannot have a UUID.")
	return nil
}

func (p Complex32) DBID() int64 {
	panic("A Complex32 cannot have a DBID.")
	return 0
}

func (p Complex32) EnsureUUID() (theUUID []byte, err error) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) UUIDstr() string {
	panic("A Complex32 cannot have a UUID.")
	return ""
}

func (p Complex32) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) UUIDabbrev() string {
	panic("A Complex32 cannot have a UUID.")
	return ""
}

func (p Complex32) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Complex32 cannot have a UUID.")
	return
}

func (p Complex32) RemoveUUID() {
	panic("A Complex32 does not have a UUID.")
	return
}

func (p Complex32) Flags() int8 {
	panic("A Complex32 has no Flags.")
	return 0
}

func (p Complex32) IsBeingStored() bool {
	return false
}
func (p Complex32) SetBeingStored() {
}
func (p Complex32) ClearBeingStored() {
}

func (p Complex32) IsIdReversed() bool {
	return false
}

func (p Complex32) SetIdReversed() {}

func (p Complex32) ClearIdReversed() {}

func (p Complex32) IsLoadNeeded() bool {
	return false
}

func (p Complex32) SetLoadNeeded()   {}
func (p Complex32) ClearLoadNeeded() {}

func (p Complex32) IsValid() bool { return true }
func (p Complex32) SetValid()     {}
func (p Complex32) ClearValid()   {}

func (p Complex32) IsMarked() bool { return false }
func (p Complex32) SetMarked()    {}
func (p Complex32) ClearMarked()  {}
func (p Complex32) ToggleMarked()  {}

func (p Complex32) Mark() bool { return false }

func (p Complex32) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Complex32) SetStoredLocally()     {}
func (p Complex32) ClearStoredLocally()   {}

func (p Complex32) IsProxy() bool { return false }

func (p Complex32) IsTransient() bool { return false }

func (p Complex32) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Complex32) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a Complex32 number in JSON.")
   return
}

func (p Complex32) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Complex32 number.")
   return
}
























type String string

func (p String) IsZero() bool {
	return p == ""
}

func (p String) Type() *RType {
	return StringType
}

func (p String) This() RObject {
	return p
}

func (p String) IsUnit() bool {
	return true
}

func (p String) IsCollection() bool {
	return false
}

func (p String) String() string {
	return string(p)
}

func (p String) StringTh(th InterpreterThread) string {
	return string(p)
}

func (p String) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p String) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p String) UUID() []byte {
	panic("A String cannot have a UUID.")
	return nil
}

func (p String) DBID() int64 {
	panic("A String cannot have a DBID.")
	return 0
}

func (p String) EnsureUUID() (theUUID []byte, err error) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) UUIDstr() string {
	panic("A String cannot have a UUID.")
	return ""
}

func (p String) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) UUIDabbrev() string {
	panic("A String cannot have a UUID.")
	return ""
}

func (p String) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A String cannot have a UUID.")
	return
}

func (p String) RemoveUUID() {
	panic("A String does not have a UUID.")
	return
}

func (p String) Flags() int8 {
	panic("A String has no Flags.")
	return 0
}

func (p String) IsBeingStored() bool {
	return false
}
func (p String) SetBeingStored() {
}
func (p String) ClearBeingStored() {
}

func (p String) IsIdReversed() bool {
	return false
}

func (p String) SetIdReversed() {}

func (p String) ClearIdReversed() {}

func (p String) IsLoadNeeded() bool {
	return false
}

func (p String) SetLoadNeeded()   {}
func (p String) ClearLoadNeeded() {}

func (p String) IsValid() bool { return true }
func (p String) SetValid()     {}
func (p String) ClearValid()   {}

func (p String) IsMarked() bool { return false }
func (p String) SetMarked()    {}
func (p String) ClearMarked()  {}
func (p String) ToggleMarked()  {}

func (p String) Mark() bool { return false }

func (p String) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p String) SetStoredLocally()     {}
func (p String) ClearStoredLocally()   {}

func (p String) IsProxy() bool { return false }

func (p String) IsTransient() bool { return false }

func (p String) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p String) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = string(p)
   return
}

func (p String) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = String(tree.(string))
   return
}

/*
The Proxy type of RObject is a special relish-system-internal type.
Proxy RObjects are used as the members in an RObject collection which has been fetched from the database,
as a way of facilitating lazy loading of the actual objects in the collection.
The int64 value of the Proxy is the dbid of the real robject.

We might also use Proxy objects as the value of attributes at some times in the persistence lifecycle. Not sure when.

Oh also, we might need UUProxy objects whose underlying type might be complex128 and which represent the 
uuid of a remote object.

*/

type Proxy int64

/*
TODO This will require fetching the real object
*/
func (p Proxy) IsZero() bool {
	return false
}

func (p Proxy) Type() *RType {
	return ProxyType
}

func (p Proxy) This() RObject {
	return p
}

func (p Proxy) IsUnit() bool {
	return true
}

func (p Proxy) IsCollection() bool {
	return false
}

func (p Proxy) String() string {
	return strconv.FormatInt(int64(p),10)
}

func (p Proxy) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p Proxy) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Proxy) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Proxy) UUID() []byte {
	panic("A Proxy cannot have a UUID.")
	return nil
}

func (p Proxy) DBID() int64 {
	return int64(p)
}

func (p Proxy) EnsureUUID() (theUUID []byte, err error) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) UUIDstr() string {
	panic("A Proxy cannot have a UUID.")
	return ""
}

func (p Proxy) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) UUIDabbrev() string {
	panic("A Proxy cannot have a UUID.")
	return ""
}

func (p Proxy) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Proxy cannot have a UUID.")
	return
}

func (p Proxy) RemoveUUID() {
	panic("A Proxy does not have a UUID.")
	return
}

func (p Proxy) Flags() int8 {
	panic("A Proxy has no Flags.")
	return 0
}

func (p Proxy) IsBeingStored() bool {
	return false
}
func (p Proxy) SetBeingStored() {
}
func (p Proxy) ClearBeingStored() {
}

func (p Proxy) IsIdReversed() bool {
	return false
}

func (p Proxy) SetIdReversed() {}

func (p Proxy) ClearIdReversed() {}

func (p Proxy) IsLoadNeeded() bool {
	return false
}

func (p Proxy) SetLoadNeeded()   {}
func (p Proxy) ClearLoadNeeded() {}

func (p Proxy) IsValid() bool { return true }
func (p Proxy) SetValid()     {}
func (p Proxy) ClearValid()   {}

func (p Proxy) IsMarked() bool { return false }
func (p Proxy) SetMarked()    {}
func (p Proxy) ClearMarked()  {}
func (p Proxy) ToggleMarked()  {}

func (p Proxy) Mark() bool { return false }

func (p Proxy) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Proxy) SetStoredLocally()     {}
func (p Proxy) ClearStoredLocally()   {}

func (p Proxy) IsProxy() bool { return true }

func (p Proxy) IsTransient() bool { return false }

func (p Proxy) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Proxy) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Cannot represent a Proxy in JSON.")
   return
}

func (p Proxy) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Cannot unmarshal JSON into a Proxy.")
   return
}

/*
Example code:

import (
       "fmt"
       "strings"
)

type RObject interface {
   This() RObject
   Boogah() string
   Exclaim() string
}

type rstring string

func (s rstring) This() RObject { return s }
func (s rstring) Boogah() string { return string(s) } 
func (s rstring) Exclaim() string { return fmt.Sprintf("My length is %v",len(s)) }
func main() {

	var r rstring = "Bagooh!"
	c := strings.Contains(string(r),"goo")
	fmt.Println(r)
	fmt.Println(c)
	var s string = "I am a string."
	r2 := rstring(s)
	fmt.Println(r2)
	fmt.Println(r2.This())
	fmt.Println(r2.This().Exclaim())	
}

*/


type Byte byte

/*
TODO This will require fetching the real object
*/
func (p Byte) IsZero() bool {
	return p == 0
}

func (p Byte) Type() *RType {
	return ByteType
}

func (p Byte) This() RObject {
	return p
}

func (p Byte) IsUnit() bool {
	return true
}

func (p Byte) IsCollection() bool {
	return false
}

func (p Byte) String() string {
	return strconv.FormatUint(uint64(p),10)
}

func (p Byte) StringTh(th InterpreterThread) string {
	return strconv.FormatUint(uint64(p),10)
}

func (p Byte) BitString() string {
	return fmt.Sprintf("%08b",byte(p)) 
}

func (p Byte) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Byte) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Byte) UUID() []byte {
	panic("A Byte cannot have a UUID.")
	return nil
}

func (p Byte) DBID() int64 {
	panic("A Byte cannot have a DBID.")
	return 0
}

func (p Byte) EnsureUUID() (theUUID []byte, err error) {
	panic("A Byte cannot have a UUID.")
	return
}

func (p Byte) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Byte cannot have a UUID.")
	return
}

func (p Byte) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Byte cannot have a UUID.")
	return
}

func (p Byte) UUIDstr() string {
	panic("A Byte cannot have a UUID.")
	return ""
}

func (p Byte) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Byte cannot have a UUID.")
	return
}

func (p Byte) UUIDabbrev() string {
	panic("A Byte cannot have a UUID.")
	return ""
}

func (p Byte) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Byte cannot have a UUID.")
	return
}

func (p Byte) RemoveUUID() {
	panic("A Byte does not have a UUID.")
	return
}

func (p Byte) Flags() int8 {
	panic("A Byte has no Flags.")
	return 0
}

func (p Byte) IsBeingStored() bool {
	return false
}
func (p Byte) SetBeingStored() {
}
func (p Byte) ClearBeingStored() {
}

func (p Byte) IsIdReversed() bool {
	return false
}

func (p Byte) SetIdReversed() {}

func (p Byte) ClearIdReversed() {}

func (p Byte) IsLoadNeeded() bool {
	return false
}

func (p Byte) SetLoadNeeded()   {}
func (p Byte) ClearLoadNeeded() {}

func (p Byte) IsValid() bool { return true }
func (p Byte) SetValid()     {}
func (p Byte) ClearValid()   {}

func (p Byte) IsMarked() bool { return false }
func (p Byte) SetMarked()    {}
func (p Byte) ClearMarked()  {}
func (p Byte) ToggleMarked()  {}

func (p Byte) Mark() bool { return false }

func (p Byte) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Byte) SetStoredLocally()     {}
func (p Byte) ClearStoredLocally()   {}

func (p Byte) IsProxy() bool { return false }

func (p Byte) IsTransient() bool { return false }

func (p Byte) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Byte) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = byte(p)
   return
}

func (p Byte) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Byte(byte(tree.(float64)))
   return
}






type Bit byte

/*
TODO This will require fetching the real object
*/
func (p Bit) IsZero() bool {
	return p == 0
}

func (p Bit) Type() *RType {
	return BitType
}

func (p Bit) This() RObject {
	return p
}

func (p Bit) IsUnit() bool {
	return true
}

func (p Bit) IsCollection() bool {
	return false
}

func (p Bit) String() string {
	return strconv.FormatUint(uint64(p),10)
}

func (p Bit) StringTh(th InterpreterThread) string {
	return strconv.FormatUint(uint64(p),10)
}


func (p Bit) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p Bit) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Bit) UUID() []byte {
	panic("A Bit cannot have a UUID.")
	return nil
}

func (p Bit) DBID() int64 {
	panic("A Bit cannot have a DBID.")
	return 0
}

func (p Bit) EnsureUUID() (theUUID []byte, err error) {
	panic("A Bit cannot have a UUID.")
	return
}

func (p Bit) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Bit cannot have a UUID.")
	return
}

func (p Bit) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Bit cannot have a UUID.")
	return
}

func (p Bit) UUIDstr() string {
	panic("A Bit cannot have a UUID.")
	return ""
}

func (p Bit) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Bit cannot have a UUID.")
	return
}

func (p Bit) UUIDabbrev() string {
	panic("A Bit cannot have a UUID.")
	return ""
}

func (p Bit) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Bit cannot have a UUID.")
	return
}

func (p Bit) RemoveUUID() {
	panic("A Bit does not have a UUID.")
	return
}

func (p Bit) Flags() int8 {
	panic("A Bit has no Flags.")
	return 0
}

func (p Bit) IsBeingStored() bool {
	return false
}
func (p Bit) SetBeingStored() {
}
func (p Bit) ClearBeingStored() {
}

func (p Bit) IsIdReversed() bool {
	return false
}

func (p Bit) SetIdReversed() {}

func (p Bit) ClearIdReversed() {}

func (p Bit) IsLoadNeeded() bool {
	return false
}

func (p Bit) SetLoadNeeded()   {}
func (p Bit) ClearLoadNeeded() {}

func (p Bit) IsValid() bool { return true }
func (p Bit) SetValid()     {}
func (p Bit) ClearValid()   {}

func (p Bit) IsMarked() bool { return false }
func (p Bit) SetMarked()    {}
func (p Bit) ClearMarked()  {}
func (p Bit) ToggleMarked()  {}

func (p Bit) Mark() bool { return false }

func (p Bit) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Bit) SetStoredLocally()     {}
func (p Bit) ClearStoredLocally()   {}

func (p Bit) IsProxy() bool { return false }

func (p Bit) IsTransient() bool { return false }

func (p Bit) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p Bit) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = byte(p)
   return
}

func (p Bit) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Bit(byte(tree.(float64)))
   return
}







type CodePoint rune  // Should we call it Rune ?? Probably not. CodePoint is unambiguous, if boring.

/*
TODO This will require fetching the real object
*/
func (p CodePoint) IsZero() bool {
	return p == 0
}

func (p CodePoint) Type() *RType {
	return CodePointType
}

func (p CodePoint) This() RObject {
	return p
}

func (p CodePoint) IsUnit() bool {
	return true
}

func (p CodePoint) IsCollection() bool {
	return false
}

func (p CodePoint) String() string {
	return strconv.FormatInt(int64(p),10)
}

func (p CodePoint) StringTh(th InterpreterThread) string {
	return strconv.FormatInt(int64(p),10)
}


func (p CodePoint) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p CodePoint) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p CodePoint) UUID() []byte {
	panic("A CodePoint cannot have a UUID.")
	return nil
}

func (p CodePoint) DBID() int64 {
	panic("A CodePoint cannot have a DBID.")
	return 0
}

func (p CodePoint) EnsureUUID() (theUUID []byte, err error) {
	panic("A CodePoint cannot have a UUID.")
	return
}

func (p CodePoint) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A CodePoint cannot have a UUID.")
	return
}

func (p CodePoint) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A CodePoint cannot have a UUID.")
	return
}

func (p CodePoint) UUIDstr() string {
	panic("A CodePoint cannot have a UUID.")
	return ""
}

func (p CodePoint) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A CodePoint cannot have a UUID.")
	return
}

func (p CodePoint) UUIDabbrev() string {
	panic("A CodePoint cannot have a UUID.")
	return ""
}

func (p CodePoint) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A CodePoint cannot have a UUID.")
	return
}

func (p CodePoint) RemoveUUID() {
	panic("A CodePoint does not have a UUID.")
	return
}

func (p CodePoint) Flags() int8 {
	panic("A CodePoint has no Flags.")
	return 0
}

func (p CodePoint) IsBeingStored() bool {
	return false
}
func (p CodePoint) SetBeingStored() {
}
func (p CodePoint) ClearBeingStored() {
}

func (p CodePoint) IsIdReversed() bool {
	return false
}

func (p CodePoint) SetIdReversed() {}

func (p CodePoint) ClearIdReversed() {}

func (p CodePoint) IsLoadNeeded() bool {
	return false
}

func (p CodePoint) SetLoadNeeded()   {}
func (p CodePoint) ClearLoadNeeded() {}

func (p CodePoint) IsValid() bool { return true }
func (p CodePoint) SetValid()     {}
func (p CodePoint) ClearValid()   {}

func (p CodePoint) IsMarked() bool { return false }
func (p CodePoint) SetMarked()    {}
func (p CodePoint) ClearMarked()  {}
func (p CodePoint) ToggleMarked()  {}

func (p CodePoint) Mark() bool { return false }

func (p CodePoint) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p CodePoint) SetStoredLocally()     {}
func (p CodePoint) ClearStoredLocally()   {}

func (p CodePoint) IsProxy() bool { return false }

func (p CodePoint) IsTransient() bool { return false }

func (p CodePoint) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p CodePoint) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = rune(p)
   return
}

func (p CodePoint) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = CodePoint(rune(tree.(float64)))
   return
}


/*
Bytes
CodePoints
Bits
*/


/*
Note: Maybe we need a bool flag as to whether the slice is the original one (and thus can be persisted)
or is a slice of a slice and cannot be persisted.
If so, need it to be a struct like the Bits type.
TODO This needs more thought.
*/
type Bytes []byte

/*
 TODO This will require fetching the real object
*/
func (p Bytes) IsZero() bool {
	return len(p) == 0
}

func (p Bytes) Type() *RType {
	return BytesType
}

func (p Bytes) This() RObject {
	return p
}

func (p Bytes) IsUnit() bool {
	return false
}

func (p Bytes) IsCollection() bool {
	return true
}

func (p Bytes) String() string {
   bt := ([]byte)(p)
   return string(bt)
}

func (p Bytes) StringTh(th InterpreterThread) string {
   return p.String()
}

func (p Bytes) Debug(th InterpreterThread) string {
   bt := ([]byte)(p)
   s := ""
   if len(bt) > 16 {
	   sep := "\n   [ "	
	   for i,b := range bt {
		  if (i > 0) {
			 if (i % 8 == 0) {
		        sep = "\n     "
		     } else {
		        sep = " "	
		     }
		  }		
	      s += sep + fmt.Sprintf("%3d",b)
	   }
	   s += "\n   ]"
   	} else { // Horizontal layout
	   s = "["
	   sep := ""
	   for _,b := range bt {
	      s += sep + fmt.Sprintf("%3d",b)
	      sep = " "
	   }	   
	   s += "]"
   }
   return s
}


func (p Bytes) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p Bytes) UUID() []byte {
	panic("A Bytes cannot have a UUID.")
	return nil
}

func (p Bytes) DBID() int64 {
	panic("A Bytes cannot have a DBID.")
	return 0
}

func (p Bytes) EnsureUUID() (theUUID []byte, err error) {
	panic("A Bytes cannot have a UUID.")
	return
}

func (p Bytes) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Bytes cannot have a UUID.")
	return
}

func (p Bytes) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Bytes cannot have a UUID.")
	return
}

func (p Bytes) UUIDstr() string {
	panic("A Bytes cannot have a UUID.")
	return ""
}

func (p Bytes) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Bytes cannot have a UUID.")
	return
}

func (p Bytes) UUIDabbrev() string {
	panic("A Bytes cannot have a UUID.")
	return ""
}

func (p Bytes) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Bytes cannot have a UUID.")
	return
}

func (p Bytes) RemoveUUID() {
	panic("A Bytes does not have a UUID.")
	return
}

func (p Bytes) Flags() int8 {
	panic("A Bytes has no Flags.")
	return 0
}

func (p Bytes) IsBeingStored() bool {
	return false
}
func (p Bytes) SetBeingStored() {
}
func (p Bytes) ClearBeingStored() {
}

func (p Bytes) IsIdReversed() bool {
	return false
}

func (p Bytes) SetIdReversed() {}

func (p Bytes) ClearIdReversed() {}

func (p Bytes) IsLoadNeeded() bool {
	return false
}

func (p Bytes) SetLoadNeeded()   {}
func (p Bytes) ClearLoadNeeded() {}

func (p Bytes) IsValid() bool { return true }
func (p Bytes) SetValid()     {}
func (p Bytes) ClearValid()   {}

func (p Bytes) IsMarked() bool { return false }
func (p Bytes) SetMarked()    {}
func (p Bytes) ClearMarked()  {}
func (p Bytes) ToggleMarked()  {}

func (p Bytes) Mark() bool { return false }

func (p Bytes) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p Bytes) SetStoredLocally()     {}
func (p Bytes) ClearStoredLocally()   {}

func (p Bytes) IsProxy() bool { return false }

func (p Bytes) IsTransient() bool { return false }

func (p Bytes) Iterable() (sliceOrMap interface{}, err error) {
	return ([]byte)(p),nil
}

func (p Bytes) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   tree = ([]byte)(p)
   return
}

func (p Bytes) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   obj = Bytes(tree.([]byte))
   return
}






type Bits struct {
	b []byte
	n int 
}

/*
TODO This will require fetching the real object
*/
func (p *Bits) IsZero() bool {
	return p.n == 0
}

func (p *Bits) Type() *RType {
	return BitsType
}

func (p *Bits) This() RObject {
	return p
}

func (p *Bits) IsUnit() bool {
	return false
}

func (p *Bits) IsCollection() bool {
	return true
}

/*
How many bits of the bit string are stored in a last extra byte not all of which is used.
If the answer is 0, there is no extra partially used byte.
*/
func (p *Bits) NumRemainderBits() int {
	return p.n % 8
}

func (p *Bits) String() string {
   bt := p.b
   nWholeBytes := len(p.b)
   rem := p.NumRemainderBits() 
   if rem > 0 {
      nWholeBytes--	
   } 

   s := ""
   sep := ""
   var i int
   for i = 0; i < nWholeBytes; i++ {
	   if i > 0 && i % 8 == 0 {
	      sep = "\n"	
	   }
	   s += sep +  fmt.Sprintf("%08b",bt[i])    
	   sep = " "
   }

   if rem > 0 {
	
	  byteString := fmt.Sprintf("%08b",bt[i])   
      if i > 0 && i % 8 == 0 {
         sep = "\n"	
      }	
	  s += sep + byteString[:rem]
   }	
   s += "\n"
   return s	
}


func (p *Bits) StringTh(th InterpreterThread) string {
   return p.String()
}

func (p *Bits) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p *Bits) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p *Bits) UUID() []byte {
	panic("A Bits cannot have a UUID.")
	return nil
}

func (p *Bits) DBID() int64 {
	panic("A Bits cannot have a DBID.")
	return 0
}

func (p *Bits) EnsureUUID() (theUUID []byte, err error) {
	panic("A Bits cannot have a UUID.")
	return
}

func (p *Bits) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A Bits cannot have a UUID.")
	return
}

func (p *Bits) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A Bits cannot have a UUID.")
	return
}

func (p *Bits) UUIDstr() string {
	panic("A Bits cannot have a UUID.")
	return ""
}

func (p *Bits) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A Bits cannot have a UUID.")
	return
}

func (p *Bits) UUIDabbrev() string {
	panic("A Bits cannot have a UUID.")
	return ""
}

func (p *Bits) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A Bits cannot have a UUID.")
	return
}

func (p *Bits) RemoveUUID() {
	panic("A Bits does not have a UUID.")
	return
}

func (p *Bits) Flags() int8 {
	panic("A Bits has no Flags.")
	return 0
}

func (p *Bits) IsBeingStored() bool {
	return false
}
func (p *Bits) SetBeingStored() {
}
func (p *Bits) ClearBeingStored() {
}

func (p *Bits) IsIdReversed() bool {
	return false
}

func (p *Bits) SetIdReversed() {}

func (p *Bits) ClearIdReversed() {}

func (p *Bits) IsLoadNeeded() bool {
	return false
}

func (p *Bits) SetLoadNeeded()   {}
func (p *Bits) ClearLoadNeeded() {}

func (p *Bits) IsValid() bool { return true }
func (p *Bits) SetValid()     {}
func (p *Bits) ClearValid()   {}

func (p *Bits) IsMarked() bool { return false }
func (p *Bits) SetMarked()    {}
func (p *Bits) ClearMarked()  {}
func (p *Bits) ToggleMarked()  {}

func (p *Bits) Mark() bool { return false }

func (p *Bits) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p *Bits) SetStoredLocally()     {}
func (p *Bits) ClearStoredLocally()   {}

func (p *Bits) IsProxy() bool { return false }

func (p *Bits) IsTransient() bool { return false }

func (p *Bits) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Iterable is not implemented yet for Bits.")
}

func (p *Bits) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("Bits type does not yet support ToMapListTree")
   return
}

func (p *Bits) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("Bits type does not yet support FroMapListTree")
   return
}







type CodePoints []rune  // Should we call it Rune ?? Probably not. CodePoint is unambiguous, if boring.

/*
TODO This will require fetching the real object
*/
func (p CodePoints) IsZero() bool {
	return len(p) == 0
}

func (p CodePoints) Type() *RType {
	return CodePointsType
}

func (p CodePoints) This() RObject {
	return p
}

func (p CodePoints) IsUnit() bool {
	return false
}

func (p CodePoints) IsCollection() bool {
	return true
}

func (p CodePoints) String() string {
   bt := ([]rune)(p)
   s := ""
   if len(bt) > 16 {
	   sep := "\n   ["
	   for i,b := range bt {
	      s += sep + fmt.Sprintf("%10d",b)
		  if (i > 0) && (i % 8 == 0) {
		     sep = "\n      "
		  } else {
		     sep = " "	
		  }	
	   }
	   s += "\n   ]"
   	} else { // Horizontal layout
	   s = "["
	   sep := ""
	   for _,b := range bt {
	      s += sep + fmt.Sprintf("%v",b)
	      sep = " "
	   }	   
	   s += "]"
   }
   return s
}


func (p CodePoints) StringTh(th InterpreterThread) string {
   return p.String()
}

func (p CodePoints) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p CodePoints) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p CodePoints) UUID() []byte {
	panic("A CodePoints cannot have a UUID.")
	return nil
}

func (p CodePoints) DBID() int64 {
	panic("A CodePoints cannot have a DBID.")
	return 0
}

func (p CodePoints) EnsureUUID() (theUUID []byte, err error) {
	panic("A CodePoints cannot have a UUID.")
	return
}

func (p CodePoints) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A CodePoints cannot have a UUID.")
	return
}

func (p CodePoints) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A CodePoints cannot have a UUID.")
	return
}

func (p CodePoints) UUIDstr() string {
	panic("A CodePoints cannot have a UUID.")
	return ""
}

func (p CodePoints) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A CodePoints cannot have a UUID.")
	return
}

func (p CodePoints) UUIDabbrev() string {
	panic("A CodePoints cannot have a UUID.")
	return ""
}

func (p CodePoints) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A CodePoints cannot have a UUID.")
	return
}

func (p CodePoints) RemoveUUID() {
	panic("A CodePoints does not have a UUID.")
	return
}

func (p CodePoints) Flags() int8 {
	panic("A CodePoints has no Flags.")
	return 0
}

func (p CodePoints) IsBeingStored() bool {
	return false
}
func (p CodePoints) SetBeingStored() {
}
func (p CodePoints) ClearBeingStored() {
}

func (p CodePoints) IsIdReversed() bool {
	return false
}

func (p CodePoints) SetIdReversed() {}

func (p CodePoints) ClearIdReversed() {}

func (p CodePoints) IsLoadNeeded() bool {
	return false
}

func (p CodePoints) SetLoadNeeded()   {}
func (p CodePoints) ClearLoadNeeded() {}

func (p CodePoints) IsValid() bool { return true }
func (p CodePoints) SetValid()     {}
func (p CodePoints) ClearValid()   {}

func (p CodePoints) IsMarked() bool { return false }
func (p CodePoints) SetMarked()    {}
func (p CodePoints) ClearMarked()  {}
func (p CodePoints) ToggleMarked()  {}

func (p CodePoints) Mark() bool { return false }

func (p CodePoints) IsStoredLocally() bool { return true } // May as well think of it as safely stored. 
func (p CodePoints) SetStoredLocally()     {}
func (p CodePoints) ClearStoredLocally()   {}

func (p CodePoints) IsProxy() bool { return false }

func (p CodePoints) IsTransient() bool { return false }

func (p CodePoints) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}

func (p CodePoints) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = errors.New("CodePoints type does not yet support ToMapListTree")
   return
}

func (p CodePoints) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = errors.New("CodePoints type does not yet support FroMapListTree")
   return
}

/*
Non-persistable objects which wrap a Go object.
Used for os.File presently.
May also be used for other standard library builtin types, but may also be used
for extension of relish with Go objects in custom builds of relish.
Only usable for "unitary" as opposed to collection types.

TODO: These are not being garbage collected yet. If they get the marked flag, make
sure that NewObject sets the marked flag properly on them.
*/

type GoWrapper struct {
	runit
	GoObj interface{}
//	rtype *RType
//	flags byte
}

func (p GoWrapper) IsZero() bool {
	return p.GoObj == nil
}

func (p GoWrapper) Type() *RType {
	return p.rtype
}

func (p *GoWrapper) This() RObject {
	return p
}

/*
Hmmm. TODO
*/
func (p GoWrapper) IsUnit() bool {
	return true
}

/*
Hmmm. TODO Maybe a GoWrapper should be considered a collection???
*/
func (p GoWrapper) IsCollection() bool {
	return false
}

func (p GoWrapper) String() string {
	return fmt.Sprintf("%v", p.GoObj)
}

func (p GoWrapper) StringTh(th InterpreterThread) string {
	return p.String()
}

func (p GoWrapper) Debug(th InterpreterThread) string {
	return p.StringTh(th)
}

func (p GoWrapper) HasUUID() bool {
	return false
}

/*
   TODO We have to figure out what to do with this.
*/
func (p GoWrapper) UUID() []byte {
	panic("A GoWrapper cannot have a UUID.")
	return nil
}

func (p GoWrapper) DBID() int64 {
	panic("A GoWrapper cannot have a DBID.")
	return 0
}

func (p GoWrapper) EnsureUUID() (theUUID []byte, err error) {
	panic("A GoWrapper cannot have a UUID.")
	return
}

func (p GoWrapper) UUIDuint64s() (id uint64, id2 uint64) {
	panic("A GoWrapper cannot have a UUID.")
	return
}

func (p GoWrapper) EnsureUUIDuint64s() (id uint64, id2 uint64, err error) {
	panic("A GoWrapper cannot have a UUID.")
	return
}

func (p GoWrapper) UUIDstr() string {
	panic("A GoWrapper cannot have a UUID.")
	return ""
}

func (p GoWrapper) EnsureUUIDstr() (uuidstr string, err error) {
	panic("A GoWrapper cannot have a UUID.")
	return
}

func (p GoWrapper) UUIDabbrev() string {
	panic("A GoWrapper cannot have a UUID.")
	return ""
}

func (p GoWrapper) EnsureUUIDabbrev() (uuidstr string, err error) {
	panic("A GoWrapper cannot have a UUID.")
	return
}

func (p GoWrapper) RemoveUUID() {
	panic("A GoWrapper does not have a UUID.")
	return
}

func (p GoWrapper) Flags() int8 {
	return int8(p.flags)
}

func (p GoWrapper) IsBeingStored() bool {
	return false
}
func (p GoWrapper) SetBeingStored() {
}
func (p GoWrapper) ClearBeingStored() {
}

func (p GoWrapper) IsIdReversed() bool {
	return false
}

func (p GoWrapper) SetIdReversed() {}

func (p GoWrapper) ClearIdReversed() {}

func (p GoWrapper) IsLoadNeeded() bool {
	return false
}

func (p GoWrapper) SetLoadNeeded()   {}
func (p GoWrapper) ClearLoadNeeded() {}

func (p GoWrapper) IsValid() bool { return true }
func (p GoWrapper) SetValid()     {}
func (p GoWrapper) ClearValid()   {}


func (o GoWrapper) IsMarked() bool { return o.flags&FLAG_MARKED != 0 }
func (o *GoWrapper) SetMarked()    { o.flags |= FLAG_MARKED }
func (o *GoWrapper) ClearMarked()  { o.flags &^= FLAG_MARKED }
func (o *GoWrapper) ToggleMarked()  { o.flags ^= FLAG_MARKED }

/*
If the object is not already marked as reachable, flag it as reachable.
Return whether we had to flag it as reachable. false if was already marked reachable.
*/
func (o *GoWrapper) Mark() bool { 
   if o.IsMarked() == markSense {
       Logln(GC3_,"Mark(): Already marked with",markSense)	
   	   return false
   } 
   o.ToggleMarked()
   Logln(GC3_,"Mark(): Marked with",o.IsMarked())
	o.markAttributes()   
   return true
}



func (o *GoWrapper) markAttributes()  {

	for _, attr := range o.rtype.Attributes {

		if !attr.Part.Type.IsPrimitive {

			val, found := RT.AttrValue(nil, o, attr, false, true, false) // No need to lock - only one thread is running
			if !found {
				break
			}
			val.Mark()	
		} 
	}

	for _, typ := range o.rtype.Up {
		for _, attr := range typ.Attributes {
			if !attr.Part.Type.IsPrimitive {

				val, found := RT.AttrValue(nil, o, attr, false, true, false)  // no need to lock - only one thread is running
				if !found {
					break
				}
			    val.Mark()				
			} 
		}
	}
	return
}




func (p GoWrapper) IsStoredLocally() bool { return false } 
func (p GoWrapper) SetStoredLocally()     {}
func (p GoWrapper) ClearStoredLocally()   {}

func (p GoWrapper) IsProxy() bool { return false }

func (p GoWrapper) IsTransient() bool { return true }


func (p GoWrapper) Iterable() (sliceOrMap interface{}, err error) {
	return nil,errors.New("Expecting a collection or map.")
}


func (p GoWrapper) ToMapListTree(th InterpreterThread, includePrivate bool, visited map[RObject]bool) (tree interface{}, err error) {
   err = fmt.Errorf("Cannot represent a %v in JSON.",p.Type())
   return
}

func (p GoWrapper) FromMapListTree(th InterpreterThread, tree interface{}) (obj RObject, err error) {
   err = fmt.Errorf("Cannot unmarshal JSON into a %v.", p.Type())
   return
}



