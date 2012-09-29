// Copyright 2012 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// "code generator" for relish.
// Currently, tours the file ast (should be package ast) and creates types and methods.

package generator

import (
	"fmt"	
	"strings"
	"relish/compiler/ast"
	"relish/compiler/token"
	"relish/runtime/data"
	"relish/runtime/interp"
	"relish"
	. "relish/defs"
)


// The parser structure holds the parser's internal state.
type Generator struct {
	file *ast.File
	Interp *interp.Interpreter
	th *interp.Thread
	packageName string // Full name of origin, artifact, and package
	packagePath string // Full name of origin, artifact, and package, ending in a /
	pkg *data.RPackage // the current package which this file is being generated into
	fileNameRoot string // filename of the code file in the package, without its .rel or .rlc suffix
}

func NewGenerator(file *ast.File, fileNameRoot string) *Generator {
	interpreter := interp.NewInterpreter(data.RT)
	thread := interpreter.NewThread()
	return &Generator{file,interpreter,thread,file.Name.Name,file.Name.Name + "/",nil,fileNameRoot}
}

/*
Given a Relish intermediate-code file node tree, create objects in the runtime for data-types, methods, and constants defined in the file.
Assumes imported packages have already been loaded; thus the objects defined in files of the imported packages have been generated.
*/
func (g *Generator) GenerateCode() {	
   g.GeneratePackage()
   g.GenerateTypes()
   g.GenerateMethods()		
   g.GenerateConstants()
}

/*
Checks to see if the package which this source code file says its in already
exists in the runtime. If not, creates it. Sets the pkg variable of the Generator
to refer to the package.
*/
func (g *Generator) GeneratePackage() {
    g.pkg = data.RT.Packages[g.packageName]
    if g.pkg == nil {
	   g.pkg = data.RT.CreatePackage(g.packageName)  // This should init MMMap from core/builtin package   
	}
	g.updatePackageDependenciesAndMultiMethodMap()
}

func (g *Generator) updatePackageDependenciesAndMultiMethodMap() {
	imports := g.file.RelishImports  // package specifications
	for _,importedPackageSpec := range imports {		
		dependencyPackageName := importedPackageSpec.OriginAndArtifactName + "/pkg/" + importedPackageSpec.PackageName
		dependencyPackage, dependencyAlreadyProcessed := g.pkg.Dependencies[dependencyPackageName]
        if ! dependencyAlreadyProcessed {
		   dependencyPackage = data.RT.Packages[dependencyPackageName]
		   g.pkg.Dependencies[dependencyPackageName] = dependencyPackage

		   g.updatePackageMultiMethodMap(dependencyPackage)
	    }

	}
}


func (g *Generator) updatePackageMultiMethodMap(dependencyPackage *data.RPackage) {
   for multiMethodName,multiMethod := range dependencyPackage.MultiMethods {
   	   myMultiMethod := g.pkg.MultiMethods[multiMethodName]
   	   if myMultiMethod == nil {
   	   	   g.pkg.MultiMethods[multiMethodName] = multiMethod  // use the mm from dependency package
   	   } else if myMultiMethod != multiMethod {
	      	
   	   	   // Merge them if possible, else complain!! panic!

           if myMultiMethod.NumReturnArgs != multiMethod.NumReturnArgs {
	          panic(fmt.Sprintf("Package %s defines %s to return %d values so can't be imported directly or indirectly into package %s which defines %s to return %d values.",dependencyPackage.Name,multiMethod.Name,multiMethod.NumReturnArgs,g.pkg.Name,myMultiMethod.Name,myMultiMethod.NumReturnArgs))
	
	          // TODO Shit! I'm inheriting methods indirectly!!! Not allowed !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	
	          // Will have to actually create new multimethods for EACH new package right up front, 
	          // putting in ONLY those methods which are defined in the DIRECT dependency packages.
           }

   	   	   // To merge them, I have to know if my multiMethod really belongs to me or if borrowing
   	   	   // it so I have to copy it.

           if myMultiMethod.Pkg != g.pkg {
	          myMultiMethod = myMultiMethod.Clone(g.pkg)
	          g.pkg.MultiMethods[multiMethodName] = multiMethod	
           }

           myMultiMethod.MergeInNewMethodsFrom(multiMethod)

   	   }
   	   // else I've already got the multimethod - everything's cool
   }
}






