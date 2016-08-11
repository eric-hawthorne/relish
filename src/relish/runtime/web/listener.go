// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements a web application server for the relish language environment.

package web

/*
   listener.go - implementation of an http listener (web server) 

   with routing of URLs to relish method calls, execution of URL-mapped methods,
   and formatting and processing of method results into http-returned response content.
*/

import (
    . "relish/dbg"
    "fmt"
    "net/http"
	"html/template"
   "util/gos"
	"regexp"
	"bytes"
    "strings"
    "errors"
	. "relish/runtime/data"
	"relish/runtime/interp"
	"sync"
  "relish/rterr"
  "net/url"
)


/*
Decisions: 

1. web or webservice dialog handler functions are public-section functions found
   in the web package or web/something, web/something/else packages.

2. These methods must have a pattern of return arguments which directs the relish runtime as to how
to find, format, and return the response to a web request. The return argument pattern are as follows:

"XML" AnyObjectToBeConverted

"XML"  // pre-formatted if a string 
"""
<?xml version="1.0"?>
<sometag>
</sometag>
"""

// Can we live without the xml tag at beginning? For literal string, for file? Do we add it? probably?

"XML FILE" "some/file/path.xml"


"HTML" "some html string"

"HTML FILE" "foo.html"


"JSON" AnyObjectToBeConverted

"JSON" "literal JSON string"

"JSON FILE" "some/file/path.json"


"MEDIA image/jpeg" AnyObjectPossiblyToBeConverted

"MEDIA FILE image/jpeg" "some/file/path.jpg"


"REDIRECT" [301,302,303,or 307] UrlOrString  http response code defaults to 303 POST-Redirect-GET style.


"path/to/template.html" SomeObjToPassToTemplate

TEMPLATE "literal template text" SomeObjToPassToTemplate


"HTTP ERROR" 404  ["message"] // or 403 (no permission) etc - message defaults to ""


"""
HEADERS 
Content-Type: application/octetstream
Content-Disposition: attachment; filename="fname.ext"
"""
obj              // literally serialized out

Or HEADERS can prefix any of the other forms, to add additional http headers. Should not be inconsistent. What if is?

"""
HEADERS 
Content-Type: application/octetstream
Content-Disposition: attachment; filename="fname.ext"
"""
"MEDIA FILE image/png" 
"some/file/path.png"











TODO

Here are some possibilities for how exposed handler methods are recognized

1. All (public-section) methods in files named [some_thing_]handlers.rel
or maybe [some_thing_]dialog.rel
or maybe [some_thing_]interface.rel

2. Variant on 1 where only public methods (no __private__ or __protected__) are allowed
in those files, if the files occur in subdirs of web directory.

(Sub-question: should it really be called web?? is it only for http protocol?)

3. Have an __exposed__ code section marker which if it occurs in a file takes the place of 
the implicit __exported__ i.e. __public__ section. You cannot have a file which has both
__exposed__ methods in it and also  __exported__ i.e. __public__ ones.
So a file can either contain default (public),__protected__,__private__
or it can contain __exposed__,__protected__,__private

(Note: we don't actually know the definition of protected, private yet! Oops!)

4. Have the methods explicitly annotated in source code, such as
web foo a1 Int a2 String

5. Have any public method which occurs under the web package tree and which
returns as its first return arg a token that indicates the kind of
return value type / processing that is to take place !!!!!!!
Like the special strings below except constants, like

foo a1 Int a2 String > web.ResponseType vehicles.Car
"""
"""
   => web.XML 
      car1   

bar a1 Int a2 String > 
   how web.ResponseType
   what SomeType
"""
"""
   how = web.JSON
   what = car1

baz a1 Int a2 String > 
   how web.ResponseType
   name String
   what SomeType
"""
"""
   how = web.TEMPLATE
   name = "foo.html"  // the template file name (or do we figure out if it is an actual template)
   what = car1   

                      // allow filenames like foo.txt a/b/foo.xml etc otherwise is a raw template string

boz a1 Int a2 String > 
   how web.ResponseType
   name String
   what SomeType
"""
"""
   how = web.MEDIA
   type = "image/jpeg"

   what = car1      

Automatically is an exposed web handler function. also allow args such as disposition encoding etc    
*/



// TODO Implement Post/Redirect/Get webapp pattern http://en.wikipedia.org/wiki/Post/Redirect/Get
//
// functions in web/handlers.rel web/subpackage/handlers.rel etc must have this pattern:
//
// => "XML" obj 
//
// => "JSON" obj
//
// => "MEDIA application/octetstream" obj
//
// => "MEDIA text/plain" obj
//
// => "MEDIA image/jpeg" obj
//
// => "MEDIA mime/type" obj
//
// => """
// HEADERS 
// Content-Type: application/octetstream
// Content-Disposition: attachment; filename="fname.ext"
//    """
//    obj
//
// => "path/to/template" obj          OOPS! Can't distinguish this from mime type
//
// => "REDIRECT" "/path/on/my/webapp/site?a1=v1&a2=v2"
//
// => "REDIRECT" "http://www.foo.com/some/path"
//
// => "REDIRECT" 
//     Url
//        protocol = "https"                  // defaults to http
//        host = "www.foo.com"                // if not specified, creates a path-only url
//        port = 8080                         // defaults to 80 or 443 depending on protocol
//        path = "/some/path"                 // relative paths not supported?
//        kw = { 
//                "a1" => v1
//                "a2" => v2
//             }

//func handler(w http.ResponseWriter, r *http.Request) {
//    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
//}

/*
An interpreter for executing web dialog handler functions
*/
var interpreter *interp.Interpreter

var webPackageSrcDirPath string

func init() {
	interpreter = interp.NewInterpreter(RT)
}	

var funcMap template.FuncMap = template.FuncMap{
    "get": AttrVal, 
    "nonempty": NonEmpty,  
    // "eq": Eq,
    "iterable": Iterable,	
    "fun": InvokeRelishMultiMethod,
    "htm": HtmlPassThru,
}

func SetWebPackageSrcDirPath(path string) {
   webPackageSrcDirPath = path
}




// Note: In relish templates, in pipeline commands that take arguments, 
// only simple expressions such as . or $A are supported.
// Chains of attributes or methods or map-keys are not supported in these contexts.
// Use variable assignment actions {{$variable := pipeline}} instead.

// TODO: We have to resolve where we can look for functions that can be used in a template.
// Probable answer is in the web controller package that the template occurs in,
// or perhaps also in the root of the web packages tree i.e. the web directory.

