// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the expression and management of runtime data (objects and values) 
// in the relish language.

package data

/*
   rpackage.go - A package is a universally unique hierarchically named namespace for
                 types, methods (and global variables?)
                 Packages will also be the name-exporting boundaries.
*/

import (
	"fmt" 
    "strings"
    "sort"
)

///////////////////////////////////////////////////////////////////////////
////////// PACKAGES
///////////////////////////////////////////////////////////////////////////

/*
A package in the relish language. 
A package is the namespace and export protection domain for types, methods etc.

Each package has its own multimethod dispatch table/cache, containing only those 
method-implementations of each multimethod which are visible from the package,
given the package's import dependencies.
*/
type RPackage struct {
	runit
	Name string  // Full origin, artifact, package
	Path string  // The Name with "/" appended
	
	// include some bytes of uuidabbrev to include in ShortName()
	ShortName string // unique in the runtime and db
	
	MultiMethods  map[string]*RMultiMethod // map from method name to RMultiMethod object.
	
	ClosureMethods map[string]*RMethod  // map from method name to RMethod object
	
	Dependencies map[string]*RPackage   // Packages that this package is dependent on		

    IsTrait bool  // Package has only abstract, bodiless methods and final (non-overrideable) methods.

    IsPureTrait bool  // Package is a trait and only refers to primitive types 
                      // or types defined in a pure trait package.
                      // TODO: Compute this at end of package generation.
}

func (p *RPackage) Origin() string {
	return p.Name[:strings.Index(p.Name,"/")]
}

func (p *RPackage) Artifact() string {
	return p.Name[strings.Index(p.Name,"/")+1:strings.Index(p.Name,"/pkg/")]	
}

func (p *RPackage) OriginAndArtifact() string {
	return p.Name[:strings.Index(p.Name,"/pkg/")]	
}

func (p *RPackage) LocalPackagePath() string {
	return p.Name[strings.Index(p.Name,"/pkg/")+5:]	
}


func Origin(packageFullName string) string {
	return packageFullName[:strings.Index(packageFullName,"/")]
}

func Artifact(packageFullName string) string {
	return packageFullName[strings.Index(packageFullName,"/")+1:strings.Index(packageFullName,"/pkg/")]	
}

func OriginAndArtifact(packageFullName string) string {
	return packageFullName[:strings.Index(packageFullName,"/pkg/")]	
}

func LocalPackagePath(packageFullName string) string {
	return packageFullName[strings.Index(packageFullName,"/pkg/")+5:]	
}




/*
Debugging function. Prints the names of multimethods visible from the package.
*/
func (p *RPackage) ListMethods() {
	fmt.Println("------------")	
	fmt.Println("Multimethods visible in package", p.Name)
	fmt.Println("------------")
	var methodNames []string
	for methodName := range p.MultiMethods {
		methodNames = append(methodNames, methodName)
	}
	sort.Strings(methodNames)
	for _,methodName := range methodNames {
		fmt.Println(methodName)
	}
}


/*
Debugging function. Prints detailed debug info for a given multimethod as visible in the package.
*/
func (p *RPackage) ListMethod(name string) {
	fmt.Println("------------")	
	fmt.Println("Multimethod", name, "visible in package", p.Name)
	fmt.Println("------------")
	mm := p.MultiMethods[name]
    fmt.Println(mm.Debug(nil))
}


