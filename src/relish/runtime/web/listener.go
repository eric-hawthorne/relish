// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package implements a web application server for the relish language environment.

package web

import (
    . "relish/dbg"
    "fmt"
    "net/http"
	"html/template"
	"io/ioutil"
	"regexp"
	"bytes"
    "strings"
    "errors"
	. "relish/runtime/data"
	"relish/runtime/interp"
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
    "iterable": Iterable,	
    "fun": InvokeRelishMultiMethod,
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

func handler(w http.ResponseWriter, r *http.Request) {
   pathSegments := strings.Split(r.URL.Path, "/") 
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
	  panic("No web package has been defined in " + RT.RunningArtifact)
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
	     remainingPathSegments = remainingPathSegments[1:]     
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
   

   //args := []RObject{}  // Temporary - TODO get these from the request remaining path and keyword args or POST content	
   //resultObjects := interpreter.RunMultiMethod(handlerMethod, args) 	

   positionalArgStringValues := remainingPathSegments
   keywordArgStringValues, err := getKeywordArgs(r)
   if err != nil {
      fmt.Println(err)  
      fmt.Fprintln(w, err)
      return  
   }     

   resultObjects,err := interpreter.RunServiceMethod(handlerMethod, positionalArgStringValues, keywordArgStringValues)      
	 if err != nil {
      fmt.Println(err)  
      fmt.Fprintln(w, err)
      return  
   }   

   err = processResponse(w,r,handlerMethod.Name, resultObjects)
   if err != nil {
      fmt.Println(err)	
      fmt.Fprintln(w, err)
      return	
   }	
}

/* Returns the arguments from the combination of the URL query string (part after the ?) and the form values in the request body, 
   if the request was POST or PUT.
   TODO Does not currently do anything with the file part of multipart formdata, if any.
   // return value type is defined in net.url package: type Values map[string][]string
*/
func getKeywordArgs(r *http.Request) (args url.Values, err error) {
   err = r.ParseMultipartForm(10000000) // ok to call this even if it is not a mime/multipart request.
   if err != nil {
       return
   }
   args = r.Form
   return 
}


/*
Given an index i in a slice of the specified length, return the corresponding index
to use were the slice actually reversed. Used in processResponse because the
method call results slice is in reverse order. 
*/
func reverseIndex(i int, length int) int {
	return length - 1 - i             
	                                    
}

/*
Note should do considerably more checking of Content-Type (detected) and mimesubtype returnval,
to ensure they are consistent with the kind of processing directive. 
*/
func processResponse(w http.ResponseWriter, r *http.Request, methodName string, results []RObject) (err error) {
   nr := len(results)
   processingDirective := string(results[reverseIndex(0,nr)].(String))
   
   switch processingDirective {

	  case "XML":
       fmt.Println("XML response not implemented yet.")			
	   fmt.Fprintln(w, "XML response not implemented yet.")		
	  case "XML FILE":
       var filePath string
       if len(results) < 2 {
         err = fmt.Errorf("%s XML FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[reverseIndex(1,nr)].(String))       
        } else {
              err = fmt.Errorf("%s XML FILE response has too many return values. Should be filepath", methodName) 
              return               
        } 
        if ! strings.HasSuffix(filePath,".html") {
            err = fmt.Errorf("%s XML FILE response expecting a .xml file", methodName) 
            return   
        }
        filePath = makeAbsoluteFilePath(methodName, filePath)        
        http.ServeFile(w,r,filePath)		
	  case "HTML":
       fmt.Println("HTML response not implemented yet.")			
	   fmt.Fprintln(w, "HTML response not implemented yet.")		
	  case "HTML FILE":
       var filePath string
       if len(results) < 2 {
         err = fmt.Errorf("%s HTML FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[reverseIndex(1,nr)].(String))       
        } else {
              err = fmt.Errorf("%s HTML FILE response has too many return values. Should be filepath", methodName) 
              return               
        } 
        if ! strings.HasSuffix(filePath,".html") {
            err = fmt.Errorf("%s HTML FILE response expecting a .html file", methodName) 
            return   
        }        
        filePath = makeAbsoluteFilePath(methodName, filePath)        
        http.ServeFile(w,r,filePath)
	  case "JSON":
       fmt.Println("JSON response not implemented yet.")			
	   fmt.Fprintln(w, "JSON response not implemented yet.")		
	  case "JSON FILE":
       fmt.Println("JSON FILE response not implemented yet.")			
	   fmt.Fprintln(w, "JSON FILE response not implemented yet.")		
	  case "IMAGE":
       fmt.Println("IMAGE response not implemented yet.")			
	   fmt.Fprintln(w, "IMAGE response not implemented yet.")		
      case "IMAGE FILE":  // [mime subtype] filePath
       var filePath string
       var mimeSubtype string       
       var mimeType string
       if len(results) < 2 {
         err = fmt.Errorf("%s IMAGE FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[reverseIndex(1,nr)].(String))    

        } else if len(results) == 3 {
          mimeSubtype = string(results[reverseIndex(1,nr)].(String))
          mimeType = "image/" + mimeSubtype
          filePath = string(results[reverseIndex(2,nr)].(String))    
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
       fmt.Println("VIDEO response not implemented yet.")			
	   fmt.Fprintln(w, "VIDEO response not implemented yet.")		
	  case "VIDEO FILE":
       var filePath string
       var mimeSubtype string       
       var mimeType string
       if len(results) < 2 {
         err = fmt.Errorf("%s VIDEO FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[reverseIndex(1,nr)].(String))    

        } else if len(results) == 3 {
          mimeSubtype = string(results[reverseIndex(1,nr)].(String))
          mimeType = "video/" + mimeSubtype
          filePath = string(results[reverseIndex(2,nr)].(String))    
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
       fmt.Println("MEDIA response not implemented yet.")			
	   fmt.Fprintln(w, "MEDIA response not implemented yet.")		
	  case "MEDIA FILE":
       var filePath string   
       var mimeType string
       if len(results) < 2 {
         err = fmt.Errorf("%s MEDIA FILE response requires a filepath", methodName) 
         return       
       } else if len(results) == 2 {            
          filePath = string(results[reverseIndex(1,nr)].(String))    

        } else if len(results) == 3 {
          mimeType = string(results[reverseIndex(1,nr)].(String))
          filePath = string(results[reverseIndex(2,nr)].(String))    
        } else {
              err = fmt.Errorf("%s IMEDIA FILE response has too many return values. Should be filepath or mimetype filepath", methodName) 
              return               
        } 
        if mimeType != "" {
           w.Header().Set("Content-Type", mimeType)
        }
        filePath = makeAbsoluteFilePath(methodName, filePath)
        http.ServeFile(w,r,filePath)		
	  case "REDIRECT": // url [301,302,303,or 307]
       var urlStr string
       var code int
       if len(results) < 2 {
         err = fmt.Errorf("%s redirect requires a URL or path", methodName) 
         return       
       } else if len(results) == 2 {
          code = 303   
          // TODO Should handle a builtin URL type as well as String               
          urlStr = string(results[reverseIndex(1,nr)].(String))    

        } else if len(results) == 3 {
          code = int(results[reverseIndex(1,nr)].(Int))           
          // TODO Should handle a builtin URL type as well as String
          urlStr = string(results[reverseIndex(2,nr)].(String))    
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
          code = int(results[reverseIndex(1,nr)].(Int))         
       } else if len(results) == 3 {
          code = int(results[reverseIndex(1,nr)].(Int))           
          message = string(results[reverseIndex(2,nr)].(String))    

       } else {
              err = fmt.Errorf("%s HTTP ERROR response has too many return values. Should be e.g. 404 [message]", methodName) 
              return               
       } 
       http.Error(w, message, code)  
			
	  case "TEMPLATE": // An inline template as a string	
		if len(results) < 3 { 
		      err = fmt.Errorf("%s (with a templated response) requires return values which are a template string and an object to pass to the template", methodName)            
		       return
		}
		if len(results) > 3 { 
		     err = fmt.Errorf("%s (with a templated response) has unexpected 4th return value", methodName)	
		     return
	    }	
	    relishTemplateText := string(results[reverseIndex(1,nr)].(String))
	    obj := results[reverseIndex(2,nr)]
	    err = processTemplateResponse(w, r, methodName, "", relishTemplateText, obj)
		
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
	    if len(results) > 2 { 
	        err = fmt.Errorf("%s (with a templated response) has unexpected 3rd return value", methodName)	
          return
      }

          templateFilePath = makeAbsoluteFilePath(methodName, templateFilePath) 
	      obj := results[reverseIndex(1,nr)]
          err = processTemplateFileResponse(w, r, methodName, templateFilePath, obj) 
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
or is "absloute" with respect to the web package source directory (the web content root)
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

func processTemplateFileResponse(w http.ResponseWriter, r *http.Request, methodName string, templateFilePath string, obj RObject) (err error) {
   bytes,err := ioutil.ReadFile(templateFilePath)
    if err != nil {
       fmt.Println(err)		
       fmt.Fprintln(w, err)
       return	
    }
    relishTemplateText := string(bytes)	
    err = processTemplateResponse(w, r, methodName, templateFilePath, relishTemplateText, obj)
    return
}	

/*
templateFilePath may be the empty string (indicating an inline template). If not it is used to make error messages more specific.
*/
func processTemplateResponse(w http.ResponseWriter, r *http.Request, methodName string, templateFilePath string, relishTemplateText string, obj RObject) (err error) {

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

func ListenAndServe(portNumber int) {
    http.HandleFunc("/", handler)
    http.ListenAndServe(fmt.Sprintf(":%d",portNumber), nil)
}


/*
   Return the value of the named attribute of the object.
   TODO: If the type of the RObject is one of the map collection types,
   use the attrName as a key instead of looking up the attribute value!!!!!!

   TODO This should also handle unary function calls on the object.
*/
func AttrVal(attrName string, obj RObject) (val RObject, err error) {
    /*
    if obj.IsCollection() && (obj.(RCollection)).IsMap() {
        theMap := obj.(Map)
        ...

    } else {
    */
	return RT.AttrValByName(obj, attrName)
}

/*
Returns nil if the RObject is considered a zero/false/empty value in Relish, or returns the non-empty RObject. 
Should be appended at the end of a pipeline in an if or with.
{{if p | nonempty}}

{{with p | nonempty}} 
*/
func NonEmpty(obj RObject) RObject {
	if obj.IsZero() {
		return nil
	}
	return obj
}

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

/*
*/
func InvokeRelishMultiMethod(funcName string, args ...interface{}) (val RObject, err error) {
   return String("RELISH FUNC CALLS NOT IMPLEMENTED YET!"),nil
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

        fmt.Println(relishAction[exprStart:exprEnd])

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
/*
    relishAction = buf.String()

    b = b[0:0]
    buf = bytes.NewBuffer(b)    

    copyStart = 0

    matches = re3.FindAllStringSubmatchIndex(relishAction,-1)  // funcName
    for _,match := range matches {
        funcNameStart := match[2]
        funcNameEnd := match[3]  
        funcName := relishAction[funcNameStart:funcNameEnd]
        switch funcName {
           case "get","nonempty","iterable","and","call","html","index","js","len","not","or","print","printf","println","urlquery":
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
*/


	return buf.String()	
}