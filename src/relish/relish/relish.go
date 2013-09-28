// Copyright 2012-2013 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

/*
relish interpreter main program.

Note: This is not expected to work on Windows yet due to path separator differences.

Current usage: 

cd $RELISH_HOME/artifacts/some.codeorigin.com2007/some/artifact/path/v0001
relish path/to/package

or

cd $RELISH_HOME/artifacts/some.codeorigin.com2007/some/artifact/path/v0001/pkg/path/to/package
relish

or

relish some.codeorigin.com2007/some/artifact/path 2 path/to/package

or 

relish some.codeorigin.com2007/some/artifact/path path/to/package

which chooses the current version as specified in some/artifact/path/metadata.txt
- note that this will first check if the artifact has been installed locally, and will use the
  local copy of the artifact's metadata.txt to determine the current version. This could be
  out of date. TODO We need another command or command option to force an internet search for the 
  true authoritative current version of the artifact. 


Command line options:

-log 1|2    The logging level: 1 (some debugging info), 2 more. Very minimal logging of key runtime events if not supplied.

-web <port#>  The http listening port - if not supplied, does not listen for http requests



-db <dbname>  The database name. A SQLITE database file called <dbname>.db will be created/used in artifact data directory.
              Defaults to db1   i.e. db1.db	

-cpuprofile <filepath>.prof  Write cpu profile to file. Then use go tool pprof /opt/devel/relish/bin/relish somerun.prof 


-publish origin/artifact [version#]    Copies to shared/relish/artifacts directory tree   - served if sharing

-makecurrent origin/artifact version#

-share <port#> Serve source code (contents of the shared directory) on the specified port. Port should be 80
               or failing that 8421, or, if behind apache2 modproxy, any other port is fine but apache2 should
               present it as port 80 or port 8421. It is ok for the share port to be the same as the web port.




*/
package main

import (
        "fmt"
        "flag"
        "strings"
        "os"
//		"relish/compiler/token"
//		"relish/compiler/ast"	
//		"relish/compiler/parser"
		"relish/compiler/generator"
		"relish/runtime/native_methods/builtin"
		"relish/runtime/web"	  
		"relish/dbg"
		"relish/global_loader"
		"relish/global_publisher"		
		"regexp"
//		"strconv"
		"runtime/pprof"
)

var reVersionAtEnd *regexp.Regexp = regexp.MustCompile("/v([0-9]+\\.[0-9]+\\.[0-9]+)$")
var reVersionedPackage *regexp.Regexp = regexp.MustCompile("/v([0-9]+\\.[0-9]+\\.[0-9]+)/pkg/")