// Looking for template action escapes: "{{??????????????}}"
//
var re1 *regexp.Regexp = regexp.MustCompile("{{([^{]+)}}")

// looking for ".attrName" or "$.attrName" or "$B.attrName" etc
//
var re2 *regexp.Regexp = regexp.MustCompile(`(\$[A-Za-z0-9]*)?\.([a-z][A-Za-z0-9]*)`)

// looking for "index ." or "index $B" or index $B.foo.bar
//
var re3 *regexp.Regexp = regexp.MustCompile(`index ([^ ]+)`)

// Looking for "afunc" or " aFunc" or "aFuncName123" or " aFuncName123"
var re4 *regexp.Regexp = regexp.MustCompile(`(?:^| )([a-z][A-Za-z0-9]*)`)

var responseProcessingThread *interp.Thread = nil  // current thread - implies serial web method result processing

var responseProcessingPackage *RPackage = nil  // current web package - implies serial web method result processing

var responseProcessingMutex sync.Mutex

func handler(w http.ResponseWriter, r *http.Request) {
	
   path := r.URL.Path
   possibleDotPos := len(path)-7
   if possibleDotPos < 0 {
      possibleDotPos = 0
   }
   if (! strings.HasSuffix(path,".ico")) && (strings.LastIndex(path,".") > possibleDotPos) && (! strings.Contains(path,"?")) {
	  // Serve static content 
	
      // fmt.Fprintln(w, r.URL.Path)

      filePath := webPackageSrcDirPath + "/static" + path      
      http.ServeFile(w,r,filePath)
	
	  return
   }
 	
   pathSegments := strings.Split(path, "/") 
   if len(pathSegments) > 0 && len(pathSegments[0]) == 0 {
      pathSegments = pathSegments[1:]
   }
   var queryString string 
   // Last one or last one -1 has to have ? removed from it
   if len(pathSegments) > 0 {
	   lastPiece := pathSegments[len(pathSegments)-1]	
	   i := strings.Index(lastPiece,"?")	
	   if i > -1 {
	     queryString = lastPiece[i+1:]	
	     if i == 0 {
	        pathSegments = pathSegments[:len(pathSegments)-1]
	     } else {
		    pathSegments[len(pathSegments)-1] = lastPiece[:i]
	     }
	  } else if len(lastPiece) == 0 {
	      pathSegments = pathSegments[:len(pathSegments)-1]		
	  }
   }
   Logln(WEB_, pathSegments) 
   Logln(WEB_, queryString)



   var handlerMethod *RMultiMethod

   pkgName := RT.RunningArtifact + "/pkg/web"
   var pkg *RPackage 
   pkg = RT.Packages[pkgName]
   if pkg == nil {
	  rterr.Stop("No web package has been defined in " + RT.RunningArtifact)
   }


   //    /foo/bar

   remainingPathSegments := pathSegments[:]
   for len(remainingPathSegments) > 0 {
      name := remainingPathSegments[0]
      methodName := underscoresToCamelCase(name)

      handlerMethod = findHandlerMethod(pkg,methodName) 
      if handlerMethod != nil {
        Log(WEB_, "1. %s %s\n",pkg.Name,methodName)  
	      remainingPathSegments = remainingPathSegments[1:]
        Log(WEB_, "    remainingPathSegments: %v\n",remainingPathSegments)       
	      break
	  }
      pkgName += "/" + name
      Log(WEB_, "2. pkgName: %s\n", pkgName)       
      nextPkg := RT.Packages[pkgName]
      if nextPkg != nil {
	     remainingPathSegments = remainingPathSegments[1:]
         pkg = nextPkg
         continue  	   
      }  
      Logln(WEB_, "     package was not found in RT.Packages")           

      if strings.HasSuffix(pkgName,"/pkg/web/favicon.ico") {
         handlerMethod = findHandlerMethod(pkg,"icon")
         if handlerMethod != nil {  
            Log(WEB_, "%s %s\n",pkg.Name,methodName)  
            remainingPathSegments = remainingPathSegments[1:]      
            break
         } else {
            http.Error(w, "", http.StatusNotFound) 
            return
         }
      } 

      // Note that default only handles paths that do not proceed down to 
      // a subdirectory controller package.
      handlerMethod = findHandlerMethod(pkg,"default") 
      if handlerMethod != nil {   
	     // remainingPathSegments = remainingPathSegments[1:]     
         Log(WEB_,"3. Found default handler method in %s\n",pkg.Name) 
	     break
	  }    
      http.Error(w, "404 page or resource not found", http.StatusNotFound)	
      return	
   }
   if handlerMethod == nil {
      handlerMethod = findHandlerMethod(pkg,"index")        	
   }	
   if handlerMethod == nil {
      http.Error(w, "404 page or resource not found", http.StatusNotFound) 
      return       	
   }   

	
   // RUN THE WEB DIALOG HANDLER METHOD 	

   Log(WEB_,"Running dialog handler method: %s\n",handlerMethod.Name)   

   positionalArgStringValues := remainingPathSegments
   keywordArgStringValues, err := getKeywordArgs(r)
   if err != nil {
      fmt.Println(err)  
      fmt.Fprintln(w, err)
      return  
   }     

   //var files map[string] []*multipart.FileHeader
   //if r.MultipartForm != nil {
//	  files = r.MultipartForm.File  // Could still be nil
//   }

   // TODO TODO Should return the InterpreterThread out of here, and
   // Do the commit or rollback later.
   // Or I should demand a thread from the interpreter separately, first, pass it in to 
   // RunServiceMethod, then commit or rollback later.

   t := interpreter.NewThread(nil)

   defer interpreter.DeregisterThread(t)   


   LogM(t, PERSIST2_ ,"Running dialog handler method: %s\n",handlerMethod.Name)   
   Log(GC2_,"Running dialog handler method: %s\n",handlerMethod.Name)   
   Log(GC2_," Args: %v\n",positionalArgStringValues)   
   Log(GC2_," KW Args: %v\n",keywordArgStringValues)   


   // Get annotations of the URL-mapped method.
   // These determine what kind of transaction behaviour it will have.
   // Options:
   // <none>  Wrap method execution and response processing in an IMMEDIATE (reserved for write) TRANSACTION
   // "READ" Wrap method execution and response processing in a DEFERRED TRANSACTION which has done a trial db read.
   // "NOTX" Do not use a long transaction at all. Use AUTOCOMMIT transactions, one per db statement.
   // If not doing any persistence in the service method, use NOTX
   //
   mods, err := interpreter.GetServiceMethodModifiers(handlerMethod)
   if err != nil {
      panic(err)
   }

   if ! mods["NOTX"] {

      var transactionType string
      if mods["READ"] {
         transactionType = "DEFERRED"
      } else {
         transactionType = "EXCLUSIVE"
      }

      err := t.DBT().BeginTransaction(transactionType)
      if err != nil {
         fmt.Println(err)  
         fmt.Fprintln(w, err)
         t.SetErr(err.Error())
         return  
      }       
      t.SetTransaction(NewTransaction())   

      defer t.SetTransaction(nil)
      defer t.CommitOrRollback()
   }
   
   t.SetErr("Uncaught panic while running web app method.")

   // fmt.Printf("Began transaction now running dialog handler method: %s\n",handlerMethod.Name)   
	
   resultObjects,err := interpreter.RunServiceMethod(t, 
	                                                 handlerMethod, 
	                                                 positionalArgStringValues, 
	                                                 keywordArgStringValues,
	                                                 r)   

   // fmt.Printf("Finished running dialog handler method: %s\n",handlerMethod.Name)   
   Log(GC2_,"Finished running dialog handler method: %s\n",handlerMethod.Name)   
   Log(GC2_," Args: %v\n",positionalArgStringValues)   
   Log(GC2_," KW Args: %v\n",keywordArgStringValues)      
     
   if err != nil {
      fmt.Println(err)  
      fmt.Fprintln(w, err)
      t.SetErr(err.Error())
      return  
   }   

   err = processResponse(w,r,pkg, handlerMethod.Name, resultObjects, t)
   if err != nil {
      fmt.Println(err)	
      fmt.Fprintln(w, err)
      t.SetErr(err.Error())      
      return	
   }	

   t.SetErr("")  // Yay! We did not panic
   // fmt.Println("finished processing response of " + handlerMethod.Name)	
}


