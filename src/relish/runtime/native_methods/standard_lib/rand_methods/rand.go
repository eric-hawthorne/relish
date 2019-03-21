// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package rand_methods

/*
   rand.go - native methods for exposing Go's math/rand library in relish
*/

import (
	. "relish/runtime/data"
    "math/rand"
)

///////////
// Go Types

// None so far


/////////////////////////////////////
// relish method to go method binding

func InitRandMethods() {



/*
expFloat returns an exponentially distributed Float in the range (0, +math.MaxFloat64] 
with an exponential distribution whose rate parameter (lambda) is 1 and whose mean is 1/lambda (1) from the default Source. 
To produce a distribution with a different rate parameter, callers can adjust the output using:

sample = div expFloat() desiredRateParameter
*/
	expFloatMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/rand",nil,"expFloat", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	expFloatMethod.PrimitiveCode = expFloat   

/*
float returns, as a Float, a pseudo-random number in [0.0,1.0) from the default Source.
*/
	floatMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/rand",nil,"float", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	floatMethod.PrimitiveCode = float   




/*
int returns a non-negative pseudo-random 63-bit integer as an Int from the default Source.
*/
	intMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/rand",nil,"int", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intMethod.PrimitiveCode = int63  


/*
intn returns, as an Int, a non-negative pseudo-random number in [0,n) from the default Source. It panics if n <= 0.
*/
	intnMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/rand",nil,"intn", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{"Int"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	intnMethod.PrimitiveCode = int63n  



/*
normFloat returns a normally distributed Float in the range [-math.MaxFloat64, +math.MaxFloat64] 
with standard normal distribution (mean = 0, stddev = 1) from the default Source. 
To produce a different normal distribution, callers can adjust the output using:

sample = plus (times normFloat desiredStdDev) desiredMean
*/
	normFloatMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/rand",nil,"normFloat", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{"Float"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	normFloatMethod.PrimitiveCode = normFloat   


//func Perm(n int) []int
//func Read(p []byte) (n int, err error)

/*
Seed uses the provided seed value to initialize the default Source to a deterministic state. 
If Seed is not called, the generator behaves as if seeded by Seed(1). 
Seed values that have the same remainder when divided by 2^31-1 generate the same pseudo-random sequence. 
Seed, unlike the Rand.Seed method, is safe for concurrent use.
*/
	seedMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/rand",nil,"seed", 
		                                    []string{"n"}, 
		                                    []string{"Int"}, 
		                                    []string{}, false, 0, false)
	if err != nil {
		panic(err)
	}
	seedMethod.PrimitiveCode = seed   



/*
uint32 returns a pseudo-random 32-bit value as a Uint32 from the default Source.
*/
	uint32Method, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/rand",nil,"uint32", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{"Uint32"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uint32Method.PrimitiveCode = uint_32   



/*
uint returns a pseudo-random 64-bit value as a Uint from the default Source.

TODO UNCOMMENT WHEN SWITCH TO GO 1.8 !!!!!!


	uintMethod, err := RT.CreateMethod("shared.relish.pl2012/relish_lib/pkg/rand",nil,"uint", 
		                                    []string{}, 
		                                    []string{}, 
		                                    []string{"Uint"}, false, 0, false)
	if err != nil {
		panic(err)
	}
	uintMethod.PrimitiveCode = uint_64   
*/
 
}
///////////////////////////////////////////////////////////////////////////////////////////
// Random functions - not suitable for crypto

/*
expFloat returns an exponentially distributed Float in the range (0, +math.MaxFloat64] 
with an exponential distribution whose rate parameter (lambda) is 1 and whose mean is 1/lambda (1) from the default Source. 
To produce a distribution with a different rate parameter, callers can adjust the output using:

sample = div expFloat() desiredRateParameter
*/
func expFloat(th InterpreterThread, objects []RObject) []RObject {
   val := rand.ExpFloat64()
   return []RObject{Float(val)}
}

/*
float returns, as a Float, a pseudo-random number in [0.0,1.0) from the default Source.
*/
func float(th InterpreterThread, objects []RObject) []RObject {

	val := rand.Float64() 
   	return []RObject{Float(val)}
}

/*
int63 returns a non-negative pseudo-random 63-bit integer as an Int from the default Source.
*/
func int63(th InterpreterThread, objects []RObject) []RObject {

	val := rand.Int63() 
   	return []RObject{Int(val)}
}

/*
int63n returns, as an Int, a non-negative pseudo-random number in [0,n) from the default Source. It panics if n <= 0.
*/
func int63n(th InterpreterThread, objects []RObject) []RObject {

	n := int64(objects[0].(Int))
	val := rand.Int63n(n) 
   	return []RObject{Int(val)}	
}

/*
normFloat returns a normally distributed Float in the range [-math.MaxFloat64, +math.MaxFloat64] 
with standard normal distribution (mean = 0, stddev = 1) from the default Source. 
To produce a different normal distribution, callers can adjust the output using:

sample = plus (times normFloat desiredStdDev) desiredMean
*/
func normFloat(th InterpreterThread, objects []RObject) []RObject {
	val := rand.NormFloat64() 
   	return []RObject{Float(val)}
}

//func Perm(n int) []int
//func Read(p []byte) (n int, err error)

/*
Seed uses the provided seed value to initialize the default Source to a deterministic state. 
If Seed is not called, the generator behaves as if seeded by Seed(1). 
Seed values that have the same remainder when divided by 2^31-1 generate the same pseudo-random sequence. 
Seed, unlike the Rand.Seed method, is safe for concurrent use.
*/
func seed(th InterpreterThread, objects []RObject) []RObject {
	n := int64(objects[0].(Int))	
	rand.Seed(n)
	return []RObject{}
}

/*
uint_32 returns a pseudo-random 32-bit value as a Uint32 from the default Source.
*/

func uint_32(th InterpreterThread, objects []RObject) []RObject {
    val := rand.Uint32()
	return []RObject{Uint32(val)}
}

/*
uint_64 returns a pseudo-random 64-bit value as a Uint from the default Source.

TODO UNCOMMENT WHEN SWITCH TO GO 1.8 !!!!!!

func uint_64(th InterpreterThread, objects []RObject) []RObject {
    val := rand.Uint64()
	return []RObject{Uint(val)}
}
*/