func main() {
    var loggingLevel int
    var webListeningPort int
    var shareListeningPort int  // port on which source code will be shared by http    
    var sharedCodeOnly bool  // do not use local artifacts - only those in shared directory.
    var runningArtifactMustBeFromShared bool
    var dbName string 
    var cpuprofile string
    var publish bool

    //var fset = token.NewFileSet()
	flag.IntVar(&loggingLevel, "log", 0, "The logging level: 0 is least verbose, 2 most")	
	flag.IntVar(&webListeningPort, "web", 0, "The http listening port - if not supplied, does not listen for http requests")	
	flag.StringVar(&dbName, "db", "db1", "The database name. A SQLITE database file called <name>.db will be created/used in artifact data directory")			

	flag.BoolVar(&sharedCodeOnly, "shared", false, "Use shared version of all artifacts - ignore local/dev copy of artifacts")		
 
    flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")

	flag.IntVar(&shareListeningPort, "share", 0, "The code sharing http listening port - if not supplied, does not listen for source code sharing http requests")	    

    flag.BoolVar(&publish, "publish", false, "artifactpath version - copy specified version of artifact to shared/relish/artifacts")



    flag.Parse()


    pathParts := flag.Args() // full path to package, or originAndArtifact and path to package 
                             // (or originAndArtifact and version number if -publish)


    if cpuprofile != "" {
        f, err := os.Create(cpuprofile)
        if err != nil {
		     fmt.Println(err)
		     return 
		}
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

   	dbg.InitLogging(int32(loggingLevel))
	//relish.InitRuntime("relish.db")
    
    if ! publish {
    	builtin.InitBuiltinFunctions()	
	}

	var g *generator.Generator
	
	var relishRoot string  // This actually has to be the root of the runtime environment
	                       // i.e. /opt/relish if this is a binary distribution,
	                       // or /opt/relish/rt if this is a source distribution
	
    workingDirectory, err := os.Getwd()
    if err != nil {
	   fmt.Printf("Cannot determine working directory: %s\n",err)	
    }


    var originAndArtifact string
    var version string
    var packagePath string

    relishIndexInWd := strings.Index(workingDirectory,"/rt/artifacts") 
    if relishIndexInWd > -1 {
	   relishRoot = workingDirectory[:relishIndexInWd] + "/rt" 
    } else {
       relishIndexInWd := strings.Index(workingDirectory,"/rt/shared/relish/artifacts") 
       if relishIndexInWd > -1 {
	      relishRoot = workingDirectory[:relishIndexInWd] + "/rt"
	      runningArtifactMustBeFromShared = true
	   } else {
          relishIndexInWd := strings.Index(workingDirectory,"/relish/shared/relish/artifacts") 
          if relishIndexInWd > -1 {
	         relishRoot = workingDirectory[:relishIndexInWd] + "/relish"
 	         runningArtifactMustBeFromShared = true 
 	      } else {
             relishIndexInWd := strings.Index(workingDirectory,"/relish/artifacts") 
             if relishIndexInWd > -1 {
	            relishRoot = workingDirectory[:relishIndexInWd] + "/relish" 
	         }	      	
 	      }	
	   }
    }    

	if relishRoot != "" && ! publish {   
       match := reVersionedPackage.FindStringSubmatchIndex(workingDirectory)
       if match != nil {
	      version = workingDirectory[match[2]:match[3]]
          
	      originAndArtifact = workingDirectory[relishIndexInWd+18:match[0]]
	      packagePath = workingDirectory[match[3]+1:]	
       } else {
	       match := reVersionAtEnd.FindStringSubmatch(workingDirectory)
	       if match != nil {
		      version = match[1]	
	          originAndArtifact = workingDirectory[relishIndexInWd+18:len(workingDirectory)-len(version)-2]		
		   }
       }
    }
	
	if relishRoot == "" { // did not find it in the working directory
		
	    relishRoot = os.Getenv("RELISH_HOME")  // See if the environment variable is defined
	    if relishRoot != "" {
			slashPos := strings.LastIndex(relishRoot,"/")
			if slashPos != -1 {
			    if relishRoot[slashPos+1:] != "relish" {
				   fmt.Printf("RELISH_HOME directory must be named relish\n")
				   return			
			    }
			} else {
			   slashPos := strings.LastIndex(relishRoot,`\`)
			   if slashPos != -1 {
			      if relishRoot[slashPos+1:] != "relish" {
				     fmt.Printf("RELISH_HOME directory must be named relish\n")
				     return			
			      }			
			   } else {
			      fmt.Printf("RELISH_HOME directory path is ill-formed\n")				
			   }
			}
	        _,err := os.Stat(relishRoot)	
	        if err != nil {
	            if ! os.IsNotExist(err) {
				   fmt.Printf("Can't stat RELISH_HOME directory '%s' : %v\n", relishRoot, err)
		  	       return		       
		        }
		        fmt.Printf("RELISH_HOME directory '%s' does not exist\n")
		        return	
	         }	
				
	    } else { // RELISH_HOME not defined. Try standard locations.
			var relishRoots []string
			if os.IsPathSeparator('/') {
				relishRoots = []string{"/opt/relish","/usr/local/relish","~/relish","~/devel/relish","/opt/devel/relish","Library/relish"}
			} else { // Windows? This part is untested	
				relishRoots = []string{`C:\relish`}	
			}
			for _,potentialInstallLocation := range relishRoots {
		        _,err := os.Stat(potentialInstallLocation)	
		        if err != nil {
			        if ! os.IsNotExist(err) {
						fmt.Printf("Can't stat potential relish install location '%s': %v\n", potentialInstallLocation, err)
						return		       
			        }
		        } else {
		            relishRoot = potentialInstallLocation
		            break
		        }	
	        }				
	   }

       if relishRoot != "" {
		   // determine if is source distribution or binary by searching for "/rt/"
		   // if source, move relishRoot to the /rt subdir	   

		   _,err := os.Stat(relishRoot + "/rt")	
		   if err == nil {
		   	   relishRoot += "/rt"
		   } else {
		      if ! os.IsNotExist(err) {
					   fmt.Printf("Can't stat RELISH_HOME subdirectory '%s/rt' : %v\n", relishRoot, err)
			  	       return		       
			  }
		   } 	
	   }
	}
	if relishRoot == "" {
		fmt.Printf("RELISH_HOME environment variable is not defined and relish is not installed in a default location and working directory is not within a relish directory tree.\n")
		return		
	}
	

    if publish {
       if len(pathParts) < 2 {
          fmt.Println("Usage (example): relish -publish somorigin.com2013/artifact_name 1.0.23")
          return
       }
	   originAndArtifact = pathParts[0]
	   version = pathParts[1]

       err = global_publisher.PublishSourceCode(relishRoot, originAndArtifact, version)
	   if err != nil {
		  fmt.Println(err)
       }
       return
    }


	var loader = global_loader.NewLoader(relishRoot, sharedCodeOnly, dbName + ".db")
	
	
	
	if originAndArtifact == "" {
		
       if len(pathParts) == 3 {  // originAndArtifact version npackagePath
	
		    originAndArtifact = pathParts[0]
		    version = pathParts[1]
		    packagePath = pathParts[2]		
	
       } else if len(pathParts) == 2 {  // originAndArtifact packagePath	
		
		    originAndArtifact = pathParts[0]
		    packagePath = pathParts[1]
		
	   } else {
	      fmt.Println("Usage: relish [-log n] [-web 80] originAndArtifact [version] path/to/package")
	      return
	   }		
	

		
	} else if packagePath == "" {
		
       if len(pathParts) != 1 {
  	      fmt.Println("Usage (when in an artifact version directory): relish [-log n] [-web 80] path/to/package")
	      return
       }	
       packagePath = pathParts[0]
	
	} else {
       if len(pathParts) != 0 {
	         fmt.Println("Usage (when in a package directory): relish [-log n] [-web 80]")
	         return
       }		
    }


	if strings.HasSuffix(originAndArtifact,"/") {  // Strip trailing / if present
	   originAndArtifact = originAndArtifact[:len(originAndArtifact)-1]	
	}
    if strings.HasSuffix(packagePath,"/") {  // Strip trailing / if present
       packagePath = packagePath[:len(packagePath)-1]	
    }

    fullPackagePath := fmt.Sprintf("%s/v%s/pkg/%s",originAndArtifact,version, packagePath)
    fullUnversionedPackagePath := fmt.Sprintf("%s/pkg/%s",originAndArtifact, packagePath)
    
    g, err = loader.LoadPackage(originAndArtifact, version, packagePath, runningArtifactMustBeFromShared)

    if err != nil {
	    if version == "" {
		   fmt.Printf("Error loading package %s from current version of %s:  %v\n", packagePath, originAndArtifact, err)		
		} else {
		   fmt.Printf("Error loading package %s:  %v\n",fullPackagePath, err)
	    }
		return	
    }


    g.Interp.SetRunningArtifact(originAndArtifact) 

	sourceCodeShareDir := ""
	if shareListeningPort != 0 {
	    // sourceCodeShareDir hould be the "relish/shared" 
	    // or "relish/rt/shared" of "relish/4production/shared" or "relish/rt/4production/shared" directory.		
		sourceCodeShareDir = relishRoot + "/shared"
	}
	
	if webListeningPort != 0 {
	   if webListeningPort < 1024 && webListeningPort != 80 && webListeningPort != 443 {
			fmt.Println("Error: The web listening port must be 80, 443, or > 1023")
			return		
	   }
	
       if shareListeningPort != webListeningPort && shareListeningPort != 0 && shareListeningPort < 1024 && shareListeningPort != 80 && shareListeningPort != 443 {
	  	  fmt.Println("Error: The source-code sharing port must be 80, 443, or > 1023 (8421 is the standard if using a high port)")
		  return		
       }		
	
       err = loader.LoadWebPackages(originAndArtifact, version, runningArtifactMustBeFromShared)	
	   if err != nil {
		    if version == "" {
			   fmt.Printf("Error loading web packages from current version of %s:  %v\n", originAndArtifact, err)		
			} else {
			   fmt.Printf("Error loading web packages from version %s of %s:  %v\n", version, originAndArtifact, err)
		    }
			return	
	   }
	
	   go g.Interp.RunMain(fullUnversionedPackagePath)
	   web.SetWebPackageSrcDirPath(loader.PackageSrcDirPath(originAndArtifact + "/pkg/web"))
	  
	   if shareListeningPort == webListeningPort {
	      web.ListenAndServe(webListeningPort, sourceCodeShareDir)
	   } else {
          if shareListeningPort != 0 {
	         web.ListenAndServeSourceCode(shareListeningPort, sourceCodeShareDir) 
	      }			
	      web.ListenAndServe(webListeningPort, "")	
	   }
	
	} else if shareListeningPort != 0 {
	   if shareListeningPort < 1024 && shareListeningPort != 80 && shareListeningPort != 443 {
			fmt.Println("Error: The source-code sharing port must be 80, 443, or > 1023 (8421 is the standard if using a high port)")
			return		
	   }		
	
	   go g.Interp.RunMain(fullUnversionedPackagePath)	

       web.ListenAndServeSourceCode(shareListeningPort, sourceCodeShareDir) 	

	
	} else {
	   g.Interp.RunMain(fullUnversionedPackagePath)
	}
}