// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   rtype.go - types attribute specifications and relation specifications, and 
              also various runtime-internal indexes on these.
*/

import (
	"fmt"
	"strings"
	"hash"
	"hash/adler32"
	. "relish/dbg"
	. "relish/defs"	
	"sort"
	"relish/runtime/native_types"
        "sync"
)

///////////////////////////////////////////////////////////////////////////
////////// DATA TYPES
///////////////////////////////////////////////////////////////////////////

/*
The datatype of a Relish object.
Should it also have a direct map by occurrence pos to type tuples it occurs in?
So far no need for that identified.

Parameterized types are either a type-pattern, which has ParameterConstraints and ParameterTypeVarNames
or they are an actual specific type (an instantiated type pattern), with actual types for its type parameters.
*/
type RType struct {
	Name                    string
	shortName               string
	IsPrimitive             bool
	IsNative                bool  // instances are of type GoWrapper and have a reference to a Go object.
	IsStruct                bool  // instances are of type runit. They can have attributes and if not native can be persisted.
	                              // Question: Should native types like reflect.DataType be IsStruct?
	                              // Right now, they will be.
	Parents                 []*RType
	Children                []*RType
	Up                      []*RType                  // chain of supertypes in dispatch order. (Python-esque?) 
	SpecializationPathNodes []*SpecializationPathNode // All the specialization-generalization chains this
	// type belongs to, with a reference to this type's node in each chain.
	Package                *RPackage
	Attributes             []*AttributeSpec
	AttributesByName       map[string]*AttributeSpec
	NumPrimitiveAttributes int
	TotalAttributeCount    int   // Including supertype attributes - used to size instances
	IsParameterized        bool  // Is a parameterized specific type or a parameterized type-pattern
	IsPattern              bool  // True if this is a parameterized type-pattern with type parameter constraints 
	                             // but no actual type parameters (might it have some parameter constraints and some
                                 // actual parameters?)
	ParameterConstraints   []*RType   // One or more types which type parameters must be same as or a subtype of
	ParameterTypeVarNames  []string   // Type variable names - the use is to identify which parameter types must be the same
	ActualParameters       []*RType   // An instantiated parameterized type lists its actual types of its parameters
	                                  //   Perhaps some of these can be nil if mixed constraints and actuals in type Spec.
	IsPrivate              bool   // Is private - visible only in package which defines it.
}


var TypesMutex sync.RWMutex

/*
   Returns true iff t is a strict subtype of t2.
*/
func (t *RType) Less(t2 *RType) bool {

	// Note do we need to move to type constrained *nil*s ??? Complicated.
	if t == NothingType {
		return t2 != NothingType
	}

    if t2 == AnyType && t != AnyType {
    	return true
    }	
	for _, ti := range t.Up {
		if ti == t2 {
			return true
		}
	}
	if t2 == StructType {
		return t.IsStruct
	}

	return false
}

/*
   Returns true iff t is identical to or a strict subtype of t2.
   This is the type assignment compatibility predicate.
*/
func (t *RType) LessEq(t2 *RType) bool {
	return (t == t2) || t2 == AnyType || t == NothingType || t.Less(t2)
}


/*
Convenience function for parameterized types with a single parameter, such as List_of_X, Channel of X
*/
func (t *RType) ElementType() *RType {
	return t.ActualParameters[0]
}

/*
Convenience function for map types.
*/
func (t *RType) KeyType() *RType {
	return t.ActualParameters[0]	
}

/*
Convenience function for map types.
*/
func (t *RType) ValType() *RType {
	return t.ActualParameters[1]	
}

/*
Convenience function for parameterized types with a single parameter, such as List_of_X, Channel of X
*/
func (t *RType) SetElementType(elementType *RType) {
	t.ActualParameters = []*RType{elementType}
}

/*
Convenience function for map types.
*/
func (t *RType) SetKeyAndValTypes(keyType *RType, valType *RType) {
	t.ActualParameters = []*RType{keyType, valType}	
}


/*
Convenience function for parameterized types with a single parameter, such as List_of_X, Channel of X
*/
func (t *RType) ElementTypeConstraint() *RType {
	return t.ParameterConstraints[0]
}

/*
Convenience function for map types.
*/
func (t *RType) KeyTypeConstraint() *RType {
	return t.ParameterConstraints[0]	
}

/*
Convenience function for map types.
*/
func (t *RType) ValTypeConstraint() *RType {
	return t.ParameterConstraints[1]	
}


/*
   Traverses the supertypes lattice to produce a list of all the supertypes in a
   partial order.
   Order is left to right breadth first.
*/
func computeUpChain(parents []*RType, ups []*RType, visited map[*RType]bool) []*RType {

	grandParents := make([]*RType, 0, len(parents)*3)
	for _, parent := range parents {
		if !visited[parent] {
			ups = append(ups, parent)
			grandParents = append(grandParents, parent.Parents...)
			visited[parent] = true
		}
	}
	if len(grandParents) > 0 {
		ups = computeUpChain(grandParents, ups, visited)
	}
	return ups
}

/*
   Notation:

   // List

   []Byte

   [ 19 213 54 77 ]

   // Set 

   {}Point

   { albatross b c d }

   // Map

   {}String=>Point

   { 
      albatross => b
      c         => d
   }

   // Ordered Set 

   {>}String
   {<}String
   {<lastName<firstName}Person
   {>seniority}Person                  // a unary function
   {?outRanks}Person                   // a binary function

   // Key-ordered Map

   {<}String=>Point

   { 
      albatross => b
      c         => d
      albatross => b
   }

For tests, the following are FALSE:
FALSE bool
0 Bit,Byte,Int8,Uint16,Int16,Uint32,Int32,Uint64,Int
0.0 Float32,Float 
"" String     utf8 code units




        f := *(*float)(unsafe.Pointer(&i)) 

		type word uint32 
		func (w word) float32() float32 { return math.Float32frombits(uint32(w)) } 
		func (w word) int32() int32 { return int32(w) } 
		func (w word) uint32() uint32 { return uint32(w) }

TODO The {} and [] names are not correct. These types are called List_of_Float,Set_of_Float etc.

*/
func newRType(name string, shortName string, parents []*RType) *RType {
	var isPrimitive bool
	switch name {
	//
	// Should we define Int for Int64 Uint for Uint64 and Float for Float64
	//
	case "Int", "Uint", "Int32", "Uint32", "Int16", "Uint16", "Int8", "Uint8", "Bit", "Bool", "CodePoint","Mutex","OwnedMutex","RwMutex":
		fallthrough
	case "[]Int", "[]Uint", "[]Int32", "[]Uint32", "[]Int16", "[]Uint16", "[]Int8", "[]Byte", "[]Bit", "[]Bool", "[]CodePoint":
		fallthrough
	case "{}Int", "{}Uint", "{}Int32", "{}Uint32", "{}Int16", "{}Uint16", "{}Int8", "{}Byte", "{}CodePoint":
		fallthrough
	case "Float", "Float32", "Complex", "Complex32":
		fallthrough
	case "[]Float", "[]Float32", "[]Complex", "[]Complex32":
		fallthrough
	case "{}Float", "{}Float32", "{}Complex", "{}Complex32":
		fallthrough		
	case "Time":
		fallthrough		
	case "{}Time", "[]Time":		
		fallthrough
	case "Nothing":		
		fallthrough		
	case "String":
		isPrimitive = true

	default:
		isPrimitive = false
	}

	upChain := computeUpChain(parents, make([]*RType, 0, len(parents)+5), make(map[*RType]bool))
	Logln(GENERATE_, fmt.Sprintf("UPCHAIN for %s:", name))
	for _, typ := range upChain {
		Logln(GENERATE_, fmt.Sprint(typ))
	}
	Logln(GENERATE_, fmt.Sprintf("END OF UPCHAIN for %s", name))

    if shortName == "" {
	   shortName = name
    }

	typ := &RType{Name: name,
		shortName: shortName,
		IsPrimitive: isPrimitive,
		IsNative: native_types.NativeType[name],
		Parents:     parents,
		Children:    make([]*RType, 0, 5),
		Up:          upChain,
		SpecializationPathNodes: make([]*SpecializationPathNode, 0, 10),
		Attributes:              make([]*AttributeSpec, 0, 10),
		AttributesByName:        make(map[string]*AttributeSpec),
	}
	
	for _,parent := range parents {
		parent.Children = append(parent.Children, typ)
	}
	
	return typ
}

