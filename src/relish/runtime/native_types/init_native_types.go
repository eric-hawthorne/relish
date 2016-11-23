// Copyright 2012-2014 EveryBitCounts Software Services Inc. All rights reserved.
// Use of this source code is governed by the GNU LESSER GPL v3 license, found in the LICENSE_LGPL3 file.

package native_types

/*
   init_native_types.go - metadata for relish types whose instances are GoWrapper objects wrapping
                          a Go-native object instance.
*/


/*
IMPORTANT

If you add a type whose instances are native-Go-implemented, you must add the fully qualified type name to 
this map, so that the RType object will configure itself as having .IsNative = true
This will cause a GoWrapper object to be instantiated upon a constructor call for the type. 
You must then implement a native Go constructor function in a .go file in 
the approprate native_methods/extensions package, to actually construct and initialize the Go object instance.
*/
var NativeType = map [string] bool {
	
	// relish standard library types with Go native instances.
	// DO NOT MODIFY THIS LIST OF MAP ENTRIES !!
	"shared.relish.pl2012/relish_lib/pkg/files/File" : true,
	"shared.relish.pl2012/relish_lib/pkg/http_srv/UploadedFile" : true,	
	"shared.relish.pl2012/relish_lib/pkg/http_srv/Cookie" : true,	
	"shared.relish.pl2012/relish_lib/pkg/http_srv/Request" : true,
	"shared.relish.pl2012/relish_lib/pkg/reflect/DataType" : true,	
	"shared.relish.pl2012/relish_lib/pkg/reflect/Attribute" : true,		
				
	// OK TO MODIFY THE ENTRIES FROM HERE DOWN !!
	// Add extensions types which need a GoWrapper as the instance here. 

    "gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp" : true,
}