/*
Handles requests on the special explore port for methods of the explorer_api.
TODO Really don't like that we have a separate, near duplicate handler function here for
explorer_api serving. DRY violation!!
*/
func explorerHandler(w http.ResponseWriter, r *http.Request) {
	
   path := r.URL.Path
	
   pathSegments := strings.Split(path, "/") 
   if len(pathSegments) > 0 && len(pathSegments[0]) == 0 {
      pathSegments = pathSegments[1:]
   }
   var queryString string 
   // Last one or last one -1 has to have ? removed from it
   if len(pathSegments) > 0 {
	   lastPiece := pathSegments[len(pathSegments)-1]	
	   i := strings.Index(lastPiece,"?")	
	   if i > -1 {
	     queryString = lastPiece[i+1:]	
	     if i == 0 {
	        pathSegments = pathSegments[:len(pathSegments)-1]
	     } else {
		    pathSegments[len(pathSegments)-1] = lastPiece[:i]
	     }
	  } else if len(lastPiece) == 0 {
	      pathSegments = pathSegments[:len(pathSegments)-1]		
	  }
   }
   Logln(WEB_, pathSegments) 
   Logln(WEB_, queryString)



   var handlerMethod *RMultiMethod

   pkgName := "shared.relish.pl2012/explorer_api/pkg/web"
   var pkg *RPackage 
   pkg = RT.Packages[pkgName]
   if pkg == nil {
	  rterr.Stop("No web package has been defined in shared.relish.pl2012/explorer_api")
   }


   //    /foo/bar

   remainingPathSegments := pathSegments[:]
   for len(remainingPathSegments) > 0 {
      name := remainingPathSegments[0]
      methodName := underscoresToCamelCase(name)

      handlerMethod = findHandlerMethod(pkg,methodName) 
      if handlerMethod != nil {
        Log(WEB_, "1. %s %s\n",pkg.Name,methodName)  
	      remainingPathSegments = remainingPathSegments[1:]
        Log(WEB_, "    remainingPathSegments: %v\n",remainingPathSegments)       
	      break
	  }
      pkgName += "/" + name
      Log(WEB_, "2. pkgName: %s\n", pkgName)       
      nextPkg := RT.Packages[pkgName]
      if nextPkg != nil {
	     remainingPathSegments = remainingPathSegments[1:]
         pkg = nextPkg
         continue  	   
      }  
      Logln(WEB_, "     package was not found in RT.Packages")           

      if strings.HasSuffix(pkgName,"/pkg/web/favicon.ico") {
         handlerMethod = findHandlerMethod(pkg,"icon")
         if handlerMethod != nil {  
            Log(WEB_, "%s %s\n",pkg.Name,methodName)  
            remainingPathSegments = remainingPathSegments[1:]      
            break
         } else {
            http.Error(w, "", http.StatusNotFound) 
            return
         }
      } 

      // Note that default only handles paths that do not proceed down to 
      // a subdirectory controller package.
      handlerMethod = findHandlerMethod(pkg,"default") 
      if handlerMethod != nil {   
	     // remainingPathSegments = remainingPathSegments[1:]     
         Log(WEB_,"3. Found default handler method in %s\n",pkg.Name) 
	     break
	  }    
      http.Error(w, "404 page or resource not found", http.StatusNotFound)	
      return	
   }
   if handlerMethod == nil {
      handlerMethod = findHandlerMethod(pkg,"index")        	
   }	
   if handlerMethod == nil {
      http.Error(w, "404 page or resource not found", http.StatusNotFound) 
      return       	
   }   

	
   // RUN THE WEB DIALOG HANDLER METHOD 	

   Log(WEB_,"Running dialog handler method: %s\n",handlerMethod.Name)   

   positionalArgStringValues := remainingPathSegments
   keywordArgStringValues, err := getKeywordArgs(r)
   if err != nil {
      fmt.Println(err)  
      fmt.Fprintln(w, err)
      return  
   }     

   //var files map[string] []*multipart.FileHeader
   //if r.MultipartForm != nil {
//	  files = r.MultipartForm.File  // Could still be nil
//   }

   // TODO TODO Should return the InterpreterThread out of here, and
   // Do the commit or rollback later.
   // Or I should demand a thread from the interpreter separately, first, pass it in to 
   // RunServiceMethod, then commit or rollback later.

   t := interpreter.NewThread(nil)

   defer interpreter.DeregisterThread(t)

   t.DBT().BeginTransaction("EXCLUSIVE")

   t.SetTransaction(NewTransaction())   
   t.SetErr("Uncaught panic while running explorer web app method.")
   defer t.SetTransaction(nil)

   defer t.CommitOrRollback()
	
   resultObjects,err := interpreter.RunServiceMethod(t, 
	                                                 handlerMethod, 
	                                                 positionalArgStringValues, 
	                                                 keywordArgStringValues,
	                                                 r)   

   if err != nil {
      fmt.Println(err)  
      fmt.Fprintln(w, err)
      return  
   }   
   
   err = processResponse(w,r,pkg, handlerMethod.Name, resultObjects, t)
   if err != nil {
      fmt.Println(err)  
      fmt.Fprintln(w, err)
      t.SetErr(err.Error())      
      return   
   }  

   t.SetErr("")  // Yay! We did not panic


}



