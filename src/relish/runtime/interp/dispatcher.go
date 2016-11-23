// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU GPL v3 license, found in the LICENSE_GPL3 file.

// this package is concerned with the operation of the intermediate-code interpreter
// in the relish language.

package interp

/*
   dispatcher.go -  multi-method dispatcher

   The gist of multi-method dispatch in relish is as follows:

   The types of all actual arguments at runtime to a method call are used to select
   which method implementation, of the method-name, will be executed.

   The tuple of argument types is matched with the input-parameter type-signature of
   each visible method-implementation, to select the best implementation to run on the
   arguments.

   The matching considers 3 factors:
   - ** Type compatibility ** - actual arguments must be type-compatible with method type-signature
   - ** Closest match **, in terms of specialization-distance down the program's multiple-inheritance
     type lattice, between the argument type tuple and the method type-signature.
   - ** Type-specificity ** of method-implementation's type signature:
     If a tie exists with two method-implementations equally closely matching the actual argument types,
     then the tie is resolved by picking the method-implementation whose type signature is
     more specific in an absolute sense in the program's type lattice.

   A few more details:

   When a multi-method of a given name is called: 
   1. The most-specific type of each actual argument at runtime to a method call is collected
      to form a unique type-tuple. 
      - A tree index of type-tuples, and 
      - a cache of most recent tuples starting at a given type, 
      speeds this type-tuple finding or creation process.
   2. The type-tuple of the actual arguments is matched against the input-parameter signature type-tuple
      of each method-declaration which: 
      - has the same method-name, 
      - has the same number of input parameters as the number of actual arguments, and
      - is visible (directly or by a chain of package imports) from within the 
        currently executing package.
   3. A particular method-declaration (and its method implementation) is selected for execution
      if its input-parameter signature type-tuple is:
      a) Type-compatible with the actual-arguments type tuple; that is, each type in the method parameters signature
         is the same as or a supertype of the corresponding type in the actual arguments type-tuple.
      b) Closest above the actual-arguments type-tuple in the program's type lattice. Specialization distance is measured
         between each argument-type and the corresponding method-parameter-type pair, and the distances are averaged.
         All different paths up the type lattice between the argument type and the parameter type are distance-measured.
         The average type-specialization-distance between the two type-tuples is the sum of all lattice-path distances 
         divided by the  total number of lattice paths found between the pairs of types. 
      c) In the case of a tie (more than one method's signature is equally close in type-distance above the
          actual-arguments type-tuple), then a secondary test picks the method of those tied whose
          parameters signature type-tuple is the more specific in the type lattice as a whole; that is
          the signature whose type-tuple is the furthest type-specialization-distance from the top "Any" type, again
          considering the average length of all paths down the lattice between "Any" and each type in the method type-signature.
   4. For any given method-name executed in a package, this assessment is done only once for each encountered
      actual-arguments type-tuple in the program run. The method-implementation chosen for the actual-arguments type-tuple
      is cached, and subsequently, a single map lookup selects the best-matching method-implementation whenever the 
      method-name is called again with arguments with that tuple of types.    
*/


import (
   . "relish/runtime/data"
   . "relish/dbg"
   "sync"
)



type dispatcher struct {
   // typeTupleTree *TypeTupleTreeNode  // obsolete
   typeTupleTrees []*TypeTupleTreeNode  // new   
   emptyTypeTuple *RTypeTuple // a cached type tuple, to speed dispatch in special cases.

}

func newDispatcher(rt *RuntimeEnv) (d *dispatcher) {
   emptyTypeTuple := rt.TypeTupleTrees[0].GetTypeTuple(nil, []RObject{})
   d =  &dispatcher{rt.TypeTupleTrees, emptyTypeTuple}
   
   return
}

// TODO Note: It is a real problem that we are having to synchronize 
// method dispatch. Perhaps the best solution would be to give each interpreter thread
// its own method lookup table and typetuple table. 
// That in itself would impose substantial inefficiencies. Hmmmm
// Would it be more efficient to mutex lock only the dynamic dispatch and the insertions into the mm.CachedMethods
// hashtable, and to put a deferred recover call around the GetTypeTuple and lookup of mm.CachedMethods[typeTuple]
// where the recover would try the lookup again inside the mutex lock. So try again, synchronized, to look up 
// method in cache after a memory fault occurs in trying to look up without locking.
// Actually, type tuple creation/search will have to be assessed separately from cached method lookup.
//
var dispatchMutex sync.Mutex