func (t *RType) HasSubtypes() bool {
	return len(t.Children) > 0
}

/*
Computes and returns the closure of the type's direct and indirect subtypes, in breadth first, left-to-right order.
*/
func (t *RType) SubtypeClosure() (subtypes []*RType) {
    subtypes = make([]*RType,len(t.Children))
    copy(subtypes,t.Children)

	for _,st := range t.Children {
		st.subtypeClosure1(&subtypes)
	}
	return
}

func (t *RType) subtypeClosure1(subtypes *[]*RType) {
	var uniqueChildren [] *RType
	Outerloop: 
	   for _,st := range t.Children {
		  // Is st already listed in the closure? 
		   for _,st1 := range *subtypes {
			   if st1 == st {
			   	   continue Outerloop
			   }
		    }
		    uniqueChildren = append(uniqueChildren, st)
		}
    *subtypes = append(*subtypes,uniqueChildren...)
    for _,c := range uniqueChildren {
    	c.subtypeClosure1(subtypes)
    }
}

/*
The universally unique full artifact-and-package-qualified name of the type.
*/
func (t RType) String() string {
	return t.Name
}

/*
A local runtime and local db unique name for the type.
*/
func (t RType) ShortName() string {
    return t.shortName
}

/*
Given a full name or possibly package-abbreviation-qualified shortName of a type,
return just the local (last) part of the type name.
*/
func LocalTypeName(typeFullNameOrShortName string) string {
   slashPos := strings.LastIndex(typeFullNameOrShortName,"/") 
   if slashPos == -1 {
      return typeFullNameOrShortName
   }
   return typeFullNameOrShortName[slashPos+1:]
}

/*
Returns the short name of the package in which the type is defined.
Returns "" if the type is a relish built-in type.
*/
func PackageShortName(typeShortName string) string {
   slashPos := strings.LastIndex(typeShortName,"/") 
   if slashPos == -1 {
      return ""
   }
   return typeShortName[:slashPos]   
}

/*
   Return the sqlite column type name of a primitive type.
   Note that we are using SQL types that might work in something like MySQL, PostgreSQL etc too.
*/
func (t *RType) DbColumnType() (sqlType string) {
	switch t {
	case IntType, UintType:
		sqlType = "BIGINT"
	case Int32Type, Uint32Type, CodePointType:
		sqlType = "INT"
	case Int16Type, Uint16Type:
		sqlType = "SMALLINT"
	case Int8Type, ByteType, BitType:
		sqlType = "TINYINT"
	case BoolType:
		sqlType = "BOOLEAN"
	case FloatType, ComplexType, Complex32Type:
		sqlType = "DOUBLE"
	case Float32Type:
		sqlType = "FLOAT"
	case StringType:
		sqlType = "TEXT"
	default:
		sqlType = ""
	}
	return
}


/*
   Note that this does not check for attribute name overriding (hiding) of supertype attributes.
   We have to decide what is allowed here and what the semantics is.
   Do we do Eiffel-style aliasing of attribute names? Do we disallow overriding
   of an attribute? Do we allow creating an alternate getter/setter in subtype? If so,
   how do we combine the application of the getter/setters?
*/
func (t *RType) addAttribute(attr *AttributeSpec) (err error) {
/*	if len(t.Attributes) == 10 {
		err = fmt.Errorf("A type '%s' cannot have more than 10 attributes.", t.Name)
		return
	}
*/

    existingAttr, found := t.GetAttribute(attr.Part.Name)

// fmt.Printf("=== adding Attribute %s to type %s. Already found = %v\n",attr.Part.Name,t.Name,found)

//	_, found := t.AttributesByName[attr.Part.Name]
	if found {
		if existingAttr.WholeType == t {
		   err = fmt.Errorf("Attempt to redefine attribute '%s' of type '%s'.", attr.Part.Name, t.Name)			
		} else {
		   err = fmt.Errorf("Can't add attribute '%s' to type '%s'. Supertype '%s' has an attribute or relation of same name.", attr.Part.Name, t.Name, existingAttr.WholeType.Name)
	    }
		return
	}
	t.Attributes = append(t.Attributes, attr)
	t.AttributesByName[attr.Part.Name] = attr

	if attr.Part.Type.IsPrimitive {
		t.NumPrimitiveAttributes++
	}

	return
}

/*
This is a HIGHLY UNOPTIMIZED way of getting the attribute spec from the type given the attribute name.
It searches the type's AttributesByName hashtable, and if not found there, traverses the type's Up chain
of supertypes, searching each's  AttributesByName hashtable in turn until an attribute of the specified name
is found. Returns nil if not found

We have to see how we can cache this dispatch, similar to how we cache multi-method dispatch. TODO TODO
*/
func (t *RType) GetAttribute(attrName string) (attr *AttributeSpec, found bool) {
	attr, found = t.AttributesByName[attrName]
	if !found {
		for _, supertype := range t.Up {
			attr, found = supertype.AttributesByName[attrName]
			if found {
				return
			}
		}
	}
	return
}


/*
Return the specialization depth of the type from the top of the inheritance lattice.
The depth is an average of the depth up potentially multiple specialization paths up the lattice.
*/
func (t *RType) SpecializationDepth() float64 {

	var sumSquaredTypeSpecificity uint32 = 0
	var nDimensions uint32 = 0
	
	for _, pathNode := range t.SpecializationPathNodes {
		nDimensions++
		sumSquaredTypeSpecificity += pathNode.Level * pathNode.Level
	}
    
    // nDimensions should never be 0 - all types have a specialization path node.

	return float64(sumSquaredTypeSpecificity) / float64(nDimensions)
}




