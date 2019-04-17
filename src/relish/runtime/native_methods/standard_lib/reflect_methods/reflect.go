// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package reflect_methods

/*
   reflect.go - native methods for reflection into language structure (metadata) of relish data.
   These methods are used by types defined in the relish standard library 'reflect' package. 
*/

import (
	. "relish/runtime/data"
    "fmt"
    "sort"
    "strconv"
    "relish"
    "strings"
    "relish/dbg"
    "sync"
    "os"
)

///////////
// Go Types

// None so far


/////////////////////////////////////
// relish method to go method binding

func InitReflectMethods() {

    // typeNames structs Int collections Int primitives Int reflect Int reverseNames > [] String
    // """
    //  Should be alphabetical.
    //
    //  Option semantics: -1 means must be not
    //                     0 means don't care
    //                     1 means must be
    //
    //   Other things to consider filtering on:
    //      builtin, relish lib
    // """
    //
	typeNamesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"typeNames", 
		                                    []string{"structs","collections","primitives","reflect","reverseNames"}, 
		                                    []string{"Int","Int","Int","Int","Bool"}, 
		                                    []string{"List_of_String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	typeNamesMethod.PrimitiveCode = typeNames

/*    
    relish.pl2012/shareware/biblio_file/pkg/publishing/core/BookCase

    becomes

    BookCase ~~~ publishing/core, shareware/biblio_file, relish.pl2012
*/
	backwardsTypeNameMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"backwardsTypeName", 
		                                    []string{"typeName"}, 
		                                    []string{"String"}, 
		                                    []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	backwardsTypeNameMethod.PrimitiveCode = backwardsTypeNameNM



/*    
    BookCase ~~~ publishing/core, shareware/biblio_file, relish.pl2012

    becomes
  
    relish.pl2012/shareware/biblio_file/pkg/publishing/core/BookCase
*/
 	forwardsTypeNameMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"forwardsTypeName", 
		                                    []string{"reversedTypeName"}, 
		                                    []string{"String"}, 
		                                    []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	forwardsTypeNameMethod.PrimitiveCode = forwardsTypeNameNM  


    // type name String > ?DataType
	typeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"type", []string{"name"}, []string{"String"}, []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	typeMethod.PrimitiveCode = typ

    // attributeNames d dataType includeSimple Bool includeComplex Bool includeInherited Bool > [] String

	attributeNamesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"attributeNames", []string{"d","simple","complex","inherited"}, []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType","Bool","Bool","Bool"}, []string{"List_of_String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	attributeNamesMethod.PrimitiveCode = attributeNames 

    // attribute d DataType attributeName String > ?Attribute 
	attributeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"attribute", 
		                                    []string{"d","attributeName"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType","String"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/Attribute"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	attributeMethod.PrimitiveCode = attribute



	// typeOf a Any > DataType
	typeOfMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"typeOf", 
		                                    []string{"a"}, 
		                                    []string{"Any"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	typeOfMethod.PrimitiveCode = typeOf

	// isa a Any d DataType > Bool
	isaMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"isa", 
		                                    []string{"a","d"}, 
		                                    []string{"Any","shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"Bool"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	isaMethod.PrimitiveCode = isa

   
	// supertypes d DataType > [] DataType
	// """
	//  Direct supertypes of d
	// """
	supertypesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"supertypes", 
		                                    []string{"d","DataType"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	supertypesMethod.PrimitiveCode = supertypes

	// subtypes d DataType > [] DataType
	// """
	//  Direct subtypes of d
	// """
	subtypesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"subtypes", 
		                                    []string{"d","DataType"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	subtypesMethod.PrimitiveCode = subtypes 


	// supertypeClosure d DataType > [] DataType
	// """
	//  All direct and indirect supertypes of d, not including d itself.
	// """
	supertypeClosureMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"supertypeClosure", 
		                                    []string{"d","DataType"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	supertypeClosureMethod.PrimitiveCode = supertypeClosure

	// subtypeClosure d DataType > [] DataType
	// """
	//  All direct and indirect subtypes of d, not including d itself.
	// """
	subtypeClosureMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"subtypeClosure", 
		                                    []string{"d","DataType"}, 
		                                    []string{"shared.relish.pl2012/relish_lib/pkg/reflect/DataType"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	subtypeClosureMethod.PrimitiveCode = subtypeClosure 	


	// attrVal obj Any attr Attribute > val Any found Bool
	// """
	//  The value of the specified attribute of the object. May be a collection.
	//  If the object instance has no defined value for the specifeid attribute, found is false.
	// """
	attrValMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"attrVal", 
		                                    []string{"obj","attr"}, 
		                                    []string{"Any","shared.relish.pl2012/relish_lib/pkg/reflect/Attribute"}, 
		                                    []string{"Any","Bool"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	attrValMethod.PrimitiveCode = attrVal  	



   //////////////////////////////
   // reflectId methods



 
/*
 	label obj Any name String > reflectId String
*/   
 	labelMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"label", 
		                                    []string{"obj","name"}, 
		                                    []string{"Any","String"}, 
		                                    []string{"String"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	labelMethod.PrimitiveCode = label
 
 /*
 	unlabel name String
*/   
 	unlabelMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"unlabel", 
		                                    []string{"name"}, 
			                                []string{"String"}, 	                                    
		                                    []string{}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	unlabelMethod.PrimitiveCode = unlabel
 
   /*
   reflectIdByName name String > reflectId String
   """
    Given an object name, which is either a tempCache name (Todo) or a perstence-dubbed name, return
    the reflectId of the object.
   """
   */ 	
 	reflectIdByNameMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"reflectIdByName", 
		                                    []string{"name"}, 
		                                    []string{"String"}, 
		                                    []string{"String"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	reflectIdByNameMethod.PrimitiveCode = reflectIdByName	



   /*
   objectNames prefix String > [] String
   """
    Returns the list of dub'bed or labelled names of objects.
    The names are returned in lexicographic order.
    If a non-empty prefix string is supplied, only names which start with the prefix are returned.
   """
   */ 	
 	objectNamesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"objectNames", 
		                                    []string{"prefix"}, 
		                                    []string{"String"}, 
		                                    []string{"List_of_String"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	objectNamesMethod.PrimitiveCode = objectNames	



   /*
   select typeName String queryConditions String > List
   """
    Returns the list of objects which are compatible with the specified data type
    (specified by its full artifact-and-package-qualified name), and which meet the queryConditions.
    If the typeName is valid and denotes type T, the returned list will be of type [] T, whereas
    if the typeName is invalid, an empty list of type [] Any will be returned. 
   """
   */ 	
 	selectMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"select", 
		                                    []string{"typeName","queryConditions"}, 
		                                    []string{"String","String"}, 
		                                    []string{"List"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	selectMethod.PrimitiveCode = selectByTypeAndConditions	







 
 
 	clearReflectIdsMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"clearReflectIds", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	clearReflectIdsMethod.PrimitiveCode = clearReflectIds
 


   /*
   ensureReflectId obj Any > reflectId String
   """
    Ensures a reflectId exists for the (non-primitive) object and returns that reflectId, using which
    the object can later be retrieved.
   """ 
   */
 	ensureReflectIdMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"ensureReflectId", 
		                                    []string{"obj"}, 
		                                    []string{"Any"}, 
		                                    []string{"String"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	ensureReflectIdMethod.PrimitiveCode = ensureReflectId



   /*
   objectByReflectId reflectId String > obj Any
   """
    Given the reflectId, return the relish object, which may or may nor be persistent.
    If given reflectId "0", returns relish NIL RObject.
    If given an invalid reflectId or one that no longer is mapped to an object,
    also returns NIL
   """ 
   */
 	objectByReflectIdMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"objectByReflectId", 
		                                    []string{"reflectId"}, 
		                                    []string{"String"}, 
		                                    []string{"Any"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	objectByReflectIdMethod.PrimitiveCode = objectByReflectId

 	
 	
   // First, make sure we have the appropriate list types in existence.
 	
   simpleAttrDescriptorType := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/SimpleAttrDescriptor"]
   simpleAttrDescriptorListType, err := RT.GetListType(simpleAttrDescriptorType) 
 	if err != nil {
 		panic(err)
 	}
 	
   complexAttrDescriptorType := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/ComplexAttrDescriptor"]
   complexAttrDescriptorListType, err := RT.GetListType(complexAttrDescriptorType) 
 	if err != nil {
 		panic(err)
 	}

	getSimpleAttributesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"getSimpleAttributes", 
		                                    []string{"reflectId"}, 
		                                    []string{"String"}, 
		                                    []string{simpleAttrDescriptorListType.Name}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	getSimpleAttributesMethod.PrimitiveCode = getSimpleAttributes


	getComplexAttributesMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"getComplexAttributes", 
		                                    []string{"reflectId"}, 
		                                    []string{"String"}, 
		                                    []string{complexAttrDescriptorListType.Name}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	getComplexAttributesMethod.PrimitiveCode = getComplexAttributes


	getCollectionInfoMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"getCollectionInfo", 
		                                    []string{"reflectId"}, 
		                                    []string{"String"}, 
		                                    []string{"String","Int","Int","Bool","Bool","String","String","List_of_String","List_of_String"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	getCollectionInfoMethod.PrimitiveCode = getCollectionInfo




	pauseMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"pause", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	pauseMethod.PrimitiveCode = pause


	resumeMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"resume", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	resumeMethod.PrimitiveCode = resume



	pausedMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"paused", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{"Bool"}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	pausedMethod.PrimitiveCode = paused


	exitMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/reflect",nil,"exit", 
		                                    []string{"code"}, 
		                                    []string{"Int"}, 
		                                    []string{}, 
		                                    false, 0, false)
	if err != nil {
		panic(err)
	}
	exitMethod.PrimitiveCode = exit

}


 
///////////////////////////////////////////////////////////////////////////////////////////
// Reflection functions


    // typeNames structs Int collections Int primitives Int reflect Int reverseNames > [] String
    // """
    //  Should be alphabetical.
    //
    //  Option semantics: -1 means must be not
    //                     0 means don't care
    //                     1 means must be
    //
    //   Other things to consider filtering on:
    //      builtin, relish lib
    // """
    //
func typeNames(th InterpreterThread, objects []RObject) []RObject {

    includeStructs := int(objects[0].(Int))
    includeCollections := int(objects[1].(Int))   
    includePrimitive := int(objects[2].(Int))
    includeReflect := int(objects[3].(Int))       
    reverseNames := bool(objects[4].(Bool))      

    typeNameList, err := RT.Newrlist(StringType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    var typeNameSlice []string

    // fmt.Println("typeNames",includeStructs,includeCollections,includePrimitive,includeReflect)
    for typeName,typ := range RT.Types {
    	// fmt.Println(typeName)

        okStruct := false
        isStruct :=  typ.Less(StructType)
    	if includeStructs == 0 {
            okStruct = true
        } else if includeStructs == 1 && isStruct {
            okStruct = true
    	} else if includeStructs == -1 && ! isStruct {
    		okStruct = true
    	}

        okColl := false
        isColl :=  typ.Less(CollectionType)
    	if includeCollections == 0 {
            okColl = true
        } else if includeCollections == 1 && isColl {
            okColl = true
    	} else if includeCollections == -1 && ! isColl {
    		okColl = true
    	}

    	okPrim := false
    	if includePrimitive == 0 {
            okPrim = true
        } else if includePrimitive == 1 && typ.IsPrimitive {
            okPrim = true
    	} else if includePrimitive == -1 && ! typ.IsPrimitive {
    		okPrim = true
    	}

        okReflect := false
        isReflect := (strings.Index(typeName,"shared.relish.pl2012/relish_lib/pkg/reflect/") != -1)
    	if includeReflect == 0 {
            okReflect = true
        } else if includeReflect == 1 && isReflect {
            okReflect = true
    	} else if includeReflect == -1 && ! isReflect {
    		okReflect = true
    	}

        if okStruct && okColl && okPrim && okReflect {
           if reverseNames {
           	  typeName = backwardsTypeName(typeName)
           }
    	   typeNameSlice = append(typeNameSlice, typeName)
        }
    } 





    sort.Strings(typeNameSlice)

    for _,typeName := range typeNameSlice {
    	typeNameList.AddSimple(String(typeName))
    }

	return []RObject{typeNameList}
}


func backwardsTypeNameNM(th InterpreterThread, objects []RObject) []RObject {

    typeName := string(objects[0].(String))
    reversed := backwardsTypeName(typeName)
	return []RObject{String(reversed)}
}   

func forwardsTypeNameNM(th InterpreterThread, objects []RObject) []RObject {

    reversedTypeName := string(objects[0].(String))
    typeName := forwardsTypeName(reversedTypeName)
	return []RObject{String(typeName)}
} 


/*    
    relish.pl2012/shareware/biblio_file/pkg/publishing/core/BookCase

    becomes

    BookCase ~~~ publishing/core, shareware/biblio_file, relish.pl2012
*/
func backwardsTypeName(typeName string) (backwardsName string) {
   if strings.Contains(typeName,"Map_of_") {
      backwardsName = typeName
      return
   }	
   collectionPrefixes,typeName := extractCollectionPrefixes(typeName)
   slashPos := strings.Index(typeName,"/")
   if slashPos == -1 {
       backwardsName = typeName
       return
   }
   pkgPos := strings.Index(typeName,"/pkg/")

   lastSlashPos := strings.LastIndex(typeName,"/")   
   name := typeName[lastSlashPos+1:]
   origin := typeName[:slashPos]
   artifact := typeName[slashPos+1:pkgPos]
   packagePath := typeName[pkgPos+5:lastSlashPos]
   
   backwardsName = collectionPrefixes + name + " ~~~ " + packagePath + ", " + artifact + ", " + origin

   return
}

/*    
    BookCase ~~~ publishing/core, shareware/biblio_file, relish.pl2012

    becomes
  
    relish.pl2012/shareware/biblio_file/pkg/publishing/core/BookCase
*/
func forwardsTypeName(backwardsTypeName string) (typeName string) {
    if strings.Contains(backwardsTypeName,"Map_of_") {
    	typeName = backwardsTypeName
    	return
    } 	
    collectionPrefixes,backwardsTypeName := extractCollectionPrefixes(backwardsTypeName)	
    pieces := strings.Split(backwardsTypeName,", ")
    if len(pieces) == 1 {
    	typeName = collectionPrefixes + pieces[0]
    } else {
    	nameAndPackage := strings.Split( pieces[0]," ~~~ ")
    	name := nameAndPackage[0]
    	packagePath := nameAndPackage[1]
    	artifact := pieces[1]
    	origin := pieces[2]

        typeName = collectionPrefixes + origin + "/" + artifact + "/pkg/" + packagePath + "/" + name
    }
    return
}

/*
Extract any List_of_Set_of_List_of from beginning of type name.
*/
func extractCollectionPrefixes(typeName string) (collectionPrefixes string, restOfTypeName string) {
   setPos := strings.LastIndex(typeName,"Set_of_")
   listPos := strings.LastIndex(typeName,"List_of_")
   var splitPos int
   if setPos > listPos {
   	   splitPos = setPos + 7
   } else if setPos < listPos {
   	   splitPos = listPos + 8   	
   } else {
   	 restOfTypeName = typeName
   	 return
   }
   collectionPrefixes = typeName[0:splitPos]
   restOfTypeName = typeName[splitPos:]
   return
}

/* Ensure the result is a forwards (official relish) type name,
   whether or not the input name is in backwards (human readable) format.
*/
func normalizeTypeName(possiblyBackwardsTypeName string) (typeName string) {
	if strings.Index(possiblyBackwardsTypeName,", ") != -1 {
		typeName = forwardsTypeName(possiblyBackwardsTypeName)
	} else {
		typeName = possiblyBackwardsTypeName
	}
	return
}


// typ name String > ?DataType
//
func typ(th InterpreterThread, objects []RObject) []RObject {
	
	name := string(objects[0].(String))

	datatype,err := ensureDataType(name) 
	if err != nil {
		datatype = NIL
	}
	return []RObject{datatype}
}


    // attributeNames d dataType includeSimple Bool includeComplex Bool includeInherited Bool > [] String
//
func attributeNames(th InterpreterThread, objects []RObject) []RObject {
	
	datatype := objects[0].(*GoWrapper)
	t := datatype.GoObj.(*RType)

	includeSimple := bool(objects[1].(Bool))
	includeComplex := bool(objects[2].(Bool))		
	includeInherited := bool(objects[3].(Bool))

    attrNameList, err := RT.Newrlist(StringType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    attrNameSlice := t.AttributeNames(includeSimple, includeComplex, includeInherited)

    for _,attrName := range attrNameSlice {
    	attrNameList.AddSimple(String(attrName))
    }

	return []RObject{attrNameList}
}


    // attribute d DataType attributeName String > Attribute
//
func attribute(th InterpreterThread, objects []RObject) []RObject {
	
	datatype := objects[0].(*GoWrapper)
	t := datatype.GoObj.(*RType)

	attributeName := string(objects[1].(String))

    attribute, err := ensureAttribute(t, attributeName)
	if err != nil {
		attribute = NIL
	}
	return []RObject{attribute}
}


	// typeOf a Any > DataType
//
func typeOf(th InterpreterThread, objects []RObject) []RObject {
	
    t := objects[0].Type()

	datatype,err := ensureDataType(t.Name) 
	if err != nil {
		panic(err)
	}

	return []RObject{datatype}
}


	// isa a Any d DataType > Bool
func isa(th InterpreterThread, objects []RObject) []RObject {
	
    t := objects[0].Type()
    datatype := objects[1].(*GoWrapper)	
	t2 := datatype.GoObj.(*RType)
	typeCompatible := t.LessEq(t2)
	
	return []RObject{Bool(typeCompatible)}
}



	// supertypes d DataType > [] DataType
	// """
	//  Direct supertypes of d
	// """
func supertypes(th InterpreterThread, objects []RObject) []RObject {
	
    datatype := objects[0].(*GoWrapper)	
	t := datatype.GoObj.(*RType)

    dataTypeType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/DataType"]
    if ! typFound {
    	panic("reflect.DataType is not defined.")
    }
    supertypeList, err := RT.Newrlist(dataTypeType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,st := range t.Parents {

	   datatype,err := ensureDataType(st.Name) 
	   if err != nil {
		   panic(err)
	   }    	
       supertypeList.AddSimple(datatype)
    }

	return []RObject{supertypeList}
}






	// subtypes d DataType > [] DataType
	// """
	//  Direct subtypes of d
	// """
func subtypes(th InterpreterThread, objects []RObject) []RObject {
	
    datatype := objects[0].(*GoWrapper)	
	t := datatype.GoObj.(*RType)

    dataTypeType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/DataType"]
    if ! typFound {
    	panic("reflect.DataType is not defined.")
    }
    subtypeList, err := RT.Newrlist(dataTypeType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,st := range t.Children {

	   datatype,err := ensureDataType(st.Name) 
	   if err != nil {
		   panic(err)
	   }    	
       subtypeList.AddSimple(datatype)
    }

	return []RObject{subtypeList}
}



	// supertypeClosure d DataType > [] DataType
	// """
	//  All direct and indirect supertypes of d, not including d itself.
	// """
func supertypeClosure(th InterpreterThread, objects []RObject) []RObject {
	
    datatype := objects[0].(*GoWrapper)	
	t := datatype.GoObj.(*RType)

    dataTypeType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/DataType"]
    if ! typFound {
    	panic("reflect.DataType is not defined.")
    }
    supertypeList, err := RT.Newrlist(dataTypeType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,st := range t.Up {

	   datatype,err := ensureDataType(st.Name) 
	   if err != nil {
		   panic(err)
	   }    	
       supertypeList.AddSimple(datatype)
    }

	return []RObject{supertypeList}
}

	// subtypeClosure d DataType > [] DataType
	// """
	//  All direct and indirect subtypes of d, not including d itself.
	// """
func subtypeClosure(th InterpreterThread, objects []RObject) []RObject {
	
    datatype := objects[0].(*GoWrapper)	
	t := datatype.GoObj.(*RType)

    dataTypeType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/DataType"]
    if ! typFound {
    	panic("reflect.DataType is not defined.")
    }
    subtypeList, err := RT.Newrlist(dataTypeType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,st := range t.SubtypeClosure() {

	   datatype,err := ensureDataType(st.Name) 
	   if err != nil {
		   panic(err)
	   }    	
       subtypeList.AddSimple(datatype)
    }

	return []RObject{subtypeList}
}


	// attrVal obj Any attr Attribute > val Any found Bool
	// """
	//  The value of the specified attribute of the object. May be a collection.
	// """
func attrVal(th InterpreterThread, objects []RObject) []RObject {
	
	obj := objects[0]
	attribute := objects[1].(*GoWrapper)	
	attr := attribute.GoObj.(*AttributeSpec)
		
    val, found := RT.AttrVal(th, obj, attr)
    if ! found {
    	val = NIL
    } 

	return []RObject{val, Bool(found)}
}




///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions

/*
  Given the full name of a relish datatype, ensure that there is a reflect.DataType object
  for it and return that object. There is a singleton reflect.DataType for each RType whose
  reflected DataType has been ensured.
  A hashtable in the RunTime is first searched, then if the reflect.DataType is not found,
  a new one is created and put in the RunTime hashtable.
  The reflect.DataType is a *GoWrapper with its GoObj pointing to the underlying *RType.
*/
func ensureDataType(typeName string) (datatype RObject, err error) {

    rtype, found := RT.Types[typeName]
    if ! found {
    	err = fmt.Errorf("No data type has name '%s'", typeName)
    	return
    }

    datatype,found = RT.ReflectedDataTypes[typeName]
    if ! found {
       datatype, err = RT.NewObject("shared.relish.pl2012/relish_lib/pkg/reflect/DataType")
       if err != nil {
    	  return
       }
       dt :=  datatype.(*GoWrapper)
       dt.GoObj = rtype
       RT.ReflectedDataTypes[typeName] = datatype    

       // Now set the attributes of the reflect.DataType object

       nameAttr, found := datatype.Type().GetAttribute("name")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.DataType not have a name attribute?")
    	  return
       }
       RT.RestoreAttr(datatype,  nameAttr, String(typeName)) 
    } 
	return 
}



/*
  Given a reflect.DataType and the name of a direct or inherited attribute, ensure that there is a reflect.Attribute object
  for the attribute and return that relish object. There is a singleton reflect.Attribute for each RType whose
  reflected DataType has been ensured.
  A hashtable in the RunTime is first searched, then if the reflect.Attribute is not found,
  a new one is created and put in the RunTime hashtable.
  The reflect.Attribute is a *GoWrapper with its GoObj pointing to the underlying *AttributeSpec.
*/
func ensureAttribute(t *RType, attrName string) (attribute RObject, err error) {


    attr, found := t.GetAttribute(attrName)
    if ! found {
    	err = fmt.Errorf("DataType '%s' has no direct or inherited attribute '%s'.", t.Name, attrName)
    	return
    }

    attrKey := attr.WholeType.Name + "_|_" + attr.Part.Name

    attribute,found = RT.ReflectedAttributes[attrKey]
    if ! found {
       attribute, err = RT.NewObject("shared.relish.pl2012/relish_lib/pkg/reflect/Attribute")
       if err != nil {
    	  return
       }
       att :=  attribute.(*GoWrapper)
       att.GoObj = attr
       RT.ReflectedAttributes[attrKey] = attribute

       // Now set the attributes of the reflect.Attribute object from the underlying AttributeSpec and RelEnd.
       // NOTE: Or should we leave these as native accessor methods? 

       nameAttr, found := attribute.Type().GetAttribute("name")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.Attribute not have a name attribute?")
    	  return
       }
       RT.RestoreAttr(attribute,  nameAttr, String(attrName)) 


       partType := attr.Part.Type
       var partDataType RObject
       partDataType, err = ensureDataType(partType.Name) 
       if err != nil {
       	  return
       }
       typeAttr, found := attribute.Type().GetAttribute("type")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.Attribute not have a type attribute?")
    	  return
       }
       RT.RestoreAttr(attribute,  typeAttr, partDataType) 

       minArityAttr, found := attribute.Type().GetAttribute("minArity")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.Attribute not have a minArity attribute?")
    	  return
       }
       RT.RestoreAttr(attribute,  minArityAttr, Int(int64(attr.Part.ArityLow))) 

       maxArityAttr, found := attribute.Type().GetAttribute("maxArity")
       if ! found {
    	  err = fmt.Errorf("Hmm. Why does reflect.Attribute not have a maxArity attribute?")
    	  return
       }
       RT.RestoreAttr(attribute,  maxArityAttr, Int(int64(attr.Part.ArityHigh)))        

       if attr.IsRelation() {
          inverseAttr, found := attribute.Type().GetAttribute("inverse")
          if ! found {
    	     err = fmt.Errorf("Hmm. Why does reflect.Attribute not have an inverse attribute?")
    	     return
          }
 
          var inverseAttribute RObject

          inverseAttrKey := inverseAttr.WholeType.Name + "_|_" + inverseAttr.Part.Name

          inverseAttribute,found = RT.ReflectedAttributes[inverseAttrKey]
          if ! found {
             inverseAttribute, err = ensureAttribute(inverseAttr.WholeType, inverseAttr.Part.Name)
             if err != nil {
       	        return
             }       	

             // Now set the inverse attributes for both Attribute objects.

             RT.RestoreAttr(attribute,  inverseAttr, inverseAttribute)             

             RT.RestoreAttr(inverseAttribute,  inverseAttr, attribute) 
          }  // If inverseAttribute already exists, it will be the one setting up the inverses on both sides.
       }
    }
	return 
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////
// Data object instance exploration methods - use a reflectId to access objects 



/*
label obj Any name String > reflectId String
*/
func label(th InterpreterThread, objects []RObject) []RObject {
	
	obj := objects[0]
	name := string(objects[1].(String))	
    reflectId := label1(obj, name) 

	return []RObject{String(reflectId)}
}


/*
unlabel name String 
*/
func unlabel(th InterpreterThread, objects []RObject) []RObject {
	
	name := string(objects[0].(String))	
    unlabel1(name) 

	return []RObject{}
}




/*
clearReflectIds 
"""
 Clear the maps from transient dubbed names and reflectIds to objects.
 Should be done at the end of a debugging session to avoid build up of reflectId cruft in the 
 target runtime, but how do we know when the "end" of a debugging session is?
"""
*/
func clearReflectIds(th InterpreterThread, objects []RObject) []RObject {
	clearReflectIds1()
	return []RObject{}
}

/*
reflectIdByName name String > reflectId String
"""
 Given an object name, which is either a tempCache name (Todo) or a perstence-dubbed name, return
 the reflectId of the object.
"""
*/
func reflectIdByName(th InterpreterThread, objects []RObject) []RObject {

	name := string(objects[0].(String))	

    reflectId := reflectIdByName1(th, name)

	return []RObject{String(reflectId)}
}


/*
objectNames prefix String > [] String
"""
 Given a prefix (which may be the empty String), return a lexicographically ordered list of 
 object names that match the prefix (or of all object names if the prefix is empty.)
 The names are made up of both dubbed and labelled object names.
"""
*/
func objectNames(th InterpreterThread, objects []RObject) []RObject {

	prefix := string(objects[0].(String))	

    nameList, err := RT.Newrlist(StringType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    nameSlice := objectNames1(th, prefix)
    for _,name := range nameSlice {
    	nameList.AddSimple(String(name))
    }

	return []RObject{nameList}
}


/*
select typeName String queryConditions String > List
"""
 Return a list of the persistent objects which are compatible with the type and meet the query conditions.
 TODO Consider promoting this method to builtin method status.
"""
*/
func selectByTypeAndConditions(th InterpreterThread, objects []RObject) []RObject {

	typeName := string(objects[0].(String))	
	queryConditions := string(objects[1].(String))	

	typeName = normalizeTypeName(typeName)	// typeName could be in backwards human readable order.

    objectList := selectByTypeAndConditions1(th, typeName, queryConditions)

	return []RObject{objectList}
}







/*
ensureReflectId obj Any > reflectId String
"""
 Ensures a reflectId exists for the (non-primitive) object and returns that reflectId, using which
 the object can later be retrieved.
""" 
*/
func ensureReflectId(th InterpreterThread, objects []RObject) []RObject {

	obj := objects[0]

    reflectId := ensureReflectId1(obj)

	return []RObject{String(reflectId)}
}


/*
objectByReflectId reflectId String > obj Any
"""
 Given the reflectId, return the relish object, which may or may nor be persistent.
 If given reflectId "0", returns relish NIL RObject.
 If given an invalid reflectId or one that no longer is mapped to an object,
 also returns NIL
""" 
*/
func objectByReflectId(th InterpreterThread, objects []RObject) []RObject {

	reflectId := string(objects[0].(String))	

    obj := objectByReflectId1(reflectId)
    if obj == nil {
    	obj = NIL
    }
	return []RObject{obj}
}


/*
getSimpleAttributes reflectId > [] SimpleAttrDescriptor 
"""
 Get unary atomic-primitive-typed attributes of the object.

      SimpleAttrDescriptor
      """
       A descriptor of a unary primitive attribute and its value for some object instance.
       The value has been converted to type String.
      """
         attrName String
         typeName String
         val String 
"""
*/

func getSimpleAttributes(th InterpreterThread, objects []RObject) []RObject {

	reflectId := string(objects[0].(String))	
    obj := objectByReflectId1(reflectId)

    simpleAttrDescrType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/SimpleAttrDescriptor"]
    if ! typFound {
    	panic("reflect.SimpleAttrDescriptor is not defined.")
    }

    attrDescrList, err := RT.Newrlist(simpleAttrDescrType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }


    if obj != nil && obj != NIL {
    	
       attrs,vals := simpleAttrs(th, obj)

       for i, attr := range attrs {

           val := vals[i]

	       descr, err := RT.NewObject("shared.relish.pl2012/relish_lib/pkg/reflect/SimpleAttrDescriptor")
	       if err != nil {
	    	  panic(err)
	       }

	       attrNameAttr, found := descr.Type().GetAttribute("attrName")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.SimpleAttrDescriptor not have an attrName attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  attrNameAttr, String( attr.Part.Name )   )  

	       typeNameAttr, found := descr.Type().GetAttribute("typeName")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.SimpleAttrDescriptor not have an typeName attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  typeNameAttr, String( attr.Part.Type.Name )   )

	       valAttr, found := descr.Type().GetAttribute("val")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.SimpleAttrDescriptor not have an val attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  valAttr, String( val.String() )   )


           attrDescrList.AddSimple(descr)
       }
    }

	return []RObject{attrDescrList}
}







/*
getComplexAttributes reflectId >

[ [attrName, minArity, maxArity, 
   inverseAttrName, inverseMinArity, inversMaxArity, 
   typeName, valIsObject, valIsCollection
   [val1, val2, ...] 
  ]
  ...
]

			ComplexAttrDescriptor
			"""
			 Temporary
			"""
			   attrName String
			   typeName String
			   minArity Int
			   maxArity Int
			   valIsObject Bool
			   valIsCollection Bool
			   inverseAttrName String
			   inverseMinArity Int
			   inverseMaxArity Int
			   vals [] String
*/
func getComplexAttributes(th InterpreterThread, objects []RObject) []RObject {

	reflectId := string(objects[0].(String))	
	// fmt.Println(reflectId)
    obj := objectByReflectId1(reflectId)

    complexAttrDescrType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/ComplexAttrDescriptor"]
    if ! typFound {
    	panic("reflect.ComplexAttrDescriptor is not defined.")
    }

    attrDescrList, err := RT.Newrlist(complexAttrDescrType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    // fmt.Println(obj)
    
    if obj != nil && obj != NIL {
    	
       attrs,vals := complexAttrs(th, obj)
       // fmt.Println(len(attrs))
       for i, attr := range attrs {

           val := vals[i]

	       descr, err := RT.NewObject("shared.relish.pl2012/relish_lib/pkg/reflect/ComplexAttrDescriptor")
	       if err != nil {
	    	  panic(err)
	       }

	       attrNameAttr, found := descr.Type().GetAttribute("attrName")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an attrName attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  attrNameAttr, String( attr.Part.Name )   )  

	       typeNameAttr, found := descr.Type().GetAttribute("typeName")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an typeName attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  typeNameAttr, String( attr.Part.Type.Name )   )


	       minArityAttr, found := descr.Type().GetAttribute("minArity")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an minArity attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  minArityAttr, Int( int64(attr.Part.ArityLow) )   )


	       maxArityAttr, found := descr.Type().GetAttribute("maxArity")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have a maxArity attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  maxArityAttr, Int( int64(attr.Part.ArityHigh) )   )
	       
	       
	       valIsObjectAttr, found := descr.Type().GetAttribute("valIsObject")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an valIsObject attribute?")
	    	  panic(err)
	       }
	       isObject := ! attr.Part.Type.IsPrimitive
	       RT.RestoreAttr(descr,  valIsObjectAttr, Bool(isObject)   )

	       valIsCollectionAttr, found := descr.Type().GetAttribute("valIsCollection")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an valIsCollection attribute?")
	    	  panic(err)
	       }       
	       isIndependentCollection := false
	       if val.IsCollection() {
	          if val.(RCollection).Owner() == nil {
                isIndependentCollection = true
             }
          }
	       RT.RestoreAttr(descr,  valIsCollectionAttr, Bool(isIndependentCollection)   )

	       

           inverseAttr := attr.Inverse
           inverseAttrName := ""
           if inverseAttr != nil {
	           inverseAttrName = inverseAttr.Part.Name

		       inverseMinArityAttr, found := descr.Type().GetAttribute("inverseMinArity")
		       if ! found {
		    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an inverseMinArity attribute?")
		    	  panic(err)
		       }
		       RT.RestoreAttr(descr,  inverseMinArityAttr, Int( int64(inverseAttr.Part.ArityLow) )   )


		       inverseMaxArityAttr, found := descr.Type().GetAttribute("inverseMaxArity")
		       if ! found {
		    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have a inverseMaxArity attribute?")
		    	  panic(err)
		       }
		       RT.RestoreAttr(descr,  inverseMaxArityAttr, Int( int64(inverseAttr.Part.ArityHigh) )   )		
		   }       

	       inverseAttrNameAttr, found := descr.Type().GetAttribute("inverseAttrName")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an inverseAttrName attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  inverseAttrNameAttr, String( inverseAttrName )   )




           // TODO NOW HANDLE MULTIPLE VALS ETC


	       valsAttr, found := descr.Type().GetAttribute("vals")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.simpleAttrDescriptor not have a vals attribute?")
	    	  panic(err)
	       }
	

	
	       var valStr string
	
	       if attr.IsCollection() && val != NIL {
	          // fmt.Println(attr.Part.Name, "is multivalued")
			    primitive := attr.Part.Type.IsPrimitive
		
		      // val must be a collection - iterate over it !! TODO
	          valColl := val.(RCollection)
	          for obj := range valColl.Iter(th) {
			      if primitive {
				      valStr = obj.String()
				    } else {
					   valStr = ensureReflectId1(obj)
				    }

				    err = RT.AddToAttr(th, descr, valsAttr, String(valStr), false, th.EvaluationContext(), false) 
				  
                  // fmt.Println("adding ",valStr)				  
				  
			       if err != nil {
					    panic(err)
				  }
		      }
		   } else { // not multi-valued, or multi-valued attribute that is uninitialized so far.
	          // fmt.Println(attr.Part.Name, "is not multivalued or is an uninitialized multi-valued attribute")		      
			  valStr = ensureReflectId1(val)
			  err = RT.AddToAttr(th, descr, valsAttr, String(valStr), false, th.EvaluationContext(), false) 
		      if err != nil {
				 panic(err)
			  }		   
		   }

           attrDescrList.AddSimple(descr)
       }
    }

	return []RObject{attrDescrList}
}


func getCollectionInfo(th InterpreterThread, objects []RObject) []RObject {

	reflectId := string(objects[0].(String))	
	// fmt.Println(reflectId)
    obj := objectByReflectId1(reflectId)

    coll, isColl := obj.(RCollection)
    if ! isColl {
       panic("getCollectionInfo called on a non-collection object")
    }

    keyList, err := RT.Newrlist(StringType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }    
    
    valList, err := RT.Newrlist(StringType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }
    
    collectionKind, minArity, maxArity, keyIsObject, valIsObject, keyType, valType, keys, vals := getCollectionInfo1(th, coll)
    
    if collectionKind == "Map" {
       if keyIsObject {
          for _,key := range keys {
             keyList.AddSimple(String(ensureReflectId1(key)))
          }       
       } else {
          for _,key := range keys {
             keyList.AddSimple(String(key.String()))
          }
       }
    }
    if valIsObject {
       for _,val := range vals {
          valList.AddSimple(String(ensureReflectId1(val)))
       }       
    } else {
       for _,val := range vals {
          valList.AddSimple(String(val.String()))
       }
    }    
    
	return []RObject{String(collectionKind),
	                 Int(minArity),
	                 Int(maxArity),
	                 Bool(keyIsObject),
	                 Bool(valIsObject),
	                 String(keyType),
	                 String(valType),
	                 keyList, 
	                 valList,
	                }    
}
    

func getCollectionInfo1(th InterpreterThread, coll RCollection) (collectionKind string, 
     minArity int64, maxArity int64, 
     keyIsObject bool, valIsObject bool, 
     keyType string, valType string, 
     keys []RObject, vals []RObject) {
        
    minArity = coll.MinCard() 
    maxArity = coll.MaxCard() 
	 valType = coll.ElementType().Name   
    valIsObject = ! coll.ElementType().IsPrimitive	     
    if coll.IsMap() {
       collectionKind = "Map"
       theMap := coll.(Map)
       keyType = theMap.KeyType().Name
       keyIsObject = ! theMap.KeyType().IsPrimitive
       for key := range coll.Iter(th) {
          keys = append(keys, key)
       	 val,_ := theMap.Get(th, key)       
          vals = append(vals, val)           
       }  
    } else if coll.IsSet() {
       collectionKind = "Set"
       for val := range coll.Iter(th) {
          vals = append(vals, val)          
       }       
    } else {
       collectionKind = "List"
       for val := range coll.Iter(th) {
          vals = append(vals, val)
       }       
    }                
    return
}
            

/*
getCollectionElements reflectId >

[ [attrName, minArity, maxArity, 
   inverseAttrName, inverseMinArity, inversMaxArity, 
   typeName, valIsObject, valIsCollection
   [val1, val2, ...] 
  ]
  ...
]

			ComplexAttrDescriptor
			"""
			 Temporary
			"""
			   attrName String
			   typeName String
			   minArity Int
			   maxArity Int
			   valIsObject Bool
			   valIsCollection Bool
			   inverseAttrName String
			   inverseMinArity Int
			   inverseMaxArity Int
			   vals [] String

func getCollectionElements(th InterpreterThread, objects []RObject) []RObject {

	reflectId := string(objects[0].(String))	
	// fmt.Println(reflectId)
    obj := objectByReflectId1(reflectId)

    complexAttrDescrType, typFound := RT.Types["shared.relish.pl2012/relish_lib/pkg/reflect/ComplexAttrDescriptor"]
    if ! typFound {
    	panic("reflect.ComplexAttrDescriptor is not defined.")
    }

    attrDescrList, err := RT.Newrlist(complexAttrDescrType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    // fmt.Println(obj)
    
    if obj != nil && obj != NIL {
    	
       attrs,vals := complexAttrs(obj)
       // fmt.Println(len(attrs))
       for i, attr := range attrs {

           val := vals[i]

	       descr, err := RT.NewObject("shared.relish.pl2012/relish_lib/pkg/reflect/ComplexAttrDescriptor")
	       if err != nil {
	    	  panic(err)
	       }

	       attrNameAttr, found := descr.Type().GetAttribute("attrName")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an attrName attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  attrNameAttr, String( attr.Part.Name )   )  

	       typeNameAttr, found := descr.Type().GetAttribute("typeName")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an typeName attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  typeNameAttr, String( attr.Part.Type.Name )   )


	       minArityAttr, found := descr.Type().GetAttribute("minArity")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an minArity attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  minArityAttr, Int( int64(attr.Part.ArityLow) )   )


	       maxArityAttr, found := descr.Type().GetAttribute("maxArity")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have a maxArity attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  maxArityAttr, Int( int64(attr.Part.ArityHigh) )   )
	       
	       
	       valIsObjectAttr, found := descr.Type().GetAttribute("valIsObject")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an valIsObject attribute?")
	    	  panic(err)
	       }
	       isObject := ! attr.Part.Type.IsPrimitive
	       RT.RestoreAttr(descr,  valIsObjectAttr, Bool(isObject)   )

	       valIsCollectionAttr, found := descr.Type().GetAttribute("valIsCollection")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an valIsCollection attribute?")
	    	  panic(err)
	       }       
	       isIndependentCollection := false
	       if val.IsCollection() {
	          if val.(RCollection).Owner() == nil {
                isIndependentCollection = true
             }
          }
	       RT.RestoreAttr(descr,  valIsCollectionAttr, Bool(isIndependentCollection)   )

	       

           inverseAttr := attr.Inverse
           inverseAttrName := ""
           if inverseAttr != nil {
	           inverseAttrName = inverseAttr.Part.Name

		       inverseMinArityAttr, found := descr.Type().GetAttribute("inverseMinArity")
		       if ! found {
		    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an inverseMinArity attribute?")
		    	  panic(err)
		       }
		       RT.RestoreAttr(descr,  inverseMinArityAttr, Int( int64(inverseAttr.Part.ArityLow) )   )


		       inverseMaxArityAttr, found := descr.Type().GetAttribute("inverseMaxArity")
		       if ! found {
		    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have a inverseMaxArity attribute?")
		    	  panic(err)
		       }
		       RT.RestoreAttr(descr,  inverseMaxArityAttr, Int( int64(inverseAttr.Part.ArityHigh) )   )		
		   }       

	       inverseAttrNameAttr, found := descr.Type().GetAttribute("inverseAttrName")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.ComplexAttrDescriptor not have an inverseAttrName attribute?")
	    	  panic(err)
	       }
	       RT.RestoreAttr(descr,  inverseAttrNameAttr, String( inverseAttrName )   )




           // TODO NOW HANDLE MULTIPLE VALS ETC


	       valsAttr, found := descr.Type().GetAttribute("vals")
	       if ! found {
	    	  err = fmt.Errorf("Hmm. Why does reflect.simpleAttrDescriptor not have a vals attribute?")
	    	  panic(err)
	       }
	

	
	       var valStr string
	
	       if attr.IsMultiValued() {
	          // fmt.Println(attr.Part.Name, "is multivalued")
			    primitive := attr.Part.Type.IsPrimitive
		
		      // val must be a collection - iterate over it !! TODO
	          valColl := val.(RCollection)
	          for obj := range valColl.Iter(th) {
			      if primitive {
				      valStr = obj.String()
				    } else {
					   valStr = ensureReflectId1(obj)
				    }

				    err = RT.AddToAttr(th, descr, valsAttr, String(valStr), false, th.EvaluationContext(), false) 
				  
                  // fmt.Println("adding ",valStr)				  
				  
			       if err != nil {
					    panic(err)
				  }
		      }
		   } else {
	          // fmt.Println(attr.Part.Name, "is not multivalued")		      
			  valStr = ensureReflectId1(val)
			  err = RT.AddToAttr(th, descr, valsAttr, String(valStr), false, th.EvaluationContext(), false) 
		      if err != nil {
				 panic(err)
			  }		   
		   }

           attrDescrList.AddSimple(descr)
       }
    }

	return []RObject{attrDescrList}
}


*/













/*

// TBD

isCollection reflectId > Bool
collectionElementType reflectId > typeName
isCollectionOfObjects reflectId > Bool

collectionElements reflectId > [val1, val2,...]
*/


/////////////////////////////////////////////////////////////////////////////////////////////////////////// 
// Helper functions for Data object instance exploration methods - use a reflectId to access objects 
//
// Notes about reflectIds
// ======================
// reflectIds are used as string tokens that represent a particular structured or collection
// object in the current relish program runtime. They can refer to a persistent or non-persistent
// object. The reflectId gives a way for web services to refer to any relish object that is maintained
// in memory in the runtime, whether the object is also persistent or not.
//
// "" is an invalid reflectId, representing the result of a failed search for a reflectId
// "0" is the reflectId of NIL
// reflectIds of regular structured or collection objects are string representations
// of the integers from 1 up in sequence with the order that the object first had a 
// reflectId assigned in the current program run. However, the sequence can be cleared
// and a new sequence can start from some higher integer.

/*
Given an object name, which is either a tempCache name (Todo) or a perstence-dubbed name, return
the reflectId of the object. Note: The transient dubbed namespace is searched before the 
persistent name.
Returns the empty string, an invalid reflectId, if the object is not found by name.
*/
func reflectIdByName1(th InterpreterThread, objectName string) (reflectId string) {

   // Try the label namespace first...	
   reflectId,found := reflectIdsByName[objectName]
   if found {
   	  return
   }

   // Not a transient dubbed name - see if is a persistence-system object name.

   relish.EnsureDatabase()

	found, err := th.DBT().ObjectNameExists(objectName) 
	if err != nil {
		panic(err)
	}
	if ! found {
		return
	}
	obj, err := th.DBT().FetchByName(objectName, 0)
	if err != nil {
		panic(err)
	}
    
	reflectId = ensureReflectId1(obj) 
    return
}


/*
Sorted, label names and persistent dub names, matching the prefix.
*/
func objectNames1(th InterpreterThread, prefix string) (names []string) {

    names, err := th.DBT().ObjectNames(prefix) 
    if err != nil {
	   panic(err)
    }
 
    transNames := transientNames(prefix)
    names = append(transNames, names...)

    sort.Strings(names)    
    return
}

/*
Returns a relish list of the resulting objects.
If the type is not found, an empty list of Any type is returned.
*/
func selectByTypeAndConditions1(th InterpreterThread, typeName string, queryConditions string) RObject {

    if queryConditions == "" {
       queryConditions = "1=1"
    }
    
    t, typeFound := RT.Types[typeName]
    if ! typeFound {
       objectList, err := RT.Newrlist(AnyType, 0, -1, nil, nil, nil)
       if err != nil {
	      panic(err)
       }
	   return objectList
    }

    objectList, err := RT.Newrlist(t, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    queryArgs := []RObject{} 
    radius := 1
    objs := []RObject{} 
    mayContainProxies, err := th.DBT().FetchN(t, queryConditions, queryArgs, nil, radius, &objs)		
    if err != nil {
	    dbg.Log(dbg.ALWAYS_,"Error in selectByTypeAndConditions: %s\n",err)
    }	else {
       objectList.ReplaceContents(objs)
       objectList.SetMayContainProxies(mayContainProxies)
    }
    return objectList
}
 


/*
Given the reflectId, return the relish object, which may or may nor be persistent.
If given reflectId "0", returns relish NIL RObject.
If given an invalid reflectId or one that no longer is mapped to an object,
returns Go nil.
*/
func objectByReflectId1(reflectId string) RObject {	
   if reflectId == "0" {
   	  return NIL
   }
   return objectsByReflectId[reflectId]
}

/*
Given a non-primitive object, return its reflectId. Give the object a reflectId if it
does not already have one.

IMPORTANT NOTE: Presence of a reflectId for an object does not prevent
the object from being garbage-collected (its attribute associations removed etc)
by the relish garbage collector. It does prevent the object from being collected
by Go's garbage collector until the reflectIds maps are cleared.
So if you don't make sure that objects with reflectIds are still referenced by
a relish program variable directly or indirectly, you could retrieve a broken
relish object (with no attribute values) when you get the object by reflectId later.
*/
func ensureReflectId1(obj RObject) (reflectId string) {
   if obj == NIL {
      reflectId = "0"
      return
   }
   if obj == nil {
      panic("ensureReflectId1: Attempt to set a reflectId for RObject which is Go nil.")
   }
   reflectId,found := reflectIdsByObject[obj]
   if ! found {
      reflectId = strconv.FormatUint(idGen.NextID(),10)
      objectsByReflectId[reflectId] = obj
      reflectIdsByObject[obj] = reflectId
      //fmt.Println("Assigning reflectId",reflectId,"to object",obj)
   }
   return
}





/*
Fetches both attribute metadata and values of the simple attributes of the object.
*/
func simpleAttrs(th InterpreterThread, obj RObject) (attrs []*AttributeSpec, vals []RObject) {
   return getAttrs(th, obj,true,false)
}

/*
Fetches both attribute metadata and values of the complex attributes of the object.
*/
func complexAttrs(th InterpreterThread, obj RObject) (attrs []*AttributeSpec, vals []RObject) {
   return getAttrs(th, obj,false,true)
}

/*
Fetches both attribute metadata and values of the simple or complex attributes of the object.
*/
func getAttrs(th InterpreterThread, obj RObject, includeSimple bool, includeComplex bool) (attrs []*AttributeSpec, vals []RObject) {
	includeInherited := true
    attrNames := obj.Type().AttributeNames(includeSimple, includeComplex, includeInherited)
    for _,attrName := range attrNames {  
    	attr,found := obj.Type().GetAttribute(attrName)
    	if ! found {
    		panic("getAttrs: attribute of type unexpectedly not found.")
    	}
    	attrs = append(attrs, attr)

	    val, found := RT.AttrVal(th, obj, attr)
	    if ! found {
	    	val = NIL
	    }     
	    vals = append(vals, val)	
    } 
    return
}





















/*
Associates the object with a name. The association is in memory only, and
actually associates the name with the reflectId of the object, from which
the object can be fetched by the reflection interface as long as reflectIds
have not been cleared.

IMPORTANT NOTE: Presence of an object in this name cache does not prevent
the object from being garbage-collected (its attribute associations removed etc)
by the relish garbage collector. It does prevent the object from being collected
by Go's garbage collector until the reflectIds maps are cleared.
*/
func label1(obj RObject, name string) (reflectId string) {
   reflectId = ensureReflectId1(obj)
   reflectIdsByName[name] = reflectId
   return
}

func unlabel1(name string)  {
   delete(reflectIdsByName,name)
}

func transientNames(prefix string) (names []string) {
   if prefix == "" {
	   for name := range reflectIdsByName {
		  names = append(names,name)
	   }
   } else {
	   for name := range reflectIdsByName {
	      if strings.HasPrefix(name,prefix) {
	      	 names = append(names,name)
	      }
	   }
   }
   sort.Strings(names)
   return
}

/*
   Removes all reflectId assignments.
*/
func clearReflectIds1()  {
   clearTransientNameCache()	
   objectsByReflectId  = make(map[string]RObject)
   reflectIdsByObject  = make(map[RObject]string)
}

	
var idGen *IdGenerator = NewIdGenerator()

var objectsByReflectId map[string]RObject = make(map[string]RObject)

var reflectIdsByObject map[RObject]string = make(map[RObject]string)


// Need some methods to clearReflectIds and clearTempCache

/*
   Removes all reflectId assignments.
*/
func clearTransientNameCache()  {
   reflectIdsByName  = make(map[string]string)
}


var reflectIdsByName map[string]string = make(map[string]string)



// pause resume

var reflectPauseMutex sync.Mutex
var reflectPauseCond = sync.NewCond(&reflectPauseMutex)
var isReflectPaused bool = false;

func pause(th InterpreterThread, objects []RObject) []RObject {
   reflectPauseCond.L.Lock()
   isReflectPaused = true;
   reflectPauseCond.Wait() 
   reflectPauseCond.L.Unlock()  
   return []RObject{}
}

func resume(th InterpreterThread, objects []RObject) []RObject {
   reflectPauseCond.L.Lock()
   reflectPauseCond.Broadcast()
   isReflectPaused = false
   reflectPauseCond.L.Unlock()
   return []RObject{}
}


/*
Whether the relish runtime (at least one of its threads anyway) is
paused and needs a resume.
*/
func paused(th InterpreterThread, objects []RObject) []RObject {
   reflectPauseCond.L.Lock()
   defer reflectPauseCond.L.Unlock()
   return []RObject{Bool(isReflectPaused)}
}


/*
Causes the relish runtime to terminate, returning the code to the OS.
Code 0 should be used to indicate normal termination.
Code 1 for abnormal termination (due to error).
*/
func exit(th InterpreterThread, objects []RObject) []RObject {
   code := int(int64(objects[0].(Int)))
   os.Exit(code)
   return []RObject{}
}