/*
   The main dispatch function.
   First looks up a cache (map) of method implementations keyed by typetuples.
   a method will be found in this cache if the type tuple of the arguments
   has had the multimethod called on it before.
   If there is a cache miss, uses a multi-argument dynamic dispatch algorithm
   to find the best-matching method implementation (then caches the find under
   the type-tuple of the arguments for next time.)
   Returns the best method implementation for the types of the argument objects,
   or nil if the multimethod has no method signature which is compatible with
   the types of the argument objects.
   Also returns the type-tuple of the argument objects, which can be used to
   report the lack of a compatible method.
*/
func (d *dispatcher) GetMethod(mm *RMultiMethod, args []RObject) (*RMethod,*RTypeTuple) {
   dispatchMutex.Lock()
   typeTuple := d.typeTupleTrees[len(args)].GetTypeTuple(mm,args)
   method,found := mm.CachedMethods[typeTuple]
   if !found {
      method = d.dynamicDispatch(mm,typeTuple)
      if method != nil {
         mm.CachedMethods[typeTuple] = method
      }
   }
   dispatchMutex.Unlock()
   return method,typeTuple
}

	

/*
Special, degenerate case when we know there is only a single method implementation associated with the multimethod.
Returns the method, or nil if there is none. May return a method of any arity.
Note that this process may result in a multimethod whose method of non-zero arity is in its CachedMethods map as having zero arity. No real
big deal as long as this method is only ever called on multimethods that actually only have a single method implementation.
*/
func (d *dispatcher) GetSingletonMethod(mm *RMultiMethod) *RMethod {
   dispatchMutex.Lock()   
   method,found := mm.CachedMethods[d.emptyTypeTuple]
   if !found {
      for _,methods := range mm.Methods {
         method = methods[0]
         break
      }      
      if method != nil {
         mm.CachedMethods[d.emptyTypeTuple] = method
      }
   }
   dispatchMutex.Unlock()
   return method   
}






/*
   Same as GetMethod but for types instead of object instances.
*/
func (d *dispatcher) GetMethodForTypes(mm *RMultiMethod, types ...*RType) (*RMethod,*RTypeTuple) {
   dispatchMutex.Lock()     
   typeTuple := d.typeTupleTrees[len(types)].GetTypeTupleFromTypes(types)
   method,found := mm.CachedMethods[typeTuple]
   if !found {
      method = d.dynamicDispatch(mm,typeTuple)
      if method != nil {
         mm.CachedMethods[typeTuple] = method
      }
   }
   dispatchMutex.Unlock()   
   return method,typeTuple
}

/*
   Find the method implementation of the multimethod whose parameter-type
   signature is more general than but minimal Euclidean distance
   (in multi-dimensional type specialization space) from the type-tuple
   of the actual argument objects.
   Specificity of the method that is chosen is determined in two ways.
   First, the method which is minimally different in types from the argument types
   is found. If more than one method is equally close in type signature to the
   argument types (measuring Euclidean distance down the specialization paths),
   then the tie is broken by selecting the method whose signature is most specific
   in types compared to the top types in the ontology known by the process.
   If there is still a tie, the method which was encountered first (in the multimethod's
   list of methods of a particular arity) is chosen. This is somewhat arbitrary.

   TODO Should I search downward in specialization chains from the type-tuple signature of     
   each correct-arity method of the multimethod to find the argument type tuple,
   or up the supertype path from each type in the argument type-tuple?

   TODO This can no doubt be optimized. Try upwards first.  

   Returns the most specific type-compatible method or nil if none is found.


*/
func (d *dispatcher) dynamicDispatch(mm *RMultiMethod, argTypeTuple *RTypeTuple) *RMethod {
   candidateMethods,found := mm.Methods[len(argTypeTuple.Types)]
   if ! found {

      
      Log(INTERP2_, "No '%s' method has arity %v.\n",mm.Name,len(argTypeTuple.Types))      
      return nil
   }
   
   var minSpecializationDistance float64 = 99999
   var maxSupertypeSpecificity float64 = 99999
   var closestCandidateMethod *RMethod = nil
   for _,candidateMethod := range candidateMethods {
      // DEBUG fmt.Printf("Checking for match with %v.\n",candidateMethod)
      specializationDistance,supertypeSpecificity,incompatible := argTypeTuple.SpecializationDistanceFrom(candidateMethod.Signature)
      // DEBUG fmt.Printf("specializationDistance=%v, supertypeSpecificity=%v, incompatible=%v\n",specializationDistance,supertypeSpecificity,incompatible)
      if incompatible {
         continue
      }
      if specializationDistance < minSpecializationDistance {
         closestCandidateMethod = candidateMethod
         minSpecializationDistance = specializationDistance
         maxSupertypeSpecificity = supertypeSpecificity
      } else if specializationDistance == minSpecializationDistance {
         if supertypeSpecificity > maxSupertypeSpecificity {
             closestCandidateMethod = candidateMethod
             maxSupertypeSpecificity = supertypeSpecificity
         }
      }
   }
   return closestCandidateMethod
}