/*
SPECIAL-PURPOSE METHOD
Used in guessing a type to instantiate for JSON unmarshalling.
Return the subtype of t which has the most (direct and inherited) attributes whose names are keys of the map argument.
If there is a tie, return the most general of the tied types (which, note, could be t itself.)
Also returns the number of attribute names that were matched by the winning subtype.
*/
func (t *RType) BestMatchingSubtype(m map[string]interface{}) (subType *RType, maxMatchingAttrs int) {
   n0, unmatched := t.countAllAttributes(m)
   maxMatchingAttrs = n0
   subType = t
   for _,st := range t.Children {
      st1, n := st.bestMatchingSubtype(unmatched, n0)
      if n > maxMatchingAttrs	{
	     subType = st1
	     maxMatchingAttrs = n
      } else if n == maxMatchingAttrs {
	      // subType = the most general of the types subtype and st1	
	      if st1.SpecializationDepth() < subType.SpecializationDepth() {
		     subType = st1
	      }
      }
   }
   return
}

func (t *RType) bestMatchingSubtype(unmatched []string, nMatchedSoFar int) (subType *RType, maxMatchingAttrs int) {
   nDirect, used1 := t.countDirectAttributes(unmatched)
   maxMatchingAttrs = nMatchedSoFar + nDirect
   subType = t
   for _,st := range t.Children {
      st1, n := st.bestMatchingSubtype(used1, nDirect)
      if n > maxMatchingAttrs	{
	      subType = st1
	      maxMatchingAttrs = n
      } else if n == maxMatchingAttrs {
	      // subType = the most general of the types subtype and st1
	      if st1.SpecializationDepth() < subType.SpecializationDepth() {
		     subType = st1
	      }		
     }
   }
   return
}



/*
Returns how many of the attrNames are the names of direct attributes of the type.
Also returns a list of the unmatched attrNames.
*/
func (t *RType) countDirectAttributes(attrNames []string) (n int, unmatched []string) {
   for _,attrName := range attrNames {
      _,found := t.AttributesByName[attrName]
      if found {
 	     n++
      }	else {
	     unmatched = append(unmatched, attrName)	
      }
   }	
   return
}

/*
Returns how many of the map keys are the names of attributes of the type or its supertypes.
Also returns a list of the unmatched attrNames.
*/
func (t *RType) countAllAttributes(m map[string]interface{}) (n int, unmatched []string) {
   for key := range m {
      _,found := t.GetAttribute(key)
      if found {
 	     n++
      }	else {
	     unmatched = append(unmatched, key)
      }
   }	
   return
}

/*
Return the names of the attributes of the datatype, in alphabetical order.
Arguments determine whether simple, complex attributes are returned and whether inherited attributes are
returned (also subject to the simple &/or complex filter).
*/
func (t *RType) AttributeNames(includeSimple bool, includeComplex bool, includeInherited bool) []string {

    var attrNameSlice []string

	for attrName,attr := range t.AttributesByName {

        if (! includeComplex) && attr.IsComplex() {
        	continue
        }
        if (! includeSimple) && attr.IsSimple() {
        	continue
        }        
		attrNameSlice = append(attrNameSlice, attrName)
	}
	if includeInherited {
		for _, supertype := range t.Up {
	       for attrName,attr := range supertype.AttributesByName {
	          if (! includeComplex) && attr.IsComplex() {
	        	 continue
	          }
	          if (! includeSimple) && attr.IsSimple() {
	          	 continue
	          }  
		      attrNameSlice = append(attrNameSlice, attrName)
	       }
		}		
	}

    sort.Strings(attrNameSlice)

	return attrNameSlice
}


/*
Return a string name of the underlying implementation type of an instance of the type.
Currently only works for collection types.

"list", "sortedlist"

"set", "sortedset"

"map","sortedmap","stringmap","sortedstringmap","uint64map","int64map"

TODO Finish implementing this !!!

*/
func (t *RType) CollectionImplementationType() string {
    if strings.HasPrefix(t.Name, "List_of_") {
       return "list"
    } else if strings.HasPrefix(t.Name, "Set_of_") {
       return "set"
    } 
    return ""
}




/*
A specification of a type of object having an attribute whose value is a given allowable number of a given type
of objects.
*/
type AttributeSpec struct {
	WholeType   *RType
	Part        RelEnd
	IsTransient bool
	IsForwardRelation bool
	IsReverseRelation bool
	Inverse *AttributeSpec

    Index map[*RType]int  // The index in an object of each type where the value of this attribute is found.

	PublicReadable, PackageReadable, SubtypeReadable, PublicWriteable, PackageWriteable, SubtypeWriteable bool
	Reassignable, CollectionMutable, Mutable, DeeplyMutable                                               bool
}

func (attr *AttributeSpec) IsRelation() bool {
	return attr.IsForwardRelation || attr.IsReverseRelation
}

func (attr *AttributeSpec) IsOneWay() bool {
	return ! attr.IsForwardRelation && ! attr.IsReverseRelation
}

/*
   Either a multi-valued attribute or an independent collection.
   In either case, the immediate value of the attribute will be an RCollection.
*/
func (attr *AttributeSpec) IsCollection() bool {
	return attr.Part.CollectionType != ""
}

func (attr *AttributeSpec) IsMultiValued() bool {
	return attr.Part.CollectionType != "" && attr.Part.ArityHigh != 1
}

func (attr *AttributeSpec) IsSimple() bool {
	return attr.Part.Type.IsPrimitive && ! attr.IsCollection()
}

func (attr *AttributeSpec) IsComplex() bool {
	return ! attr.IsSimple()
}

func (attr *AttributeSpec) IsIndependentCollection() bool {
	return attr.Part.CollectionType != "" && attr.Part.ArityHigh == 1
}






func (a *AttributeSpec) IsReassignable() bool {
	return a.Reassignable
}

/*
If this is a multivalued attribute, it means the collection is immutable.
Or is this intended to apply only to a collection-valued attribute? or both?
*/
func (a *AttributeSpec) IsCollectionMutable() bool {
	return a.CollectionMutable
}

/*
Is the attibute-value object itself mutable (or the objects in the collection if this is multivalued attribute.)
*/
func (a *AttributeSpec) IsMutable() bool {
	return a.CollectionMutable
}

/*
Whether the tree of objects referred to by the attribute-value object is mutable.
Presumeably, this just means "is mutable through this reference"
*/
func (a *AttributeSpec) IsDeeplyMutable() bool {
	return a.DeeplyMutable
}

func (a *AttributeSpec) IsPublicReadable() bool {
	return a.PublicReadable
}

/*
Currently always true. If it was false, I guess we would mean an attribute that can
be read in a getter method of the type, but nowhere else. e.g. a hidden bookkeeping attribute.
That concept is not implemented yet. May or may not ever be.
*/
func (a *AttributeSpec) IsPackageReadable() bool {
	return a.PackageReadable
}

func (a *AttributeSpec) IsSubtypeReadable() bool {  // Not implemented yet. May not make sense.
	return a.SubtypeReadable
}

func (a *AttributeSpec) IsPublicWriteable() bool {
	return a.PublicWriteable
}

