// Copyright TBD

package modbus_methods

/*
   modbus.go - native methods for relish extensions library 'modbus' package. 
*/

import (
   . "relish/runtime/data"
   modbus "relish/runtime/native_methods/extensions/protocols/modbus"
   "strconv"
)


func InitModbusMethods() {
   // -------------------------------------------------------------------
   // Modbus 

   // readHoldingRegisters
   //    modbus Modbus
   //    addr Uint32
   //    numRegister Uint32
   // >
   //    pdu Bytes
   readHoldingRegistersMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"readHoldingRegisters", []string{"modbus","addr","numRegister"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Uint32","Uint32"}, []string{"Bytes"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   readHoldingRegistersMethod.PrimitiveCode = readHoldingRegisters

   // writeMultipleRegisters
   //    modbus Modbus
   //    addr Uint32
   //    values []Uint32 // should be []Uint16
   // >
   //    pdu Bytes
   writeMultipleRegistersMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"writeMultipleRegisters", []string{"modbus","addr","values"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Uint32","List_of_Uint"}, []string{"Bytes"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   writeMultipleRegistersMethod.PrimitiveCode = writeMultipleRegisters

   // writeSingleRegister
   //    modbus Modbus
   //    addr Uint32
   //    value Uint32 // should be Uint16
   // >
   //    pdu Bytes
   writeSingleRegisterMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"writeSingleRegister", []string{"modbus","addr","value"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Uint32","Uint32"}, []string{"Bytes"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   writeSingleRegisterMethod.PrimitiveCode = writeSingleRegister

   // diagnostic
   //    modbus Modbus
   //    value Int32
   // >
   //    pdu Bytes
   diagnosticMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"diagnostic", []string{"modbus","value"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Int32"}, []string{"Bytes"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   diagnosticMethod.PrimitiveCode = diagnostic

   // query
   //    mb Modbus
   //    command Bytes
   // >
   //    response Bytes
   //    err String  
   queryMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"query", []string{"mb","command"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/Modbus","Bytes"}, []string{"Bytes","String"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   queryMethod.PrimitiveCode = query


   //--------------------------------------------------------------
   // Encoding Methods

   // toWord
   //    num Int32
   // >
   //    upper Byte
   //    lower Byte
   toWordMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"toWord", []string{"num"}, []string{"Int32"}, []string{"Bytes","Byte"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   toWordMethod.PrimitiveCode = toWord

   // toInt
   //    arr Bytes
   // >
   //    num Int32
   toIntMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"toInt", []string{"arr"}, []string{"Bytes"}, []string{"Int32"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   toIntMethod.PrimitiveCode = toInt

   // toInt
   //    arr Bytes
   // >
   //    num Int32
   toInt2301Method, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"toInt2301", []string{"arr"}, []string{"Bytes"}, []string{"Int32"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   toInt2301Method.PrimitiveCode = toInt2301


   //--------------------------------------------------------------
   // ModbusTCP

   // connect
   //    modbus ModbusTcp
   //    ipAddr String
   //    port Uint32
   //    slaveAddr Uint32
   // >
   //    err String
   connectMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"connect", []string{"modbus","ipAddr","port","slaveAddr"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp","String","Uint32","Uint32"}, []string{"String"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   connectMethod.PrimitiveCode = connect

   // close
   //    modbus ModbusTcp
   closeMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"close", []string{"modbus"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp"}, []string{}, false, 0, false)
   if err != nil {
      panic(err)
   }
   closeMethod.PrimitiveCode = close


   // maintainOpenConnection
   //    ipAddr String
   //    port Uint32
   maintainOpenConnectionMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"maintainOpenConnection", []string{"ipAddr","port"}, []string{"String","Uint32"}, []string{}, false, 0, false)
   if err != nil {
      panic(err)
   }
   maintainOpenConnectionMethod.PrimitiveCode = maintainOpenConnection

   // discardOpenConnection
   //    ipAddr String
   //    port Uint32
   discardOpenConnectionMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"discardOpenConnection", []string{"ipAddr","port"}, []string{"String","Uint32"}, []string{}, false, 0, false)
   if err != nil {
      panic(err)
   }
   discardOpenConnectionMethod.PrimitiveCode = discardOpenConnection



   // send
   //    modbus Modbus
   //    pdu Bytes
   // >
   //    err String
   sendMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"send", []string{"modbus","pdu"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp","Bytes"}, []string{"String"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   sendMethod.PrimitiveCode = send

   // read
   //    modbus Modbus
   // >
   //    response Bytes
   //    err String
   readMethod, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"read", []string{"modbus"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp"}, []string{"Bytes","String"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   readMethod.PrimitiveCode = read



   modbusTcpInitMethod1, err := RT.CreateMethod("gait.bcit.ca2012/protocols/pkg/modbus",nil,"gait.bcit.ca2012/protocols/pkg/modbus/initModbusTcp", []string{"modbus","addressMode","queryTimeout", "queryRetries"}, []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp","String","Uint","Uint32"},  []string{"gait.bcit.ca2012/protocols/pkg/modbus/ModbusTcp"}, false, 0, false)
   if err != nil {
      panic(err)
   }
   modbusTcpInitMethod1.PrimitiveCode = initModbusTcp 

}




 
///////////////////////////////////////////////////////////////////////////////////////////
// Modbus functions


// readHoldingRegisters
//    modbus Modbus
//    addr Uint32
//    numRegister Uint32
// >
//    pdu Bytes
func readHoldingRegisters(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   addr := uint32(objects[1].(Uint32))
   numRegister := uint32(objects[2].(Uint32))
   modbus := wrapper.GoObj.(modbus.Modbus)
   pdu := modbus.ReadHoldingRegisters(addr,numRegister)
   return []RObject{Bytes(pdu)}
}


// writeMultipleRegisters
//    modbus Modbus
//    addr Uint32
//    values []Uint32
// >
//    pdu Bytes
func writeMultipleRegisters(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   addr := uint32(objects[1].(Uint32))
   
   collection := objects[2].(RCollection)
   length := collection.Length()
   values := make([]uint16,length)
   i := 0
   for val := range collection.Iter(th) {
      values[i] = uint16(uint32(val.(Uint32))) // can we convert from Uint32 to uint16?
      i++
   }

   modbus := wrapper.GoObj.(modbus.Modbus)
   pdu := modbus.WriteMultipleRegisters(addr,values)
   return []RObject{Bytes(pdu)}
}


// writeSingleRegisters
//    modbus Modbus
//    addr Uint32
//    value Uint32
// >
//    pdu Bytes
func writeSingleRegister(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   addr := uint32(objects[1].(Uint32))
   value := uint16(uint32(objects[2].(Uint32))) // can we convert from Uint32 to uint16?
   modbus := wrapper.GoObj.(modbus.Modbus)
   pdu := modbus.WriteSingleRegister(addr,value)
   return []RObject{Bytes(pdu)}
}


// diagnostic
//    modbus Modbus
//    value Int32
// >
//    pdu Bytes
func diagnostic(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   value := int32(objects[1].(Int32))
   modbus := wrapper.GoObj.(modbus.Modbus)
   pdu := modbus.Diagnostic(value)
   return []RObject{Bytes(pdu)}
}


// query
//    mb Modbus
//    command Bytes
// >
//    response Bytes
//    err String
func query(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   command := objects[1].(Bytes)
   c := ([]byte)(command)
   mb := wrapper.GoObj.(modbus.Modbus)
   response, err := modbus.Query(mb, c) //modbus package function, not a function of a modbus object
   errStr := ""
   if err != nil {
      errStr = err.Error()
   }   
   return []RObject{Bytes(response),String(errStr)}
}


///////////////////////////////////////////////////////////////////////////////////////////
// Encoding functions


// toWord
//    num Int32
// >
//    upper Byte
//    lower Byte
func toWord(th InterpreterThread, objects []RObject) []RObject {
   
   num := int32(objects[0].(Int32))
   upper, lower := modbus.ToWord(num)
   return []RObject{Byte(upper),Byte(lower)}
}

// toInt
//    arr Bytes
// >
//    num Int32
func toInt(th InterpreterThread, objects []RObject) []RObject {
   
   arr := objects[0].(Bytes)
   a := ([]byte)(arr)
   num := modbus.ToInt(a)
   return []RObject{Int32(num)}
}

// toInt2301
//    arr Bytes
// >
//    num Int32
func toInt2301(th InterpreterThread, objects []RObject) []RObject {
   
   arr := objects[0].(Bytes)
   a := ([]byte)(arr)
   num := modbus.ToInt_2301(a)
   return []RObject{Int32(num)}
}

///////////////////////////////////////////////////////////////////////////////////////////
// ModbusTcp functions


// connect
//    modbus ModbusTcp
//    ipAddr String
//    port Uint32
//    slaveAddr Uint32
// >
//    err String
func connect(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   ipAddr := string(objects[1].(String))
   port := uint32(objects[2].(Uint32))
   slaveAddr := uint32(objects[3].(Uint32))
   modbus := wrapper.GoObj.(*modbus.ModbusTCP)
   err := modbus.Connect(ipAddr,port,slaveAddr)
   errStr := ""
   if err != nil {
      errStr = err.Error()
   }
   return []RObject{String(errStr)}
}


// close
//    modbus ModbusTcp
func close(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   modbus := wrapper.GoObj.(*modbus.ModbusTCP)
   modbus.Close()
   return []RObject{}
}


// maintainOpenConnection
//    ipAddr String
//    port Uint32
func maintainOpenConnection(th InterpreterThread, objects []RObject) []RObject {
   ipAddr := string(objects[0].(String))
   port := uint32(objects[1].(Uint32))
   
   ipAddrAndPort := ipAddr+":"+strconv.FormatUint(uint64(port), 10) 
     
   modbus.MaintainOpenConnection(ipAddrAndPort)

   return []RObject{}
}

// discardOpenConnection
//    ipAddr String
//    port Uint32
func discardOpenConnection(th InterpreterThread, objects []RObject) []RObject {
   ipAddr := string(objects[0].(String))
   port := uint32(objects[1].(Uint32))
   
   ipAddrAndPort := ipAddr+":"+strconv.FormatUint(uint64(port), 10) 
     
   modbus.DiscardOpenConnection(ipAddrAndPort)

   return []RObject{}
}



// send
//    modbus Modbus
//    pdu Bytes
// >
//    err String
func send(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   pdu := objects[1].(Bytes)
   p := ([]byte)(pdu)
   modbus := wrapper.GoObj.(*modbus.ModbusTCP)
   err := modbus.Send(p)
   errStr := ""
   if err != nil {
      errStr = err.Error()
   }
   return []RObject{String(errStr)}
}


// read
//    modbus Modbus
// >
//    response Bytes
//    err String
func read(th InterpreterThread, objects []RObject) []RObject {
   
   wrapper := objects[0].(*GoWrapper)
   modbus := wrapper.GoObj.(*modbus.ModbusTCP)
   response, err := modbus.Read()
   errStr := ""
   if err != nil {
      errStr = err.Error()
   }
   return []RObject{Bytes(response),String(errStr)}
}





///////////////////////////////////////////////////////////////////////////////////////////
// Type init functions

func initModbusTcp(th InterpreterThread, objects []RObject) []RObject {
   
   modbusWrapper := objects[0].(*GoWrapper)

   addressMode := string(objects[1].(String))   
   queryTimeout := uint64(objects[2].(Uint))    
   queryRetries := uint32(objects[3].(Uint32))

   modbusTcp := modbus.MakeModbusTCP( addressMode, queryTimeout, queryRetries )

   modbusWrapper.GoObj = modbusTcp

   return []RObject{modbusWrapper}
}