/* Returns the arguments from the combination of the URL query string (part after the ?) and the form values in the request body, 
   if the request was POST or PUT.
   TODO Does not currently do anything with the file part of multipart formdata, if any.
   // return value type is defined in net.url package: type Values map[string][]string
*/
func getKeywordArgs(r *http.Request) (args url.Values, err error) {
   err = r.ParseMultipartForm(10000000) // ok to call this even if it is not a mime/multipart request.
   if err == http.ErrNotMultipart {
      err = nil
   } else if err != nil {
       return
   }
   args = r.Form
   return 
}

/*
headers is expected to be one or more \n terminated http header lines.
Sends these to the ResponseWriter.
*/
func sendHeaders(w http.ResponseWriter, headers string) (err error) {
   headerList := strings.Split(headers,"\n")
   for _,header := range headerList {
      header = strings.TrimSpace(header)
      if len(header) == 0 {
         break
      }
      headerNameVal := strings.Split(header,":") 
      if len(headerNameVal) != 2 {
         err = fmt.Errorf(`Malformed http header '%s'`, header) 
         return 
      }
      headerName := strings.TrimSpace(headerNameVal[0])
      headerVal := strings.TrimSpace(headerNameVal[1])  
      w.Header().Set(headerName, headerVal)          
   }
   return
}