func (a *AttributeSpec) IsPackageWriteable() bool {
	return a.PackageWriteable
}

func (a *AttributeSpec) IsSubtypeWriteable() bool {   // Not implemented yet. May not make sense.
	return a.SubtypeWriteable
}



/*
One end of a relation - specifies arity and type constraints and a few other details.
*/
type RelEnd struct {
	Name           string
	Type           *RType
	ArityLow       int32
	ArityHigh      int32
	CollectionType string         ///  "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap" ""
	OrderAttr      *AttributeSpec // which primitive attribute of other is it ordered by when retrieving? nil if none

	OrderMethod      *RMultiMethod
	OrderMethodArity int32 // 1 or 2 if applicable
	IsAscending      bool  // ascending order if ordered collection? or descending order

	// Protection    string // "public" "protected" "package" "private"
	DependentPart bool   // delete of parent results in delete of attribute value
}

/*
TODO: Implement
*/
func (e *RelEnd) IsPublicReadable() bool {
	return true
}

/*
   Return the sqlite column definition of a primitive-type attribute.
*/
func (end RelEnd) DbColumnDef() (colDef string) {
	if end.Type == ComplexType {
		colDef = end.Name + "_r DOUBLE,\n" + end.Name + "_i DOUBLE"
	} else if end.Type == Complex32Type {
		colDef = end.Name + "_r DOUBLE,\n" + end.Name + "_i DOUBLE"		
	} else if end.Type == TimeType {
		colDef = end.Name + " TEXT,\n" + end.Name + "_loc TEXT"
	} else {
		colDef = end.Name + " "
		if end.ArityHigh == 1 {
			colDef += end.Type.DbColumnType()
		} else {
			colDef += "BLOB"
		}
	}
	return
}

/*
   Return the sqlite column definition of the part-end of a multi-valued primitive-type attribute
   or the value column(s) of a primitive value collection.
*/
func (typ *RType) DbCollectionColumnDef() (colDef string) {
	if typ == ComplexType {
		colDef = "val_r DOUBLE,\nval_i DOUBLE"
	} else if typ == Complex32Type {
		colDef = "val_r DOUBLE,\nval_i DOUBLE"		
	} else if typ == TimeType {
		colDef = "val TEXT,\nval_loc TEXT"
	} else {
		colDef = "val "
		colDef += typ.DbColumnType()
	}
	return
}


/*
   Return the sqlite column name or names needed to construct a collection element insert statement.
   Also return the SQL variables part of the insert statement appropriate for inserting that type 
   of val into the table.
*/
func (typ *RType) DbCollectionColumnInsert() (sqlFragment string,varsFragment String) {
	if typ == ComplexType {
		sqlFragment = "val_r,val_i"
		varsFragment = "?,?"
	} else if typ == Complex32Type {
		sqlFragment = "val_r,val_i"	
		varsFragment = "?,?"			
	} else if typ == TimeType {
		sqlFragment = "val,val_loc"
		varsFragment = "?,?"		
	} else {
		sqlFragment = "val"
		varsFragment = "?"		
	}
	return
}


func (typ *RType) DbCollectionRemove() (sqlFragment string) {
	if typ == ComplexType {
		sqlFragment = "val_r = ? AND val_i = ?"
	} else if typ == Complex32Type {
      sqlFragment = "val_r = ? AND val_i = ?"		
	} else if typ == TimeType {
		sqlFragment = "val = ? AND val_loc = ?"	
	} else {
		sqlFragment = "val = ?"
	}
	return
}


func (typ *RType) DbCollectionUpdate() (sqlFragment string) {
	if typ == ComplexType {
		sqlFragment = "val_r = ?,val_i = ?"
	} else if typ == Complex32Type {
      sqlFragment = "val_r = ?,val_i = ?"		
	} else if typ == TimeType {
		sqlFragment = "val = ?,val_loc = ?"	
	} else {
		sqlFragment = "val = ?"
	}
	return
}







/*
   The name for this attribute that will become the db table name.
   (only applicable if the attribute's Part Type is non-primitive.)

   Currently the name is e.g. "Cart___wheel__Wheel"
*/
func (attr *AttributeSpec) ShortName() string {
	if attr.IsIndependentCollection() {
	   return fmt.Sprintf("%s___%s__%s_of_%s", attr.WholeType.ShortName(),
		  attr.Part.Name, attr.Part.CollectionType, attr.Part.Type.ShortName())	
	} else {
	   return fmt.Sprintf("%s___%s__%s", attr.WholeType.ShortName(),
		  attr.Part.Name, attr.Part.Type.ShortName())
	}
}



/*
   Create a new relish type.
   The parent types are the direct supertypes of the type.
   The parent types are assumed to exist.
   Returns an error if a type with the name already exists. (What scope?)
   TODO How do we handle incremental compilation that includes redefinitions of types
*/
func (rt *RuntimeEnv) CreateType(typeName string, typeShortName string, parentTypeNames []string) (*RType, error) {
	if _, found := rt.Types[typeName]; found {
		return nil, fmt.Errorf("Attempt to redefine type '%s'.", typeName)
	}
	var parentTypes []*RType = make([]*RType, len(parentTypeNames))
	for i, parentTypeName := range parentTypeNames {
		if parentType, found := rt.Types[parentTypeName]; found {
			parentTypes[i] = parentType
		} else {
			return nil, fmt.Errorf("Defining type '%s' but parent type '%s' does not exist.", typeName, parentTypeName)
		}
	}
	typ := newRType(typeName, typeShortName, parentTypes)

	// Make or extend specialization paths.   
	if len(parentTypes) == 0 {
		// If the new type is a top (most general) type in the type lattice, give it a
		// top-most specialization-path node.
		//
		typ.startSpecializationPath()

	} else { // The new type has parent types

		for _, parentType := range parentTypes {
			if len(parentType.Children) == 0 {

				// If this new type is the first child of the parent type, 
				// then existing specialization-paths involving the parent type
				// will currently end at the parent type. i.e. bottom node in each
				// specialization-path is the parent type.
				// Extend each specialization-path that ends at the parent type
				// down to reach this new sub-type.
				// 
				for _, parentNode := range parentType.SpecializationPathNodes {
					parentNode.extendPath(typ)
				}
			} else {
				// The parent type of the new node already has other children types.
				// So that means the parent type's specialization paths do not
				// end at the parent type but continue down to each of its existing
				// children types.
				//
				// To accommodate the new sibling subtype, we need to create
				// copies of the parent type's specialization paths, or
				// more precisely, of the prefixes of the paths that end at the
				// parent type, then extend each copied prefix path down to reach
				// this new sub-type. 
				// Detail: See discussion below of how many copies
				// need to be made of the parent type's specialization paths.
				//

				// a map from "checksum" hash of specialization-path to specialization-path
				// used to determine the set of upwards-distinct specialization paths that
				// must thus be copied and extended down to the new subtype.
				//
				pathMap := make(map[uint32]*SpecializationPathNode)
				for _, parentNode := range parentType.SpecializationPathNodes {

					copiedNode, pathHash := parentNode.copyUpwards()

					// Now it may be the case that there are different paths up from 
					// the parent type which are identical to each other. They only
					// begin to differ lower down the specialization hierarchy than
					// the parent type. 
					// We only want to keep a copy of one path from each group of
					// upwards-equivalent paths. That one copy will be the new 
					// specialization path that goes down to the new type.
					//
					_, found := pathMap[pathHash]
					if !found {
						pathMap[pathHash] = copiedNode
					}
				}
				for _, copiedNode := range pathMap {
					copiedNode.glue()
					copiedNode.extendPath(typ)
				}
			}
		}
	}
	rt.Types[typeName] = typ
	rt.Typs[typ.ShortName()] = typ
	return typ, nil

}

