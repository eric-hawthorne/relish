// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// Utility function(s) for handling relish runtime errors.

package rterr

import (
   "os"
   "fmt"
   "relish/dbg"
   "relish/compiler/ast"
   "relish/compiler/token"
   "runtime/debug"
)

// FOR DEBUGGING relish: Set to true to enable a print of the stack to stderr after reporting the error message.
const STACKTRACE = false

/*
   An entity that is located in a relish source code file, and can return a reference to the ast.File node
   for the file it is located in.
*/
type CodeFileLocated interface {
	CodeFile() *ast.File 
}

type Positioned interface {
   Pos() token.Pos
}

/*
Prints the error message on os.Stdout, prefixed with the phrase: Runtime Error: 
then exits the relish process with status 1 (meaning abnormal exit as opposed to status 0 which would mean an exit with success.)
*/
func Stop(errorMessage interface{}) {
	fmt.Fprintln(os.Stdout, "Runtime Error:", errorMessage)
	if STACKTRACE {
    	debug.PrintStack()
    }
	os.Exit(1)
}

/*
Prints the error message on os.Stdout, prefixed with the phrase: Runtime Error: 
then exits the relish process with status 1 (meaning abnormal exit as opposed to status 0 which would mean an exit with success.)
*/
func Stop1(cfl CodeFileLocated, p Positioned, errorMessage interface{}) {

	errorMsgStr := fmt.Sprint("Runtime Error: ",errorMessage)
	file := cfl.CodeFile()
	if file == nil {
		Stop(errorMessage)
	}
	position := file.Position(p.Pos())	
	errorMsgStr = dbg.FmtErr(position, errorMsgStr)
	fmt.Fprintln(os.Stdout,errorMsgStr)
	if STACKTRACE {
    	debug.PrintStack()
	}
	os.Exit(1)
}

/*
Prints the formatted error message on os.Stdout, prefixed with the phrase: Runtime Error: 
Formats the message just like Printf except also adds a carriage return after the formatted message. 
then exits the relish process with status 1 (meaning abnormal exit as opposed to status 0 which would mean an exit with success.)
*/
func Stopf(errorMessage string, args ...interface{}) {
	errorMessage = "Runtime Error: " + errorMessage 
	fmt.Fprintf(os.Stdout, errorMessage, args...)
	fmt.Fprintln(os.Stdout)
	if STACKTRACE {
    	debug.PrintStack()
	}
	os.Exit(1)
}

/*
Prints the formatted error message on os.Stdout, prefixed with the phrase: Runtime Error: 
Formats the message just like Printf except also adds a carriage return after the formatted message. 
then exits the relish process with status 1 (meaning abnormal exit as opposed to status 0 which would mean an exit with success.)
*/
func Stopf1(cfl CodeFileLocated, p Positioned, errorMessage string, args ...interface{}) {
	file := cfl.CodeFile()	
	if file == nil {
		Stopf(errorMessage, args...)
	}	
	errorMessage = "Runtime Error: " + errorMessage 
	errorMessage = fmt.Sprintf(errorMessage, args...)
	position := file.Position(p.Pos())	
	errorMessage = dbg.FmtErr(position, errorMessage)
	fmt.Fprintln(os.Stdout, errorMessage)
	if STACKTRACE {
    	debug.PrintStack()
	}
	os.Exit(1)
}