// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package builtin

/*
   builtin_functions.go - functions with Go-native implementaiton 
                          which require no package specification. They are visible in every package.
*/

import (
	"fmt"
	"relish"
	. "relish/runtime/data"
	"strconv"
	"strings"
	"bytes"
	"io"
	"unicode/utf8"
	"relish/rterr"
    "time"
    "crypto/sha256"
    "encoding/base64"
	"os"
	"bufio"
	"net/smtp"
	"util/smtp_util"
	"net/url"	
	"io/ioutil"
	"encoding/csv"
	"encoding/json"
	"sort"
	"os/exec"
	"sync"
)

// Reader for reading from standard input
var buf *bufio.Reader = bufio.NewReader(os.Stdin)

var relishRuntimeRoot string

func InitBuiltinFunctions(relishRoot string) {

    relishRuntimeRoot = relishRoot

	dbgMethod, err := RT.CreateMethod("",nil,"dbg", []string{"p"}, []string{"Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	dbgMethod.PrimitiveCode = builtinDbg

	debugMethod, err := RT.CreateMethod("",nil,"debug", []string{"p"}, []string{"Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	debugMethod.PrimitiveCode = builtinDebug

    /*
    Returns "" if the object has all its mandatory attributes initialized to non-nil, and
    also is fully connected according to relation declarations.
    Returns an appropriate error message if there is a missing mandatory attribute or a
    relationship cardinality violation.
    */
	incompleteMethod, err := RT.CreateMethod("",nil,"incomplete", []string{"p"}, []string{"Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	incompleteMethod.PrimitiveCode = builtinIncomplete



	printMethod, err := RT.CreateMethod("",nil,"print", []string{"p"}, []string{"Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	printMethod.PrimitiveCode = builtinPrint

	print2Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2"}, []string{"Any", "Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print2Method.PrimitiveCode = builtinPrint

	print3Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3"}, []string{"Any", "Any", "Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print3Method.PrimitiveCode = builtinPrint

	print4Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4"}, []string{"Any", "Any", "Any", "Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print4Method.PrimitiveCode = builtinPrint

	print5Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"Any", "Any", "Any", "Any", "Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print5Method.PrimitiveCode = builtinPrint

	print6Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print6Method.PrimitiveCode = builtinPrint

	print7Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print7Method.PrimitiveCode = builtinPrint

	print8Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7","p8"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any","Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print8Method.PrimitiveCode = builtinPrint

	print9Method, err := RT.CreateMethod("",nil,"print", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7","p8","p9"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any","Any"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	print9Method.PrimitiveCode = builtinPrint	


	inputMethod, err := RT.CreateMethod("",nil,"input", []string{"message"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	inputMethod.PrimitiveCode = builtinInput
	
	
	dbidMethod, err := RT.CreateMethod("",nil,"dbid", []string{"obj"}, []string{"Any"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	dbidMethod.PrimitiveCode = builtinDBID	
	
	uuidMethod, err := RT.CreateMethod("",nil,"uuid", []string{"obj"}, []string{"Any"}, []string{"Bytes"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uuidMethod.PrimitiveCode = builtinUUID
	
	uuidStrMethod, err := RT.CreateMethod("",nil,"uuidStr", []string{"obj"}, []string{"Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uuidStrMethod.PrimitiveCode = builtinUUIDstr
	
	isPersistedLocallyMethod, err := RT.CreateMethod("",nil,"isPersistedLocally", []string{"obj"}, []string{"Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	isPersistedLocallyMethod.PrimitiveCode = builtinIsPersistedLocally	
	
	hasUuidMethod, err := RT.CreateMethod("",nil,"hasUuid", []string{"obj"}, []string{"Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	hasUuidMethod.PrimitiveCode = builtinHasUUID		
	
	
	execMethod, err := RT.CreateMethod("",nil,"exec", []string{"p"}, []string{"String"}, []string{"Bytes","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	execMethod.PrimitiveCode = builtinExec

	exec2Method, err := RT.CreateMethod("",nil,"exec", []string{"p1", "p2"}, []string{"String", "Any"}, []string{"Bytes","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	exec2Method.PrimitiveCode = builtinExec

	exec3Method, err := RT.CreateMethod("",nil,"exec", []string{"p1", "p2", "p3"}, []string{"String", "Any", "Any"}, []string{"Bytes","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	exec3Method.PrimitiveCode = builtinExec

	exec4Method, err := RT.CreateMethod("",nil,"exec", []string{"p1", "p2", "p3", "p4"}, []string{"String", "Any", "Any", "Any"}, []string{"Bytes","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	exec4Method.PrimitiveCode = builtinExec

	exec5Method, err := RT.CreateMethod("",nil,"exec", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"String", "Any", "Any", "Any", "Any"}, []string{"Bytes","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	exec5Method.PrimitiveCode = builtinExec

	exec6Method, err := RT.CreateMethod("",nil,"exec", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"String", "Any", "Any", "Any", "Any", "Any"}, []string{"Bytes","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	exec6Method.PrimitiveCode = builtinExec

	exec7Method, err := RT.CreateMethod("",nil,"exec", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Bytes","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	exec7Method.PrimitiveCode = builtinExec
	
	
		
		
	execNoWaitMethod, err := RT.CreateMethod("",nil,"execNoWait", []string{"p"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	execNoWaitMethod.PrimitiveCode = builtinExecNoWait

	execNoWait2Method, err := RT.CreateMethod("",nil,"execNoWait", []string{"p1", "p2"}, []string{"String", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	execNoWait2Method.PrimitiveCode = builtinExecNoWait

	execNoWait3Method, err := RT.CreateMethod("",nil,"execNoWait", []string{"p1", "p2", "p3"}, []string{"String", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	execNoWait3Method.PrimitiveCode = builtinExecNoWait

	execNoWait4Method, err := RT.CreateMethod("",nil,"execNoWait", []string{"p1", "p2", "p3", "p4"}, []string{"String", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	execNoWait4Method.PrimitiveCode = builtinExecNoWait

	execNoWait5Method, err := RT.CreateMethod("",nil,"execNoWait", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"String", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	execNoWait5Method.PrimitiveCode = builtinExecNoWait

	execNoWait6Method, err := RT.CreateMethod("",nil,"execNoWait", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"String", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	execNoWait6Method.PrimitiveCode = builtinExecNoWait

	execNoWait7Method, err := RT.CreateMethod("",nil,"execNoWait", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	execNoWait7Method.PrimitiveCode = builtinExecNoWait		
	
	
	panicMethod, err := RT.CreateMethod("",nil,"panic", []string{"message"}, []string{"String"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	panicMethod.PrimitiveCode = builtinPanic
	



	email1Method, err := RT.CreateMethod("",nil,
	                                     "sendEmail", 
	                                     []string{"smtpServerAddr","from","recipient","subject","messageBody"}, 
	                                     []string{"String","String","String","String","String"}, 
	                                     []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	email1Method.PrimitiveCode = builtinSendEmail



	email2Method, err := RT.CreateMethod("",nil,
	                                     "sendEmail", 
	                                     []string{"smtpServerAddr","from","recipients","subject","messageBody"}, 
	                                     []string{"String","String","List_of_String","String","String"}, 
	                                     []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	email2Method.PrimitiveCode = builtinSendEmail


	email3Method, err := RT.CreateMethod("",nil,
	                                     "sendEmail", 
	                                     []string{"smtpServerAddr","user","password","from","recipient","subject","messageBody"}, 
	                                     []string{"String","String","String","String","String","String","String"}, 
	                                     []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	email3Method.PrimitiveCode = builtinSendEmail


	email4Method, err := RT.CreateMethod("",nil,
	                                     "sendEmail", 
	                                     []string{"smtpServerAddr","user","password","from","recipients","subject","messageBody"}, 
	                                     []string{"String","String","String","String","List_of_String","String","String"}, 
	                                     []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	email4Method.PrimitiveCode = builtinSendEmail




    ////////////////////////////////////////////////////////////
    // http client functions

    // httpGet url String > responseBody String err String
    //
    //
	httpGetMethod, err := RT.CreateMethod("relish/pkg/http",nil,"httpGet", []string{"url"}, []string{"String"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	httpGetMethod.PrimitiveCode = builtinHttpGet

    // httpGet url String timeoutSeconds Int > responseBody String err String
    //
    //
	httpGet2Method, err := RT.CreateMethod("relish/pkg/http",nil,"httpGet", []string{"url","timeoutSeconds"}, []string{"String","Int"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	httpGet2Method.PrimitiveCode = builtinHttpGet


    // httpPost 
    //    url String 
    //    keysVals {} String > [] String
    // > 
    //    responseBody String 
    //    err String
    //
	httpPostMethod, err := RT.CreateMethod("relish/pkg/http",nil,"httpPost", []string{"url","keysVals"}, []string{"String","Map"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	httpPostMethod.PrimitiveCode = builtinHttpPost

    // httpPost 
    //    url String 
    //    keysVals {} String > [] String
    //    timeoutSeconds Int
    // > 
    //    responseBody String 
    //    err String
    //
	httpPost2Method, err := RT.CreateMethod("relish/pkg/http",nil,"httpPost", []string{"url","keysVals","timeoutSeconds"}, []string{"String","Map","Int"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	httpPost2Method.PrimitiveCode = builtinHttpPost


	// httpPost 
	//    url String 
	//    mimeType String
	//    bodyContentToPost String
	// > 
	//    responseBody String 
	//    err String
	// """
	//  Posts the data with the specified mime type as the Content-type: header. 
	// """	
	httpPost3Method, err := RT.CreateMethod("relish/pkg/http",nil,"httpPost", []string{"url","mimeType","bodyContent"}, []string{"String","String","String"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	httpPost3Method.PrimitiveCode = builtinHttpPost


	// httpPost 
	//    url String 
	//    mimeType String
	//    bodyContentToPost String
	//    timeoutSeconds Int
	// > 
	//    responseBody String 
	//    err String
	// """
	//  Posts the data with the specified mime type as the Content-type: header. 
	// """	
	httpPost4Method, err := RT.CreateMethod("relish/pkg/http",nil,"httpPost", []string{"url","mimeType","bodyContent","timeoutSeconds"}, []string{"String","String","String","Int"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	httpPost4Method.PrimitiveCode = builtinHttpPost




















	// csvRead
	//    content String 
	// > 
	//    records [] [] String
	//    err String
	// """
	//  
	// """
	//
	// csvRead
	//    content String 
	//    trimSpace Bool
	// > 
	//    records [] [] String
	//    err String
	// """
	//  
	// """
	//
	// csvRead
	//    content String 
	//    trimSpace Bool
	//    fieldsPerRecord Int
	// > 
	//    records [] [] String
	//    err String
	// """
	//  If trimSpace is supplied and true, leading and trailing space is trimmed from each field value.
	//  If fieldsPerRecord is supplied and negative, no check is made and each row (each record) can
	//  have a different number of fields.
	//  If fieldsPerRecord is 0 (the default) each record must have same number of fields
	//  as the first record (first row)
	//  If fieldsPerRecord is supplied and is a positive integer, it specifies the required number 
	//  of fields per record
	// """
	//
	// Note: The declared list type spec is not correct since not expressible yet.
	//
	csvReadMethod, err := RT.CreateMethod("relish/pkg/csv",nil,"csvRead", []string{"content"}, []string{"String"}, []string{"List","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	csvReadMethod.PrimitiveCode = builtinCsvRead

	csvRead2Method, err := RT.CreateMethod("relish/pkg/csv",nil,"csvRead", []string{"content","trimSpace"}, []string{"String","Bool"}, []string{"List","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	csvRead2Method.PrimitiveCode = builtinCsvRead

	csvRead3Method, err := RT.CreateMethod("relish/pkg/csv",nil,"csvRead", []string{"content","trimSpace","fieldsPerRecord"}, []string{"String","Bool","Int"}, []string{"List","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	csvRead3Method.PrimitiveCode = builtinCsvRead	
	
	// csvWrite
	//    records [] [] String 
	// > 
	//    content String 
	//    err String
	// """
	//  Each record (i.e. data row) is a list of Strings, one String per field.
	//  Writes the records out in CSV format.
	//  If there is an error, the returned content string will be incomplete and
	//  the second return value will contain an error message.
	//  If no error, the second return value will be the empty String ""  
	// """
	//
	// csvWrite
	//    records [] [] String 
	//    useCrLf Bool
	// > 
	//    content String 
	//    err String
	// """
	//  Each record (i.e. data row) is a list of Strings, one String per field.
	//  Writes the records out in CSV format.
	//  If there is an error, the returned content string will be incomplete and
	//  the second return value will contain an error message.
	//  If no error, the second return value will be the empty String ""  
	//  If userCrLf is supplied and true, each line terminates with \r\n instead of
	//  the default \n.  \r\n is the Windows-compatible text file line ending,
	//  whereas the default \n is the linux and macosx compatible text file format.
	// """
	//
	// Note: The declared list type spec is not correct since not expressible yet.
	//
	csvWriteMethod, err := RT.CreateMethod("relish/pkg/csv",nil,"csvWrite", []string{"records"}, []string{"List"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	csvWriteMethod.PrimitiveCode = builtinCsvWrite

	csvWrite2Method, err := RT.CreateMethod("relish/pkg/csv",nil,"csvWrite", []string{"records","useCrLf"}, []string{"List","Bool"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	csvWrite2Method.PrimitiveCode = builtinCsvWrite








	/*
	jsonMarshal val Any > jsonEncoded String err String
	"""
	 Marshals the argument value/object/collection including only the public
	 attributes of structured objects.
	"""

	jsonMarshal val Any includePrivate bool > jsonEncoded String err String  
	"""
	 Marshals the argument value/object/collection including the public
	 and if indicated also the private attributes of structured objects.
	"""

	*/	
	jsonMarshalMethod, err := RT.CreateMethod("relish/pkg/json",nil,"jsonMarshal", []string{"val"}, []string{"Any"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	jsonMarshalMethod.PrimitiveCode = builtinJsonMarshal	
	
	jsonMarshal2Method, err := RT.CreateMethod("relish/pkg/json",nil,"jsonMarshal", []string{"val","includePrivate"}, []string{"Any","Bool"}, []string{"String","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	jsonMarshal2Method.PrimitiveCode = builtinJsonMarshal	
	
	

	/*
	jsonUnmarshal json String > val Any err String 
	jsonUnmarshal json String obj Any > obj Any err String 
	"""


	 Decodes the json-encoded string argument into a relish object or object tree, which is returned.
	 If the second argument is supplied, attempts to populate the supplied object, 
	 which must be a collection, a map, or a structured object. In that case, the object argument itself
	 is returned, after being populated with attribute/collection values from the JSON string.
	"""
	*/
	jsonUnmarshalMethod, err := RT.CreateMethod("relish/pkg/json",nil,"jsonUnmarshal", []string{"json"}, []string{"String"}, []string{"Any","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	jsonUnmarshalMethod.PrimitiveCode = builtinJsonUnmarshal	
	
	jsonUnmarshal2Method, err := RT.CreateMethod("relish/pkg/json",nil,"jsonUnmarshal", []string{"json","prototype"}, []string{"String","Any"}, []string{"Any","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	jsonUnmarshal2Method.PrimitiveCode = builtinJsonUnmarshal


	rtPathMethod, err := RT.CreateMethod("relish/pkg/env",nil,"rtPath", []string{}, []string{}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	rtPathMethod.PrimitiveCode = builtinRtPath



	eqMethod, err := RT.CreateMethod("",nil,"eq", []string{"p1", "p2"}, []string{"RelishPrimitive", "RelishPrimitive"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	eqMethod.PrimitiveCode = builtinEq
	
	eq1Method, err := RT.CreateMethod("",nil,"eq", []string{"p1", "p2"}, []string{"Any", "RelishPrimitive"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	eq1Method.PrimitiveCode = builtinEqNo
	
	eq2Method, err := RT.CreateMethod("",nil,"eq", []string{"p1", "p2"}, []string{"RelishPrimitive", "Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	eq2Method.PrimitiveCode = builtinEqNo	
	
	eqObjMethod, err := RT.CreateMethod("",nil,"eq", []string{"p1", "p2"}, []string{"Any", "Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	eqObjMethod.PrimitiveCode = builtinEqObj
	
	
	
	neqMethod, err := RT.CreateMethod("",nil,"neq", []string{"p1", "p2"}, []string{"RelishPrimitive", "RelishPrimitive"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	neqMethod.PrimitiveCode = builtinNeq
	
	neq1Method, err := RT.CreateMethod("",nil,"neq", []string{"p1", "p2"}, []string{"Any", "RelishPrimitive"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	neq1Method.PrimitiveCode = builtinNeqNo
	
	neq2Method, err := RT.CreateMethod("",nil,"neq", []string{"p1", "p2"}, []string{"RelishPrimitive", "Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	neq2Method.PrimitiveCode = builtinNeqNo	
	
	neqObjMethod, err := RT.CreateMethod("",nil,"neq", []string{"p1", "p2"}, []string{"Any", "Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	neqObjMethod.PrimitiveCode = builtinNeqObj	
	
			

	ltNumMethod, err := RT.CreateMethod("",nil,"lt", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	ltNumMethod.PrimitiveCode = builtinLtNum

	ltTimeMethod, err := RT.CreateMethod("",nil,"lt", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	ltTimeMethod.PrimitiveCode = builtinLtTime	

	ltStrMethod, err := RT.CreateMethod("",nil,"lt", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	ltStrMethod.PrimitiveCode = builtinLtStr
	
	
	
	lessNumMethod, err := RT.CreateMethod("",nil,"less", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessNumMethod.PrimitiveCode = builtinLtNum

	lessTimeMethod, err := RT.CreateMethod("",nil,"less", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessTimeMethod.PrimitiveCode = builtinLtTime	

	lessStrMethod, err := RT.CreateMethod("",nil,"less", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessStrMethod.PrimitiveCode = builtinLtStr	
	
	beforeTimeMethod, err := RT.CreateMethod("",nil,"before", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	beforeTimeMethod.PrimitiveCode = builtinLtTime	
	
	
	

	lteNumMethod, err := RT.CreateMethod("",nil,"lte", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lteNumMethod.PrimitiveCode = builtinLteNum

	lteTimeMethod, err := RT.CreateMethod("",nil,"lte", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lteTimeMethod.PrimitiveCode = builtinLteTime	

	lteStrMethod, err := RT.CreateMethod("",nil,"lte", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lteStrMethod.PrimitiveCode = builtinLteStr
	
	
	
	lessEqNumMethod, err := RT.CreateMethod("",nil,"lessEq", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessEqNumMethod.PrimitiveCode = builtinLteNum

	lessEqTimeMethod, err := RT.CreateMethod("",nil,"lessEq", []string{"p1", "p2"}, []string{"Time","Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessEqTimeMethod.PrimitiveCode = builtinLteTime	

	lessEqStrMethod, err := RT.CreateMethod("",nil,"lessEq", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lessEqStrMethod.PrimitiveCode = builtinLteStr	
	
	
	
	

	
	gtNumMethod, err := RT.CreateMethod("",nil,"gt", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gtNumMethod.PrimitiveCode = builtinGtNum

	gtTimeMethod, err := RT.CreateMethod("",nil,"gt", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gtTimeMethod.PrimitiveCode = builtinGtTime	

	gtStrMethod, err := RT.CreateMethod("",nil,"gt", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gtStrMethod.PrimitiveCode = builtinGtStr	
	
	
	
	
	greaterNumMethod, err := RT.CreateMethod("",nil,"greater", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterNumMethod.PrimitiveCode = builtinGtNum

	greaterTimeMethod, err := RT.CreateMethod("",nil,"greater", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterTimeMethod.PrimitiveCode = builtinGtTime	

	greaterStrMethod, err := RT.CreateMethod("",nil,"greater", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterStrMethod.PrimitiveCode = builtinGtStr
		
	afterTimeMethod, err := RT.CreateMethod("",nil,"after", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	afterTimeMethod.PrimitiveCode = builtinGtTime		




	gteNumMethod, err := RT.CreateMethod("",nil,"gte", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gteNumMethod.PrimitiveCode = builtinGteNum

	gteTimeMethod, err := RT.CreateMethod("",nil,"gte", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gteTimeMethod.PrimitiveCode = builtinGteTime	

	gteStrMethod, err := RT.CreateMethod("",nil,"gte", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gteStrMethod.PrimitiveCode = builtinGteStr	
	
	
	
	
	
	greaterEqNumMethod, err := RT.CreateMethod("",nil,"greaterEq", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterEqNumMethod.PrimitiveCode = builtinGteNum

	greaterEqTimeMethod, err := RT.CreateMethod("",nil,"greaterEq", []string{"p1", "p2"}, []string{"Time", "Time"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterEqTimeMethod.PrimitiveCode = builtinGteTime	

	greaterEqStrMethod, err := RT.CreateMethod("",nil,"greaterEq", []string{"p1", "p2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	greaterEqStrMethod.PrimitiveCode = builtinGteStr
		












    /////////////////////////////////////////////////////////
    // Numeric arithmetic functions
    //
    // Note: We have an issue here about return type
    //
    // Is the language going to allow co-variant return types in different methods of same multimethod?
    //
    // What does that imply for static type checking of return values?
    //

	timesMethod, err := RT.CreateMethod("",nil,"times", []string{"p1", "p2"}, []string{"Numeric", "Numeric"}, []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timesMethod.PrimitiveCode = builtinTimes


	plusMethod, err := RT.CreateMethod("",nil,"plus", []string{"p1", "p2"}, []string{"Numeric", "Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	plusMethod.PrimitiveCode = builtinPlus	

	sumNumericMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"List_of_Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumNumericMethod.PrimitiveCode = builtinSum	

	sumFloatMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"List_of_Float"},  []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumFloatMethod.PrimitiveCode = builtinSum

	sumIntMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"List_of_Int"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumIntMethod.PrimitiveCode = builtinSum

	sumUintMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"List_of_Uint"},  []string{"Uint"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumUintMethod.PrimitiveCode = builtinSum	

	sumNumericSetMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"Set_of_Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumNumericSetMethod.PrimitiveCode = builtinSum	

	sumFloatSetMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"Set_of_Float"},  []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumFloatSetMethod.PrimitiveCode = builtinSum

	sumIntSetMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"Set_of_Int"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumIntSetMethod.PrimitiveCode = builtinSum

	sumUintSetMethod, err := RT.CreateMethod("",nil,"sum", []string{"nums"}, []string{"Set_of_Uint"},  []string{"Uint"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	sumUintSetMethod.PrimitiveCode = builtinSum	



	minusMethod, err := RT.CreateMethod("",nil,"minus", []string{"p1", "p2"}, []string{"Numeric", "Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	minusMethod.PrimitiveCode = builtinMinus	

	divMethod, err := RT.CreateMethod("",nil,"div", []string{"p1", "p2"}, []string{"Numeric", "Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	divMethod.PrimitiveCode = builtinDiv	

	modMethod, err := RT.CreateMethod("",nil,"mod", []string{"p1", "p2"}, []string{"Integer", "Integer"},  []string{"Integer"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	modMethod.PrimitiveCode = builtinMod	

	negMethod, err := RT.CreateMethod("",nil,"neg", []string{"p1"}, []string{"Numeric"},  []string{"Numeric"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	negMethod.PrimitiveCode = builtinNeg		


    ////////////////////////////////////////////////////////////
    // Time and Duration functions

    // now location String > Time
    //
    // t = now "America/Los_Angeles"       t = now "Local"      t = now "UTC"
    //
	timeNowMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"now", []string{"loc"}, []string{"String"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeNowMethod.PrimitiveCode = builtinTimeNow	

    // sleep durationNs Int
    //
	sleepMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"sleep", []string{"ns"}, []string{"Int"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	sleepMethod.PrimitiveCode = builtinSleep

    // tick durationNs Int > InChannel of Time
    //
	tickMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"tick", []string{"ns"}, []string{"Int"},  []string{"InChannel"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	tickMethod.PrimitiveCode = builtinTick
	
    // gc
    //
	gcMethod, err := RT.CreateMethod("",nil,"gc", []string{}, []string{},  []string{}, false, 0, false)
	if err != nil {
		panic(err)
	}
	gcMethod.PrimitiveCode = builtinGC


    // plus t Time durationNs Int > Time
    //
	timePlusMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"plus", []string{"t", "ns"}, []string{"Time", "Int"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timePlusMethod.PrimitiveCode = builtinTimePlus	

    // addDate t Time years Int months Int days Int > Time
    //
	timeAddDateMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"addDate", []string{"t", "years","months","days"}, []string{"Time", "Int", "Int", "Int"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeAddDateMethod.PrimitiveCode = builtinTimeAddDate

    // minus t Time durationNs Int > Time
    //
	timeMinusDurationMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minus", []string{"t", "ns"}, []string{"Time", "Int"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeMinusDurationMethod.PrimitiveCode = builtinTimeMinusDuration	

    // minus t2 Time t2 Time > durationNs Int
    //
	timeMinusTimeMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minus", []string{"t", "t"}, []string{"Time", "Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeMinusTimeMethod.PrimitiveCode = builtinTimeMinusTime	

    // since t Time > durationNs Int
    //
	timeSinceMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"since", []string{"t"}, []string{"Time"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeSinceMethod.PrimitiveCode = builtinTimeSince	

    // timeIn t Time location String > Time
    //
	timeInMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"timeIn", []string{"t","location"}, []string{"Time","String"}, []string{"Time"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInMethod.PrimitiveCode = builtinTimeIn

    // location t Time > locationName String 
    //
    // Returns IANA time location name (long-form timezone name) for the time.
    //
	timeLocationMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"location", []string{"t"}, []string{"Time"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeLocationMethod.PrimitiveCode = builtinTimeLocation


    // hours n Int > durationNs Int
    // 
	hoursMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"hours", []string{"n"}, []string{"Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	hoursMethod.PrimitiveCode = builtinHours

    // minutes n Int > durationNs Int
    // 
	minutesMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minutes", []string{"n"}, []string{"Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	minutesMethod.PrimitiveCode = builtinMinutes	

    // seconds n Int > durationNs Int
    // 
	secondsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"seconds", []string{"n"}, []string{"Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	secondsMethod.PrimitiveCode = builtinSeconds

    // milliseconds n Int > durationNs Int
    // 
	millisecondsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"milliseconds", []string{"n"}, []string{"Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	millisecondsMethod.PrimitiveCode = builtinMilliseconds	






// date t Time > year Int month Int day Int
//
	timeDateMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"date", []string{"t"}, []string{"Time"}, []string{"Int","Int","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeDateMethod.PrimitiveCode = builtinTimeDate

// clock t Time > hour Int min Int sec Int
//
	timeClockMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"clock", []string{"t"}, []string{"Time"}, []string{"Int","Int","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeClockMethod.PrimitiveCode = builtinTimeClock

// day t Time > dayOfMonth Int  
// """
//  0..31
// """
//
	timeDayMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"day", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeDayMethod.PrimitiveCode = builtinTimeDay


// hour t Time >  Int  
// """
//  0..23
// """
//
	timeHourMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"hour", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeHourMethod.PrimitiveCode = builtinTimeHour


// minute t Time >  Int  
// """
//  0..59
// """
//
	timeMinuteMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minute", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeMinuteMethod.PrimitiveCode = builtinTimeMinute


// second t Time >  Int  
// """
//  0..59
// """
//
	timeSecondMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"second", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeSecondMethod.PrimitiveCode = builtinTimeSecond

// nanosecond t Time >  Int  
// """
//  0..999999999
// """
//
	timeNanosecondMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"nanosecond", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeNanosecondMethod.PrimitiveCode = builtinTimeNanosecond


// weekday t Time >  Int  
// """
//  0..6  Sunday = 0 
// """
//
	timeWeekdayMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"weekday", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeWeekdayMethod.PrimitiveCode = builtinTimeWeekday


// year t Time >  Int  
//
	timeYearMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"year", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeYearMethod.PrimitiveCode = builtinTimeYear


// month t Time >  Int  
// """
//  1..12  January = 1
// """
//
	timeMonthMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"month", []string{"t"}, []string{"Time"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeMonthMethod.PrimitiveCode = builtinTimeMonth
	

// zone t Time > name String offset Int
// """
//  EST secondsEastOfUTC
// """  
//
	timeZoneMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"zone", []string{"t"}, []string{"Time"}, []string{"String","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeZoneMethod.PrimitiveCode = builtinTimeZone


// format t Time layout String > String
// """
//  Like Go time layouts (http://golang.org/pkg/time/)
// """  
//
	timeFormatMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"format", []string{"t","layout"}, []string{"Time","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeFormatMethod.PrimitiveCode = builtinTimeFormat








    // duration hours Int minutes Int > durationNs Int
    // 
	durationHoursMinutesMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"duration", []string{"h","m"}, []string{"Int","Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	durationHoursMinutesMethod.PrimitiveCode = builtinDuration

    // duration hours Int minutes Int seconds Int > durationNs Int
    // 
	durationHoursMinutesSecondsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"duration", []string{"h","m","s"}, []string{"Int","Int","Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	durationHoursMinutesSecondsMethod.PrimitiveCode = builtinDuration	

    // duration hours Int minutes Int seconds Int nanoseconds Int > durationNs Int
    // 
	durationHoursMinutesSecondsNsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"duration", []string{"h","m","s","ns"}, []string{"Int","Int","Int","Int"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	durationHoursMinutesSecondsNsMethod.PrimitiveCode = builtinDuration		


    // hoursEquivalentOf durationNs Int > Float
    //
    // Returns the duration as a floating point number of hours.
    // 
	hoursEquivalentOfMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"hoursEquivalentOf", []string{"ns"}, []string{"Int"}, []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	hoursEquivalentOfMethod.PrimitiveCode = builtinHoursEquivalentOf

    // minutesEquivalentOf durationNs Int > Float
    //
    // Returns the duration as a floating point number of minutes.
    // 
	minutesEquivalentOfMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"minutesEquivalentOf", []string{"ns"}, []string{"Int"}, []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	minutesEquivalentOfMethod.PrimitiveCode = builtinMinutesEquivalentOf

    // secondsEquivalentOf durationNs Int > Float
    //
    // Returns the duration as a floating point number of seconds.
    // 
	secondsEquivalentOfMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"secondsEquivalentOf", []string{"ns"}, []string{"Int"}, []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	secondsEquivalentOfMethod.PrimitiveCode = builtinSecondsEquivalentOf	

    // timeParts durationNs Int > h Int m Int s Int ns Int
    //
    // Returns the duration as the number of hours, remaining minutes, remaining seconds, 
    // and remaining nanoseconds
    // 
	timePartsMethod, err := RT.CreateMethod("relish/pkg/datetime",nil,"timeParts", []string{"ns"}, []string{"Int"}, []string{"Int","Int","Int","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timePartsMethod.PrimitiveCode = builtinTimeParts	




    ///////////////////////////////////////////////////////////
    // Persistence functions

	dubMethod, err := RT.CreateMethod("",nil,"dub", []string{"obj", "name"}, []string{"NonPrimitive", "String"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	dubMethod.PrimitiveCode = builtinDub

	summonMethod, err := RT.CreateMethod("",nil,"summon", []string{"name"}, []string{"String"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	summonMethod.PrimitiveCode = builtinSummon	

	summonByIdMethod, err := RT.CreateMethod("",nil,"summon", []string{"dbid"}, []string{"Int"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	summonByIdMethod.PrimitiveCode = builtinSummonById
	
	existsMethod, err := RT.CreateMethod("",nil,"exists", []string{"name"}, []string{"String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	existsMethod.PrimitiveCode = builtinExists

    // delete object from the local database and the in-memory cache. Mark object as not stored locally.
    // NOTE!! Does not clean up attribute/relation/collection association tables in the database.
    // So disassociate associations first, then delete!
    //
	deleteMethod, err := RT.CreateMethod("",nil,"delete", []string{"obj"}, []string{"NonPrimitive"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	deleteMethod.PrimitiveCode = builtinDelete	
	
	
	renameObjectMethod, err := RT.CreateMethod("",nil,"rename", []string{"oldName","newName"}, []string{"String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	renameObjectMethod.PrimitiveCode = builtinRenameObject	

    // err = begin  // Begins a db transaction. On success returns an empty string.
    //
	beginMethod, err := RT.CreateMethod("",nil,"begin", []string{}, []string{}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	beginMethod.PrimitiveCode = builtinBeginTransaction

    // err = begin "READ" // Begins a db transaction usable only for reading. On success returns an empty string.
    //
	begin2Method, err := RT.CreateMethod("",nil,"begin", []string{"transactionType"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	begin2Method.PrimitiveCode = builtinBeginTransaction


    // err = local  // Begins a db transaction which TODO: never contributes new or fetched-from-db objects to the global
    //              // in-memory object cache. On success returns an empty string.
    //
	localMethod, err := RT.CreateMethod("",nil,"local", []string{}, []string{}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	localMethod.PrimitiveCode = builtinBeginLocalTransaction	

    // err = commit  // Commits the in-progress db transaction. On success returns an empty string.
    //               // TODO: If succeeds and is not a local transaction, adds new and fetched-from-db objects to the
    //               // global in-memory object cache if they were not already there.
    //
	commitMethod, err := RT.CreateMethod("",nil,"commit", []string{}, []string{}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	commitMethod.PrimitiveCode = builtinCommitTransaction	

    // err = rollback  // Rolls back the in-progress db transaction. On success returns an empty string.
    //                 // TODO - NEED TO DETERMINE POLICY FOR REVERTING CHANGES MADE TO IN MEMORY OBJECTS IF
    //                 // THIS WAS A NON-LOCAL TRANSACTION.
    //                 // PROBABLY WHAT WE HAVE TO DO IS DECLARE ALL OBJECTS THAT WERE MODIFIED/RELATED DURING
    //                 // THE TRANSACTION (IE DIRTY OBJECTS) AS BEING INVALID-IN-MEMORY I.E. THEY NEED TO BE
    //                 // RESTORED FROM THE DB BEFORE THEIR STATE IS ACCESSED AGAIN. IS THIS RESTORATION DONE
    //                 // RIGHT NOW UPON ROLLBACK, OR ON DEMAND IN GET-ATTR-VAL OPERATION?
    //
	rollbackMethod, err := RT.CreateMethod("",nil,"rollback", []string{}, []string{}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	rollbackMethod.PrimitiveCode = builtinRollbackTransaction	


    ///////////////////////////////////////////////////////////
    // Context map functions. The context is a non-persistent map from String name to object, which
    // serves as a substitute for global variables.
    // Requesting, setting and removing the context objects is thread-safe.
    // Being in the context map is one way to keep a non-persistent object around (un-garbage-collected) in 
    // between method executions.


	contextMethod, err := RT.CreateMethod("",nil,"context", []string{"name"}, []string{"String"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	contextMethod.PrimitiveCode = builtinContext	

	contextPutMethod, err := RT.CreateMethod("",nil,"contextPut", []string{"obj", "name"}, []string{"Any", "String"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	contextPutMethod.PrimitiveCode = builtinContextPut

	contextExistsMethod, err := RT.CreateMethod("",nil,"contextExists", []string{"name"}, []string{"String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	contextExistsMethod.PrimitiveCode = builtinContextExists	

	contextRemoveMethod, err := RT.CreateMethod("",nil,"contextRemove", []string{"name"}, []string{"String"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	contextRemoveMethod.PrimitiveCode = builtinContextRemove	


    ///////////////////////////////////////////////////////////////////
    // Boolean Logical functions - not and or

	notMethod, err := RT.CreateMethod("",nil,"not", []string{"p"}, []string{"Any"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	notMethod.PrimitiveCode = builtinNot    



/*
   TODO Replace these with proper variadic function creations.
   Also, and and or should not evaluate their arguments right away, but should only evaluate each one conditionally
   based on the truth of the one before, as in Lisp.
*/
	and2Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2"}, []string{"Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and2Method.PrimitiveCode = builtinAnd

	and3Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3"}, []string{"Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and3Method.PrimitiveCode = builtinAnd

	and4Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3", "p4"}, []string{"Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and4Method.PrimitiveCode = builtinAnd

	and5Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and5Method.PrimitiveCode = builtinAnd

	and6Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and6Method.PrimitiveCode = builtinAnd

	and7Method, err := RT.CreateMethod("",nil,"and", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	and7Method.PrimitiveCode = builtinAnd

	

	or2Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2"}, []string{"Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or2Method.PrimitiveCode = builtinOr

	or3Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3"}, []string{"Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or3Method.PrimitiveCode = builtinOr

	or4Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3", "p4"}, []string{"Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or4Method.PrimitiveCode = builtinOr

	or5Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3", "p4", "p5"}, []string{"Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or5Method.PrimitiveCode = builtinOr

	or6Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3", "p4", "p5", "p6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or6Method.PrimitiveCode = builtinOr

	or7Method, err := RT.CreateMethod("",nil,"or", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	or7Method.PrimitiveCode = builtinOr

    









	//////////////////////////////////////////////////////////////////////////////////////
	// String methods - TODO Put them in their own package
	
	// contains s String substr String > Bool	
	//
	stringContainsMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"contains", []string{"s","substr"}, []string{"String","String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringContainsMethod.PrimitiveCode = builtinStringContains	




/*
replace s String old String new String > String
"""
  Returns a copy of the string s with all non-overlapping instances of old replaced by new. 
"""

replace s String old String new String n Int > String
"""
  Returns a copy of the string s with the first n non-overlapping instances of old replaced by new. 
  If n < 0, there is no limit on the number of replacements.
"""
*/


    stringReplace3Method, err := RT.CreateMethod("relish/pkg/strings",nil,"replace", []string{"s","old","new"}, []string{"String","String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringReplace3Method.PrimitiveCode = builtinStringReplace

    stringReplace4Method, err := RT.CreateMethod("relish/pkg/strings",nil,"replace", []string{"s","old","new","n"}, []string{"String","String","String","Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringReplace4Method.PrimitiveCode = builtinStringReplace




/*
urlPathPartEncode s > String
"""
  Returns a copy of the string that has been encoded to be a legal part of a URL path.
"""
*/
    urlPathPartEncodeMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"urlPathPartEncode", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	urlPathPartEncodeMethod.PrimitiveCode = builtinUrlPathPartEncode

    urlPathPartEncodeMethod2, err := RT.CreateMethod("relish/pkg/strings",nil,"urlPathPartEncode", []string{"s","encodeDash"}, []string{"String","Bool"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	urlPathPartEncodeMethod2.PrimitiveCode = builtinUrlPathPartEncode




/*
urlPathPartDecode s > String
"""
  Returns a copy of the string which has had url path part encodings decoded. 
"""
*/
    urlPathPartDecodeMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"urlPathPartDecode", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	urlPathPartDecodeMethod.PrimitiveCode = builtinUrlPathPartDecode

    urlPathPartDecodeMethod2, err := RT.CreateMethod("relish/pkg/strings",nil,"urlPathPartDecode", []string{"s","decodeDash"}, []string{"String","Bool"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	urlPathPartDecodeMethod2.PrimitiveCode = builtinUrlPathPartDecode














// join a []String
// """
//  Concatenates the elements of a to create a single string. T
// """
//
// join a []String sep String > String
//
// Should be 
// join a []Any sep String > String
// but we don't have covariant collection type compatibility in input args yet. should have, for immutable input args	


    stringJoin1Method, err := RT.CreateMethod("relish/pkg/strings",nil,"join", []string{"a"}, []string{"List_of_String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringJoin1Method.PrimitiveCode = builtinStringJoin



    stringJoin2Method, err := RT.CreateMethod("relish/pkg/strings",nil,"join", []string{"a","sep"}, []string{"List_of_String","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringJoin2Method.PrimitiveCode = builtinStringJoin


    stringJoin2AnyMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"join", []string{"a","sep"}, []string{"List_of_Any","String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringJoin2AnyMethod.PrimitiveCode = builtinStringJoin


    stringSplitMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"split", []string{"a","sep"}, []string{"String","String"}, []string{"List_of_String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringSplitMethod.PrimitiveCode = builtinStringSplit


	
	/*

	containsAny s String chars String > Bool

	containsRune s String r Rune > bool

	count s String sep String > Int  // or Int32 or UInt or UInt32 ? Which? Why?

	equalFold s String t String > Bool

	fields s String > []String

	// func FieldsFunc(s string, f func(rune) bool) []string

	hasPrefix s String prefix String > Bool

	hasSuffix s String suffix String > Bool

	index s String t String > Int
	"""
	 index of first occurrence of t in s or -1 if t not found in s 
	"""

	indexAny s String chars String > Int

	// func IndexFunc(s string, f func(rune) bool) int

	indexRune s string r Rune > Int

	join a []string sep String > String	
	
	*/
	
    // substituting values for %s 
    //
    stringFill2Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2"}, []string{"String", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill2Method.PrimitiveCode = builtinStringFill

	stringFill3Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3"}, []string{"String", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill3Method.PrimitiveCode = builtinStringFill

	stringFill4Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4"}, []string{"String", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill4Method.PrimitiveCode = builtinStringFill

	stringFill5Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5"}, []string{"String", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill5Method.PrimitiveCode = builtinStringFill

	stringFill6Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5", "s6"}, []string{"String", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill6Method.PrimitiveCode = builtinStringFill

	stringFill7Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill7Method.PrimitiveCode = builtinStringFill	

	stringFill8Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill8Method.PrimitiveCode = builtinStringFill	

		stringFill9Method, err := RT.CreateMethod("relish/pkg/strings",nil,"fill", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9"}, []string{"String", "Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFill9Method.PrimitiveCode = builtinStringFill		


    // length of a string
    //
    // in bytes
	stringLenMethod, err := RT.CreateMethod("",nil,"len", []string{"c"}, []string{"String"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringLenMethod.PrimitiveCode = builtinStringLen
	
    // in characters (codepoints)
    //
	stringNumCodePointsMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"numCodePoints", []string{"c"}, []string{"String"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringNumCodePointsMethod.PrimitiveCode = builtinStringNumCodePoints
	
	// concatenation of strings
	
    // s = cat s1 String s2 String s3 String s4 String s5 String s6 String s7 String s8 String s9 String > String  
	
    stringCat2Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2"}, []string{"Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat2Method.PrimitiveCode = builtinStringCat

	stringCat3Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3"}, []string{"Any", "Any", "Any"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat3Method.PrimitiveCode = builtinStringCat

	stringCat4Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4"}, []string{"Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat4Method.PrimitiveCode = builtinStringCat

	stringCat5Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5"}, []string{"Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat5Method.PrimitiveCode = builtinStringCat

	stringCat6Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5", "s6"}, []string{"Any", "Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat6Method.PrimitiveCode = builtinStringCat

	stringCat7Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat7Method.PrimitiveCode = builtinStringCat	

	stringCat8Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat8Method.PrimitiveCode = builtinStringCat	

		stringCat9Method, err := RT.CreateMethod("relish/pkg/strings",nil,"cat", []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9"}, []string{"Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any", "Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringCat9Method.PrimitiveCode = builtinStringCat	
	
	
    stringHasPrefixMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"hasPrefix", []string{"s1", "s2"}, []string{"String", "String"},  []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringHasPrefixMethod.PrimitiveCode = builtinStringHasPrefix	
	
    stringHasSuffixMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"hasSuffix", []string{"s1", "s2"}, []string{"String", "String"}, []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringHasSuffixMethod.PrimitiveCode = builtinStringHasSuffix	
	
	// index of substring in string
	
/*	
	index s String t String > Int
	"""
	 index of first occurrence of t in s or -1 if t not found in s 
	"""
*/
    stringIndexMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"index", []string{"s1", "s2"}, []string{"String", "String"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringIndexMethod.PrimitiveCode = builtinStringIndex
	
    stringLastIndexMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"lastIndex", []string{"s1", "s2"}, []string{"String", "String"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringLastIndexMethod.PrimitiveCode = builtinStringLastIndex	


    // substring or slice
/*
	slice s String start Int end Int > String
*/
    stringSlice2Method, err := RT.CreateMethod("relish/pkg/strings",nil,"slice", []string{"s", "start"}, []string{"String", "Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringSlice2Method.PrimitiveCode = builtinStringSlice	
	
    stringSlice3Method, err := RT.CreateMethod("relish/pkg/strings",nil,"slice", []string{"s", "start", "end"}, []string{"String", "Int", "Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringSlice3Method.PrimitiveCode = builtinStringSlice	
	
    stringFirstMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"first", []string{"s", "n"}, []string{"String", "Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringFirstMethod.PrimitiveCode = builtinStringFirst	
	
    stringLastMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"last", []string{"s", "n"}, []string{"String", "Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringLastMethod.PrimitiveCode = builtinStringLast	

    stringLowerMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"lower", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringLowerMethod.PrimitiveCode = builtinStringLower	

    stringUpperMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"upper", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringUpperMethod.PrimitiveCode = builtinStringUpper	

    stringTitleMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"title", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringTitleMethod.PrimitiveCode = builtinStringTitle	

    stringTrimSpaceMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"trimSpace", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringTrimSpaceMethod.PrimitiveCode = builtinStringTrimSpace			

	



	/*
	    returns a String - the base64-encoded version of the input argument Bytes
	*/
	stringBase64EncodeMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"base64Encode", []string{"b","urlSafe"}, []string{"Bytes","Bool"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringBase64EncodeMethod.PrimitiveCode = builtinStringBase64Encode

	/*
	    returns a String - the base64-encoded version of the input argument Bytes
	*/
	stringBase64EncodeMethod2, err := RT.CreateMethod("relish/pkg/strings",nil,"base64Encode", []string{"b","urlSafe","lineLen"}, []string{"Bytes","Bool","Int"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringBase64EncodeMethod2.PrimitiveCode = builtinStringBase64Encode


	/*
	    returns a Bytes - the decoding of the input argument base-64 String
	*/
	stringBase64DecodeMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"base64Decode", []string{"s","urlSafe"}, []string{"String","Bool"}, []string{"Bytes","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringBase64DecodeMethod.PrimitiveCode = builtinStringBase64Decode

	
	
	/*
	    returns a String - the base64-encoded sha256 hash of the input argument String.
	*/
	stringBase64HashMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"base64Hash", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringBase64HashMethod.PrimitiveCode = builtinStringHashSha256Base64		

	/*
	    returns a String - the hexadecimal-encoded sha256 hash of the input argument String.
	*/
	stringHexHashMethod, err := RT.CreateMethod("relish/pkg/strings",nil,"hexHash", []string{"s"}, []string{"String"}, []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringHexHashMethod.PrimitiveCode = builtinStringHashSha256Hex			
	
	
	
	
	
	
	
	
	
	// at s String i Int > Byte
	//
	stringAtMethod, err := RT.CreateMethod("",nil,"at", []string{"s","i"}, []string{"String","Int"}, []string{"Byte"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringAtMethod.PrimitiveCode = builtinStringAt		
	
	
	// runeAt s String i Int > codePoint Int width Int
	//
	stringRuneAtMethod, err := RT.CreateMethod("",nil,"runeAt", []string{"s","i"}, []string{"String","Int"}, []string{"Int","Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringRuneAtMethod.PrimitiveCode = builtinStringRuneAt	

	
    ///////////////////////////////////////////////////////////////////
    // Bytes functions
	
    // length of a byte-slice
    //
    // in bytes
	bytesLenMethod, err := RT.CreateMethod("",nil,"len", []string{"b"}, []string{"Bytes"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	bytesLenMethod.PrimitiveCode = builtinBytesLen	
	
	// at bytes Bytes i Int > Byte
	//
	bytesAtMethod, err := RT.CreateMethod("",nil,"at", []string{"bytes","i"}, []string{"Bytes","Int"}, []string{"Byte"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	bytesAtMethod.PrimitiveCode = builtinBytesAt		
	

	// set bytes Bytes i Int b Byte
	// set bytes Bytes i Int b Int	
	// set bytes Bytes i Int b Int32
	// set bytes Bytes i Int b Uint
	// set bytes Bytes i Int b Uint32
	//	
	bytesSetMethod, err := RT.CreateMethod("",nil,"set", []string{"bytes","i","b"}, []string{"Bytes","Int","Integer"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	bytesSetMethod.PrimitiveCode = builtinBytesSet


    // sub-array or slice
/*
	slice b Bytes start Int > Bytes
*/
    bytesSlice2Method, err := RT.CreateMethod("",nil,"slice", []string{"s", "start"}, []string{"Bytes", "Int"}, []string{"Bytes"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	bytesSlice2Method.PrimitiveCode = builtinBytesSlice	
	
/*
	slice b Bytes start Int end Int > String
*/	
    bytesSlice3Method, err := RT.CreateMethod("",nil,"slice", []string{"s", "start", "end"}, []string{"Bytes", "Int", "Int"}, []string{"Bytes"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	bytesSlice3Method.PrimitiveCode = builtinBytesSlice	

		
	
    ///////////////////////////////////////////////////////////////////
    // Collection functions	
	
	// len coll Collection > Int	
	//
	lenMethod, err := RT.CreateMethod("",nil,"len", []string{"c"}, []string{"Collection"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	lenMethod.PrimitiveCode = builtinLen


	// cap coll Collection > Int	
	//
	capMethod, err := RT.CreateMethod("",nil,"cap", []string{"c"}, []string{"Collection"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	capMethod.PrimitiveCode = builtinCap	


	// contains coll Collection val Any > Bool	
	//
	containsMethod, err := RT.CreateMethod("",nil,"contains", []string{"c","v"}, []string{"Collection","Any"},  []string{"Bool"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	containsMethod.PrimitiveCode = builtinContains


	// clear coll Collection 
	//
	clearMethod, err := RT.CreateMethod("",nil,"clear", []string{"c"}, []string{"Collection"},  nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	clearMethod.PrimitiveCode = builtinClear

	/*
	Defers (disables) auto-sorting on add element.
	*/
	deferSortingMethod, err := RT.CreateMethod("",nil,"deferSorting", []string{"c"}, []string{"Collection"},  nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	deferSortingMethod.PrimitiveCode = builtinDeferSorting

	/*
	Resumes auto-sorting. Does a sort.
	*/
	resumeSortingMethod, err := RT.CreateMethod("",nil,"resumeSorting", []string{"c"}, []string{"Collection"},  nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	resumeSortingMethod.PrimitiveCode = builtinResumeSorting	


	// asList coll Collection of T > List of T	
	//
	asListMethod, err := RT.CreateMethod("",nil,"asList", []string{"c"}, []string{"Collection"},  []string{"List"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	asListMethod.PrimitiveCode = builtinAsList


	// asList coll Collection of T oqlSelectClause String > List of T	
	//
	// Assumes that the collection is stored locally.
	// Returns a list formed by executing a select query including the conditions in the select clause
	// joined with the condition of membership in the collection. 
	//
	asList2Method, err := RT.CreateMethod("",nil,"asList", []string{"c","selectConditions"}, []string{"Collection","String"},  []string{"List"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	asList2Method.PrimitiveCode = builtinAsList2	


	// asList 
	//    coll Collection of T 
	//    selectClauseWithArgs List
	// > 
	//    List of T	
	//
	// Assumes that the collection is stored locally.
	// Returns a list formed by executing a select query including the conditions in the select clause
	// joined with the condition of membership in the collection. 
	//
	asList3Method, err := RT.CreateMethod("",nil,"asList", []string{"c","selectConditionsWithArgs"}, []string{"Collection","List"},  []string{"List"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	asList3Method.PrimitiveCode = builtinAsList2	


/*
	slice s List of T start Int end Int > List of T

	One idea: Can we have shared-array slice operators AND copy slice operators?

	Cannot proceed with these things overall til sort our parameterized types.

    listSlice2Method, err := RT.CreateMethod("",nil,"slice", []string{"s", "start"}, []string{"List_of_Any", "Int"}, []string{"List_of_Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	listSlice2Method.PrimitiveCode = builtinListSlice	
	
    listSlice3Method, err := RT.CreateMethod("",nil,"slice", []string{"s", "start", "end"}, []string{"List_of_Any", "Int", "Int"}, []string{"List_of_Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	listSlice3Method.PrimitiveCode = builtinListSlice	
	
    listFirstMethod, err := RT.CreateMethod("",nil,"first", []string{"s", "n"}, []string{"List_of_Any", "Int"}, []string{"List_of_Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	listFirstMethod.PrimitiveCode = builtinListFirst	
	
    listLastMethod, err := RT.CreateMethod("",nil,"last", []string{"s", "n"}, []string{"List_of_Any", "Int"}, []string{"List_of_Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	listLastMethod.PrimitiveCode = builtinListLast		

*/

    sortMethod, err := RT.CreateMethod("",nil,"sort", []string{"c"}, []string{"Collection"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	sortMethod.PrimitiveCode = builtinSort
	
    swapMethod, err := RT.CreateMethod("",nil,"swap", []string{"c","i","j"}, []string{"Collection","Int","Int"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	swapMethod.PrimitiveCode = builtinSwap
	
	/*
	   Returns a new set whose element type is the same as c1's.
	   c2's element type must be same as or a subtype of c1's.
	   The returned set contains the (unordered) set-union of the elements of c1 and c2.
	*/
	unionMethod, err := RT.CreateMethod("",nil,"union", []string{"c1","c2"}, []string{"Collection", "Collection"}, []string{"Set"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	unionMethod.PrimitiveCode = builtinUnion

	/*
	   Returns a new set whose element type is the same as c1's.
	   c2's element type can be anything.
	   The returned set contains the (unordered) set-intersection of the elements of c1 and c2.
	*/
	intersectionMethod, err := RT.CreateMethod("",nil,"intersection", []string{"c1","c2"}, []string{"Collection", "Collection"}, []string{"Set"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intersectionMethod.PrimitiveCode = builtinIntersection	

	/*
	   Returns a new set whose element type is the same as c1's.
	   c2's element type can be anything.
	   The returned set contains the an unordered set of c1's elements, except any which also occur in c2.
	*/
	setMinusMethod, err := RT.CreateMethod("",nil,"minus", []string{"c1","c2"}, []string{"Collection", "Collection"}, []string{"Set"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	setMinusMethod.PrimitiveCode = builtinSetMinus		

	/*
	   Adds the elements from c2 to collection c1, which must be a mutable collection.
	   c2's element type must be same as or a subtype of c1's.
	*/
	extendMethod, err := RT.CreateMethod("",nil,"extend", []string{"c1","c2"}, []string{"Collection", "Collection"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	extendMethod.PrimitiveCode = builtinExtend

	
    ///////////////////////////////////////////////////////////////////
    // Concurrency functions	

	// len c Channel > Int	
	//
	channelLenMethod, err := RT.CreateMethod("",nil,"len", []string{"c"}, []string{"Channel"}, []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelLenMethod.PrimitiveCode = builtinChannelLen    

	// cap c Channel > Int	
	//
	channelCapMethod, err := RT.CreateMethod("",nil,"cap", []string{"c"}, []string{"Channel"},  []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelCapMethod.PrimitiveCode = builtinChannelCap    	


	// from c InChannel of T > T	
	//
	channelFromMethod, err := RT.CreateMethod("",nil,"<-", []string{"c"}, []string{"InChannel"},  []string{"Any"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelFromMethod.PrimitiveCode = builtinFrom   	

/* Deprecated in favour of ch <- operator

    // to c OutChannel of T obj T 	
	//
	channelToMethod, err := RT.CreateMethod("",nil,"to", []string{"c","v"}, []string{"OutChannel","Any"}, 1, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelToMethod.PrimitiveCode = builtinTo 
*/


	mutexLockMethod, err := RT.CreateMethod("",nil,"lock", []string{"m"}, []string{"Mutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	mutexLockMethod.PrimitiveCode = builtinMutexLock

	mutexUnlockMethod, err := RT.CreateMethod("",nil,"unlock", []string{"m"}, []string{"Mutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	mutexUnlockMethod.PrimitiveCode = builtinMutexUnlock
	
	
	
	ownedMutexLockMethod, err := RT.CreateMethod("",nil,"lock", []string{"m"}, []string{"OwnedMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	ownedMutexLockMethod.PrimitiveCode = builtinOwnedMutexLock

	ownedMutexUnlockMethod, err := RT.CreateMethod("",nil,"unlock", []string{"m"}, []string{"OwnedMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	ownedMutexUnlockMethod.PrimitiveCode = builtinOwnedMutexUnlock	
	
	

	rwmutexLockMethod, err := RT.CreateMethod("",nil,"lock", []string{"m"}, []string{"RwMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexLockMethod.PrimitiveCode = builtinRWMutexLock

	rwmutexUnlockMethod, err := RT.CreateMethod("",nil,"unlock", []string{"m"}, []string{"RwMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexUnlockMethod.PrimitiveCode = builtinRWMutexUnlock

	rwmutexRLockMethod, err := RT.CreateMethod("",nil,"rlock", []string{"m"}, []string{"RwMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexRLockMethod.PrimitiveCode = builtinRWMutexRLock

	rwmutexRUnlockMethod, err := RT.CreateMethod("",nil,"runlock", []string{"m"}, []string{"RwMutex"}, nil, false, 0, false)
	if err != nil {
		panic(err)
	}
	rwmutexRUnlockMethod.PrimitiveCode = builtinRWMutexRUnlock




    /////////////////////////////////////////////////////////////////////
    // Type init functions

	timeInit1Method, err := RT.CreateMethod("",nil,"initTime", []string{"t","timeString"}, []string{"Time","String"},  []string{"Time","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit1Method.PrimitiveCode = builtinInitTimeParse    
	
	timeInit2Method, err := RT.CreateMethod("",nil,"initTime", []string{"t","timeString","layout"}, []string{"Time","String","String"},  []string{"Time","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit2Method.PrimitiveCode = builtinInitTimeParse	

	timeInit3Method, err := RT.CreateMethod("",nil,"initTime", []string{"t","timeString","layout","location"}, []string{"Time","String","String","String"},  []string{"Time","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit3Method.PrimitiveCode = builtinInitTimeParse	


	timeInit4Method, err := RT.CreateMethod("",nil,"initTime", []string{"t","year","month","day","hour","min","sec","nsec","loc"}, []string{"Time","Int","Int","Int","Int","Int","Int","Int","String"},  []string{"Time","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	timeInit4Method.PrimitiveCode = builtinInitTimeDate	



	channelInit0Method, err := RT.CreateMethod("",nil,"initChannel", []string{"c"}, []string{"Channel"}, []string{"Channel"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelInit0Method.PrimitiveCode = builtinInitChannel 

	channelInit1Method, err := RT.CreateMethod("",nil,"initChannel", []string{"c","n"}, []string{"Channel","Int"}, []string{"Channel"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	channelInit1Method.PrimitiveCode = builtinInitChannel	

	bytesInitMethod, err := RT.CreateMethod("",nil,"initBytes", []string{"b","n"}, []string{"Bytes","Int"},  []string{"Bytes"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	bytesInitMethod.PrimitiveCode = builtinInitBytes 
	
	bytesInitMethod2, err := RT.CreateMethod("",nil,"initBytes", []string{"b","n","c"}, []string{"Bytes","Int","Int"},  []string{"Bytes"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	bytesInitMethod2.PrimitiveCode = builtinInitBytes	
	
	bytesFromStringInitMethod, err := RT.CreateMethod("",nil,"initBytes", []string{"b","s"}, []string{"Bytes","String"},  []string{"Bytes"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	bytesFromStringInitMethod.PrimitiveCode = builtinInitBytesFromString	


	stringInitMethod, err := RT.CreateMethod("",nil,"initString", []string{"s","o"}, []string{"String","Any"},  []string{"String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	stringInitMethod.PrimitiveCode = builtinInitString  
	
	

	intInitMethod, err := RT.CreateMethod("",nil,"initInt", []string{"i","s"}, []string{"Int","String"},  []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intInitMethod.PrimitiveCode = builtinInitInt	

	intInitMethod2, err := RT.CreateMethod("",nil,"initInt", []string{"i","f"}, []string{"Int","Float"},  []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intInitMethod2.PrimitiveCode = builtinInitIntFromFloat	

	intInitMethod3, err := RT.CreateMethod("",nil,"initInt", []string{"i","v"}, []string{"Int","Int32"},  []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intInitMethod3.PrimitiveCode = builtinInitIntFromInt32	

	intInitMethod4, err := RT.CreateMethod("",nil,"initInt", []string{"i","v"}, []string{"Int","Uint"},  []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intInitMethod4.PrimitiveCode = builtinInitIntFromUint	

	intInitMethod5, err := RT.CreateMethod("",nil,"initInt", []string{"i","v"}, []string{"Int","Uint32"},  []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intInitMethod5.PrimitiveCode = builtinInitIntFromUint32		

	intInitMethod6, err := RT.CreateMethod("",nil,"initInt", []string{"i","v"}, []string{"Int","Byte"},  []string{"Int","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intInitMethod6.PrimitiveCode = builtinInitIntFromByte
	
	
	uintInitMethod6, err := RT.CreateMethod("",nil,"initUint", []string{"i","v"}, []string{"Uint","Byte"},  []string{"Uint","String"}, false, 0, false)
	if err != nil {
		panic(err)  
	}
	uintInitMethod6.PrimitiveCode = builtinInitUintFromByte	
	
	uint32InitMethod6, err := RT.CreateMethod("",nil,"initUint32", []string{"i","v"}, []string{"Uint32","Byte"},  []string{"Uint32","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uint32InitMethod6.PrimitiveCode = builtinInitUint32FromByte	
	
	

	int32InitMethod, err := RT.CreateMethod("",nil,"initInt32", []string{"i","s"}, []string{"Int32","String"},  []string{"Int32","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	int32InitMethod.PrimitiveCode = builtinInitInt32

	int32InitMethod2, err := RT.CreateMethod("",nil,"initInt32", []string{"i","v"}, []string{"Int32","Int"},  []string{"Int32","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	int32InitMethod2.PrimitiveCode = builtinInitInt32FromInt


	uintInitMethod, err := RT.CreateMethod("",nil,"initUint", []string{"i","s"}, []string{"Uint","String"},  []string{"Uint","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uintInitMethod.PrimitiveCode = builtinInitUint		

	uintInitMethod2, err := RT.CreateMethod("",nil,"initUint", []string{"i","v"}, []string{"Uint","Int"},  []string{"Uint","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uintInitMethod2.PrimitiveCode = builtinInitUintFromInt


	uint32InitMethod, err := RT.CreateMethod("",nil,"initUint32", []string{"i","s"}, []string{"Uint32","String"},  []string{"Uint32","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uint32InitMethod.PrimitiveCode = builtinInitUint32

	uint32InitMethod2, err := RT.CreateMethod("",nil,"initUint32", []string{"i","v"}, []string{"Uint32","Int"},  []string{"Uint32","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uint32InitMethod2.PrimitiveCode = builtinInitUint32FromInt
	
	
	
	byteInitMethod, err := RT.CreateMethod("",nil,"initByte", []string{"i","v"}, []string{"Byte","Int"},  []string{"Byte","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	byteInitMethod.PrimitiveCode = builtinInitByteFromInt	

	byteInitMethod2, err := RT.CreateMethod("",nil,"initByte", []string{"i","v"}, []string{"Byte","Uint"},  []string{"Byte","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	byteInitMethod2.PrimitiveCode = builtinInitByteFromUint
	
	byteInitMethod3, err := RT.CreateMethod("",nil,"initByte", []string{"i","v"}, []string{"Byte","Uint32"},  []string{"Byte","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	byteInitMethod3.PrimitiveCode = builtinInitByteFromUint32
			

	
	floatInitMethod, err := RT.CreateMethod("",nil,"initFloat", []string{"f","s"}, []string{"Float","String"},  []string{"Float","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	floatInitMethod.PrimitiveCode = builtinInitFloat	


	floatInitMethod2, err := RT.CreateMethod("",nil,"initFloat", []string{"f","i"}, []string{"Float","Int"},  []string{"Float","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	floatInitMethod2.PrimitiveCode = builtinInitFloatFromInt	

	
	
	boolInitMethod, err := RT.CreateMethod("",nil,"initBool", []string{"b","s"}, []string{"Bool","String"},  []string{"Bool","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	boolInitMethod.PrimitiveCode = builtinInitBool
		
	boolInitMethod2, err := RT.CreateMethod("",nil,"initBool", []string{"b","i"}, []string{"Bool","Integer"},  []string{"Bool","String"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	boolInitMethod2.PrimitiveCode = builtinInitBoolFromInteger		
}


///////////////////////////////////////////////////////////////////////////////////////////
// I/O functions

func builtinPrint(th InterpreterThread, objects []RObject) []RObject {
	for i, obj := range objects {
		if i > 0 {
			fmt.Print(" ")
		}
		switch obj.(type) {
		case Int, Int32, Float, Bool, String:
			fmt.Print(obj.String())
		default:
			fmt.Print(obj.StringTh(th)) 
		}
	}
	fmt.Println()
	return nil
}

func builtinInput(th InterpreterThread, objects []RObject) []RObject {
   message := string(objects[0].(String))	
   fmt.Print(message)
   result,err := buf.ReadString('\n')
   if err != nil {
      result = err.Error()
   }
   n := len(result)
   if n > 1 && result[n-2] == '\r'  {   // Windows
   	  n -= 2
   } else {
   	  n--
   }
   result = result[:n]
   return []RObject{String(result)}
}

func builtinDbg(th InterpreterThread, objects []RObject) []RObject {
   fmt.Println(objects[0].Debug(th))
   return []RObject{}
}

func builtinDebug(th InterpreterThread, objects []RObject) []RObject {
   s := objects[0].Debug(th)
   return []RObject{String(s)}
}






//////////////////////////////////////////////////////////////////////////////////////////////
// Comparison functions

/*
Equality operator.
TODO !!!
Really have to study which kinds of equality we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinEq(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Bool:
		switch obj2.(type) {
		case Bool:
			val = Bool(obj1.(Bool) == obj2.(Bool))	
		default:
			val = Bool(false)			
		}			
	case Int:
		switch obj2.(type) {
		case Int:
			val = Bool(obj1.(Int) == obj2.(Int))
		case Int32:
			val = Bool(int64(obj1.(Int)) == int64(obj2.(Int32)))
		case Float:
			val = Bool(float64(obj2.(Float)) == float64(obj1.(Int)))			
		case String:
			val = Bool(string(obj2.(String)) == strconv.FormatInt(int64(obj1.(Int)), 10))
		case Uint:
			val = Bool(int64(obj1.(Int)) == int64(obj2.(Uint)))
		case Uint32:
			val = Bool(int64(obj1.(Int)) == int64(obj2.(Uint32)))			
		default:
			val = Bool(false)
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Bool(obj1.(Int32) == obj2.(Int32))
		case Int:
			val = Bool(int64(obj1.(Int32)) == int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj2.(Float)) == float64(obj1.(Int32)))			
		case String:
			val = Bool(string(obj2.(String)) == strconv.FormatInt(int64(obj1.(Int32)), 10))
		case Uint:
			val = Bool(int64(obj1.(Int32)) == int64(obj2.(Uint)))
		case Uint32:
			val = Bool(int64(obj1.(Int32)) == int64(obj2.(Uint32)))				
		default:
			val = Bool(false)
		}
	case Float:
		switch obj2.(type) {
		case Int32:
			val = Bool(float64(obj1.(Float)) == float64(obj2.(Int32)))		
		case Int:
			val = Bool(float64(obj1.(Float)) == float64(obj2.(Int)))	
		case Float:
			val = Bool(float64(obj2.(Float)) == float64(obj1.(Float)))			
		case String:
			val = Bool(string(obj2.(String)) == strconv.FormatFloat(float64(obj1.(Float)), 'G', -1, 64)) 
		case Uint32:
			val = Bool(float64(obj1.(Float)) == float64(obj2.(Uint32)))		
		case Uint:
			val = Bool(float64(obj1.(Float)) == float64(obj2.(Uint)))					
		default:
			val = Bool(false)	
		}	
	case String:
		switch obj2.(type) {
		case String:
			val = Bool(string(obj1.(String)) == string(obj2.(String)))
		case Int:
			val = Bool(string(obj1.(String)) == strconv.FormatInt(int64(obj2.(Int)), 10))
		case Int32:
			val = Bool(string(obj1.(String)) == strconv.FormatInt(int64(obj2.(Int32)), 10))
		case Float:
			val = Bool(string(obj1.(String)) == strconv.FormatFloat(float64(obj2.(Float)), 'G', -1, 64)) 		
		case Uint:
			val = Bool(string(obj1.(String)) == strconv.FormatUint(uint64(obj2.(Uint)), 10))
		case Uint32:
			val = Bool(string(obj1.(String)) == strconv.FormatUint(uint64(obj2.(Uint32)),10))					
		default:
			val = Bool(false)
		}
	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Bool(obj1.(Uint) == obj2.(Uint))
		case Uint32:
			val = Bool(uint64(obj1.(Uint)) == uint64(obj2.(Uint32)))
		case Int:
			val = Bool(int64(obj1.(Uint)) == int64(obj2.(Int)))
		case Int32:
			val = Bool(int64(obj1.(Uint)) == int64(obj2.(Int32)))			
		case Float:
			val = Bool(float64(obj2.(Float)) == float64(obj1.(Uint)))			
		case String:
			val = Bool(string(obj2.(String)) == strconv.FormatUint(uint64(obj1.(Uint)), 10))
		default:
			val = Bool(false)
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Bool(obj1.(Uint32) == obj2.(Uint32))
		case Uint:
			val = Bool(uint64(obj1.(Uint32)) == uint64(obj2.(Uint)))			
		case Int32:
			val = Bool(int64(obj1.(Uint32)) == int64(obj2.(Int32)))
		case Int:
			val = Bool(int64(obj1.(Uint32)) == int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj2.(Float)) == float64(obj1.(Uint32)))			
		case String:
			val = Bool(string(obj2.(String)) == strconv.FormatUint(uint64(obj1.(Uint32)), 10))
		default:
			val = Bool(false)
		}
	case RTime:
		switch obj2.(type) {
		case RTime:
			val = Bool(time.Time(obj1.(RTime)).Equal(time.Time(obj2.(RTime))))				
//		case String:
// Add a Parse of an ISO8601 time string, and allow String, RTime combination as well
		default:
			val = Bool(false)	
		}	
		
	}
	return []RObject{val}
}

/*
A primitive is not eq to a non-primitive.
*/
func builtinEqNo(th InterpreterThread, objects []RObject) []RObject {	
	return []RObject{Bool(false)}
}

/*
Non-primitives evaluate to eq if their RObject interfaces are == in Go.
This is probably not what we want eventually.
*/
func builtinEqObj(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	val := Bool(obj1 == obj2)
	return []RObject{val}
}




/*
Inequality operator.
TODO !!!
Really have to study which kinds of equality we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.

Note: I guess the only reason we're not just throwing ! on eq is one less function call and one less
slice allocation.
*/
func builtinNeq(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Bool:
		switch obj2.(type) {
		case Bool:
			val = Bool(obj1.(Bool) != obj2.(Bool))	
		default:
			val = Bool(true)			
		}			
	case Int:
		switch obj2.(type) {
		case Int:
			val = Bool(obj1.(Int) != obj2.(Int))
		case Int32:
			val = Bool(int64(obj1.(Int)) != int64(obj2.(Int32)))
		case Float:
			val = Bool(float64(obj2.(Float)) != float64(obj1.(Int)))			
		case String:
			val = Bool(string(obj2.(String)) != strconv.FormatInt(int64(obj1.(Int)), 10))
		case Uint:
			val = Bool(int64(obj1.(Int)) != int64(obj2.(Uint)))
		case Uint32:
			val = Bool(int64(obj1.(Int)) != int64(obj2.(Uint32)))		
		default:
			val = Bool(true)
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Bool(obj1.(Int32) != obj2.(Int32))
		case Int:
			val = Bool(int64(obj1.(Int32)) != int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj2.(Float)) != float64(obj1.(Int32)))			
		case String:
			val = Bool(string(obj2.(String)) != strconv.FormatInt(int64(obj1.(Int32)), 10))
		case Uint:
			val = Bool(int64(obj1.(Int32)) != int64(obj2.(Uint)))
		case Uint32:
			val = Bool(int64(obj1.(Int32)) != int64(obj2.(Uint32)))		
		default:
			val = Bool(true)
		}
	case Float:
		switch obj2.(type) {
		case Int32:
			val = Bool(float64(obj1.(Float)) != float64(obj2.(Int32)))		
		case Int:
			val = Bool(float64(obj1.(Float)) != float64(obj2.(Int)))	
		case Float:
			val = Bool(float64(obj2.(Float)) != float64(obj1.(Float)))			
		case String:
			val = Bool(string(obj2.(String)) != strconv.FormatFloat(float64(obj1.(Float)), 'G', -1, 64)) 
		case Uint32:
			val = Bool(float64(obj1.(Float)) != float64(obj2.(Uint32)))		
		case Uint:
			val = Bool(float64(obj1.(Float)) != float64(obj2.(Uint)))	
		default:
			val = Bool(true)	
		}	
	case String:
		switch obj2.(type) {
		case String:
			val = Bool(string(obj1.(String)) != string(obj2.(String)))
		case Int:
			val = Bool(string(obj1.(String)) != strconv.FormatInt(int64(obj2.(Int)), 10))
		case Int32:
			val = Bool(string(obj1.(String)) != strconv.Itoa(int(obj2.(Int32))))
		case Float:
			val = Bool(string(obj1.(String)) != strconv.FormatFloat(float64(obj2.(Float)), 'G', -1, 64)) 		
		case Uint:
			val = Bool(string(obj1.(String)) != strconv.FormatUint(uint64(obj2.(Uint)), 10))
		case Uint32:
			val = Bool(string(obj1.(String)) != strconv.FormatUint(uint64(obj2.(Uint32)),10))						
		default:
			val = Bool(true)
		}
	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Bool(obj1.(Uint) != obj2.(Uint))
		case Uint32:
			val = Bool(uint64(obj1.(Uint)) != uint64(obj2.(Uint32)))
		case Int:
			val = Bool(int64(obj1.(Uint)) != int64(obj2.(Int)))
		case Int32:
			val = Bool(int64(obj1.(Uint)) != int64(obj2.(Int32)))			
		case Float:
			val = Bool(float64(obj2.(Float)) != float64(obj1.(Uint)))			
		case String:
			val = Bool(string(obj2.(String)) != strconv.FormatUint(uint64(obj1.(Uint)), 10))
		default:
			val = Bool(false)
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Bool(obj1.(Uint32) != obj2.(Uint32))
		case Uint:
			val = Bool(uint64(obj1.(Uint32)) != uint64(obj2.(Uint)))
		case Int32:
			val = Bool(int64(obj1.(Uint32)) != int64(obj2.(Int32)))
		case Int:
			val = Bool(int64(obj1.(Uint32)) != int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj2.(Float)) != float64(obj1.(Uint32)))			
		case String:
			val = Bool(string(obj2.(String)) != strconv.FormatUint(uint64(obj1.(Uint32)), 10))
		default:
			val = Bool(false)
		}			
	case RTime:
		switch obj2.(type) {
		case RTime:
			val = Bool( ! time.Time(obj1.(RTime)).Equal(time.Time(obj2.(RTime))))				
//		case String:
// Add a Parse of an ISO8601 time string, and allow String, RTime combination as well
		default:
			val = Bool(true)	
		}	
		
	}
	return []RObject{val}
}

/*
A primitive is not eq to a non-primitive.
*/
func builtinNeqNo(th InterpreterThread, objects []RObject) []RObject {	
	return []RObject{Bool(true)}
}

/*
Non-primitives evaluate to eq if their RObject interfaces are == in Go.
This is probably not what we want eventually.
*/
func builtinNeqObj(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	val := Bool(obj1 != obj2)
	return []RObject{val}
}





/*
Less-than operator for numeric types.
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinLtNum(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Bool(int64(obj1.(Int)) < int64(obj2.(Int)))
		case Int32:
			val = Bool(int64(obj1.(Int)) < int64(obj2.(Int32)))
		case Float:
			val = Bool(float64(obj1.(Int)) < float64(obj2.(Float)))
		case Uint:
			val = Bool(int64(obj1.(Int)) < int64(obj2.(Uint)))
		case Uint32:
			val = Bool(int64(obj1.(Int)) < int64(obj2.(Uint32)))			
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Bool(int32(obj1.(Int32)) < int32(obj2.(Int32)))
		case Int:
			val = Bool(int64(obj1.(Int32)) < int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj1.(Int32)) < float64(obj2.(Float)))
		case Uint:
			val = Bool(int64(obj1.(Int32)) < int64(obj2.(Uint)))
		case Uint32:
			val = Bool(int64(obj1.(Int32)) < int64(obj2.(Uint32)))				
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Bool(float64(obj1.(Float)) < float64(obj2.(Float)))
		case Int:
			val = Bool(float64(obj1.(Float)) < float64(obj2.(Int)))
		case Int32:
			val = Bool(float64(obj1.(Float)) < float64(obj2.(Int32)))
		case Uint32:
			val = Bool(float64(obj1.(Float)) < float64(obj2.(Uint32)))		
		case Uint:
			val = Bool(float64(obj1.(Float)) < float64(obj2.(Uint)))			
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Bool(obj1.(Uint) < obj2.(Uint))
		case Uint32:
			val = Bool(uint64(obj1.(Uint)) < uint64(obj2.(Uint32)))
		case Int:
			val = Bool(int64(obj1.(Uint)) < int64(obj2.(Int)))
		case Int32:
			val = Bool(int64(obj1.(Uint)) < int64(obj2.(Int32)))			
		case Float:
			val = Bool(float64(obj2.(Float)) > float64(obj1.(Uint)))			
		case String:
			val = Bool(string(obj2.(String)) > strconv.FormatUint(uint64(obj1.(Uint)), 10))
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Bool(obj1.(Uint32) < obj2.(Uint32))
		case Uint:
			val = Bool(uint64(obj1.(Uint32)) < uint64(obj2.(Uint)))
		case Int32:
			val = Bool(int64(obj1.(Uint32)) < int64(obj2.(Int32)))
		case Int:
			val = Bool(int64(obj1.(Uint32)) < int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj2.(Float)) > float64(obj1.(Uint32)))			
		case String:
			val = Bool(string(obj2.(String)) > strconv.FormatUint(uint64(obj1.(Uint32)), 10))
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	default: rterr.Stop("lt is not defined for argument types")
	}
	return []RObject{val}
}


func builtinLteNum(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinLtNum(th,objects) 
}





/*
Less-than operator for time types.
*/
func builtinLtTime(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case RTime:
		switch obj2.(type) {
		case RTime:
			val = Bool(time.Time(obj1.(RTime)).Before(time.Time(obj2.(RTime))))
		default:
			rterr.Stop("lt is not defined for argument types")
		}
	default: rterr.Stop("lt is not defined for argument types")
	}
	return []RObject{val}
}


func builtinLteTime(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinLtTime(th,objects) 
}


/*
Less-than operator (strings).
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinLtStr(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0].(String)
	obj2 := objects[1].(String)
	var val RObject
	val = Bool(string(obj1) < string(obj2))
	return []RObject{val}
}


func builtinLteStr(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinLtStr(th,objects) 
}


/*
Less-than operator for numeric types.
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinGtNum(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Bool(int64(obj1.(Int)) > int64(obj2.(Int)))
		case Int32:
			val = Bool(int64(obj1.(Int)) > int64(obj2.(Int32)))
		case Float:
			val = Bool(float64(obj1.(Int)) > float64(obj2.(Float)))
		case Uint:
			val = Bool(int64(obj1.(Int)) > int64(obj2.(Uint)))
		case Uint32:
			val = Bool(int64(obj1.(Int)) > int64(obj2.(Uint32)))			
		default:
		rterr.Stop("gt is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Bool(int32(obj1.(Int32)) > int32(obj2.(Int32)))
		case Int:
			val = Bool(int64(obj1.(Int32)) > int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj1.(Int32)) > float64(obj2.(Float)))
		case Uint:
			val = Bool(int64(obj1.(Int32)) > int64(obj2.(Uint)))
		case Uint32:
			val = Bool(int64(obj1.(Int32)) > int64(obj2.(Uint32)))				
		default:
		rterr.Stop("gt is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Bool(float64(obj1.(Float)) > float64(obj2.(Float)))
		case Int:
			val = Bool(float64(obj1.(Float)) > float64(obj2.(Int)))
		case Int32:
			val = Bool(float64(obj1.(Float)) > float64(obj2.(Int32)))
		case Uint32:
			val = Bool(float64(obj1.(Float)) > float64(obj2.(Uint32)))		
		case Uint:
			val = Bool(float64(obj1.(Float)) > float64(obj2.(Uint)))					
		default:
		rterr.Stop("gt is not defined for argument types")
		}

	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Bool(obj1.(Uint) > obj2.(Uint))
		case Uint32:
			val = Bool(uint64(obj1.(Uint)) > uint64(obj2.(Uint32)))
		case Int:
			val = Bool(int64(obj1.(Uint)) > int64(obj2.(Int)))
		case Int32:
			val = Bool(int64(obj1.(Uint)) > int64(obj2.(Int32)))			
		case Float:
			val = Bool(float64(obj2.(Float)) < float64(obj1.(Uint)))			
		case String:
			val = Bool(string(obj2.(String)) < strconv.FormatUint(uint64(obj1.(Uint)), 10))
		default:
		   rterr.Stop("gt is not defined for argument types")
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Bool(obj1.(Uint32) > obj2.(Uint32))
		case Uint:
			val = Bool(uint64(obj1.(Uint32)) > uint64(obj2.(Uint)))
		case Int32:
			val = Bool(int64(obj1.(Uint32)) > int64(obj2.(Int32)))
		case Int:
			val = Bool(int64(obj1.(Uint32)) > int64(obj2.(Int)))
		case Float:
			val = Bool(float64(obj2.(Float)) < float64(obj1.(Uint32)))			
		case String:
			val = Bool(string(obj2.(String)) < strconv.FormatUint(uint64(obj1.(Uint32)), 10))
		default:
		   rterr.Stop("gt is not defined for argument types")
		}	



		default: rterr.Stop("gt is not defined for argument types")	
	}
	return []RObject{val}
}


func builtinGteNum(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinGtNum(th,objects) 
}


/*
Greater-than operator for time types.
*/
func builtinGtTime(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case RTime:
		switch obj2.(type) {
		case RTime:
			val = Bool(time.Time(obj1.(RTime)).After(time.Time(obj2.(RTime))))
		default:
			rterr.Stop("gt is not defined for argument types")
		}
	default: rterr.Stop("gt is not defined for argument types")
	}
	return []RObject{val}
}


func builtinGteTime(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinGtTime(th,objects) 
}


/*
Greater-than operator (strings).
TODO !!!
Really have to study which kinds of order we want to support and with which function names,
by studying other languages. LISP, Java, Go, Python etc.
*/
func builtinGtStr(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0].(String)
	obj2 := objects[1].(String)
	var val RObject
	val = Bool(string(obj1) > string(obj2))
	return []RObject{val}
}


func builtinGteStr(th InterpreterThread, objects []RObject) []RObject {
	if bool( (builtinEq(th,objects))[0].(Bool)) {
		return []RObject{Bool(true)}
	}
	return builtinGtStr(th,objects) 
}


//////////////////////////////////////////////////////////////////////////////////////////////
// Arithmetic functions

/*
Multiply operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinTimes(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) * int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) * int64(obj2.(Int32)))
		case Float:
			val = Float(float64(obj2.(Float)) * float64(obj1.(Int)))
		case Uint:
			val = Int(int64(obj1.(Int)) * int64(uint64(obj2.(Uint))))
		case Uint32:
			val = Int(int64(obj1.(Int)) * int64(uint32(obj2.(Uint32))))				
		default:
		    rterr.Stop("times is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int(int32(obj1.(Int32)) * int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) * int64(obj2.(Int)))
		case Float:
			val = Float(float64(obj2.(Float)) * float64(obj1.(Int32)))
		case Uint32:
			val = Int(int64(obj1.(Int32)) * int64(uint32(obj2.(Uint32))))
		case Uint:
			val = Int(int64(obj1.(Int32)) * int64(uint64(obj2.(Uint))))				
		default:
		    rterr.Stop("times is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Float(float64(obj1.(Float)) * float64(obj2.(Float)))
		case Int:
			val = Float(float64(obj1.(Float)) * float64(obj2.(Int)))
		case Int32:
			val = Float(float64(obj1.(Float)) * float64(obj2.(Int32)))
		case Uint:
			val = Float(float64(obj1.(Float)) * float64(uint64(obj2.(Uint))))
		case Uint32:
			val = Float(float64(obj1.(Float)) * float64(uint32(obj2.(Uint32))))				
		default:
		    rterr.Stop("times is not defined for argument types")
		}
	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Uint(uint64(obj1.(Uint)) * uint64(obj2.(Uint)))
		case Uint32:
			val = Uint(uint64(obj1.(Uint)) * uint64(obj2.(Uint32)))		
		case Float:
			val = Float(float64(obj2.(Float)) * float64(obj1.(Uint)))				
		case Int:
			val = Int(int64(uint64(obj1.(Uint))) * int64(obj2.(Int)))
		case Int32:
			val = Int(int64(uint64(obj1.(Uint))) * int64(obj2.(Int32)))
		default:
		    rterr.Stop("times is not defined for argument types")
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Uint(uint64(obj1.(Uint32)) * uint64(obj2.(Uint32)))					
		case Uint:
			val = Uint(uint64(obj1.(Uint32)) * uint64(obj2.(Uint)))
		case Float:
			val = Float(float64(obj2.(Float)) * float64(obj1.(Uint32)))			
		case Int32:
			val = Int(int64(uint32(obj1.(Uint32))) * int64(obj2.(Int32)))
		case Int:
			val = Int(int64(uint64(obj1.(Uint32))) * int64(obj2.(Int)))	
		default:
		    rterr.Stop("times is not defined for argument types")
		}

	default:
		rterr.Stop("times is not defined for argument types")
	}
	return []RObject{val}
}


/*
Arithmetic Addition operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinPlus(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	return []RObject{plus(obj1,obj2)}
}

/*
Arithmetic Addition operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func plus(obj1, obj2 RObject) RObject {
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) + int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) + int64(obj2.(Int32)))
		case Float:
			val = Float(float64(obj2.(Float)) + float64(obj1.(Int)))
		case Uint:
			val = Int(int64(obj1.(Int)) + int64(uint64(obj2.(Uint))))
		case Uint32:
			val = Int(int64(obj1.(Int)) + int64(uint32(obj2.(Uint32))))	

		default:
		    rterr.Stop("plus is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int(int32(obj1.(Int32)) + int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) + int64(obj2.(Int)))
		case Float:
			val = Float(float64(obj2.(Float)) + float64(obj1.(Int32)))
		case Uint32:
			val = Int(int64(obj1.(Int32)) + int64(uint32(obj2.(Uint32))))
		case Uint:
			val = Int(int64(obj1.(Int32)) + int64(uint64(obj2.(Uint))))					
		default:
		    rterr.Stop("plus is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Float(float64(obj1.(Float)) + float64(obj2.(Float)))
		case Int:
			val = Float(float64(obj1.(Float)) + float64(obj2.(Int)))
		case Int32:
			val = Float(float64(obj1.(Float)) + float64(obj2.(Int32)))
		case Uint:
			val = Float(float64(obj1.(Float)) + float64(uint64(obj2.(Uint))))
		case Uint32:
			val = Float(float64(obj1.(Float)) + float64(uint32(obj2.(Uint32))))				
		default:
		    rterr.Stop("plus is not defined for argument types")
		}
	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Uint(uint64(obj1.(Uint)) + uint64(obj2.(Uint)))
		case Uint32:
			val = Uint(uint64(obj1.(Uint)) + uint64(obj2.(Uint32)))		
		case Float:
			val = Float(float64(obj2.(Float)) + float64(obj1.(Uint)))					
		case Int:
			val = Int(int64(uint64(obj1.(Uint))) + int64(obj2.(Int)))
		case Int32:
			val = Int(int64(uint64(obj1.(Uint))) + int64(obj2.(Int32)))
		default:
		    rterr.Stop("plus is not defined for argument types")
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Uint(uint64(obj1.(Uint32)) + uint64(obj2.(Uint32)))					
		case Uint:
			val = Uint(uint64(obj1.(Uint32)) + uint64(obj2.(Uint)))
		case Float:
			val = Float(float64(obj2.(Float)) + float64(obj1.(Uint32)))				
		case Int32:
			val = Int(int64(uint32(obj1.(Uint32))) + int64(obj2.(Int32)))
		case Int:
			val = Int(int64(uint64(obj1.(Uint32))) + int64(obj2.(Int)))	
		default:
		    rterr.Stop("plus is not defined for argument types")
		}

	default:
		rterr.Stop("plus is not defined for argument types")
	}
	return val
}

/*
Arithmetic sum (of a collection of numbers) operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.

TODO  IMPLEMENT

var ListOfNumericType *RType
var SetOfNumericType *RType

var ListOfFloatType *RType
var SetOfFloatType *RType

var ListOfIntType *RType
var SetOfIntType *RType

var ListOfUIntType *RType
var SetOfUIntType *RType

*/
func builtinSum(th InterpreterThread, objects []RObject) []RObject {
	collection := objects[0].(RCollection)
	t := collection.ElementType()
	var n RObject	
	if collection.Length() == 0 {
		switch t {
		   case FloatType:
			  n = Float(0.)
		   case IntType, Int32Type:
			  n = Int(0)
		   case UintType, Uint32Type:
			  n = Uint(0)
		   case NumericType:
			  n = Float(0.)	
	  	   default:
		     rterr.Stop("wrong element type for sum")				
		}
	} else {
		first := true
		for val := range collection.Iter(th) {
			if first {
				n = val
				first = false
			} else {
			   n = plus(n, val)
		    }
		}
	}	
	
	return []RObject{n}
}


/*
Arithmetic Subtraction operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinMinus(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) - int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) - int64(obj2.(Int32)))
		case Float:
			val = Float(float64(obj1.(Int)) - float64(obj2.(Float)))
		case Uint:
			val = Int(int64(obj1.(Int)) - int64(uint64(obj2.(Uint))))
		case Uint32:
			val = Int(int64(obj1.(Int)) - int64(uint32(obj2.(Uint32))))	
		default:
		    rterr.Stop("minus is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int(int32(obj1.(Int32)) - int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) - int64(obj2.(Int)))
		case Float:
			val = Float(float64(obj1.(Int32)) - float64(obj2.(Float)))
		case Uint32:
			val = Int32(int32(obj1.(Int32)) - int32(uint32(obj2.(Uint32))))
		case Uint:
			val = Int(int64(obj1.(Int32)) - int64(uint64(obj2.(Uint))))				
		default:
		    rterr.Stop("minus is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Float(float64(obj1.(Float)) - float64(obj2.(Float)))
		case Int:
			val = Float(float64(obj1.(Float)) - float64(obj2.(Int)))
		case Int32:
			val = Float(float64(obj1.(Float)) - float64(obj2.(Int32)))
		case Uint:
			val = Float(float64(obj1.(Float)) - float64(uint64(obj2.(Uint))))
		case Uint32:
			val = Float(float64(obj1.(Float)) - float64(uint32(obj2.(Uint32))))	

		default:
		    rterr.Stop("minus is not defined for argument types")
		}
	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Uint(uint64(obj1.(Uint)) - uint64(obj2.(Uint)))
		case Uint32:
			val = Uint(uint64(obj1.(Uint)) - uint64(obj2.(Uint32)))	
		case Float:
			val = Float(float64(obj1.(Uint)) - float64(obj2.(Float)))						
		case Int:
			val = Int(int64(uint64(obj1.(Uint))) - int64(obj2.(Int)))
		case Int32:
			val = Int(int64(uint64(obj1.(Uint))) - int64(obj2.(Int32)))
		default:
		    rterr.Stop("minus is not defined for argument types")
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Uint32(uint32(obj1.(Uint32)) - uint32(obj2.(Uint32)))					
		case Uint:
			val = Uint(uint64(obj1.(Uint32)) - uint64(obj2.(Uint)))
		case Float:
			val = Float(float64(obj1.(Uint32)) - float64(obj2.(Float)))				
		case Int32:
			val = Int32(int32(uint32(obj1.(Uint32))) - int32(obj2.(Int32)))
		case Int:
			val = Int(int64(uint64(obj1.(Uint32))) - int64(obj2.(Int)))	
		default:
		    rterr.Stop("minus is not defined for argument types")
		}

	default:
		rterr.Stop("minus is not defined for argument types")
	}
	return []RObject{val}
}

/*
Arithmetic division operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinDiv(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) / int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) / int64(obj2.(Int32)))
		case Float:
			val = Float(float64(obj1.(Int)) / float64(obj2.(Float)))
		case Uint:
			val = Int(int64(obj1.(Int)) / int64(uint64(obj2.(Uint))))
		case Uint32:
			val = Int(int64(obj1.(Int)) / int64(uint32(obj2.(Uint32))))	

		default:
		    rterr.Stop("div is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int32(int32(obj1.(Int32)) / int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) / int64(obj2.(Int)))
		case Float:
			val = Float(float64(obj1.(Int32)) / float64(obj2.(Float)))
		case Uint32:
			val = Int32(int32(obj1.(Int32)) / int32(uint32(obj2.(Uint32))))
		case Uint:
			val = Int(int64(obj1.(Int32)) / int64(uint64(obj2.(Uint))))				
		default:
		    rterr.Stop("div is not defined for argument types")
		}
	case Float:
		switch obj2.(type) {
		case Float:
			val = Float(float64(obj1.(Float)) / float64(obj2.(Float)))
		case Int:
			val = Float(float64(obj1.(Float)) / float64(obj2.(Int)))
		case Int32:
			val = Float(float64(obj1.(Float)) / float64(obj2.(Int32)))
		case Uint:
			val = Float(float64(obj1.(Float)) / float64(uint64(obj2.(Uint))))
		case Uint32:
			val = Float(float64(obj1.(Float)) / float64(uint32(obj2.(Uint32))))	
		default:
		    rterr.Stop("div is not defined for argument types")
		}
	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Uint(uint64(obj1.(Uint)) / uint64(obj2.(Uint)))
		case Uint32:
			val = Uint(uint64(obj1.(Uint)) / uint64(obj2.(Uint32)))		
		case Float:
			val = Float(float64(obj1.(Uint)) / float64(obj2.(Float)))					
		case Int:
			val = Int(int64(uint64(obj1.(Uint))) / int64(obj2.(Int)))
		case Int32:
			val = Int(int64(uint64(obj1.(Uint))) / int64(obj2.(Int32)))
		default:
		    rterr.Stop("div is not defined for argument types")
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Uint32(uint32(obj1.(Uint32)) / uint32(obj2.(Uint32)))					
		case Uint:
			val = Uint(uint64(obj1.(Uint32)) / uint64(obj2.(Uint)))
		case Float:
			val = Float(float64(obj1.(Uint32)) / float64(obj2.(Float)))				
		case Int32:
			val = Int32(int32(uint32(obj1.(Uint32))) / int32(obj2.(Int32)))
		case Int:
			val = Int(int64(uint64(obj1.(Uint32))) / int64(obj2.(Int)))	
		default:
		    rterr.Stop("div is not defined for argument types")
		}

	default:
		rterr.Stop("div is not defined for argument types")
	}
	return []RObject{val}
}

/*
Arithmetic modulo operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinMod(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	obj2 := objects[1]
	var val RObject
	switch obj1.(type) {
	case Int:
		switch obj2.(type) {
		case Int:
			val = Int(int64(obj1.(Int)) % int64(obj2.(Int)))
		case Int32:
			val = Int(int64(obj1.(Int)) % int64(obj2.(Int32)))
		case Uint:
			val = Int(int64(obj1.(Int)) % int64(uint64(obj2.(Uint))))
		case Uint32:
			val = Int(int64(obj1.(Int)) % int64(uint32(obj2.(Uint32))))	
		default:
		    rterr.Stop("mod is not defined for argument types")
		}
	case Int32:
		switch obj2.(type) {
		case Int32:
			val = Int32(int32(obj1.(Int32)) % int32(obj2.(Int32)))
		case Int:
			val = Int(int64(obj1.(Int32)) % int64(obj2.(Int)))
		case Uint32:
			val = Int32(int32(obj1.(Int32)) % int32(uint32(obj2.(Uint32))))
		case Uint:
			val = Int(int64(obj1.(Int32)) % int64(uint64(obj2.(Uint))))			

		default:
		    rterr.Stop("mod is not defined for argument types")
		}
	case Uint:
		switch obj2.(type) {
		case Uint:
			val = Uint(uint64(obj1.(Uint)) % uint64(obj2.(Uint)))
		case Uint32:
			val = Uint(uint64(obj1.(Uint)) % uint64(obj2.(Uint32)))			
		case Int:
			val = Int(int64(uint64(obj1.(Uint))) % int64(obj2.(Int)))
		case Int32:
			val = Int(int64(uint64(obj1.(Uint))) % int64(obj2.(Int32)))
		default:
		    rterr.Stop("mod is not defined for argument types")
		}
	case Uint32:
		switch obj2.(type) {
		case Uint32:
			val = Uint32(uint32(obj1.(Uint32)) % uint32(obj2.(Uint32)))					
		case Uint:
			val = Uint(uint64(obj1.(Uint32)) % uint64(obj2.(Uint)))
		case Int32:
			val = Int32(int32(uint32(obj1.(Uint32))) % int32(obj2.(Int32)))
		case Int:
			val = Int(int64(uint64(obj1.(Uint32))) % int64(obj2.(Int)))	
		default:
		    rterr.Stop("mod is not defined for argument types")
		}


	default:
		rterr.Stop("mod is not defined for argument types")
	}
	return []RObject{val}
}



/*
Numeric negation operator. Note: Result type varies (is covariant with argument type variation). All results are Numeric but result is a different
subtype of numeric depending on the argument types.
*/
func builtinNeg(th InterpreterThread, objects []RObject) []RObject {
	obj1 := objects[0]
	var val RObject
	switch obj1.(type) {
	case Int:
			val = Int(-int64(obj1.(Int)))
	case Int32:
			val = Int32(-int32(obj1.(Int32)))
	case Float:
			val = Float(-float64(obj1.(Float)))
	default:
		    rterr.Stop("neg is not defined for argument type")
	}
	return []RObject{val}
}



////////////////////////////////////////////////////////////
// Time and Duration functions

// now location String > Time
//
// t = now "America/Los_Angeles"       t = now "Local"      t = now "UTC"
//
// Will throw a runtime error and halt the relish program if the location string is invalid.
//
func builtinTimeNow(th InterpreterThread, objects []RObject) []RObject {
	loc := string(objects[0].(String))
    location,err := time.LoadLocation(loc) 
    if err != nil {
    	rterr.Stop(err.Error())
    }
    t := time.Now().In(location)
	return []RObject{RTime(t)}
}	


// sleep durationNs Int
//
func builtinSleep(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))
	// th.AllowGC("sleep")
	th.AllowGC()	
    time.Sleep(d)
    // th.DisallowGC("sleep")
    th.DisallowGC()    

	return []RObject{}
}	


// tick durationNs Int > Inchannel of Time
//
func builtinTick(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))
    ch := time.Tick(d)

    c := &TimeChannel{Ch: ch}
 
	return []RObject{c}
}	

func builtinGC(th InterpreterThread, objects []RObject) []RObject {
   th.GC()
	return []RObject{}
}





// plus t Time durationNs Int > Time
//
func builtinTimePlus(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
	d := time.Duration(int64(objects[1].(Int)))
    t2 := t.Add(d)
	return []RObject{RTime(t2)}
}	

// addDate t Time years Int months Int days Int > Time
//
func builtinTimeAddDate(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
	years := int(objects[1].(Int))
	months := int(objects[2].(Int))
	days := int(objects[3].(Int))		
    t2 := t.AddDate(years,months,days)
	return []RObject{RTime(t2)}
}	

// minus t Time durationNs Int > Time
//
func builtinTimeMinusDuration(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))	
	d := time.Duration(int64(objects[1].(Int)))
    t2 := t.Add(-d)
	return []RObject{RTime(t2)}	
}	

// minus t2 Time t1 Time > durationNs Int
//
func builtinTimeMinusTime(th InterpreterThread, objects []RObject) []RObject {
	t2 := time.Time(objects[0].(RTime))
	t1 := time.Time(objects[1].(RTime))	
    d := t2.Sub(t1)
	return []RObject{Int(int64(d))}			
}		

// since t Time > durationNs Int
//
func builtinTimeSince(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    d := time.Since(t)
	return []RObject{Int(int64(d))}		
}	

// location t Time > locationName String
//
func builtinTimeLocation(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    loc := t.Location().String()
	return []RObject{String(loc)}		
}	


// timeIn t Time location String > Time 
// 
// Will throw a runtime error and halt the relish program if the location string is invalid.
//
func builtinTimeIn(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
	loc := string(objects[1].(String))	
    location,err := time.LoadLocation(loc) 
    if err != nil {
    	rterr.Stop(err.Error())
    }
    t2 := t.In(location)
	return []RObject{RTime(t2)}	
}	

// hours n Int > durationNs Int
// 
func builtinHours(th InterpreterThread, objects []RObject) []RObject {
	h := int64(objects[0].(Int))	
	d := h * 3600 * 1000000000 
	return []RObject{Int(d)}		
}	

// minutes n Int > durationNs Int
// 
func builtinMinutes(th InterpreterThread, objects []RObject) []RObject {
	m := int64(objects[0].(Int))	
	d := m * 60 * 1000000000 
	return []RObject{Int(d)}		
}		

// seconds n Int > durationNs Int
// 
func builtinSeconds(th InterpreterThread, objects []RObject) []RObject {
	s := int64(objects[0].(Int))	
	d := s * 1000000000 	
	return []RObject{Int(d)}		
}		

// milliseconds n Int > durationNs Int
// 
func builtinMilliseconds(th InterpreterThread, objects []RObject) []RObject {
	ms := int64(objects[0].(Int))	
	d := ms * 1000000 	
	return []RObject{Int(d)}		
}	
	


// date t Time > year Int month Int day Int
//
func builtinTimeDate(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    year, month, day := t.Date()
	return []RObject{Int(year),Int(month),Int(day)}
}

// clock t Time > hour Int min Int sec Int
//
func builtinTimeClock(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    hour, min, sec := t.Clock()
	return []RObject{Int(hour),Int(min),Int(sec)}
}

// day t Time > dayOfMonth Int  
// """
//  0..31
// """
//
func builtinTimeDay(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    day := t.Day()
	return []RObject{Int(day)}
}

// hour t Time >  Int  
// """
//  0..23
// """
//
func builtinTimeHour(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    hour := t.Hour()
	return []RObject{Int(hour)}
}

// minute t Time >  Int  
// """
//  0..59
// """
//
func builtinTimeMinute(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    min := t.Minute()
	return []RObject{Int(min)}
}

// second t Time >  Int  
// """
//  0..59
// """
//
func builtinTimeSecond(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    sec := t.Second()
	return []RObject{Int(sec)}
}

// nanosecond t Time >  Int  
// """
//  0..999999999
// """
//
func builtinTimeNanosecond(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    nsec := t.Nanosecond()
	return []RObject{Int(nsec)}
}

// weekday t Time >  Int  
// """
//  0..6  Sunday = 0 
// """
//
func builtinTimeWeekday(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    wd := t.Weekday()
	return []RObject{Int(int64(int(wd)))}
}

// year t Time >  Int  
//
func builtinTimeYear(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    year := t.Year()
	return []RObject{Int(year)}
}

// month t Time >  Int  
// """
//  1..12  January = 1
// """
//
func builtinTimeMonth(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    month := t.Month()
	return []RObject{Int(int64(int(month)))}
}


// zone t Time > name String offset Int
// """
//  EST secondsEastOfUTC
// """  
//
func builtinTimeZone(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
    zoneName,offset := t.Zone()
	return []RObject{String(zoneName),Int(offset)}
}

// format t Time layout String > String
// """
//  Like Go time layouts (http://golang.org/pkg/time/)
// """  
//
func builtinTimeFormat(th InterpreterThread, objects []RObject) []RObject {
	t := time.Time(objects[0].(RTime))
	layout := string(objects[1].(String))	
    s := t.Format(layout)
	return []RObject{String(s)}
}



// duration hours Int minutes Int > durationNs Int
// 
// duration hours Int minutes Int seconds Int > durationNs Int
// 	
// duration hours Int minutes Int seconds Int nanoseconds Int > durationNs Int
// 
func builtinDuration(th InterpreterThread, objects []RObject) []RObject {
	var h,m,s,ns int64
	h = int64(objects[0].(Int))
	m = int64(objects[1].(Int))
	if len(objects) > 2 {
	   s = int64(objects[0].(Int))
	}
	if len(objects) > 3 {
	   ns = int64(objects[0].(Int))
	}		
	d := h * 3600 * 1000000000 + m * 60 * 1000000000 + s * 1000000000 + ns
	return []RObject{Int(d)}	
}		

func builtinHoursEquivalentOf(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))	
	h := d.Hours()
    return []RObject{Float(h)}	
}

func builtinMinutesEquivalentOf(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))	
	m := d.Minutes()	
    return []RObject{Float(m)}		
}

func builtinSecondsEquivalentOf(th InterpreterThread, objects []RObject) []RObject {
	d := time.Duration(int64(objects[0].(Int)))	
	s := d.Seconds()
    return []RObject{Float(s)}		
}

func builtinTimeParts(th InterpreterThread, objects []RObject) []RObject {
	d := int64(objects[0].(Int))
    h := d / (3600 * 1000000000)	
    excess := d % (3600 * 1000000000)	
    m := excess / (60 * 1000000000)   
    excess = excess % (60 * 1000000000)
    s := excess / 1000000000
    ns := excess % 1000000000
    return []RObject{Int(h),Int(m),Int(s),Int(ns)}	    
}



/////////////////////////////////////////////////////////////////////
// Persistence functions

/*
dub NonPrimitive String 

dub car1 "FEC 092"

Gives the object a name in the local object-persistence database.
This confers persistence on the object if it is not already persistent.

Maybe should return the DBID eventually. Does not right now. TODO

NOTE: Currently, it is an error to dub the same object with a name that already is in use
in the database, which implies it is an error to attempt to dub the object a second time
with the same name.

However, currently, it is legal to give an object multiple names.

Q: Should that become illegal?
*/
func builtinDub(th InterpreterThread, objects []RObject) []RObject {

	relish.EnsureDatabase()
	obj := objects[0]
	name := objects[1].String()

    // Now ok to dub an already persistent object.
	// if obj.IsStoredLocally() {
	// 	rterr.Stopf("Object %v is already persistent. Cannot re-dub (re-persist) as '%s'.", obj, name)
	// }

	// Ensure that the name is not already used for a persistent object in this database.

	found, err := th.DBT().ObjectNameExists(name)
	if err != nil {
		panic(err)
	} else if found {
		rterr.Stopf("The object name '%s' is already in use.", name)
	}

	// Ensure that the object is persisted.

	err = th.DBT().EnsurePersisted(th, obj)
	if err != nil {
		panic(err)
	}

	// Now we have to associate the object with its name in the database. 

	err = th.DBT().NameObject(obj, name)
	if err != nil {
		panic(err)
	}

	return nil
}


func builtinRenameObject(th InterpreterThread, objects []RObject) []RObject {

	relish.EnsureDatabase()
	oldName := objects[0].String()
	newName := objects[1].String()

    var errStr string

	// Ensure that the name is not already used for a persistent object in this database.

	found, err := th.DBT().ObjectNameExists(oldName)
	if err != nil {
		panic(err)
	} else if ! found {
		errStr = fmt.Sprintf("'%s' is not the name of an object. rename cancelled.", oldName)
	}
	
	if errStr != "" {
		found, err = th.DBT().ObjectNameExists(newName)
		if err != nil {
			panic(err)
		} else if found {
			errStr = fmt.Sprintf("The proposed new object name '%s' is already in use. rename cancelled.", newName)
		}	
    }
    
    if errStr == "" {
	    // Now we have to rename the object in the database. 

	    err = th.DBT().RenameObject(oldName, newName)
		if err != nil {
			panic(err)
		}	   
    }

	return []RObject{String(errStr)}
}


/*
summon String > NonPrimitive

car1 = Car: summon "FEC 092"

TODO Add an optional radius argument controlling fetch eagerness of non-primitive attributes and objects related to the object fetched. 
Right now is using radius=0 meaning fetch the object with its primitive (and non-multi) valued attribute values. Other attributes are
currently fetched from DB lazily.

*/
func builtinSummon(th InterpreterThread, objects []RObject) []RObject {
	relish.EnsureDatabase()
	name := objects[0].String()

	obj, err := th.DBT().FetchByName(name, 0)
	if err != nil {
		panic(err)
	}

	return []RObject{obj}
}


func builtinSummonById(th InterpreterThread, objects []RObject) []RObject {
	relish.EnsureDatabase()
	dbid := int64(objects[0].(Int))

	obj, err := th.DBT().Fetch(dbid, 0)
	if err != nil {
		panic(err)
	}

	return []RObject{obj}
}

/*
exists String > Bool

if exists "FEC 092"
   car1 = Car: summon "FEC 092"


*/
func builtinExists(th InterpreterThread, objects []RObject) []RObject {
	relish.EnsureDatabase()
	name := objects[0].String()

	found, err := th.DBT().ObjectNameExists(name) 
	if err != nil {
		panic(err)
	}

	return []RObject{Bool(found)}
}


func builtinDelete(th InterpreterThread, objects []RObject) []RObject {

	relish.EnsureDatabase()
	obj := objects[0]

	err := th.DBT().Delete(obj) 
	if err != nil {
		panic(err)
	}

	return nil
}


var transactionMutex sync.Mutex


// err = begin  // Begins a db transaction. On success returns an empty string.
//
func builtinBeginTransaction(th InterpreterThread, objects []RObject) []RObject {

    transactionType := "EXCLUSIVE"
    if len(objects) > 0 {
	   transactionTypeKeyword := objects[0].String()
	   if transactionTypeKeyword == "READ" {
	   	   transactionType = "DEFERRED"
	   }
    }
	th.AllowGC()	
        //fmt.Println("bBT1")
	transactionMutex.Lock()
        //fmt.Println("bBT2")
	th.DisallowGC()
	defer transactionMutex.Unlock()
	relish.EnsureDatabase()
    var errStr string

    if th.Transaction() != nil {
    	errStr = "Cannot begin a transaction. Goroutine is already participating in an active transaction."
    } else {
	    transactionMutex.Unlock()    	

            //fmt.Println("bBT3")
	    err := th.DBT().BeginTransaction(transactionType)

            //fmt.Println("bBT4")
	    th.AllowGC()		    
	    transactionMutex.Lock()
            //fmt.Println("bBT5")
	    th.DisallowGC()	    	    
		
		if err != nil {
			errStr = err.Error()
		} else {
		    th.SetTransaction(NewTransaction())
	    }
    }
	return []RObject{String(errStr)}
}





// err = local  // Begins a db transaction which TODO: never contributes new or fetched-from-db objects to the global
//              // in-memory object cache. On success returns an empty string.
//
// TODO: Implement MAYBE. THIS IS OBSOLETE !!
//
func builtinBeginLocalTransaction(th InterpreterThread, objects []RObject) []RObject {
	relish.EnsureDatabase()
    var errStr string
    panic("UNIMPLEMENTED")
    err := th.DBT().BeginTransaction("IMMEDIATE")
	if err != nil {
		errStr = err.Error()
	}

	return []RObject{String(errStr)}
}



// err = commit  // Commits the in-progress db transaction. On success returns an empty string.
//               // On failure, retries after 2 seconds, then again after 4 more seconds,
//               // then if still failed, tries a rollback.         
//               // If failed, whether or not successfully rolled back, leaves the dirty persistent objects 
//               // pointing to the RolledBackTransaction, so they will be refreshed from DB.
//               // If succeeded, the dirty persistent objects are marked as clean.
//
//               // OBSOLETE? TODO: If succeeds and is not a local transaction, adds new and fetched-from-db objects to the
//               // global in-memory object cache if they were not already there.
//
func builtinCommitTransaction(th InterpreterThread, objects []RObject) []RObject {
	th.AllowGC()
        //fmt.Println("bCT1")
	transactionMutex.Lock()
        //fmt.Println("bCT2")
	th.DisallowGC()
	defer transactionMutex.Unlock()	
	relish.EnsureDatabase()
    var errStr string

    if th.Transaction() == nil {
    	errStr = "Cannot commit transaction. Goroutine is not participating in an active transaction."
    } else {
            //fmt.Println("bCT3")
	    err := th.DBT().CommitTransaction()
            //fmt.Println("bCT4")
		if err != nil {
			errStr = err.Error()

                        //fmt.Println("bCT Error",errStr)
           rollBackErrStr := rollbackTransactionCore(th)
           if rollBackErrStr != "" {
              errStr = fmt.Sprintf("%s. Rollback also failed: %s",errStr,rollBackErrStr)
              th.DBT().ReleaseDB()
              // Roll back the state of the in memory objects anyway, forcing re-load
              // of objects from database.
              th.Transaction().RollBack()	
              th.SetTransaction(nil)	
              //fmt.Println("bCT Rolled back with DB released.")
           }
		} else {  // db commit succeeded
                   //fmt.Println("bCT5")
		   th.Transaction().Commit()	
		   th.SetTransaction(nil)	
                   //fmt.Println("bCT6")
		}	
    }
	return []RObject{String(errStr)}
}


// err = rollback  // Rolls back the in-progress db transaction. On success returns an empty string.
//                 // TODO - NEED TO DETERMINE POLICY FOR REVERTING CHANGES MADE TO IN MEMORY OBJECTS IF
//                 // THIS WAS A NON-LOCAL TRANSACTION.
//                 // PROBABLY WHAT WE HAVE TO DO IS DECLARE ALL OBJECTS THAT WERE MODIFIED/RELATED DURING
//                 // THE TRANSACTION (IE DIRTY OBJECTS) AS BEING INVALID-IN-MEMORY I.E. THEY NEED TO BE
//                 // RESTORED FROM THE DB BEFORE THEIR STATE IS ACCESSED AGAIN. IS THIS RESTORATION DONE
//                 // RIGHT NOW UPON ROLLBACK, OR ON DEMAND IN GET-ATTR-VAL OPERATION?
//
func builtinRollbackTransaction(th InterpreterThread, objects []RObject) []RObject {
	th.AllowGC()
	transactionMutex.Lock()
	th.DisallowGC()
	defer transactionMutex.Unlock()	
	relish.EnsureDatabase()
    errStr := rollbackTransactionCore(th)

    if errStr != "" {
    	th.DBT().ReleaseDB()
        // Roll back the state of the in memory objects anyway, forcing re-load
        // of objects from database.
        th.Transaction().RollBack()	
        th.SetTransaction(nil)	
    }


	return []RObject{String(errStr)}
}

func rollbackTransactionCore(th InterpreterThread) string {
    var errStr string

    if th.Transaction() == nil {
    	errStr = "Cannot rollback transaction. Goroutine is not participating in an active transaction."
    } else {
	    err := th.DBT().RollbackTransaction()
		if err != nil {
			errStr = err.Error()
		} else {
		   th.Transaction().RollBack()	
		   th.SetTransaction(nil)			
		}
    }
	return errStr
}


//////////////////////////////////////////////////////////////////////
// Non-persistent global context functions


func builtinContext(th InterpreterThread, objects []RObject) []RObject {

	name := objects[0].String()
	val := RT.ContextGet(name)
	return []RObject{val}
}

func builtinContextExists(th InterpreterThread, objects []RObject) []RObject {

	name := objects[0].String()
	found := RT.ContextExists(name)
	return []RObject{Bool(found)}
}

func builtinContextPut(th InterpreterThread, objects []RObject) []RObject {

	obj := objects[0]
	name := objects[1].String()
	RT.ContextPut(obj, name)

	return nil
}

func builtinContextRemove(th InterpreterThread, objects []RObject) []RObject {

	name := objects[0].String()
	RT.ContextRemove(name)
	return nil
}



//////////////////////////////////////////////////////////////////////
// Boolean logic functions

/*
Boolean logic not. 
Return type is Bool
*/
func builtinNot(th InterpreterThread, objects []RObject) []RObject {
	return []RObject{Bool(objects[0].IsZero())}
}

/*
Boolean logic and. Note THIS IS NOT LAZY evaluator of arguments, 
unlike LISP and. Darn. Maybe change that someday.  
Returns false if any argument is zero,
otherwise returns the last argument.
Return type is Any
*/
func builtinAnd(th InterpreterThread, objects []RObject) []RObject {

    var obj RObject
	for _,obj = range objects {
       if obj.IsZero() {
       	 return []RObject{Bool(false)}
       }
	}
	return[]RObject{obj}
}

/*
Boolean logic or. Note THIS IS NOT LAZY evaluator of arguments, 
unlike LISP or. Darn. Maybe change that someday.  
Returns false if all arguments are zero,
otherwise returns the first argument which is non-zero.
Return type is Any
*/
func builtinOr(th InterpreterThread, objects []RObject) []RObject {

    var obj RObject
	for _,obj = range objects {
       if ! obj.IsZero() {
       	 return[]RObject{obj}
       }
	}

	return []RObject{Bool(false)}
}


///////////////////////////////////////////////////////////////////////  
// Collection functions

func builtinLen(th InterpreterThread, objects []RObject) []RObject {
	coll := objects[0].(RCollection)
	var val RObject
	val = Int(coll.Length())
	return []RObject{val}
}

func builtinCap(th InterpreterThread, objects []RObject) []RObject {
	coll := objects[0].(RCollection)
	var val RObject
	val = Int(coll.Cap())
	return []RObject{val}
}

func builtinContains(th InterpreterThread, objects []RObject) []RObject {
	coll := objects[0].(RCollection)
    found := coll.Contains(th, objects[1])
	return []RObject{Bool(found)}
}

func builtinClear(th InterpreterThread, objects []RObject) []RObject {
	coll,isRemovableCollection := objects[0].(RemovableCollection)
    if ! isRemovableCollection {
    	rterr.Stop("Can only apply clear to a mutable,clearable collection or map.")
    }

    owner := coll.Owner()
    if (owner == nil) {
        RT.ClearCollection(th, coll)
    } else {
	    err := RT.ClearAttr(th, owner, coll.Attribute())
	    if err != nil {
	        rterr.Stop(err)
		}	    	
    }

	return []RObject{}
}

/*
Defers (disables) auto-sorting on add element.
*/
func builtinDeferSorting(th InterpreterThread, objects []RObject) []RObject {
	coll,isSortableMixin := objects[0].(SortableMixin)
    if ! isSortableMixin || ! coll.IsSorting() {
    	rterr.Stop("Can only apply deferSorting to a sorting collection or map.")
    }
    coll.SetSortingDeferred(true)
	return []RObject{}
}

/*
Resumes auto-sorting. Does a sort.
TODO TODO !!! Will have to figure out the best way to sort in the db to go along with this!!

Firstly, the bulk of this needs to move out of here into being implemented by the collection, and the
rt methods for manipulating collections.

Strategy should be:
While sorting is deferred, objects should be inserted into the collection association table as if
it were an unsorted "list".

Then, after the in-memory sort, a persistence operation called reorder should happen,
which goes through all collection elements in order in memory and updates the corresponding database
row to have the correct ordinal value.
*/
func builtinResumeSorting(th InterpreterThread, objects []RObject) []RObject {
	sortable,isSortableMixin := objects[0].(SortableMixin)
	coll,isCollection := objects[0].(RCollection)	
    if ! isSortableMixin || !isCollection || ! coll.IsSorting() {
    	rterr.Stop("Can only apply resumeSorting to a sorting collection or map.")
    }
    sortable.SetSortingDeferred(false)

	RT.SetEvalContext(coll, th.EvaluationContext())
    defer RT.UnsetEvalContext(coll)
    sort.Sort(sortable)

	return []RObject{}
}

/*
Copy the elements of the collection to form a list with the same element type constraint as
the collection.
*/
func builtinAsList(th InterpreterThread, objects []RObject) []RObject {
	coll := objects[0].(RCollection)

    list, err := RT.Newrlist(coll.ElementType(),0,-1,nil,nil,nil)
	if err != nil {
		panic(err)
	}

	for obj := range coll.Iter(th) {
		list.AddSimple(obj)
	}


	return []RObject{list}
}


/*
In the database, select the elements of the persistent collection which meet the select clause conditions,
to form a list with the same element type constraint as
the collection.
*/
func builtinAsList2(th InterpreterThread, objects []RObject) []RObject {
	coll := objects[0].(RCollection)
    qExpr := objects[1]

    if coll.ElementType().IsPrimitive {
		  rterr.Stop("asList with query argument can only be applied to a collection of objects, not to a collection of primitive values.")	
    }

    list, err := RT.Newrlist(coll.ElementType(),0,-1,nil,nil,nil)
	if err != nil {
		panic(err)
	}

	query := ""
	queryArgs := []RObject{}

	isList := false
	qS,isString := qExpr.(String)
	if isString {
	 query = string(qS)		
	} else {
		 if qExpr.IsCollection() {
			argColl := qExpr.(RCollection)
			if argColl.IsList() {
				isList = true
				list := argColl.(List)
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
		  rterr.Stop("asList second argument must be a String or a list starting with a String.")	
		}
	}
	radius := 1
	if strings.HasPrefix(query, "lazy: ") {
	 query = query[6:]
	 radius = 0	 
	}

	objs := []RObject{} // TODO Use the existing List's RVector somehow

	mayContainProxies, err := th.DBT().FetchN(list.ElementType(), query, queryArgs, coll, radius, &objs)		
	if err != nil {
	  rterr.Stop(err)
	}	

	list.ReplaceContents(objs)
	list.SetMayContainProxies(mayContainProxies)

	return []RObject{list}
}



/*
Sort the sortable collection.
*/
func builtinSort(th InterpreterThread, objects []RObject) []RObject {
	sortable,isSortableMixin := objects[0].(SortableMixin)
	coll,isCollection := objects[0].(RCollection)	
    if (! isSortableMixin) || (! isCollection) {
    	rterr.Stop("Can only apply sort to a an ordered collection or ordered map.")
    }	
	RT.SetEvalContext(coll, th.EvaluationContext())
    defer RT.UnsetEvalContext(coll)
    sort.Sort(sortable)
	return []RObject{}	
}

/*
Swap the elements at the two indexes.
*/
func builtinSwap(th InterpreterThread, objects []RObject) []RObject {
	sortable,isSortableMixin := objects[0].(SortableMixin)	
    if (! isSortableMixin) {
    	rterr.Stop("Can only apply swap to a an ordered collection or ordered map.")
    }
    i := int(objects[1].(Int))
    j := int(objects[2].(Int))    
    sortable.Swap(i,j)
	return []RObject{}	
}


/*
   Returns a new set whose element type is the same as c1's.
   c2's element type must be same as or a subtype of c1's.
   The returned set contains the (unordered) set-union of the elements of c1 and c2.
*/
func builtinUnion(th InterpreterThread, objects []RObject) []RObject {
	c1 := objects[0].(RCollection)
	c2 := objects[1].(RCollection)


    if ! c2.ElementType().LessEq(c1.ElementType()) {
		  panic("union: Second argument's element type must be same as or a subtype of first argument's element type.")	
    }

    s, err := RT.Newrset(c1.ElementType(),0,-1,nil,nil)
	if err != nil {
		panic(err)
	}    

    a := s.(AddableCollection)

	for obj := range c1.Iter(th) {
		a.AddSimple(obj)
	}

	for obj := range c2.Iter(th) {
		a.AddSimple(obj)
	}

	return []RObject{s}
}



	/*
	   Returns a new set whose element type is the same as c1's.
	   c2's element type can be anything.
	   The returned set contains the (unordered) set-intersection of the elements of c1 and c2.
	*/
func builtinIntersection(th InterpreterThread, objects []RObject) []RObject {
	c1 := objects[0].(RCollection)
	c2 := objects[1].(RCollection)

    s, err := RT.Newrset(c1.ElementType(),0,-1,nil,nil)
	if err != nil {
		panic(err)
	}    

    a := s.(AddableCollection)

	for obj := range c1.Iter(th) {
		if c2.Contains(th, obj) {
		   a.AddSimple(obj)
	    }
	}

	return []RObject{s}
}

	/*
	   Returns a new set whose element type is the same as c1's.
	   c2's element type can be anything.
	   The returned set contains the an unordered set of c1's elements, except any which also occur in c2.
	*/
func builtinSetMinus(th InterpreterThread, objects []RObject) []RObject {
	c1 := objects[0].(RCollection)
	c2 := objects[1].(RCollection)

    s, err := RT.Newrset(c1.ElementType(),0,-1,nil,nil)
	if err != nil {
		panic(err)
	}    

    a := s.(AddableCollection)
	for obj := range c1.Iter(th) {
		if ! c2.Contains(th, obj) {
		   a.AddSimple(obj)
	    }
	}

	return []RObject{s}
}


	/*
	   Appends the elements from c2 to collection c1, which must be a mutable (Addable) collection.
	   c2's element type must be same as or a subtype of c1's.
	   The elements will be appended in c2's natural iteration order, but of course that ordering
	   will be lost if c1 is an unordered collection or is a sorting collection.
	*/
func builtinExtend(th InterpreterThread, objects []RObject) []RObject {
	c1, isAddable := objects[0].(AddableCollection)
	if ! isAddable {
		  panic("extend: First argument is not a collection that permits element additions.")			
	}
	c2 := objects[1].(RCollection)

    if ! c2.ElementType().LessEq(c1.ElementType()) {
		  panic("extend: Second argument's element type must be same as or a subtype of first argument's element type.")	
    }

	for obj := range c2.Iter(th) {
		c1.AddSimple(obj)
	}

	return []RObject{}
}




    /*
    Returns "" if the object has all its mandatory attributes initialized to non-nil, and
    also is fully connected (and within arity constraints) according to relation declarations.
    Returns an appropriate error message if there is a missing mandatory attribute or a
    relationship cardinality violation.
    */
func builtinIncomplete(th InterpreterThread, objects []RObject) []RObject {
	obj := objects[0]	
	typ := obj.Type()

    msg := ""
    sep := ""
    link := ""
    valType := ""
    for _,attr := range typ.Attributes {
    	if attr.Part.ArityLow > 0 {
            if attr.IsRelation() {
            	link = "relation"
            	valType = "associated object"
            } else {
            	link = "attribute"
            	valType = "value"
            }
			val, found := RT.AttrVal(th, obj, attr)
			if ! found {
				msg += sep + fmt.Sprintf("Object %v has no %s for %s %s", obj, valType, link, attr.Part.Name)
                sep = "\n"
			} else if attr.Part.ArityHigh != 1 { // A multi-valued attribute.
                coll := val.(RCollection)
                n := int32(coll.Length())
                if n == 0 {                	
				   msg += sep + fmt.Sprintf("Object %v has no %s for %s %s", obj, valType, link, attr.Part.Name)
                   sep = "\n"
                } else if n < attr.Part.ArityLow {
				   msg += sep + fmt.Sprintf("Object %v has fewer than the allowed minimum %d %ss for %s %s", obj, attr.Part.ArityLow, valType, link, attr.Part.Name)
                   sep = "\n"
        		} else if attr.Part.ArityHigh != -1 && n > attr.Part.ArityHigh {
				   msg += sep + fmt.Sprintf("Object %v has more than the allowed maximum %d %ss for %s %s", obj, attr.Part.ArityHigh, valType, link, attr.Part.Name)
                   sep = "\n"	        			
        		}
			}
    	}
    }


    for _,typ := range typ.Up {
        for _,attr := range typ.Attributes {
        	if attr.Part.ArityLow > 0 {
                if attr.IsRelation() {
                	link = "relation"
                	valType = "associated object"
                } else {
                	link = "attribute"
                	valType = "value"
                }
				val, found := RT.AttrVal(th, obj, attr)
				if ! found {
					msg += sep + fmt.Sprintf("Object %v has no %s for %s %s", obj, valType, link, attr.Part.Name)
                    sep = "\n"
				} else if attr.Part.ArityHigh != 1 { // A multi-valued attribute.
	                coll := val.(RCollection)
	                n := int32(coll.Length())
	                if n == 0 {                	
					   msg += sep + fmt.Sprintf("Object %v has no %s for %s %s", obj, valType, link, attr.Part.Name)
                       sep = "\n"
                    } else if n < attr.Part.ArityLow {
					   msg += sep + fmt.Sprintf("Object %v has fewer than the allowed minimum %d %ss for %s %s", obj, attr.Part.ArityLow, valType, link, attr.Part.Name)
                       sep = "\n"
	        		} else if attr.Part.ArityHigh != -1 && n > attr.Part.ArityHigh {
					   msg += sep + fmt.Sprintf("Object %v has more than the allowed maximum %d %ss for %s %s", obj, attr.Part.ArityHigh, valType, link, attr.Part.Name)
                       sep = "\n"	        			
	        		}
				}
        	}
        }
    }
	return []RObject{String(msg)}
}










///////////////////////////////////////////////////////////////////////
// Concurrency functions

/*

from ch InChannel of T > T

val = from ch

TODO DUMMY Implementation 
*/
func builtinFrom(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(RInChannel)
	// th.AllowGC("From Channel")	
	th.AllowGC()		
    val := c.From()
    // th.DisallowGC("From Channel")
    th.DisallowGC()    
	if val.IsUnit() || val.IsCollection() || val.Type() == ClosureType {
		   RT.DecrementInTransitCount(val)
	}
	return []RObject{val}
}


/*

to 
   ch OutChannel of T 
   val T

to ch val


func builtinTo(objects []RObject) []RObject {
	c := objects[0].(*Channel)
    val := objects[1]
    // TODO do a runtime type-compatibility check of val's type with c.ElementType
	c.Ch <- val

	return []RObject{}
}
*/

func builtinChannelLen(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*Channel)
	var val RObject
	val = Int(c.Length())
	return []RObject{val}
}

func builtinChannelCap(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*Channel)
	var val RObject
	val = Int(c.Cap())
	return []RObject{val}
}

func builtinMutexLock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*Mutex)
	// th.AllowGC("MutexLock")
	th.AllowGC()	
	c.Lock()
    // th.DisallowGC("MutexLock")	
    th.DisallowGC()	    
	return []RObject{}
}

func builtinMutexUnlock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*Mutex)
	c.Unlock()
	return []RObject{}
}


func builtinOwnedMutexLock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*OwnedMutex)
	// th.AllowGC("MutexLock")
	th.AllowGC()	
	c.Lock(th)
    // th.DisallowGC("MutexLock")	
    th.DisallowGC()	    
	return []RObject{}
}

func builtinOwnedMutexUnlock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*OwnedMutex)
	c.Unlock(th)
	return []RObject{}
}


func builtinRWMutexLock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	// th.AllowGC("RWMutexLock")	
	th.AllowGC()	
	c.Lock()
    // th.DisallowGC("RWMutexLock")	
    th.DisallowGC()    	
	return []RObject{}
}

func builtinRWMutexUnlock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.Unlock()
	return []RObject{}
}

func builtinRWMutexRLock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	// th.AllowGC("RWMutexRLock")	
	th.AllowGC()		
	c.RLock()
    // th.DisallowGC("RWMutexRLock")	
    th.DisallowGC()    	
	return []RObject{}
}

func builtinRWMutexRUnlock(th InterpreterThread, objects []RObject) []RObject {
	c := objects[0].(*RWMutex)
	c.RUnlock()
	return []RObject{}
}

/////////////////////////////////////////////////////////////// 
// String functions

// length in bytes

func builtinStringLen(th InterpreterThread, objects []RObject) []RObject {
	obj := objects[0].(String)
    s := string(obj)
	var val RObject
	val = Int(int64(len(s)))
	return []RObject{val}
}

// count of unicode codepoints

func builtinStringNumCodePoints(th InterpreterThread, objects []RObject) []RObject {
	obj := objects[0].(String)
    s := string(obj)
	var val RObject
	val = Int(int64(utf8.RuneCountInString(s)))
	return []RObject{val}
}

/*
contains operator (strings).

contains s String substr String > Bool	

*/
func builtinStringContains(th InterpreterThread, objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.Contains(string(s), string(substr)) )
	return []RObject{val}
}

func builtinStringHasPrefix(th InterpreterThread, objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.HasPrefix(string(s), string(substr)) )
	return []RObject{val}
}

func builtinStringHasSuffix(th InterpreterThread, objects []RObject) []RObject {
	s := objects[0].(String)
	substr := objects[1].(String)
	var val RObject
	val = Bool( strings.HasSuffix(string(s), string(substr)) )
	return []RObject{val}
}


/*
replace s String old String new String > String
"""
  Returns a copy of the string s with all non-overlapping instances of old replaced by new. 
"""

replace s String old String new String n Int > String
"""
  Returns a copy of the string s with the first n non-overlapping instances of old replaced by new. 
  If n < 0, there is no limit on the number of replacements.
"""
*/

func builtinStringReplace(th InterpreterThread, objects []RObject) []RObject {
    var n int = -1
	s := string(objects[0].(String))
	oldPiece := string(objects[1].(String))
	newPiece := string(objects[2].(String))	
	if len(objects) == 4 {
		n = int(objects[3].(Int))
	}
	s2 := strings.Replace(s, oldPiece, newPiece, n)
	return []RObject{String(s2)}
}


func builtinUrlPathPartEncode(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	
	encodeDash := true
	if len(objects) == 2 {
	   	encodeDash = bool(objects[1].(Bool))
	}
	
	var buf bytes.Buffer
    for _,c := range s {
	   if c > 255 {
		  panic("Cannot encode non ISO-LATIN characters in a URL path.")
	   }
	   if c > 127 || c < 32 {
		   buf.WriteString(fmt.Sprintf("%%%02X",c))
	   } else if encodeDash {
		   switch c {
			  case  '$','&','+',',',':',';','=','?','@','"','\'','<','>','#','%','{','}','|','\\','^','[',']','`':
		         buf.WriteString(fmt.Sprintf("%%%02X",c))
		      case '.':
			    buf.WriteString("_~*")		
		      case '-':
			    buf.WriteString("*~*")						
		      case '/':
			    buf.WriteString("~*~")		
		      case ' ':
			    buf.WriteRune('-')	
		      default:
		         buf.WriteRune(c)
	       }
	   } else {		
		   switch c {
			  case  '$','&','+',',','/',':',';','=','?','@',' ','"','\'','<','>','#','%','{','}','|','\\','^','[',']','`':
		         buf.WriteString(fmt.Sprintf("%%%02X",c))			
		      default:
		         buf.WriteRune(c)
	       }
	   }	  
	}
	s2 := buf.String()	
    return []RObject{String(s2)}	
}

func builtinUrlPathPartDecode(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	decodeDash := true
	if len(objects) == 2 {
	   	decodeDash = bool(objects[1].(Bool))
	}	
	
	var buf bytes.Buffer
	var d0,d1,r rune
	digitsToCollect := 0
	charsToSkip := 0
	n := len(s)	
    for i,c := range s {
	   if charsToSkip > 0 {
		  charsToSkip --
	   } else if c == '%' {
		  digitsToCollect = 2
	   } else if digitsToCollect == 2 {
	      if c >= '0' && c <= '9' {
		     d0 = c-'0' 
		  } else if c >= 'A' && c <= 'F' {
		     d0 = 10 + c - 'A' 			 
		  } else if c >= 'a' && c <= 'f' {
		     d0 = 10 + c - 'a'		     
          } else {
	         panic("Bad % escape sequence in URL path part.")
	      } 
	   } else if digitsToCollect == 1 {
	      if c >= '0' && c <= '9' {
		     d1 = c-'0' 
		  } else if c >= 'A' && c <= 'F' {
		     d1 = 10 + c - 'A' 			 
		  } else if c >= 'a' && c <= 'f' {
		     d1 = 10 + c - 'a'		     
          } else {
	         panic("Bad % escape sequence in URL path part.")
	      }
	      r =  16 * d0 + d1
	      buf.WriteRune(r)
	   } else if decodeDash {
            if c == '-' {
               buf.WriteRune(' ')
            } else if c == '*' {
		       if i < n - 2 && s[i+1] == '~' && s[i+2] == '*' {
	              buf.WriteRune('-')
	              charsToSkip = 2			      
			   } else {
				  buf.WriteRune('*')
			   }
		    } else if c == '_' {
		       if i < n - 2 && s[i+1] == '~' && s[i+2] == '*' {
	              buf.WriteRune('.')
	              charsToSkip = 2			      
			   } else {
				  buf.WriteRune('*')
			   }
            } else if c == '~' {
		       if i < n - 2 && s[i+1] == '*' && s[i+2] == '~' {
	              buf.WriteRune('/')
	              charsToSkip = 2			      
			   } else {
				  buf.WriteRune('~')
			   }			   
			} else {	
	           buf.WriteRune(c)
	        }								   
	   } else {	
	      buf.WriteRune(c)		
	   }
	}
	s2 := buf.String()	
	
    return []RObject{String(s2)}	
}












/*
join a []String
"""
 Concatenates the elements of a to create a single string. T
"""

join a []Any sep String > String
"""
 Concatenates the elements of a to create a single string. The separator string sep
 is placed between elements in the resulting string.
"""
*/

func builtinStringJoin(th InterpreterThread, objects []RObject) []RObject {
    list := objects[0].(List)
    objSlice := list.AsSlice(th)
    stringSlice := make([]string,len(objSlice))
    for i,obj := range objSlice {
    	stringSlice[i] = obj.String()
    }
    sep := ""
	if len(objects) == 2 {
		sep = string(objects[1].(String))
    }
	s2 := strings.Join(stringSlice, sep)
	return []RObject{String(s2)}
}

/*
split a String sep String > [] String
*/
func builtinStringSplit(th InterpreterThread, objects []RObject) []RObject {
    s := string(objects[0].(String))
    sep := string(objects[1].(String))

    stringSlice := strings.Split(s,sep)

    elementList, err := RT.Newrlist(StringType, 0, -1, nil, nil, nil)
    if err != nil {
	   panic(err)
    }

    for _,element := range stringSlice {
	   elementList.AddSimple(String(element))
	} 

	return []RObject{elementList}
}



/*
TODO IMPORTANT

Implement a variadic builtin function called fill which is used for python/go style string variable substitution.
e.g.

s = fill "Hello, my name is %s %s and I am %d years old." firstName lastName 19


   s = fill """
"""
Hello!

My name is %s %s 
and I am %d years old.
"""
            firstName
            lastName
            19 

Also, should provide a template builtin function

result = template templateString singleArgObject
*/


func builtinStringFill(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))

	nFillers := len(objects) - 1
    nSlots := strings.Count(s, "%s")
	if nSlots != nFillers {
		snippet := s
		if len(snippet) > 70 {
			chars := strings.Split(s,"")
			if len(chars) > 25 {
			   firstChars := chars[:25]
			   snippet = strings.Join(firstChars,"") + "..."
		    }
		}
        rterr.Stopf("fill: String '%s' contains %d %%s slots but there are %d objects to fill slots.",snippet, nSlots, nFillers)
	} 
    s = strings.Replace(s,"%s","|!@|?!|@?|",-1)

	for i := 1; i <= nSlots; i++ {
       filler := objects[i].String()
       s = strings.Replace(s,"|!@|?!|@?|", filler, 1)
	}
    return []RObject{String(s)}
}



		// concatenation of strings

	    // s = cat s1 String s2 String s3 String s4 String s5 String s6 String s7 String s8 String s9 String > String  

func builtinStringCat(th InterpreterThread, objects []RObject) []RObject {
	s := objects[0].String()
    n := len(objects) 
	for i := 1; i < n; i++ {
		s += objects[i].String()
	}
    return []RObject{String(s)}
}

		// index of substring in string

	/*	
		index s String t String > Int
		"""
		 index of first occurrence of t in s or -1 if t not found in s 
		"""
	*/
	
func builtinStringIndex(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	substr := string(objects[1].(String))
	var val RObject
    val = Int(int64(strings.Index(s, substr)))	
	return []RObject{val}
}


func builtinStringLastIndex(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	substr := string(objects[1].(String))
	var val RObject
    val = Int(int64(strings.LastIndex(s, substr)))	
	return []RObject{val}
}


	    // substring or slice
	/*
		slice s String start Int end Int > String
		
		slice s String start Int > String		
	*/

/*
Counts in bytes.
*/
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

/*
Counts in utf-8 encoded codepoints.
Does not mind if the actual string is shorter than n codepoints.
*/
func builtinStringFirst(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	n := int(objects[1].(Int))
    i := 0
    end := len(s)
    var pos int
	for pos, _ = range s {
	    if i >= n {
		   end = pos
		   break
	    }
	    i++
	}
	substr := s[0:end]
    return []RObject{String(substr)}	
}


/*
Counts in utf-8 encoded codepoints.
Does not mind if the actual string is shorter than n codepoints.
*/
func builtinStringLast(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	n := int(objects[1].(Int))
    if n == 0 {
      return []RObject{String("")}	
    }

    nRunes := utf8.RuneCountInString(s)
    n = nRunes - n
    if n < 0 {
	   n = 0
    }
	
    i := 0
    end := len(s)
    var pos int
    var start int
	for pos, _ = range s {
	    if i >= n {
		   start = pos
		   break
	    }
	    i++
	}
	substr := s[start:end]
    return []RObject{String(substr)}	
}




	
func builtinStringLower(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	lowered := strings.ToLower(s)  
    return []RObject{String(lowered)}	
}	

func builtinStringUpper(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	capitalized := strings.ToUpper(s)      
    return []RObject{String(capitalized)}	
}	

func builtinStringTitle(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	t := strings.Title(strings.ToLower(s))   
    return []RObject{String(t)}	
}	

func builtinStringTrimSpace(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	trimmed := strings.TrimSpace(s)  
    return []RObject{String(trimmed)}	
}	


func builtinStringBase64Encode(th InterpreterThread, objects []RObject) []RObject {
	b := ([]byte)(objects[0].(Bytes))
	urlSafe := bool(objects[1].(Bool)) 
	var lineLen int = -1
	if len(objects) == 3 {
	   lineLen = int(int64(objects[2].(Int)))
   }
	
   var b64 string	
	if urlSafe {
	   b64 = base64.URLEncoding.EncodeToString(b)   
	} else {
	   b64 = base64.StdEncoding.EncodeToString(b)      
   }
   
   
   if lineLen > 0 {
      var buf bytes.Buffer
      var s string
      var n int
      for b64 != "" {
         n = len(b64)
         if n <= lineLen {
            s = b64
            b64 = ""
         } else {
            s = b64[0:lineLen]
            b64 = b64[lineLen:]
         } 
         buf.WriteString(s)
         buf.WriteString("\n")
      }
      b64 = buf.String()
   }
   
   return []RObject{String(b64)}	
}


func builtinStringBase64Decode(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	urlSafe := bool(objects[1].(Bool)) 	

   var errStr string
   var err error
   var decoded []byte
   
   if urlSafe {
	   decoded,err = base64.URLEncoding.DecodeString(s)  
   } else {
	   decoded,err = base64.StdEncoding.DecodeString(s)        
   }
	if err != nil {
	   errStr = err.Error()
	}   
   return []RObject{Bytes(decoded),String(errStr)}	
}



func builtinStringHashSha256Base64(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	hasher := sha256.New()
	hasher.Write([]byte(s))
	sha := hasher.Sum(nil)
	b64 := base64.URLEncoding.EncodeToString(sha)     
    return []RObject{String(b64)}	
}

func builtinStringHashSha256Hex(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	hasher := sha256.New()
	hasher.Write([]byte(s))
	sha := hasher.Sum(nil)
	hex := fmt.Sprintf("%x",sha)   
    return []RObject{String(hex)}	
}

/*
at s String i Int > Byte
"""
 Return the Byte at the specified index in the String.
"""
*/ 
func builtinStringAt(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	i := int(int64(objects[1].(Int)))	
	
    defer stringAtErrHandle(i, s)	
	b := s[i]
  
    return []RObject{Byte(b)}	
}


func stringAtErrHandle(i int, s string) {
      r := recover()	
      if r != nil {
          panic(fmt.Sprintf("Error: index [%d] is out of range. String length is %d.",i,len(s)))
          // rterr.Stopf("Error: index [%d] is out of range. String length is %d.",i,len(s))
      }	
}





/*
runeAt s String i Int > codePoint Int width Int
"""
 Return the unicode codepoint value of the UTF-8 encoded character that starts 
 at the specified byte-index in the String.
 Also return the width in bytes of the UTF-8 encoding of the character.
 At end of string, returns RuneError 0
 At an invalid encoding, returns RuneError 1
"""
*/ 
func builtinStringRuneAt(th InterpreterThread, objects []RObject) []RObject {
	s := string(objects[0].(String))
	i := int(int64(objects[1].(Int)))	
	
    cp, w := utf8.DecodeRuneInString(s[i:])
	
  
    return []RObject{Int(cp),Int(w)}	
}


/////////////////////////////////////////////////////////////// 
// Bytes functions

// length in bytes

func builtinBytesLen(th InterpreterThread, objects []RObject) []RObject {
	obj := objects[0].(Bytes)
    b := []byte(obj)
	var val RObject
	val = Int(int64(len(b)))
	return []RObject{val}
}


/*
at bytes Bytes i Int > Byte
"""
 Return the Byte at the specified index.
"""
*/ 
func builtinBytesAt(th InterpreterThread, objects []RObject) []RObject {
	s := []byte(objects[0].(Bytes))
	i := int(int64(objects[1].(Int)))	
	
    defer bytesAtErrHandle(i, s)	
	b := s[i]
  
    return []RObject{Byte(b)}	
}


/*
set bytes Bytes i Int b Byte

set bytes Bytes i Int b Int

set bytes Bytes i Int b Int32

set bytes Bytes i Int b Uint

set bytes Bytes i Int b Uint32
"""
 Set the Byte at the specified index to the value b.
"""
*/
func builtinBytesSet(th InterpreterThread, objects []RObject) []RObject {
	s := []byte(objects[0].(Bytes))
	i := int(int64(objects[1].(Int)))	
    var val byte
    switch objects[2].(type) {
       case Byte:
          val = byte(objects[2].(Byte))	
       case Int:
          val = byte(int64(objects[2].(Int)))      	
       case Uint:
          val = byte(uint64(objects[2].(Uint)))     	
       case Int32:
       	  val = byte(int32(objects[2].(Int32)))
       case Uint32:
          val = byte(uint32(objects[2].(Uint32)))
       // TODO Add smaller Integer types	
    }	
	
    defer bytesAtErrHandle(i, s)	
	s[i] = val
  
    return []RObject{}	
}



func bytesAtErrHandle(i int, bts []byte) {
      r := recover()	
      if r != nil {
          panic(fmt.Sprintf("Error: index [%d] is out of range. Bytes length is %d.",i,len(bts)))
          // rterr.Stopf("Error: index [%d] is out of range. Bytes length is %d.",i,len(bts))
      }	
}


/*
Counts in bytes.
NOTE!! that byte slices are not currently persistable, so
this method creates a Bytes that re-uses the memory of the argument Bytes.
*/
func builtinBytesSlice(th InterpreterThread, objects []RObject) []RObject {
	s := []byte(objects[0].(Bytes))
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
    return []RObject{Bytes(substr)}
}


/////////////////////////////////////////////////////////////// 
// 

/*


sendEmail 
   smtpServerAddr String 
   from String 
   recipient String 
   subject String 
   messageBody String 
> 
   error String
"""
 Send an email message to a single recipient, without attempting username+password authentication 
 with the sending smtp mail server.
 Returns an empty String "" if it succeeds, otherwise a mail protocol error message.
"""


sendEmail 
   smtpServerAddr String 
   from String 
   recipients [] String 
   subject String 
   messageBody String 
> 
   error String
"""
 Send an email message to possibly multiple recipients, 
 without attempting username+password authentication with the sending smtp mail server.
 Returns an empty String "" if it succeeds, otherwise a mail protocol error message.
"""


sendEmail 
   smtpServerAddr String 
   sendingMailAccountUserName String
   sendingMailAccountPassword String
   from String 
   recipient String 
   subject String 
   messageBody String 
> 
   error String
"""
 Send an email message to a single recipient, first authenticating with the sending smtp mail server
 to establish that we are a permitted sender by using "plain auth" username+password authentication.
 Returns an empty String "" if it succeeds, otherwise a mail protocol error message.
"""


sendEmail 
   smtpServerAddr String 
   sendingMailAccountUserName String
   sendingMailAccountPassword String
   from String 
   recipients [] String 
   subject String 
   messageBody String 
> 
   error String
"""
 Send an email message to possibly multiple recipients, first authenticating with the sending smtp mail server
 to establish that we are a permitted sender by using "plain auth" username+password authentication.
 Returns an empty String "" if it succeeds, otherwise a mail protocol error message.
"""


email4Method, err := RT.CreateMethod("",nil,
                                     "sendEmail", 
                                     []string{"smtpServerAddr","user","password","from","recipients","subject","messageBody"}, 
                                     []string{"String","String","String","String","List_of_String","String","String"}, 
                                     []string{"String"}, false, 0, false)
*/
func builtinSendEmail(th InterpreterThread, objects []RObject) []RObject {
   var auth smtp.Auth = nil
   var serverAddr string 
   var serverName string // without the :port part
   var recipients [] string
   var errStr string
   var offset int = 0
   serverAddr = string(objects[0].(String))
   colonPos := strings.LastIndex(serverAddr,":")
   if colonPos > 0 {
      serverName = serverAddr[:colonPos]	
   } else {
      serverName = serverAddr
      serverAddr += ":25"	
   }
  
   if len(objects) == 7 { // username and password authentication 
	  userName := string(objects[1].(String))
	  password := string(objects[2].(String))
	  if userName != "" {
	     auth = smtp.PlainAuth("", userName, password, serverName)
	  }
	  offset = 2
   } 
   from := string(objects[1 + offset].(String))
   switch objects[2 + offset].(type) {
      case String:
	     recipients = []string{string(objects[2 + offset].(String))}
	  default:	
	     coll :=  objects[2 + offset].(RCollection)
	     for val := range coll.Iter(th) {
		     recipients = append(recipients, string(val.(String)))
	     }
   }
   subject := string(objects[3 + offset].(String)) 
   body := string(objects[4 + offset].(String))

   fromHeader := "From: " + from
   toHeader := "To: "
   sep := ""
   for _,recipient := range recipients {
	  toHeader += sep + recipient
	  sep = ","
   }
   headers := "Subject: " + subject + "\r\n" + fromHeader + "\r\n" + toHeader +  "\r\n\r\n"

   smtp_util.AllowMailingToInsecureTlsSmtpServers()
   
   err := smtp_util.SendMail(serverAddr, 
	                    auth,
                        from,
                        recipients,
                        []byte(headers + body))
   
   if err != nil {
      	errStr = err.Error()
   }

   return []RObject{String(errStr)}
}




// httpGet url String > responseBody String err String
//
// or httpGet url String timeoutSeconds Int
//
// timeoutSeconds, which defaults to 60 if not supplied, is the time that will be waited
// until the first response header records come back to the client. This time includes
// connection establishment, TLS negotiation, and redirects.
//
//
func builtinHttpGet(th InterpreterThread, objects []RObject) []RObject {
	
	urlStr := string(objects[0].(String))

    var timeoutNanos time.Duration = time.Minute

	if len(objects) > 1 {
       timeoutSeconds := int(objects[1].(Int))
       timeoutNanos = time.Duration(timeoutSeconds) * time.Second
	}
	
    contentStr := ""
    errStr := ""

    client := th.HttpClient()
    client.Timeout = timeoutNanos

	res, err := client.Get(urlStr)
	if err != nil {
		errStr = err.Error()
	} else {
		content, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			errStr = err.Error()
		}
		contentStr = string(content)
    }	
	return []RObject{String(contentStr),String(errStr)}
}



// httpPost 
//    url String 
//    keysVals {} String > Any
// > 
//    responseBody String 
//    err String
// """
//  The keysVals map should be a map from String to (value or collection of values)
//  The values are converted to their default String representation to be used as
//  http post argument values.
// """
//
//
// httpPost 
//    url String 
//    mimeType String
//    bodyContentToPost String
// > 
//    responseBody String 
//    err String
// """
//  Posts the data with the specified mime type as the Content-type: header. 
// """
//
//
// httpPost 
//    url String 
//    keysVals {} String > Any
//    timeoutSeconds Int
// > 
//    responseBody String 
//    err String
// """
//  The keysVals map should be a map from String to (value or collection of values)
//  The values are converted to their default String representation to be used as
//  http post argument values.
// """
//
//
// httpPost 
//    url String 
//    mimeType String
//    bodyContentToPost String
//    timeoutSeconds Int
// > 
//    responseBody String 
//    err String
// """
//  Posts the data with the specified mime type as the Content-type: header. 
// """

//
func builtinHttpPost(th InterpreterThread, objects []RObject) []RObject {

    var timeoutNanos time.Duration = time.Minute

    contentStr := ""
    errStr := ""

	urlStr := string(objects[0].(String))

    hasArgumentsMap := true

    switch len(objects) {
    case 2:
    case 3:
    	timeoutInt, isInt := objects[2].(Int)
    	if isInt {
    		timeoutSeconds := int(timeoutInt)
            timeoutNanos = time.Duration(timeoutSeconds) * time.Second
    	} else {
    		hasArgumentsMap = false
    	}
    case 4:
    	hasArgumentsMap = false
        timeoutSeconds := int(objects[3].(Int))
        timeoutNanos = time.Duration(timeoutSeconds) * time.Second
    }
	if hasArgumentsMap {  // 2nd arg is a map of form-input keys-values

		theMap := objects[1].(Map)
		
		values := url.Values{}
	    for key := range theMap.Iter(th) {
		   keyStr := string(key.(String))
		   val, _ := theMap.Get(key)	
	       if val.IsCollection() {
			    coll := val.(RCollection)
			    for obj := range coll.Iter(th) {
				   valStr := obj.String()
			       values.Add(keyStr, valStr)			
			    }
		    } else {
		       values.Set(keyStr, val.String())
		   }
	    }

        client := th.HttpClient()
        client.Timeout = timeoutNanos

		res, err := client.PostForm(urlStr, values)
		if err != nil {
			errStr = err.Error()
		} else {
			content, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				errStr = err.Error()
			}
			contentStr = string(content)
	    }	
    } else {  // httpPost url mimeType bodyContentToPost [timeoutSeconds]
              // - posting raw content with mime-type identified

        // For now we only handle a String as argument.
        // Later should handle a [] Byte too.

        bodyMimeType := string(objects[1].(String))
        bodyContentToPost := string(objects[2].(String))

        stringReader := strings.NewReader(bodyContentToPost) 

        client := th.HttpClient()
        client.Timeout = timeoutNanos

		res, err := client.Post(urlStr, bodyMimeType, stringReader)
		if err != nil {
			errStr = err.Error()
		} else {
			content, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				errStr = err.Error()
			}
			contentStr = string(content)
	    }	
    }
	return []RObject{String(contentStr),String(errStr)}
}



// csvRead
//    content String 
// > 
//    records [] [] String
//    err String
// """
//  
// """
//
// csvRead
//    content String 
//    trimSpace Bool
// > 
//    records [] [] String
//    err String
// """
//  
// """
//
// csvRead
//    content String 
//    trimSpace Bool
//    fieldsPerRecord Int
// > 
//    records [] [] String
//    err String
// """
//  If trimSpace is supplied and true, leading and trailing space is trimmed from each field value.
//  If fieldsPerRecord is supplied and negative, no check is made and each row (each record) can
//  have a different number of fields.
//  If fieldsPerRecord is 0 (the default) each record must have same number of fields
//  as the first record (first row)
//  If fieldsPerRecord is supplied and is a positive integer, it specifies the required number 
//  of fields per record
// """
//
func builtinCsvRead(th InterpreterThread, objects []RObject) []RObject {
   content := string(objects[0].(String))
   trim := false
   if len(objects) > 1 {
      trim = bool(objects[1].(Bool))
   }

   stringReader := strings.NewReader(content) 
   csvReader := csv.NewReader(stringReader)
   csvReader.TrailingComma = true

   if len(objects) > 2 {  // -1 = no check, 0 = same as first row, 1+ = number of columns
   	   csvReader.FieldsPerRecord = int(int64(objects[2].(Int)))
   }


   errStr := ""
   recordsList, err := RT.Newrlist(ListOfStringType, 0, -1, nil, nil, nil)
   if err != nil {
      errStr = err.Error()
   } else {
   	   for {
          record, err := csvReader.Read()
          if err == io.EOF {
          	 break
          } else if err != nil {
          	 errStr = err.Error()
          	 break
          }    

		  recordList, err := RT.Newrlist(StringType, 0, -1, nil, nil, nil)
		  if err != nil {
			 errStr = err.Error()
			 break
		  }
	      for _,field := range record {
	      	 if trim {
	      	 	field = strings.TrimSpace(field)
	      	 }
		     recordList.AddSimple(String(field))
	      }
	      recordsList.AddSimple(recordList)		  
	   }
   }
   return []RObject{recordsList, String(errStr)}
}
	
	
// csvWrite
//    records [] [] String 
// > 
//    content String 
//    err String
// """
//  Each record (i.e. data row) is a list of Strings, one String per field.
//  Writes the records out in CSV format.
//  If there is an error, the returned content string will be incomplete and
//  the second return value will contain an error message.
//  If no error, the second return value will be the empty String ""  
// """
//
// csvWrite
//    records [] [] String 
//    useCrLf Bool
// > 
//    content String 
//    err String
// """
//  Each record (i.e. data row) is a list of Strings, one String per field.
//  Writes the records out in CSV format.
//  If there is an error, the returned content string will be incomplete and
//  the second return value will contain an error message.
//  If no error, the second return value will be the empty String ""  
//  If useCrLf is supplied and true, each line terminates with \r\n instead of
//  the default \n.  \r\n is the Windows-compatible text file line ending,
//  whereas the default \n is the linux and macosx compatible text file format.
// """
//
func builtinCsvWrite(th InterpreterThread, objects []RObject) []RObject {

	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)
	content := ""
	errStr := ""
	recordsList := objects[0].(RCollection) 
	if len(objects) == 2 {
		csvWriter.UseCRLF = bool(objects[1].(Bool))
	}
	for recList := range recordsList.Iter(th) {
		recordList := recList.(RCollection)
		var record []string
		for field := range recordList.Iter(th) {
           record = append(record, string(field.(String)))
		}
        err := csvWriter.Write(record)
        if err != nil {
        	errStr = err.Error()
        	break
        }

	}
	csvWriter.Flush()
	content = buf.String()

	return []RObject{String(content),String(errStr)}
}
	
	
	
	
	
	
/*
jsonMarshal val Any > jsonEncoded String err String
"""
 Marshals the argument value/object/collection including only the public
 attributes of structured objects.
"""

jsonMarshal val Any includePrivate bool > jsonEncoded String err String  
"""
 Marshals the argument value/object/collection including the public
 and if indicated also the private attributes of structured objects.
"""

*/	
func builtinJsonMarshal(th InterpreterThread, objects []RObject) []RObject {
   obj := objects[0]
   var includePrivate bool
   var errStr string
   if len(objects) == 2 {
      	includePrivate = bool(objects[1].(Bool))
   } 
   encoded, err := JsonMarshal(th, obj, includePrivate) 
   if err != nil {
      errStr = err.Error()
   }
   return []RObject{String(encoded), String(errStr)}
}	

	
/*
jsonUnmarshal json String > val Any err String 
jsonUnmarshal json String obj Any > obj Any err String 
"""


 Decodes the json-encoded string argument into a relish object or object tree, which is returned.
 If the second argument is summplied, attempts to populate the supplied object, 
 which must be a collection, a map, or a structured object. In that case, the object argument itself
 is returned, after being populated with attribute/collection values from the JSON string.
"""
*/	
func builtinJsonUnmarshal(th InterpreterThread, objects []RObject) []RObject {
   content := string(objects[0].(String))
   b := []byte(content)
   var v interface{}
   var errStr string
   var resultObj RObject
   err := json.Unmarshal(b, &v)
   if err != nil {
      errStr = err.Error()
   } else if len(objects) == 2 {
	  prototypeObj := objects[1]
	  resultObj, err = prototypeObj.FromMapListTree(th, v) 
      if err != nil {
         errStr = err.Error()	
      }
   } else {
      resultObj = treeFromGoToRelish(v)
   }
   return []RObject{resultObj, String(errStr)}
}

/*
From a Go map/slice/primitive-value tree, create and return a relish map-list-tree.
*/
func treeFromGoToRelish(v interface{}) RObject {
	if v == nil {
		return NIL
	}	
	switch v.(type) {
	 case bool:
		return Bool(v.(bool))
	 case float64:
		return Float(v.(float64))	
	 case string:
		return String(v.(string)) 
	 case []interface{}:
		relishList, err := RT.Newrlist(AnyType, 0, -1, nil, nil, nil)
		if err != nil {
		   panic(err)
		}
		theList := v.([]interface{})
	    for _,val := range theList {
		   obj := treeFromGoToRelish(val)
		   relishList.AddSimple(obj)		
		}
		return relishList
     case map[string]interface{}:
        relishMap,err := RT.Newmap(StringType, AnyType, 0, -1, nil, nil, nil)	
		if err != nil {
		   panic(err)
		}	
	    theMap := v.(map[string]interface{})
	    for key, val := range theMap {
		   obj := treeFromGoToRelish(val)
		   relishMap.PutSimple(String(key),obj)
	    }
		return relishMap 	
	 default: panic("Unexpected value type.")	
    }
    return nil
}


/*

rtPath > String 
"""
 Returns the runtime root directory of the running relish program.
 Either .../relish (for binary relish distros) or .../relish/rt (for source distros).
"""

*/
func builtinRtPath(th InterpreterThread, objects []RObject) []RObject {
   var path string
   pkg := th.CallingPackage()
   if pkg != nil {   
	   if strings.HasPrefix(pkg.Name, "shared.relish.pl2012/dev_tools/pkg/web/playground") {  
	      path = relishRuntimeRoot
	   }
   }
      
   return []RObject{String(path)}
}



/*

dbid obj Any > Int 

*/
func builtinDBID(th InterpreterThread, objects []RObject) []RObject {
   obj := objects[0]
   if ! (obj.HasUUID() && obj.IsBeingStored()) {
	   rterr.Stop("Requested DBID of a non-persistent object.")
   }
   id := obj.DBID()
   
   return []RObject{Int(id)}
}


/*

uuid obj Any > Bytes

*/
func builtinUUID(th InterpreterThread, objects []RObject) []RObject {
   obj := objects[0]
   if ! obj.HasUUID() {
	   rterr.Stop("Requested UUID of an object that has none.")
   }
   uuid := obj.UUID()

   return []RObject{Bytes(uuid)}
}


/*

uuidStr obj Any > String 

*/
func builtinUUIDstr(th InterpreterThread, objects []RObject) []RObject {
   obj := objects[0]
   if ! obj.HasUUID() {
	   rterr.Stop("Requested UUID of an object that has none.")
   }
   uuid := obj.UUIDstr()

   return []RObject{String(uuid)}
}


/*

isPersistedLocally obj Any > Bool

Object storage transaction must have committed for this to be true.
Not true inside the first transaction that is persisting the object.

*/
func builtinIsPersistedLocally(th InterpreterThread, objects []RObject) []RObject {
   obj := objects[0]
   persisted :=  (obj.HasUUID() && obj.IsStoredLocally()) 
   
   return []RObject{Bool(persisted)}
}


/*

hasUuid obj Any > Bool

*/
func builtinHasUUID(th InterpreterThread, objects []RObject) []RObject {
   obj := objects[0]
   return []RObject{Bool(obj.HasUUID())}
}



/*

outputBytes err = exec "commmand" "arg1" "arg2"  

*/
func builtinExec(th InterpreterThread, objects []RObject) []RObject {
   command := string(objects[0].(String))	
   var args []string
   for _,obj := range objects[1:] {
	  args = append(args, string(obj.String()))
   }
   cmd := exec.Command(command, args...)
   output, err := cmd.CombinedOutput()

   var errStr string
   if err != nil {
      errStr = err.Error()
   }

   return []RObject{Bytes(output), String(errStr)}
}


/*

err = execNoWait "commmand" "arg1" "arg2"  

*/
func builtinExecNoWait(th InterpreterThread, objects []RObject) []RObject {
   command := string(objects[0].(String))	
   var args []string
   for _,obj := range objects[1:] {
	  args = append(args, string(obj.String()))
   }
   cmd := exec.Command(command, args...)
   err := cmd.Start()

   var errStr string
   if err != nil {
      errStr = err.Error()
   }

   return []RObject{String(errStr)}
}


func builtinPanic(th InterpreterThread, objects []RObject) []RObject {
   message := string(objects[0].(String))	

   panic(message)

   return []RObject{}
}









/* ISO 8601 format as in W3C

t err = Time "2012-09-23T14:23Z"

t err = Time "2012-09-23T14:23:00Z"

t err = Time "2012-09-23T14:23:00.000Z"

t err = Time "2012-09-23 14:23 UTC"

t err = Time "2012-09-23 14:23:00 America/New York"

t err = Time "2012-09-23 14:23:00.000 Local" 

t err = Time "2012-09-23 14:23:00.000 " "2006-01-02 15:04:05.999 "

t err = Time "2012-09-23 14:23:00.000 " "2006-01-02 15:04:05.999" "America/Los Angeles"

// What is the status of these ones????

initTime t0 Time location String > t Time err String

initTime t0 Time location String > t Time err String

initTime t0 Time inKey String location String > t Time err String

*/

func builtinInitTimeParse(th InterpreterThread, objects []RObject) []RObject {
// ignore first Time arg
	if len(objects) == 2 {
		return initTimeParse1(objects[1].String())
	} else if len(objects) == 4 {
		return initTimeParse3(objects[1].String(),objects[2].String(), objects[3].String())
	} 
	// len = 3
	return initTimeParse2(objects[1].String(),objects[2].String())
}

func initTimeParse1(timeString string) []RObject {

   layout := "2006-01-02T15:04Z"		
	
   t, err := time.Parse(layout, timeString)
   if err != nil {
      layout = "2006-01-02T15:04:05Z"
      t, err = time.Parse(layout, timeString)	
	  if err != nil {
	     layout = "2006-01-02T15:04:05.999Z"
	     t, err = time.Parse(layout, timeString)
	     if err != nil {
             pieces := strings.SplitN(timeString, " " , 3) 
             if len(pieces) < 3 {
			    var tZero time.Time 
			    t := RTime(tZero)
		        return []RObject{t, String("Invalid date-time string format - see relish language reference")}	
             }
             timeString = pieces[0] + " " + pieces[1]
		     locationName := pieces[2]
		     loc, err := time.LoadLocation(locationName) 
		     if err != nil {
			    var tZero time.Time 
			    t := RTime(tZero)
		        return []RObject{t, String(err.Error())}	
		     }	

	         layout = "2006-01-02 15:04"	
             tUtc, err := time.Parse(layout, timeString)		
		     if err != nil {
	            layout = "2006-01-02 15:04:05"				
                tUtc, err = time.Parse(layout, timeString)		
		        if err != nil {			
	               layout = "2006-01-02 15:04:05.999"
                   tUtc, err = time.Parse(layout, timeString)					
		           if err != nil {				  
			          var tZero time.Time 
			          t := RTime(tZero)
		              return []RObject{t, String(err.Error())}	
		           }
		        }
		     }
		     y := tUtc.Year()
		     m := tUtc.Month()
		     d := tUtc.Day()
		     hh := tUtc.Hour()
		     mm := tUtc.Minute()
		     ss := tUtc.Second()
		     ns := tUtc.Nanosecond() 

		     t = time.Date(y, m, d, hh, mm, ss, ns, loc) 
	     }		  
	  }  
   }
   return []RObject{RTime(t),String("")}
}


func initTimeParse2(timeString string, layout string) []RObject {
   t, err := time.Parse(layout, timeString)
   if err != nil {
	  var tZero time.Time 
	  t := RTime(tZero)
      return []RObject{t, String(err.Error())}	
   }		
   return []RObject{RTime(t),String("")}
}


func initTimeParse3(timeString string, layout string, locationName string) []RObject {

   loc, err := time.LoadLocation(locationName) 
   if err != nil {
	  var tZero time.Time 
	  t := RTime(tZero)
      return []RObject{t, String(err.Error())}	
   }

   t, err := time.ParseInLocation(layout, timeString, loc)
   if err != nil {
	  var tZero time.Time 
	  t := RTime(tZero)
      return []RObject{t, String(err.Error())}	
   }		
   return []RObject{RTime(t),String("")}
}



/*
t err = Time 2012 12 30 18 36 29 0 "UTC"

initTime t0 Time year Int month Int day Int hour Int min Int sec Int nsec Int location String > t Time err String
*/
func builtinInitTimeDate(th InterpreterThread, objects []RObject) []RObject {

   // ignore first Time argument
  
   year := int(objects[1].(Int))
   monthInt := int(objects[2].(Int))
   month := time.Month(monthInt)
   day := int(objects[3].(Int))
   hour := int(objects[4].(Int))
   min := int(objects[5].(Int))
   sec := int(objects[6].(Int))
   nsec := int(objects[7].(Int))
   locationName := objects[8].String()
   loc, err := time.LoadLocation(locationName) 
   if err != nil {
	  var tZero time.Time 
	  t := RTime(tZero)
      return []RObject{t, String(err.Error())}	
   }

   t := time.Date(year, month, day, hour, min, sec, nsec, loc) 

   return []RObject{RTime(t),String("")}
}



func builtinInitChannel(th InterpreterThread, objects []RObject) []RObject {
   
    c := objects[0].(*Channel)

	var n int
	if len(objects) == 2 {
		n = int(objects[1].(Int))
		if n < 0 {
			rterr.Stop("Channel capacity cannot be specified to be less than zero.")
		}
	} 
	c.Ch = make(chan RObject, n)
	return []RObject{c}
}



func builtinInitBytes(th InterpreterThread, objects []RObject) []RObject {

	var n int
	var c int
	if len(objects) >= 2 {
		n = int(objects[1].(Int))
		if n < 0 {
			rterr.Stop("Bytes length cannot be specified to be less than zero.")
		}
		if len(objects) == 3 {
			c = int(objects[2].(Int))
			if c < n {
				rterr.Stop("Bytes capacity cannot be specified to be less than length.")
			}			
		} else {
			c = n
		}
	} 
	b := Bytes(make([]byte,n,c))
	return []RObject{b}
}

func builtinInitBytesFromString(th InterpreterThread, objects []RObject) []RObject {
   
    s := string(objects[1].(String))

    b := ([]byte)(s)
	return []RObject{Bytes(b)}
}


/*
Convert any object to a String.
Note: Ignoring string methods defined for programmer-defined datatypes for the moment,
but should probably call that if it exists.
For now, call the builtin string-ification of object Go method.
*/
func builtinInitString(th InterpreterThread, objects []RObject) []RObject {
   
    // ignore the first String argument
    obj := objects[1]
	return []RObject{String(obj.String())}
}

/*
    initInt i Int s String > j Int err String
*/
func builtinInitInt(th InterpreterThread, objects []RObject) []RObject {
    valStr := string(objects[1].(String))

    // ignore the first Int argument
    var errStr string
    v64, err := strconv.ParseInt(valStr, 0, 64)  
    if err != nil {
	   errStr = err.Error()
    }
	return []RObject{Int(v64),String(errStr)}
}



/*
    initInt i Int f Float> j Int err String
*/
func builtinInitIntFromFloat(th InterpreterThread, objects []RObject) []RObject {
    val := float64(objects[1].(Float))

    // ignore the first Int argument
    var errStr string
    i := int64(val)
	return []RObject{Int(i),String(errStr)}
}

/*
    initInt i Int v Int32 > j Int err String
*/
func builtinInitIntFromInt32(th InterpreterThread, objects []RObject) []RObject {
    val := int32(objects[1].(Int32))

    // ignore the first Int argument
    var errStr string
    i := int64(val)
	return []RObject{Int(i),String(errStr)}
}

/*
    initInt i Int v Uint > j Int err String
*/
func builtinInitIntFromUint(th InterpreterThread, objects []RObject) []RObject {
    val := uint64(objects[1].(Uint))

    // ignore the first Int argument
    var errStr string
    i := int64(val)
	return []RObject{Int(i),String(errStr)}
}

/*
    initInt i Int v Uint32 > j Int err String
*/
func builtinInitIntFromUint32(th InterpreterThread, objects []RObject) []RObject {
    val := uint32(objects[1].(Uint32))

    // ignore the first Int argument
    var errStr string
    i := int64(val)
	return []RObject{Int(i),String(errStr)}
}

/*
    initInt i Int v Byte > j Int err String
*/
func builtinInitIntFromByte(th InterpreterThread, objects []RObject) []RObject {
    val := byte(objects[1].(Byte))

    // ignore the first Int argument
    var errStr string
    i := int64(val)
	return []RObject{Int(i),String(errStr)}
}



/*
    initInt32 i Int32 s String > j Int32 err String
*/
func builtinInitInt32(th InterpreterThread, objects []RObject) []RObject {
    valStr := string(objects[1].(String))

    // ignore the first Int argument
    var errStr string
    v64, err := strconv.ParseInt(valStr, 0, 32)  
    if err != nil {
	   errStr = err.Error()
    }
	return []RObject{Int32(int32(v64)),String(errStr)}
}

/*
    initInt32 i Int32 v Int > j Int32 err String
*/
func builtinInitInt32FromInt(th InterpreterThread, objects []RObject) []RObject {
    v := int64(objects[1].(Int))

    // ignore the first Int32 argument
    var errStr string

	return []RObject{Int32(int32(v)),String(errStr)}
}




/*
    initUint i Uint s String > j Uint err String
*/
func builtinInitUint(th InterpreterThread, objects []RObject) []RObject {
    valStr := string(objects[1].(String))

    // ignore the first Uint argument
    var errStr string
    v64, err := strconv.ParseUint(valStr, 0, 64)  
    if err != nil {
	   errStr = err.Error()
    }
	return []RObject{Uint(v64),String(errStr)}
}

/*
    initUint i Uint v Int > j Uint err String
*/
func builtinInitUintFromInt(th InterpreterThread, objects []RObject) []RObject {
    v := int64(objects[1].(Int))

    // ignore the first Uint argument
    var errStr string

	return []RObject{Uint(uint64(v)),String(errStr)}
}

/*
    initUint i Int v Byte > j Uint err String
*/
func builtinInitUintFromByte(th InterpreterThread, objects []RObject) []RObject {
    val := byte(objects[1].(Byte))

    // ignore the first Int argument
    var errStr string
    i := uint64(val)
	return []RObject{Uint(i),String(errStr)}
}



/*
    initUint32 i Uint32 s String > j Uint32 err String
*/
func builtinInitUint32(th InterpreterThread, objects []RObject) []RObject {
    valStr := string(objects[1].(String))

    // ignore the first Uint32 argument
    var errStr string
    v64, err := strconv.ParseUint(valStr, 0, 32)  
    if err != nil {
	   errStr = err.Error()
    }
	return []RObject{Uint32(uint32(v64)),String(errStr)}
}


/*
    initUint32 i Uint32 v Int > j Uint32 err String
*/
func builtinInitUint32FromInt(th InterpreterThread, objects []RObject) []RObject {
    v := int64(objects[1].(Int))

    // ignore the first Uint32 argument
    var errStr string

	return []RObject{Uint32(uint32(v)),String(errStr)}
}

/*
    initUint32 i Int v Byte > j Uint err String
*/
func builtinInitUint32FromByte(th InterpreterThread, objects []RObject) []RObject {
    val := byte(objects[1].(Byte))

    // ignore the first Int argument
    var errStr string
    i := uint32(val)
	return []RObject{Uint32(i),String(errStr)}
}	


/*
    initByte i Byte v Int > j Byte err String
*/
func builtinInitByteFromInt(th InterpreterThread, objects []RObject) []RObject {
    v := int64(objects[1].(Int))

    // ignore the first Byte argument
    var errStr string

	return []RObject{Byte(byte(v)),String(errStr)}
}

/*
    initByte i Byte v Uint > j Byte err String
*/
func builtinInitByteFromUint(th InterpreterThread, objects []RObject) []RObject {
    v := uint64(objects[1].(Uint))

    // ignore the first Byte argument
    var errStr string

	return []RObject{Byte(byte(v)),String(errStr)}
}

/*
    initByte i Byte v Uint32 > j Byte err String
*/
func builtinInitByteFromUint32(th InterpreterThread, objects []RObject) []RObject {
    v := uint32(objects[1].(Uint32))

    // ignore the first Byte argument
    var errStr string

	return []RObject{Byte(byte(v)),String(errStr)}
}





	


















/*
    initFloat f Float s String > g Float err String
*/
func builtinInitFloat(th InterpreterThread, objects []RObject) []RObject {
    valStr := string(objects[1].(String))

    // ignore the first Float argument
    var errStr string
    v64, err := strconv.ParseFloat(valStr, 64)  
    if err != nil {
	   errStr = err.Error()
    }
	return []RObject{Float(v64),String(errStr)}
}


/*
    initFloat f Float i Int s String > g Float err String
*/
func builtinInitFloatFromInt(th InterpreterThread, objects []RObject) []RObject {
    val := int64(objects[1].(Int))

    // ignore the first Float argument
    var errStr string
    f := float64(val)
	return []RObject{Float(f),String(errStr)}
}


func builtinInitBool(th InterpreterThread, objects []RObject) []RObject {

    // ignore the first Bool argument	
    valStr := string(objects[1].(String))
    var v64 bool
    var err error
    var errStr string    
    switch valStr {
       case "Y","y","YES","Yes","yes":
       	   v64 = true
       case "N","n","NO","No","no":
           v64 = false
       default:
           v64, err = strconv.ParseBool(valStr)  
		    if err != nil {
			   errStr = err.Error()
		    }
	}
	return []RObject{Bool(v64),String(errStr)}
}


func builtinInitBoolFromInteger(th InterpreterThread, objects []RObject) []RObject {

    
    // ignore the first Bool argument	

    var val int64
    switch objects[1].(type) {
       case Int:
          val = int64(objects[1].(Int))       	
       case Uint:
          val = int64(objects[1].(Uint))     	
       case Int32:
       	  val = int64(objects[1].(Int32))
       case Uint32:
          val = int64(objects[1].(Uint32))
       // TODO Add smaller Integer types	
    }
    if val == 0 {
    	return []RObject{Bool(false),String("")}
    } else if val == 1 {
    	return []RObject{Bool(true),String("")}    	
    }
    return []RObject{Bool(false),String(fmt.Sprintf("Invalid Integer value %d for conversion to Bool (0,1 accepted)",val))}    
}