/*
Debugging function.
*/
func (rt *RuntimeEnv) ListTypes() {
	var typeNames []string
	for typeName := range rt.Types {
		typeNames = append(typeNames, typeName)
	}
	sort.Strings(typeNames)
	for _,typeName := range typeNames {
		fmt.Println(typeName)
	}
}

/*
   Create if necessary and return the type representing a set of some element type.
   THESE ARE ALL WRONG. Should use a single type for all sets with type parameters.
*/
func (rt *RuntimeEnv) GetSetType(elementType *RType) (typ *RType, err error) {
	typeName := "Set_of_" + elementType.Name
	typeShortName := "Set_of_" + elementType.ShortName()	
        TypesMutex.Lock()
	typ, found := rt.Types[typeName]
	if !found {
		typ, err = rt.CreateType(typeName, typeShortName, []string{"Set"})
		typ.IsParameterized = true
		typ.SetElementType(elementType)		
	}
        TypesMutex.Unlock()
	return
}

/*
   Create if necessary and return the type representing a list of some element type.
   THESE ARE ALL WRONG. Should use a single type for all lists with type parameters.
*/
func (rt *RuntimeEnv) GetListType(elementType *RType) (typ *RType, err error) {
	typeName := "List_of_" + elementType.Name
	typeShortName := "List_of_" + elementType.ShortName()
        TypesMutex.Lock()
	typ, found := rt.Types[typeName]
	if !found {
		typ, err = rt.CreateType(typeName, typeShortName, []string{"List"})
		typ.IsParameterized = true		
		typ.SetElementType(elementType)
	}
        TypesMutex.Unlock()
	return
}

/*
   Create if necessary and return the type representing a map from some key type to some value type.
   THESE ARE ALL WRONG. Should use a single type for all maps with type parameters.
*/
func (rt *RuntimeEnv) GetMapType(keyType *RType, valType *RType) (typ *RType, err error) {
	typeName := "Map_of_(" + keyType.Name + ")=>(" + valType.Name + ")"
	typeShortName := "Map_of_(" + keyType.ShortName() + ")=>(" + valType.ShortName() + ")"
        TypesMutex.Lock()
	typ, found := rt.Types[typeName]
	if !found {
		typ, err = rt.CreateType(typeName, typeShortName, []string{"Map"})
		typ.IsParameterized = true		
		typ.SetKeyAndValTypes(keyType,valType)		
	}
        TypesMutex.Unlock()
	return
}

/*
Deprecated comment:
Each type should have a list (map from opposite end name?) of forwardRels (I am end 0) and backRels (where I am end 1)
*/
func (rt *RuntimeEnv) CreateRelation(typeName1 string,
	endName1 string,
	arityLow1 int32,
	arityHigh1 int32,
	collectionType1 string,
	orderFuncOrAttrName1 string,
	isAscending1 bool,	
	typeName2 string,
	endName2 string,
	arityLow2 int32,
	arityHigh2 int32,
	collectionType2 string,
	orderFuncOrAttrName2 string,
	isAscending2 bool,	
	isTransient bool,
	orderings map[string]*AttributeSpec,
    end1PublicReadable,
    end1PackageReadable,
    end1SubtypeReadable,
    end1PublicWriteable,
    end1PackageWriteable,
    end1SubtypeWriteable,
    end1Reassignable,
    end1CollectionMutable,
    end1Mutable,
    end1DeeplyMutable,
    end2PublicReadable,
    end2PackageReadable,
    end2SubtypeReadable,
    end2PublicWriteable,
    end2PackageWriteable,
    end2SubtypeWriteable,
    end2Reassignable,
    end2CollectionMutable,
    end2Mutable,
    end2DeeplyMutable bool) (type1 *RType, type2 *RType,err error) {
				


	forwardRelAttr,err := rt.CreateAttribute(typeName1,
									 	 typeName2,
										 endName2,
										 arityLow2,
										 arityHigh2,   // Is the -1 meaning N respected in here???? TODO
										 collectionType2,
				                         orderFuncOrAttrName2,
				                         isAscending2,
										 isTransient,
										 true,
										 false,
										 orderings,
										 end2PublicReadable,
									     end2PackageReadable,
									     end2SubtypeReadable,
									     end2PublicWriteable,
									     end2PackageWriteable,
									     end2SubtypeWriteable,
									     end2Reassignable,
									     end2CollectionMutable,
									     end2Mutable,
									     end2DeeplyMutable)

   if err != nil {
       return
   }

   reverseRelAttr,err := rt.CreateAttribute(typeName2,
								 	 typeName1,
									 endName1,
									 arityLow1,
									 arityHigh1,   // Is the -1 meaning N respected in here???? TODO
									 collectionType1,
			                         orderFuncOrAttrName1,
			                         isAscending1,
									 isTransient,
									 false,
									 true,
									 orderings,
									 end1PublicReadable,
								     end1PackageReadable,
								     end1SubtypeReadable,
								     end1PublicWriteable,
								     end1PackageWriteable,
								     end1SubtypeWriteable,
								     end1Reassignable,
								     end1CollectionMutable,
								     end1Mutable,
								     end1DeeplyMutable)


   if err != nil {
       return
   }

   forwardRelAttr.Inverse = reverseRelAttr
   reverseRelAttr.Inverse = forwardRelAttr

   type1 = forwardRelAttr.WholeType
   type2 = reverseRelAttr.WholeType

   return
}