/*
   Create an in-memory representation of a relish code package, for use during interpreter runtime.

   Example package path (globally unique full name of a relish code package):

   alphaworks.ibm.com1995/relish/editor/pkg/frontend/parser/stringutils/regexp

   consisting of:
   orgDomain - e.g. ibm.com - may be a subdomain e.g. compsci.berkeley.edu or research.ca.ibm.com
   orgFoundedYear - first full calendar year in which the organization owns the domain name
   artifactPath - the name of the overall project, library, or application  
   "/pkg/" separator between artifact path and package path
   packageparentdir(s) - super-directories of the package path, within the artifact, may be none
   name - the local name of the package. This is how it is known in a program that has imported it.

   Each package will have its globally unique (no really, worldwide and perpetually unique) full name, 
   its local name (e.g. regexp), and
   its short name which is a short name guaranteed to be unique in the local database e.g. P3a_regexp

   Obscure Note: As soon as we persist packages in the local db, we cannot recreate them from source into memory, 
   because they will get a different uuid and may be defined in different order giving them a different
   shortName, and package shortnames are part of the name of type tables and relation tables. 
   Therefore, on initialization of a relish application, RPackage objects are resurrected out of the
   database into memory, before relish code is loaded.
*/
func (rt *RuntimeEnv) CreatePackage(path string, isStandardLibPackage bool) *RPackage {

	typ, typFound := rt.Types["shared.relish.pl2012/core/pkg/relish/lang/Package"]
	var err error
	if !typFound {
		// Create the reflection type for packages.
		// Note: The bad thing here is we're not giving the type its package.
		// TODO Make an actual package here for the type to be in?
		typ, err = rt.CreateType("shared.relish.pl2012/core/pkg/relish/lang/Package", "lang/Package",[]string{})
		if err != nil {
			panic(fmt.Sprintf("Unable to define type 'relish.pl2012/core/pkg/relish/lang/Package' : %s", err))
		}
	}


	pkg := &RPackage{runit: runit{robject: robject{rtype: typ}},
		Name:  path,
		Path: path + "/",
        MultiMethods: make(map[string]*RMultiMethod),	
        ClosureMethods: make(map[string]*RMethod),	
        Dependencies: make(map[string]*RPackage),
        IsTrait: strings.HasSuffix(path, "_trait"),
	}
	pkg.runit.robject.this = pkg

	if _, found := rt.Packages[pkg.Name]; found {
		panic(fmt.Sprintf("Attempt to redefine package '%s'", pkg.Name))
	}

    var shortName string
    if isStandardLibPackage {    
	    shortName = pkg.Name[11:]
    } else {
	    shortName = rt.PkgNameToShortName[pkg.Name]
	}
    if shortName != "" {
       	pkg.ShortName = shortName
    } else {
		// Create locally unique short name of package
		uuidAbbrev, err := pkg.EnsureUUIDabbrev()
		if err != nil {
			panic(fmt.Sprintf("Unable to create package uuid: %v", err))
		}
		
		simpleShortName := pkg.Name[strings.LastIndex(pkg.Name,"/")+1:]		
		candidateShortName := simpleShortName

	
		if _, found := rt.PkgShortNameToName[candidateShortName]; found {		
//		if _, found := rt.Pkgs[candidateShortName]; found {
			for i := 2; i <= len(uuidAbbrev); i += 2 {
				candidateShortName = "P" + uuidAbbrev[0:i] + "_" + simpleShortName
				_, found = rt.PkgShortNameToName[candidateShortName]
				if !found {
					break
				}
			}
			if found {
				panic(fmt.Sprintf("Unable to make a locally unique short name for package '%s'", pkg.Name))
			}
		}
		pkg.ShortName = candidateShortName
	
	    rt.DBT().RecordPackageName(pkg.Name, pkg.ShortName)
    	if err != nil {
		   panic(fmt.Sprintf("Unable to record package name in db: %v", err))
	    }	
    }
    

    // Note that this package name is not a legal package name. Maybe that's ok.
    if pkg.Name == "relish.pl2012/core/inbuilt" {
    	rt.InbuiltFunctionsPackage = pkg
    } else if ! isStandardLibPackage {
       // Copy multimethod map from inbuilt functions package
       inbuiltPkg := rt.Packages["relish.pl2012/core/inbuilt"]	
       for multiMethodName, multiMethod := range inbuiltPkg.MultiMethods {
          pkg.MultiMethods[multiMethodName] = multiMethod
       }    	
    }


	rt.Packages[pkg.Name] = pkg
	rt.Pkgs[pkg.ShortName] = pkg
	return pkg
}