/*
If the type name is unqualified (has no package path), prefixes it with the current package path.
A little bit inefficient.
*/
func (g *Generator) qualifyTypeName(typeName string) string {
   if strings.LastIndex(typeName,"/") == -1 && ! BuiltinTypeName[typeName] {
       return g.packagePath + typeName  	
   }	
   return typeName
}

/*
Processes the TypeDecls list of a ast.File object (which has been created by the parser.)
Generates the runtime environment's objects for datatypes and attributes, and also ensures that db tables exist for these.
TODO prefix the g.packagePath onto the name of the type.!!!!!!!!!!!!!!
*/
func (g *Generator) GenerateTypes() {
	for _,typeDeclaration := range g.file.TypeDecls {
		
	   typeSpec := typeDeclaration.Spec
	   typeName := g.packagePath + typeSpec.Name.Name
	
	   var parentTypeNames []string
	
	   for _,parentTypeSpec := range typeSpec.SuperTypes {
		  parentTypeNames = append(parentTypeNames, g.qualifyTypeName(parentTypeSpec.Name.Name))
	   } 
		
	   // Get the type name and the supertype names	
	   theNewType, err := data.RT.CreateType(typeName, parentTypeNames)
       if err != nil {
          panic(err)
       }	

	   for _,attrDecl := range typeDeclaration.Attributes {
		  var minCard int32 = 1
		  var maxCard int32 = 1
		  
          attributeName := attrDecl.Name.Name
          multiValuedAttribute := (attrDecl.Arity != nil)
          if multiValuedAttribute {
	         minCard = int32(attrDecl.Arity.MinCard)
	         maxCard = int32(attrDecl.Arity.MaxCard)  // -1 means N
          }
          
          var collectionType string 

          var orderFuncOrAttrName string = ""
          var isAscending bool 

          if attrDecl.Type.CollectionSpec != nil {
              switch attrDecl.Type.CollectionSpec.Kind {
	             case token.SET:
			        if attrDecl.Type.CollectionSpec.IsSorting {
			           collectionType = "sortedset"
                       orderFuncOrAttrName = attrDecl.Type.CollectionSpec.OrderFunc	
                       isAscending = attrDecl.Type.CollectionSpec.IsAscending		
		            } else {
			           collectionType = "set"			
		            }		
		         case token.LIST:
			        if attrDecl.Type.CollectionSpec.IsSorting {
			           collectionType = "sortedlist"
                       orderFuncOrAttrName = attrDecl.Type.CollectionSpec.OrderFunc	
                       isAscending = attrDecl.Type.CollectionSpec.IsAscending			
		            } else {
			           collectionType = "list"			
		            }
			     case token.MAP:
			        if attrDecl.Type.CollectionSpec.IsSorting {
			           collectionType = "sortedmap"
                       orderFuncOrAttrName = attrDecl.Type.CollectionSpec.OrderFunc	
                       isAscending = attrDecl.Type.CollectionSpec.IsAscending			
		            } else {
			           collectionType = "map"			
		            }				
	           }	
          }


/*	type CollectionTypeSpec struct {
	   Kind token.Token
	   LDelim token.Pos
	   RDelim token.Pos
	   IsSorting bool
	   IsAscending bool
	   OrderFunc string
	}
*/	
 
          attributeTypeName := g.qualifyTypeName(attrDecl.Type.Name.Name)

          
/*"vector"
RelEnd
   ...
   CollectionType string // "list", "sortedlist","set", "sortedset", "map", "stringmap", "sortedmap","sortedstringmap",""
   OrderAttrName string   // which primitive attribute of other is it ordered by when retrieving? "" if none

   OrderMethod *RMultiMethod
*/


	      _,err = data.RT.CreateAttribute(typeName,
									 	 attributeTypeName,
										 attributeName,
										 minCard,
										 maxCard,   // Is the -1 meaning N respected in here???? TODO
										 collectionType,
				                         orderFuncOrAttrName,
				                         isAscending,
										 false,
										 g.Interp.Dispatcher())
		   if err != nil {
		      panic(err)
		   }

        }

        // Now ensure the persistence data model is created for the type.

        relish.EnsureDatabase()

		err = data.RT.DB().EnsureTypeTable(theNewType) 
		if err != nil {
		      panic(err)
		}
		
		// ... and for the type's attributes and relations

		err = data.RT.DB().EnsureAttributeAndRelationTables(theNewType) 
		if err != nil {
		      panic(err)
		}
    }
}