/*
Note should do considerably more checking of Content-Type (detected) and mimesubtype returnval,
to ensure they are consistent with the kind of processing directive. 
Serialize this, and accept a thread.
*/
func processResponse(w http.ResponseWriter, r *http.Request, pkg *RPackage, methodName string, results []RObject, thread *interp.Thread) (err error) {

   
   processingDirective := string(results[0].(String))

   if strings.HasPrefix(processingDirective, "HEADERS") {
      if len(results) < 2 {
         err = fmt.Errorf(`%s HEADERS directive must be followed by another result processing directive.`, methodName) 
        return         
      }
      firstLineEndPos := strings.Index(processingDirective,"\n")
      if firstLineEndPos == -1 {
         err = fmt.Errorf(`%s HEADERS directive must include some http headers, each on a separate line.`, methodName) 
        return         
      }        
      headers := processingDirective[firstLineEndPos+1:]
      if len(headers) == 0 {
         err = fmt.Errorf(`%s HEADERS directive must include some http headers, each on a separate line.`, methodName) 
        return         
      }      

      sendHeaders(w, headers)

      results = results[1:]
      processingDirective = string(results[0].(String))      
   }
   
   switch processingDirective {

    case "XML":
        fmt.Println("XML response not implemented yet.")      
        fmt.Fprintln(w, "XML response not implemented yet.")  

	  case "XML PRE":
	   var xmlContent string
     var mimeType string

       if len(results) < 3 {
         err = fmt.Errorf(`%s XML PRE response requires a mime/type e.g. "text/xml" then an xml-formatted string as third return value`, methodName) 
        return       
      } else if len(results) == 3 {  

         mimeType = string(results[1].(String))     

         xmlContent = string(results[2].(String)) 
         if ! strings.HasPrefix(xmlContent,"<?xml") {
           err = fmt.Errorf("%s XML PRE response requires an xml-formatted string as third return value", methodName) 
           return      
         }
       } else {
             err = fmt.Errorf(`%s XML PRE response has too many return values. Should be "XML PRE" then  a mime/type e.g. "text/xml" then an xml-formatted string`, methodName) 
             return               
       }           

       w.Header().Set("Content-Type", mimeType)         
       fmt.Fprintln(w, xmlContent)		

	  case "XML FILE":
       var filePath string      
       var mimeType string



       if len(results) < 3 {
         err = fmt.Errorf(`%s XML FILE response requires a mime/type e.g. "text/xml" then a filepath`, methodName) 
         return       
       } else if len(results) == 3 {  
          mimeType = string(results[1].(String))                    
          filePath = string(results[2].(String))       
        } else {
              err = fmt.Errorf(`%s XML FILE response has too many return values. Should be a mime/type e.g. "text/xml" then a filepath`, methodName) 
              return               
        } 
        if ! strings.HasSuffix(filePath,".xml") {
            err = fmt.Errorf("%s XML FILE response expecting a .xml file", methodName) 
            return   
        }

        w.Header().Set("Content-Type", mimeType)    
               
        filePath = makeAbsoluteFilePath(methodName, filePath)        
        http.ServeFile(w,r,filePath)		


	  case "HTML":
	   var htmlContent string
       if len(results) < 2 {
         err = fmt.Errorf("%s HTML response requires a html-formatted string as second return value", methodName) 
        return       
      } else if len(results) == 2 {            
         htmlContent = string(results[1].(String)) 
         if ! (strings.HasPrefix(htmlContent,"<html") || strings.HasPrefix(htmlContent,"<HTML") || strings.HasPrefix(htmlContent,"<!DOCTYPE html")) {
           err = fmt.Errorf("%s HTML response requires a html-formatted string as second return value", methodName) 
           return      
         }
       } else if ! ( len(results) == 3 && results[2].IsZero() ) {
             err = fmt.Errorf("%s HTML response has too many return values. Should be 'HTML' then a html-formatted string", methodName) 
             return               
       }		
       fmt.Fprintln(w, htmlContent)		

	  case "HTML FILE":
       var filePath string
       if len(results) < 2 {
         err = fmt.Errorf("%s HTML FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[1].(String))       
        } else if ! ( len(results) == 3 && results[2].IsZero() ) {
              err = fmt.Errorf("%s HTML FILE response has too many return values. Should be filepath", methodName) 
              return               
        } 
        if ! strings.HasSuffix(filePath,".html") {
            err = fmt.Errorf("%s HTML FILE response expecting a .html file", methodName) 
            return   
        }          
        filePath = makeAbsoluteFilePath(methodName, filePath)  
        // fmt.Println(methodName) 
        // fmt.Println(filePath)      
        http.ServeFile(w,r,filePath)

	  case "JSON":	
	   var jsonContent string
       if len(results) < 2 {
         err = fmt.Errorf("%s JSON response requires a second return value, which is to be converted to JSON", methodName) 
        return       
      } else if len(results) == 2 { 
	     obj := results[1]           
	     includePrivate := false
         jsonContent, err = JsonMarshal(thread, obj, includePrivate) 
         if err != nil {
	         err = fmt.Errorf("%s JSON encoding error: %s", methodName, err.Error()) 
             return
         }
       } else if ! ( len(results) == 3 && results[2].IsZero() ) {
             err = fmt.Errorf("%s JSON response has too many return values. Should be 'JSON' then an object/value to be converted to JSON", methodName) 
             return               
       }		
       w.Header().Set("Content-Type", "text/json")
       fmt.Fprintln(w, jsonContent)	

    case "JSON PRE":
     var jsonContent string
       if len(results) < 2 {
         err = fmt.Errorf("%s JSON PRE response requires an json-formatted string as second return value", methodName) 
         return       
       } else if len(results) == 2 {            
         jsonContent = string(results[1].(String)) 
       } else if ! ( len(results) == 3 && results[2].IsZero() ) {
             err = fmt.Errorf("%s JSON PRE response has too many return values. Should be 'XML' then an xml-formatted string", methodName) 
             return               
       }           
       w.Header().Set("Content-Type", "text/json")           
       fmt.Fprintln(w, jsonContent)           
	
	  case "JSON FILE":
        var filePath string		
        if len(results) < 2 {
          err = fmt.Errorf("%s JSON FILE response requires a filepath", methodName) 
          return       
        } else if len(results) == 2 {  
           w.Header().Set("Content-Type", "text/json")		          
           filePath = string(results[1].(String))    
	       filePath = makeAbsoluteFilePath(methodName, filePath)        
	       http.ServeFile(w,r,filePath)
       } else if ! ( len(results) == 3 && results[2].IsZero() ) {		
           err = fmt.Errorf("%s JSON FILE response has too many return values. Should be filepath", methodName) 
           return	
	   }


	  case "IMAGE":	
		var mediaContent string
        var mimeType string	
        var mimeSubtype string          	
	    if len(results) < 3 {
	        err = fmt.Errorf("%s IMAGE response requires an image-data-type (MIME subtype) then a content string", methodName) 
	        return       
        } else if len(results) == 3 {
            mimeSubtype = strings.ToLower(string(results[1].(String)))
            if mimeSubtype == "jpg" {
               mimeSubtype = "jpeg"
            }
            mimeType = "image/" + mimeSubtype    

            // TODO Need to be more flexible about the type conversions here!!!!!!!!!!!!!!!!!!
	          mediaContent = string(results[2].(String)) 
            w.Header().Set("Content-Type", mimeType)	
        } else {
             err = fmt.Errorf("%s IMAGE response has too many return values. Should be 'IMAGE' then a image-data-type (MIME subtype) then a content string", methodName) 
             return               
        }		
        fmt.Fprint(w, mediaContent)	     
	     

    case "IMAGE FILE":  // [mime subtype] filePath
       var filePath string
       var mimeSubtype string       
       var mimeType string
       if len(results) < 2 {
         err = fmt.Errorf("%s IMAGE FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[1].(String))    

        } else if len(results) == 3 {

          if results[2].IsZero() {           
            filePath = string(results[1].(String))              
          } else {
             mimeSubtype = strings.ToLower(string(results[1].(String)))
             if mimeSubtype == "jpg" {
                mimeSubtype = "jpeg"
             }             
             mimeType = "image/" + mimeSubtype
             filePath = string(results[2].(String))   
          } 
        } else {
              err = fmt.Errorf("%s IMAGE FILE response has too many return values. Should be filepath or mimesubtype filepath", methodName) 
              return               
        } 
        if mimeType != "" {
           w.Header().Set("Content-Type", mimeType)
        }
        filePath = makeAbsoluteFilePath(methodName, filePath)        
        http.ServeFile(w,r,filePath)
	  case "VIDEO":
  		var mediaContent string
          var mimeType string	
          var mimeSubtype string            	
  	    if len(results) < 3 {
  	        err = fmt.Errorf("%s VIDEO response requires an image-data-type (MIME subtype) then a content string", methodName) 
  	        return       
          } else if len(results) == 3 {
              mimeSubtype = strings.ToLower(string(results[1].(String)))
              if mimeSubtype == "mpeg4" {
                 mimeSubtype = "mp4"
              }
              mimeType = "video/" + mimeSubtype    

              // TODO Need to be more flexible about the type conversions here!!!!!!!!!!!!!!!!!!
  	          mediaContent = string(results[2].(String)) 
              w.Header().Set("Content-Type", mimeType)	
          } else {
               err = fmt.Errorf("%s VIDEO response has too many return values. Should be 'VIDEO' then a image-data-type (MIME subtype) then a content string", methodName) 
               return               
          }		
          fmt.Fprint(w, mediaContent)	
	  case "VIDEO FILE":
       var filePath string
       var mimeSubtype string       
       var mimeType string
       if len(results) < 2 {
         err = fmt.Errorf("%s VIDEO FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[1].(String))    

        } else if len(results) == 3 {
          if results[2].IsZero() {           
            filePath = string(results[1].(String))           
          } else {
            mimeSubtype = strings.ToLower(string(results[1].(String)))
            if mimeSubtype == "mpeg4" {
               mimeSubtype = "mp4"
            }            
            mimeType = "video/" + mimeSubtype
            filePath = string(results[2].(String))  
          }  
        } else {
              err = fmt.Errorf("%s VIDEO FILE response has too many return values. Should be filepath or mimesubtype filepath", methodName) 
              return               
        } 
        if mimeType != "" {
           w.Header().Set("Content-Type", mimeType)
        }
        filePath = makeAbsoluteFilePath(methodName, filePath)        
        http.ServeFile(w,r,filePath)		
	  case "MEDIA":
		var mediaContent string
        var mimeType string		
	    if len(results) < 3 {
	        err = fmt.Errorf("%s MEDIA response requires a mimetype then a content string", methodName) 
	        return       
        } else if len(results) == 3 {
            mimeType = string(results[1].(String))     

            // TODO Need to be more flexible about the type conversions here!!!!!!!!!!!!!!!!!!
	          mediaContent = string(results[2].(String)) 
            w.Header().Set("Content-Type", mimeType)	
        } else {
             err = fmt.Errorf("%s MEDIA response has too many return values. Should be 'MEDIA' then a mimetype then a content string", methodName) 
             return               
        }		
        fmt.Fprint(w, mediaContent)
	
			
	  case "MEDIA FILE":
       var filePath string   
       var mimeType string
       if len(results) < 2 {
         err = fmt.Errorf("%s MEDIA FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[1].(String))    

        } else if len(results) == 3 {
          if results[2].IsZero() {           
            filePath = string(results[1].(String))              
          } else {          
            mimeType = string(results[1].(String))
            filePath = string(results[2].(String)) 
          }   
        } else {
              err = fmt.Errorf("%s MEDIA FILE response has too many return values. Should be filepath or mimetype filepath", methodName) 
              return               
        } 
        if mimeType != "" {
           w.Header().Set("Content-Type", mimeType)
        }
        filePath = makeAbsoluteFilePath(methodName, filePath)
        http.ServeFile(w,r,filePath)		

	  case "REDIRECT": // [301,302,303,or 307] url 
       var urlStr string
       var code int
       if len(results) < 2 {
         err = fmt.Errorf("%s redirect requires a URL or path", methodName) 
         return       
       } else if len(results) == 2 {
          code = 303   
          // TODO Should handle a builtin URL type as well as String               
          urlStr = string(results[1].(String))    

        } else if len(results) == 3 {
          if results[2].IsZero() {           
             code = 303   
            // TODO Should handle a builtin URL type as well as String               
            urlStr = string(results[1].(String))             
          } else {  

            code = int(results[1].(Int))           
            // TODO Should handle a builtin URL type as well as String
            urlStr = string(results[2].(String))    
          }
        } else {
              err = fmt.Errorf("%s redirect has too many return values. Should be URL or e.g. 307 URL", methodName) 
              return               
        } 
        http.Redirect(w,r,urlStr,code)


	  case "HTTP ERROR":
       var message string
       var code int
       if len(results) < 2 {
         err = fmt.Errorf("%s HTTP ERROR response requires an http error code # e.g. 404", methodName) 
         return       
       } else if len(results) == 2 {
          code = int(results[1].(Int))         
       } else if len(results) == 3 {
          code = int(results[1].(Int))           
          message = string(results[2].(String))    

       } else {
              err = fmt.Errorf("%s HTTP ERROR response has too many return values. Should be e.g. 404 [message]", methodName) 
              return               
       } 
       if code < 400 || code > 599 {
             err = fmt.Errorf("%s HTTP ERROR requires an error code value between 400 and 599 inclusive.", methodName) 
       }
       http.Error(w, message, code)  

    case "HTTP CODE":
       var message string
       var code int
       if len(results) < 2 {
         err = fmt.Errorf("%s HTTP CODE response requires an http response code # e.g. 200", methodName) 
         return       
       } else if len(results) == 2 {
          code = int(results[1].(Int))         
       } else if len(results) == 3 {
          code = int(results[1].(Int))           
          message = string(results[2].(String))    

       } else {
              err = fmt.Errorf("%s HTTP ERROR response has too many return values. Should be e.g. 404 [message]", methodName) 
              return               
       } 
       http.Error(w, message, code)  

			
	  // Do we not need a MIME type argument for this one???
	  case "TEMPLATE": // An inline template as a string	
		if len(results) < 3 { 
		      err = fmt.Errorf("%s (with a templated response) requires return values which are a template string and an object to pass to the template", methodName)            
		       return
		}
		if len(results) > 3 { 
		     err = fmt.Errorf("%s (with a templated response) has unexpected 4th return value", methodName)	
		     return
	    }	
	    relishTemplateText := string(results[1].(String))
	    obj := results[2]
	    err = processTemplateResponse(w, r, pkg, methodName, "", relishTemplateText, obj, thread)
		

    case "HEADERS":
        fmt.Println("HEADERS response not implemented yet.")			
        fmt.Fprintln(w, "HEADERS response not implemented yet.")	
	

	  default: // Must be a template file path or raise an error
	     templateFilePath := processingDirective
		 if ! strings.Contains(templateFilePath,".") { // not a valid path
		    err = fmt.Errorf("'%s' is not a valid response content processing directive, nor a valid template file path",processingDirective)	
        return
      }
		  if len(results) < 2 { 
         err = fmt.Errorf("%s (with a templated response) has no second return value to pass to template", methodName)                       
         return
      }
	    if len(results) > 3 || (len(results) == 3 && ! results[2].IsZero() ) { 
	        err = fmt.Errorf("%s (with a templated response) has unexpected 3rd return value", methodName)	
          return
      }

      templateFilePath = makeAbsoluteFilePath(methodName, templateFilePath) 
	    obj := results[1]
      err = processTemplateFileResponse(w, r, pkg, methodName, templateFilePath, obj, thread) 
   }

   return
}

/*
Find the multimethod in the package's multimethods map. May return nil meaning not found.
TODO Have to make sure it is a public multimethod!!!!!!!!!
*/    
func findHandlerMethod(pkg *RPackage, methodName string) *RMultiMethod {
	methodName = pkg.Name + "/" + methodName
	return pkg.MultiMethods[methodName]
}

/*
Given a file path which is either relative to current src package directory 
e.g. "foo.html" "bar/foo.html"
or is "absolute" with respect to the web package source directory (the web content root)
e.g. "/bar/baz/foo.html"
Return a filesystem absolute path.
Assumes the files are stored in the /src/ directory tree of the artifact.

*/
func makeAbsoluteFilePath(methodName string, filePath string) string {
  if strings.HasPrefix(filePath,"/") {
     return webPackageSrcDirPath + filePath
  }

	slashPos := strings.LastIndex(methodName,"/")
	pkgPos := strings.Index(methodName,"/pkg/")
	packagePath := methodName[pkgPos+8:slashPos] // eliminate up to and including /pkg/web" from beginning of package path

   return webPackageSrcDirPath + packagePath + "/" + filePath
}

func processTemplateFileResponse(w http.ResponseWriter, r *http.Request, pkg *RPackage, methodName string, templateFilePath string, obj RObject, thread *interp.Thread) (err error) {
   bytes,err := gos.ReadFile(templateFilePath)
    if err != nil {
       fmt.Println(err)		
       fmt.Fprintln(w, err)
       return	
    }
    relishTemplateText := string(bytes)	
    err = processTemplateResponse(w, r, pkg, methodName, templateFilePath, relishTemplateText, obj, thread)
    return
}	

/*
templateFilePath may be the empty string (indicating an inline template). If not it is used to make error messages more specific.
This method's implementation grabs a mutex, so that the method's code executes as an exclusive critical section 
(i.e. only one execution at a time.) This is so that the responseProcessingThread can be assigned consistently so that 
relish method execution within template processing knows which interpreter thread to execute the function in.
*/
func processTemplateResponse(w http.ResponseWriter, r *http.Request, pkg *RPackage, methodName string, templateFilePath string, relishTemplateText string, obj RObject, thread *interp.Thread) (err error) {
    thread.AllowGC()
    responseProcessingMutex.Lock()
    thread.DisallowGC()
    defer responseProcessingMutex.Unlock()

     
    responseProcessingThread = thread
    responseProcessingPackage = pkg

    originalDbt := RT.PushDBT(thread.DBT())
    defer RT.PopDBT(originalDbt)

    goTemplateText := goTemplate(relishTemplateText)
    Logln(WEB2_,goTemplateText)

    tmplName := templateFilePath
    if tmplName == "" {
	    tmplName = methodName[strings.LastIndex(methodName,"/")+1:]
    }

    t := template.New(tmplName).Funcs(funcMap)


    t1,err := t.Parse(goTemplateText)
    if err != nil {
       return	
    }

    //myMap := map[string]int{"one":1,"two":2,"three":3}
    //
    //
    //myList := []string{"foo","bar","baz"}
    //
    //var ob interface{} = myList
    //t1.Execute(w, ob)    
    err = t1.Execute(w, obj)
    return
}

















/*

"XML" AnyObjectToBeConverted

"XML PRE" 
"""
<?xml version="1.0"?>
<sometag>
</sometag>
"""

// Can we live without the xml tag at beginning? For literal string, for file? Do we add it? probably?

"XML FILE" "some/file/path.xml"


"HTML"

"HTML FILE" "foo.html"


"JSON" AnyObjectToBeConverted


"JSON FILE" "some/file/path.json"


"IMAGE" ["jpeg"] ObjectPossiblyToBeConverted

"IMAGE FILE" ["jpeg"] "some/file/path.jpg"

"VIDEO" ["mpeg4"] ObjectPossiblyToBeConverted  

"VIDEO FILE" ["mpeg4"] "some/file/path.mp4"

"MEDIA" ["mime/type"] ObjectPossiblyToBeConverted

"MEDIA FILE" ["application/x-octetstream"] "some/file/path.dat"



"REDIRECT"  [code] UrlOrString


"path/to/template.html" SomeObjToPassToTemplate

"TEMPLATE" "literal template text" SomeObjToPassToTemplate


"HTTP ERROR" 404   // or 403 (no permission) etc  need/allow a string message too?


"""
HEADERS 
Content-Type: application/octetstream
Content-Disposition: attachment; filename="fname.ext"
"""
obj              // literally serialized out

Or HEADERS can prefix any of the other forms, to add additional http headers. Should not be inconsistent. What if is?

"""
HEADERS 
Content-Type: application/octetstream
Content-Disposition: attachment; filename="fname.ext"
"""
"MEDIA FILE image/png" 
"some/file/path.png"



*/




















func underscoresToCamelCase(s string) string {
	ss := strings.Split(s,"_")
	var cs string
	for i, si := range ss {
		if i == 0 {
			cs = si
		} else {
			cs += strings.Title(si)
		}
	}
	return cs
}

var handlerAdded = false; // Whether the relish webapp method-mapping handler has been added to DefaultServeMux yet
var listenerConfigMutex sync.Mutex

/*
  Starts up relish web app serving on the specified port.
  If sourceCodeShareDir is not "" it should be the "relish/shared" 
  or "relish/rt/shared" of "relish/4production/shared" or "relish/rt/4production/shared" directory. 
  In that case, also serves source code from the shared directory tree.
*/
func ListenAndServe(portNumber int, sourceCodeShareDir string) {
	if sourceCodeShareDir != "" {
		http.Handle("/relish/", http.FileServer(http.Dir(sourceCodeShareDir)))
	}
    listenerConfigMutex.Lock()
    if ! handlerAdded { 
    http.HandleFunc("/", handler)
       handlerAdded = true
    }
    listenerConfigMutex.Unlock()
    http.ListenAndServe(fmt.Sprintf(":%d",portNumber), nil)
}

/*
  Starts up relish web app serving via TLS on the specified port.
  If sourceCodeShareDir is not "" it should be the "relish/shared" 
  or "relish/rt/shared" of "relish/4production/shared" or "relish/rt/4production/shared" directory. 
  In that case, also serves source code from the shared directory tree.
*/
func ListenAndServeTLS(portNumber int, certFilePath string, keyFilePath string) {
    listenerConfigMutex.Lock()  
    if ! handlerAdded { 
       http.HandleFunc("/", handler)
       handlerAdded = true       
    }
    listenerConfigMutex.Unlock()    
    http.ListenAndServeTLS(fmt.Sprintf(":%d",portNumber), certFilePath, keyFilePath, nil)
}

/*
  Starts up relish shared source code serving on the specified port.
  sourceCodeShareDir should be the "relish/shared" 
  or "relish/rt/shared" of "relish/4production/shared" or "relish/rt/4production/shared" directory. 
  It is specifying the root directory of the shared source code tree.
*/
func ListenAndServeSourceCode(portNumber int, sourceCodeShareDir string) {
    http.ListenAndServe(fmt.Sprintf(":%d",portNumber), http.FileServer(http.Dir(sourceCodeShareDir)))  
}


func ListenAndServeExplorerApi(portNumber int) { 
   http.ListenAndServe(fmt.Sprintf(":%d",portNumber), http.HandlerFunc(explorerHandler))   
}



/*
   Return the value of the named attribute of the object.
   TODO: If the type of the RObject is one of the map collection types,
   use the attrName as a key instead of looking up the attribute value!!!!!!

   TODO This should also handle unary function calls on the object.
*/
func AttrVal(attrName string, obj RObject) (val RObject, err error) {
	// fmt.Println("Getting value of attrName",attrName)
    if obj.IsCollection() && (obj.(RCollection)).IsMap() {
        theMap := obj.(Map)
	    if theMap.KeyType() != StringType  {
           return nil,errors.New("template error: A map must have String key-type to serve as a template argument.") 		
		}   
		key := String(attrName) 
		var found bool
		val, found = theMap.Get(key) 
        if ! found {
           return nil,fmt.Errorf("template error: Map has no value for the key \"%s\".",attrName) 
        }
        return
    } 
	return RT.AttrValByName(responseProcessingThread, obj, attrName)
}

/*
Returns nil if the RObject is considered a zero/false/empty value in Relish, or returns the non-empty RObject. 
Should be appended at the end of a pipeline in an if or with.
{{if p | nonempty}}

{{with p | nonempty}} 
*/
func NonEmpty(obj RObject) RObject {
	if obj == nil || obj.IsZero() {
		return nil
	}
	return obj
}

/*
{{if eq .thing .otherThing}}

func Eq(obj RObject, obj2 RObject) bool {
	objs := []RObject{obj, obj2}

	return obj
}
*/

/*
{{range iterable .}}
*/
func Iterable(obj RObject) (iterable interface{}, err error) {
    if ! obj.IsCollection() {
        return nil,errors.New("template error: range action expects pipeline value to be a collection or map.") 
    }
    coll := obj.(RCollection)

    return coll.Iterable()
}

func HtmlPassThru(obj RObject) template.HTML {
	if obj == nil {
	   return template.HTML("")
	}
	s := obj.String()
	return template.HTML(s)
}

/*
Helper function for InvokeRelishMultiMethod below.
Convert an arbitrary go primitive literal value to the equivalent primitive RObject value,
or if the argument is already an RObject, just pass it through.
*/
func toRelishObject(a interface{}) RObject {
	var robj RObject
	var ok bool
	robj,ok = a.(RObject)
	if ! ok {
		switch a.(type) {
		case string:
			robj = String(a.(string))
		case int:
			robj = Int(a.(int))
		case float64:
			robj = Float(a.(float64))
		case bool:
			robj = Bool(a.(bool))
		case rune:
			robj = String(string(a.(rune)))			
		default: 
		    robj = String("DATA TYPE CONVERSION ERROR CONVERTING LITERAL TO RELISH VALUE!")
		}
	}
	return robj;
}

/*
*/
func InvokeRelishMultiMethod(methodName string, args ...interface{}) (val RObject, err error) {
	
   context := responseProcessingThread.EvalContext
   pkg := responseProcessingPackage

   var argObjects []RObject 

   for _,arg := range args {
	  relishArg := toRelishObject(arg)
	  argObjects = append(argObjects, relishArg)
   }


   // Which package to look for method by name from?
   // TODO We should also be able to explicitly specify a package in the prefix of the method name in the template

   multiMethod,multiMethodFound := pkg.MultiMethods[methodName] 
   if ! multiMethodFound {
	  return String(fmt.Sprintf("ERROR: relish built-in function '%s' not found!",methodName)),nil
   }

   // var multiMethod *RMultiMethod = nil // temporary

   obj := context.EvalMultiMethodCall(multiMethod, argObjects)
	
   return obj,nil
}

func goTemplate(relishTemplateText string) string {
	b := make([]byte,0,len(relishTemplateText) * 2 + 200) 
	var actionBuf [2048]byte
	
	buf := bytes.NewBuffer(b)
    copyStart := 0
    matches := re1.FindAllStringSubmatchIndex(relishTemplateText,-1)
    for _,match := range matches {
        //escapeStart := match[0]
        //escapeEnd := match[1]
        relishExprStart := match[2]
        relishExprEnd := match[3]	

        buf.WriteString(relishTemplateText[copyStart:relishExprStart])
	    pb := actionBuf[0:0]
        goExpr := goTemplateAction(pb, relishTemplateText[relishExprStart:relishExprEnd])
        buf.WriteString(goExpr)
        copyStart = relishExprEnd 
    }	
    buf.WriteString(relishTemplateText[copyStart:])
	return buf.String()
}

/*
Convert a relish template action to a Go template action.
*/
func goTemplateAction(b []byte, relishAction string) string {


	buf := bytes.NewBuffer(b)
    copyStart := 0

    matches := re2.FindAllStringSubmatchIndex(relishAction,-1)  // $ab2.vin.foo | .a1
    for i,match := range matches {
        exprStart := match[0]
        exprEnd := match[1]
        varStart := match[2]
        varEnd := match[3]	
        attrStart := match[4]
        attrEnd := match[5]

        // Use for debugging translation from relish to go template actions.
        //fmt.Println(relishAction[exprStart:exprEnd])

        buf.WriteString(relishAction[copyStart:exprStart])

        object := "."
        var goExpr string
        attr := relishAction[attrStart:attrEnd]
        if i == 0 {
	       if varStart >= 0 {
		      object = relishAction[varStart:varEnd]
		   }
		   goExpr = `get "` + attr + `" ` + object 
		} else {
			goExpr = ` | get "` + attr + `"`
		}      

        buf.WriteString(goExpr)
        copyStart = exprEnd 
    }	
    buf.WriteString(relishAction[copyStart:])

    if strings.Index(relishAction,"if ") == 0 || strings.Index(relishAction,"with ") == 0 { 
        buf.WriteString(" | nonempty")
    } else if strings.Index(relishAction,"range ") == 0 { 
        buf.WriteString(" | iterable")
    }      

    // Do the next round of substitution processing, this time fixing index expressions

    relishAction = buf.String()

    b = b[0:0]
    buf = bytes.NewBuffer(b)    

    copyStart = 0

    matches = re3.FindAllStringSubmatchIndex(relishAction,-1)  // index $ab
    for _,match := range matches {
        argEnd := match[3]  

        buf.WriteString(relishAction[copyStart:argEnd])
        if relishAction[argEnd-1] == '.' {
            buf.WriteString("Iterable") 
        } else { 
            buf.WriteString(".Iterable") 
        }  
        copyStart = argEnd
    }  
    buf.WriteString(relishAction[copyStart:])

    // Do the final round of substitution processing, this time wrapping relish function calls

// Was commented out from here down
    relishAction = buf.String()

    b = b[0:0]
    buf = bytes.NewBuffer(b)    

    copyStart = 0

    matches = re4.FindAllStringSubmatchIndex(relishAction,-1)  // funcName
    for _,match := range matches {
        funcNameStart := match[2]
        funcNameEnd := match[3]  
        funcName := relishAction[funcNameStart:funcNameEnd]
        switch funcName {
           case "if","with","else","end","range","get","nonempty","iterable","and","call","html","htm","index","js","len","not","or","print","printf","println","urlquery","template","true","false","nil":
              buf.WriteString(relishAction[copyStart:funcNameEnd])
           default:
	          buf.WriteString(relishAction[copyStart:funcNameStart])
	          buf.WriteString("fun ")
	          buf.WriteString(`"`)
	          buf.WriteString(relishAction[funcNameStart:funcNameEnd])
	          buf.WriteString(`"`)		
        }
        copyStart = funcNameEnd
    }  
    buf.WriteString(relishAction[copyStart:])
// was commented out down to here


	return buf.String()	
}