/*
Each type should have a list (map from name?) of attributes.

I need to create several auto-generated primitive-implementation methods for each attribute
(getters, setters etc, and make them multimethods!!!!!)
Is this needed in order to get dispatch, or not yet?

*/
func (rt *RuntimeEnv) CreateAttribute(typeName1 string,
	typeName2 string,
	endName2 string,
	arityLow2 int32,
	arityHigh2 int32,
	collectionType2 string,
	orderFuncOrAttrName string,
	isAscending bool,
	isTransient bool,
	isForwardRelation bool,
	isReverseRelation bool,
	orderings map[string]*AttributeSpec,
	publicReadable,
    packageReadable,
    subtypeReadable,
    publicWriteable,
    packageWriteable,
    subtypeWriteable,
    reassignable,
    collectionMutable,
    mutable,
    deeplyMutable bool) (attr *AttributeSpec, err error) {

	typ1, found := rt.Types[typeName1]
	if !found {
		err = fmt.Errorf("Type '%s' not found.", typeName1)
		return
	}
	
	typ2, found := rt.Types[typeName2]
	if !found {
		err = fmt.Errorf("Type '%s' not found.", typeName2)
		return
	}

	isTransient = isTransient || typ2 == MutexType || typ2 == RWMutexType || typ2 == OwnedMutexType

	attr = &AttributeSpec{typ1,
		RelEnd{
			Name:             endName2,
			Type:             typ2,
			ArityLow:         arityLow2,
			ArityHigh:        arityHigh2,
			CollectionType:   collectionType2,
//			OrderAttr:        orderAttr,
//			OrderMethod:      orderMethod,
//			OrderMethodArity: orderMethodArity, // 1 or 2
			IsAscending:      isAscending,
		},
		isTransient,
		isForwardRelation,
		isReverseRelation,
		nil,
		make(map[*RType]int),
		publicReadable,
	    packageReadable,
	    subtypeReadable,
	    publicWriteable,
	    packageWriteable,
	    subtypeWriteable,
	    reassignable,
	    collectionMutable,
	    mutable,
	    deeplyMutable,
	}

    if orderFuncOrAttrName != "" {
	
	    key := typeName1 + KEY_PART_SEPARATOR + endName2 + KEY_PART_SEPARATOR + typeName2 + KEY_PART_SEPARATOR + orderFuncOrAttrName
	
	    orderings[key] = attr
    }	
	
	err = typ1.addAttribute(attr)

	return
}




/*
RelEnd
   ...
   CollectionType string //  "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap" ""
   OrderAttrName string   // which primitive attribute of other is it ordered by when retrieving? "" if none

   OrderMethod *RMultiMethod
*/