/*
TODO Need to add the constraint that the method is public here !!!!
*/
func (g *Generator) isWebDialogHandlerMethod() bool {
	return strings.HasSuffix(g.fileNameRoot,"dialog") && strings.Contains(g.packagePath,"/pkg/web/")
}


func (g *Generator) GenerateMethods() {
	
	for _,methodDeclaration := range g.file.MethodDecls {

	   methodName := methodDeclaration.Name.Name
		   
	   // main functions and web dialog handler functions are explicitly package name qualified.
	   //
	// TODO   OOPS Which package are we executing when looking for web handler methods?????
	
	   if (methodName == "main") || (g.isWebDialogHandlerMethod()) {
	      methodName = g.packagePath + methodName	
	   }
	
	   var parameterNames []string
	   var parameterTypes []string
	   for _,inputArgDecl := range methodDeclaration.Type.Params {
		  parameterNames = append(parameterNames, inputArgDecl.Name.Name)
		  parameterTypes = append(parameterTypes, inputArgDecl.Type.Name.Name)
	   }
	
	   numReturnArgs := len(methodDeclaration.Type.Results)
	
	   allowRedefinition := false	
	
	   // FuncType.	Params  []*InputArgDecl // input parameter declarations. Can be empty list.
   	   // 	        Results []*ReturnArgDecl // (outgoing) result declarations; Can be empty list.	
	
	   rMethod, err := data.RT.CreateMethod(g.packageName,
	   	                                    methodName,
	   	                                    parameterNames, 
		                                    parameterTypes, 
		                                    numReturnArgs,
		                                    methodDeclaration.NumLocalVars,
		                                    allowRedefinition  )
	   if err != nil {
	       panic(err)
	   }
	
       rMethod.Code = methodDeclaration // abstract syntax tree	

	   fmt.Println(rMethod)		
    }	
}


/*
If the const name is unqualified (has no package path), prefixes it with the current package path.
A little bit inefficient.
*/
func (g *Generator) qualifyConstName(constName string) string {
   if strings.LastIndex(constName,"/") == -1 {
       return g.packagePath + constName  	
   }	
   return constName
}


func (g *Generator) GenerateConstants () {
	for _,constDeclaration := range g.file.ConstantDecls {
		

	   constantName := g.qualifyConstName(constDeclaration.Name.Name)
	   g.Interp.EvalExpr(g.th,constDeclaration.Value)
	   obj := g.th.Pop()
	   data.RT.CreateConstant(constantName,obj)
    }	
}

func (g *Generator) TestWalk() {
	p := &NodePrinter{}
	ast.Walk(p, g.file)
}

type NodePrinter struct {
}

func (p *NodePrinter) Visit(node ast.Node) (w ast.Visitor) {
	if node != nil {
	   fmt.Println("***",node)
    }
	return p
}