/*
   Given a list of typeNames, return the unique RTypeTuple object representing that
   list of types.
   First has to look up the type objects by name in the runtime environment's Types map.

   This function is used during the definition of a method implementation for a particular
   argument-types signature.
*/
func (rt *RuntimeEnv) getTypeTupleFromTypes(typeNames []string) (*RTypeTuple, error) {
    n := len(typeNames)
	mTypes := make([]*RType, len(typeNames))
	for i, typeName := range typeNames {
		typ, found := rt.Types[typeName]
		if !found {
			return nil, fmt.Errorf("Type '%v' not found.", typeName)
		}
		mTypes[i] = typ
	}

	if rt.TypeTupleTrees[n] == nil {
	   tttn := &TypeTupleTreeNode{}
	   tttn.LastTypeTuple = make(map[*RType]*RTypeTuple)
	   rt.TypeTupleTrees[n] = tttn		
	}
	return rt.TypeTupleTrees[n].GetTypeTupleFromTypes(mTypes), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////
//////// SPECIALIZATION PATHS - Used in the calculation of the absolute and relative specificity 
//////// of types and type tuples.
//////// Each specialization path is a single unique path (chain) from general to specific types 
//////// in the multiple-inheritance type lattice.
////////////////////////////////////////////////////////////////////////////////////////////////

/*
   Nodes in a chain (a doubly-linked list) representing a single path from a most general datatype
   down to a most specific datatype.
*/
type SpecializationPathNode struct {
	// DimID uint64 // May not need this.
	Up    *SpecializationPathNode
	Down  *SpecializationPathNode
	Level uint32
	Type  *RType
}

/*
   Return a copy of the specialization path from the node upwards.
   Also return a hash of the names of the types in the path from the
   node upwards.
   Note. The copy of the path is not yet "glued" to its types, in that
   while the path refers to the types, the types do not yet have a reference
   to the copied path. This is  because some copied paths produced when adding
   a new type into the inheritance structure are redundant and are thrown away.
*/
func (node *SpecializationPathNode) copyUpwards() (*SpecializationPathNode, uint32) {
	a32Hash := adler32.New()
	copiedNode := node.copyUpwards1(nil, a32Hash)
	pathHash := a32Hash.Sum32()
	return copiedNode, pathHash
}

/*
   Helper function for copyUpwards. Recursive.
*/
func (node *SpecializationPathNode) copyUpwards1(downNode *SpecializationPathNode, h hash.Hash32) *SpecializationPathNode {
	h.Write([]byte(node.Type.Name))
	copiedNode := &SpecializationPathNode{Up: nil, Down: downNode, Level: node.Level, Type: node.Type}
	if node.Up != nil {
		copiedUpNode := node.Up.copyUpwards1(copiedNode, h)
		copiedNode.Up = copiedUpNode
	}
	return copiedNode
}

/*
   Glues the specialization path to the types by giving the types a reference
   to their node in the path. Second stage of specialization-path creation.
*/
func (node *SpecializationPathNode) glue() {
	for nd := node; nd != nil; nd = nd.Up {
		node.Type.SpecializationPathNodes = append(node.Type.SpecializationPathNodes, nd)
	}
}

/*
   Call this when creating a type with no parent types.
   Gives the type a top-most specialization-path-node.
   The node has specialization-level 0, thus declaring that the type is
   at the general top of the specialization lattice.
*/
func (t *RType) startSpecializationPath() {
	node := &SpecializationPathNode{Up: nil, Down: nil, Level: 0, Type: t}
	t.SpecializationPathNodes = append(t.SpecializationPathNodes, node)
}

/*
Extend the given specialization path downward to reach the new type.
This function should be called on the bottom-most node in the specialization path. 

For a newly copied specialization-path, glue it first then extend it.
*/
func (node *SpecializationPathNode) extendPath(typ *RType) {
	newNode := &SpecializationPathNode{Up: node, Down: nil, Level: node.Level + 1, Type: typ}
	node.Down = newNode
	typ.SpecializationPathNodes = append(typ.SpecializationPathNodes, newNode)
}

/* 
   A specialization path is a unique path down the type specialization lattice.

   Procedure for adding a new subtype:
   1. If the parent type's Children list is empty, 
      extend each path down to the new type.
   2. If the parent has children
      a. Copy all paths, down to the parent only.
         (Two steps: copy, then glue to types. Second step done after uniqueness testing.)
         (while doing so, make a hash or concatenation of the type names,
         and throw away a fresh copy if it has the same hash value.
      b. Extend all copy paths.

*/

////////////////////////////////////////////////////////////////////////////////////////////////
//////// TYPE TUPLES - lists of types that appear as formal parameters, and actual arguments
//////// in definitions, and calls, respectively, of Relish methods.
////////
//////// Part of the Relish language interpreter's multi-method dispatch system.
//////// 
//////// For each unique list of types which has appeared so far in a method definition or
//////// a method call, there will be a single RTypeTuple object, and the runtime environment
//////// maintains a tree index to rapidly find that tuple object given a list of types (or 
//////// typed objects) that is being used a second or subsequent time in a method definition or 
//////// call.
////////
////////////////////////////////////////////////////////////////////////////////////////////////

type RTypeTuple struct {
	Types []*RType
}

func (tt *RTypeTuple) String() string {
	s := "("
	sep := ""
	for _, typ := range tt.Types {
		s += sep + typ.Name
		sep = ","
	}
	s += ")"
	return s
}

/*
   Returns the specialization distance 
   (Squared Euclidian distance in a multi-dimensional type-specialization space) from
   the presumed more general type tuple (superTT) to the presumed more specific type 
   tuple (subTT).
   Distances are divided by the number of dimensions.
   The second result value is the Euclidean specialization depth 
   (specificity compared to topmost general types in the ontology) of the supertype tuple.
   Third result value is set to true if there is no compatible specialization from superTT
   down to subTT.

   Assumes the two type tuples are the same length (arity).
*/
func (subTT *RTypeTuple) SpecializationDistanceFrom(superTT *RTypeTuple) (float64, float64, bool) {
	if len(superTT.Types) == 0 { // Degenerate case. No types to compare.
		return 0.0, 0.0, false
	}
	var sumSquaredLevelDiff uint32 = 0
	var sumSquaredSupertypeSpecificity uint32 = 0
	var nDimensions uint32 = 0
	for i, subType := range subTT.Types {
		superType := superTT.Types[i]
		var foundOnePathForTypeSpecialization bool = false
		
		if subType == NothingType { // type of *nil*
			nDimensions++
			var levelDiff uint32 = 101
			foundOnePathForTypeSpecialization = true	
			for _, node := range superType.SpecializationPathNodes {
			   nDimensions++	
			   sumSquaredLevelDiff += levelDiff * levelDiff		
			   sumSquaredSupertypeSpecificity += node.Level * node.Level										
			}								
		} else {
			for _, pathNode := range subType.SpecializationPathNodes {
				nDimensions++
				node := pathNode
				var found bool = false
				var levelDiff uint32

				if superType == AnyType { // Special case
					levelDiff = 100
					found = true
					foundOnePathForTypeSpecialization = true
				} else if superType == NonPrimitiveType {
					if !node.Type.IsPrimitive { // Special case 
						levelDiff = 99
						found = true
						foundOnePathForTypeSpecialization = true
					}
	   			} else if superType == StructType {
	   				if node.Type.IsStruct { // Special case 
	   					levelDiff = 98
	   					found = true
	   					foundOnePathForTypeSpecialization = true
	   				}
	   			} else {
					for levelDiff = 0; ; levelDiff++ {
						if node.Type == superType {
							found = true
							foundOnePathForTypeSpecialization = true
							break
						} else if node.Up == nil {
							break
						} else {
							node = node.Up
						}
					}
				}
				if found {
					sumSquaredLevelDiff += levelDiff * levelDiff
					sumSquaredSupertypeSpecificity += node.Level * node.Level
				}
			}
	    }
		if !foundOnePathForTypeSpecialization {
			return -1.0, -1.0, true
		}
	}
	return float64(sumSquaredLevelDiff) / float64(nDimensions), float64(sumSquaredSupertypeSpecificity) / float64(nDimensions), false
}


/*
   Returns true if the presumed sub typetuple is the same as or a specialization of the super typetuple.

   Returns false if the sub typetuple is neither the same as nor a specialization of the super typetuple.

   Same logic as SpecializationDistanceFrom without the distance calculations.

   Assumes the two type tuples are the same length (arity).
*/
func (subTT *RTypeTuple) LessEq(superTT *RTypeTuple) bool {
	if len(superTT.Types) == 0 { // Degenerate case. No types to compare.
		return true
	}
	for i, subType := range subTT.Types {
		superType := superTT.Types[i]
		var foundOnePathForTypeSpecialization bool = false
		
		if subType == NothingType { // type of *nil*

			foundOnePathForTypeSpecialization = true	
								
		} else {
			for _, pathNode := range subType.SpecializationPathNodes {
				node := pathNode
				if superType == AnyType { // Special case
					foundOnePathForTypeSpecialization = true
				} else if superType == NonPrimitiveType {
					if !node.Type.IsPrimitive { // Special case 
						foundOnePathForTypeSpecialization = true
					}
	   			} else if superType == StructType {
	   				if node.Type.IsStruct { // Special case 
	   					foundOnePathForTypeSpecialization = true
	   				}
	   			} else {
					for {
						if node.Type == superType {
							foundOnePathForTypeSpecialization = true
							break
						} else if node.Up == nil {
							break
						} else {
							node = node.Up
					    }
					}
				}

			}
	    }
		if !foundOnePathForTypeSpecialization {
			return false
		}
	}
	return true
}
















////////////////////////////////////////////////////////////////////////////
//////// The tree index of type tuples
//
// How do we ensure there is only one typetuple object per
// tuple-signature in the runtime? Would seem to need a tree index structure
// to point to the type tuples. Root is leftmost type.
////////////////////////////////////////////////////////////////////////////

type TypeTupleTreeNode struct {
	mType    *RType                          // nil if root node - this is just documentation
	nextType map[*RType][]*TypeTupleTreeNode // empty map if end of tree
	tuple    *RTypeTuple                     // nil if no typetuple ends here.
	LastTypeTuple  map[*RType]*RTypeTuple  // The most recent tuple that started here.
}

/*
   Return a reference to process-wide unique type tuple representing the 
   list of types of the argument objects.
   TBD: If the tuple is not already created, create it and insert it in the 
   TypeTupleTree (an index that helps later calls of this method find the type tuple.)

   T1 T2 T3 T4

   NIL T1
         T3
         T2 tuple
         T2
           T4
         T2 
           T3 
             T4 
               T5
         T2 
           T3 
             T4

*/
func (tttn *TypeTupleTreeNode) GetTypeTuple(mm *RMultiMethod, mObjects []RObject) *RTypeTuple {
   // fmt.Println("===GetTypeTuple===")
   var typ *RType
   m := len(mObjects)
//   for i:= 0; i < m; i++ {
//   	  if mObjects[i] == nil {
//   	  	  fmt.Println("nil")
//   	  } else {
//         fmt.Println(mObjects[i].Type().ShortName())
//      }
//   }
//   fmt.Println("---")
   if m > 0 {
      obj := mObjects[0]
	  if obj == nil {
        errMsg := fmt.Sprintf("Multimethod '%s' Argument %v is uninitialized.\nMethod Call Arguments:\n",mm, 1);

		//fmt.Println("len mObjects = ",len(mObjects),"i =",i )			
        for j, ob := range mObjects {
		   errMsg += fmt.Sprintf("%d:%v\n", j+1,ob)		       
	    }
	    panic(errMsg)			
      }

      typ = obj.Type()      
      last := tttn.LastTypeTuple[typ]
      if  last != nil {
         for i := 1; i < m; i++ {
            obj := mObjects[i]
		    if obj == nil {
	           errMsg := fmt.Sprintf("Multimethod '%s' Argument %v is uninitialized.\nMethod Call Arguments:\n", mm, i+1);

			   //fmt.Println("len mObjects = ",len(mObjects),"i =",i )			
	           for j, ob := range mObjects {
			      errMsg += fmt.Sprintf("%d:%v\n", j+1,ob)		       
		       }
		       panic(errMsg)			
	        }

            if obj.Type() != last.Types[i] {
               goto find
            }
         }
         return last
      }
   }   
   find:
	tuple := tttn.findOrCreateTypeTuple(mObjects, mObjects)
	if m > 0 {
	   tttn.LastTypeTuple[typ] = tuple
   }
	return tuple
}

/*
start with 
typeTupleTree :=  &RTypeTupleTreeNode{}
Create for T1 T2 

TODO How can this be made thread-safe?????

Does each thread need its own typeTupleTree? Or do we have a single thread that supplies typetuples and dispatch results.
*/
func (tttn *TypeTupleTreeNode) findOrCreateTypeTuple(mObjects []RObject, allObjects []RObject) *RTypeTuple {
	// tttnType := ""
	// if tttn.mType != nil {
    //     tttnType = tttn.mType.ShortName()
	// }
	// fmt.Println("(",tttnType,") findOrCreateTypeTuple:", len(mObjects))
	if len(mObjects) == 0 {
		if tttn.tuple == nil {
//		   fmt.Println("here1: allObjects types:")
		   tttn.tuple = createTypeTuple(allObjects)
//		   for i:= 0; i < len(allObjects); i++ {
//		   	  if allObjects[i] == nil {
//		   	  	  fmt.Println("nil")
//		   	  } else {
//		         fmt.Println(allObjects[i].Type().ShortName())
//		      }
//		   }

		} // else {
			// fmt.Println("existing tuple in tree found.")
		// }
		return tttn.tuple
	}

    var typ *RType
    if mObjects[0] == nil {
	   typ = AnyType
    } else {
	   typ = mObjects[0].Type()  
    }
    var newNode *TypeTupleTreeNode
	if tttn.nextType != nil { // we need to explore the tree from here.
//        fmt.Println("...this node has some nextType nodes")
		nextNodes, found := tttn.nextType[typ]
		if found {
//		    fmt.Println("...there are", len(nextNodes),"nextNodes for type", typ.ShortName())

			for _, node := range nextNodes {
				tpl := node.findOrCreateTypeTuple(mObjects[1:], allObjects)
				if tpl != nil {
//					fmt.Println("(",tttnType,") FOUND a tuple for nextType",typ.ShortName())
					return tpl
				}
			}
			// If got here need to create a new nextType node in the 
			// list under key typ
			panic(fmt.Sprintf("searched all nodes for nextType",typ.ShortName(),"without finding or creating a tuple. SHOULD NOT GET HERE!!" ))
//			fmt.Println("here2:nextType[", typ.ShortName(), "] is a new node")	   		
//			newNode := &TypeTupleTreeNode{mType: typ}
//			tttn.nextType[typ] = append(tttn.nextType[typ], newNode)
		} else {
//           fmt.Println("...this node has no nextType nodes for type", typ.ShortName())
        
		    // If got here need to create a list of nodes with a
		    // single new node for the typ, and store the list under key typ
	
//			fmt.Println("here3: nextType[", typ.ShortName(), "] is a new node")	    	
			newNode = &TypeTupleTreeNode{mType: typ}
			tttn.nextType[typ] = []*TypeTupleTreeNode{newNode}
	    }
	} else {
//		fmt.Println("...this node has no nextType nodes")

	    // If got here need to make the nextType map and
	    // create a list of nodes with a
	    // single new node for the typ, and store the list under key typ

//	    fmt.Println("here4: nextType[", typ.ShortName(), "] is a new node")	  	
	    newNode = &TypeTupleTreeNode{mType: typ}
	    tttn.nextType = make(map[*RType][]*TypeTupleTreeNode)
	    tttn.nextType[typ] = []*TypeTupleTreeNode{newNode}
    }
	tpl := newNode.findOrCreateTypeTuple(mObjects[1:], allObjects)
	return tpl
}

/*
   Given a list of objects, return a RTypeTuple containing a list
   of the types of the objects.
*/
func createTypeTuple(mObjects []RObject) *RTypeTuple {
	tt := &RTypeTuple{Types: make([]*RType, len(mObjects))}
	for i, obj := range mObjects {
		if obj == nil {
            errMsg := fmt.Sprintf("Argument %v is an uninitialized local variable.\nMethod Call Arguments:\n",i+1);
  
			//fmt.Println("len mObjects = ",len(mObjects),"i =",i )			
	        for j, ob := range mObjects {
			   errMsg += fmt.Sprintf("%d:%v\n", j+1,ob)		       
		    }
		    panic(errMsg)			
		}
		tt.Types[i] = obj.Type()
	}
	return tt
}

func (tttn *TypeTupleTreeNode) GetTypeTupleFromTypes(mTypes []*RType) *RTypeTuple {
	return tttn.findOrCreateTypeTupleFromTypes(mTypes, mTypes)
}

/*
start with 
typeTupleTree :=  &RTypeTupleTreeNode{}
Create for T1 T2 

*/
func (tttn *TypeTupleTreeNode) findOrCreateTypeTupleFromTypes(mTypes []*RType, allTypes []*RType) *RTypeTuple {
	//fmt.Printf("findOrCreateTypeTupleForTypes %v %v",mTypes,allTypes)
	if len(mTypes) == 0 {
		if tttn.tuple == nil {
			tttn.tuple = createTypeTupleFromTypes(allTypes)
		}
		return tttn.tuple
	}
	typ := mTypes[0]
    var newNode *TypeTupleTreeNode

	if tttn.nextType != nil { // we need to explore the tree from here.
		nextNodes, found := tttn.nextType[typ]
		if found {
			for _, node := range nextNodes {
				tpl := node.findOrCreateTypeTupleFromTypes(mTypes[1:], allTypes)
				if tpl != nil {
					return tpl
				}
			}
			panic(fmt.Sprintf("searched all nodes for nextType",typ.ShortName(),"without finding or creating a tuple. SHOULD NOT GET HERE!!" ))

			// If got here need to create a new nextType node in the 
			// list under key typ
			//newNode := &TypeTupleTreeNode{mType: typ}
			//tttn.nextType[typ] = append(tttn.nextType[typ], newNode)
		} else {
		   // If got here need to create a list of nodes with a
		   // single new node for the typ, and store the list under key typ
		   newNode = &TypeTupleTreeNode{mType: typ}
		   tttn.nextType[typ] = []*TypeTupleTreeNode{newNode}
		}
	} else {
	   // If got here need to make the nextType map and
	   // create a list of nodes with a
	   // single new node for the typ, and store the list under key typ

	   newNode = &TypeTupleTreeNode{mType: typ}   
	   tttn.nextType = make(map[*RType][]*TypeTupleTreeNode)
	   tttn.nextType[typ] = []*TypeTupleTreeNode{newNode}
    }
	tpl := newNode.findOrCreateTypeTupleFromTypes(mTypes[1:], allTypes)
	return tpl
}

/*
   Given a list of objects, return a RTypeTuple containing a list
   of the types of the objects.
*/
func createTypeTupleFromTypes(mTypes []*RType) *RTypeTuple {
	//fmt.Println("Making type tuple for")
	//fmt.Println(mTypes)
	tt := &RTypeTuple{Types: make([]*RType, len(mTypes))}
	for i, typ := range mTypes {
		tt.Types[i] = typ
	}
	return tt
}
