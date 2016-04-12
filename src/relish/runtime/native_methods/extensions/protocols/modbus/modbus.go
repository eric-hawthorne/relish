/*
   This is an implementation of the ModBus protocol in Go.
*/

package modbus

import (
//    "fmt"
"errors"
)

type Modbus interface {
	ReadHoldingRegisters(addr uint32, numRegister uint32) []byte
	WriteMultipleRegisters(addr uint32, values []uint16) []byte
	WriteSingleRegister(addr uint32, value uint16) []byte
	Diagnostic(value int32) []byte
	Send(pdu []byte) (err error)
	Read() (response []byte, err error)
}

/// Go time appears to always be measured in nanoseconds.
/// This constant needs to be somewhere that *everything* can easily get at.
const NANOSECONDS = 1000 * 1000 * 1000

const (
	FOUR_BYTE_MODE = "Four Byte Mode"
	TWO_BYTE_MODE  = "Two Byte Mode"
)

/// MODBUS registers are 16-bits (2 bytes).  At least all the ones that I've
/// encountered are.  I'm sure, somewhere, sometime, someone will come up with a
/// way to mess with this standard as well.
const SIZEOF_MODBUS_REGISTER = 2

// Connection error codes.

const (
	NO_CONNECTION      = "No connection established."
	ERR_TRANSACTION_ID = "Transaction ID mismatched."
	ERR_SLAVE_ADDR     = "Incorrect slave address."
	ERR_READ_TIMEOUT   = "Timed out during read attempt."
	ERR_FUNCTION_CODE  = "Incorrect function code."
	ERR_CRC            = "Incorrect CRC."
	ERR_DATA_SIZE      = "Incorrect data size."
)

/*
   Modbus function codes
*/
const (
	FC_DISCRETE_OUTPUT_COILS    = 1
	FC_DISCRETE_INPUT_COILS     = 2
	FC_READ_HOLDING_REGISTERS   = 3
	FC_READ_INPUT_REGISTERS     = 4
	FC_WRITE_SINGLE_REGISTER    = 6
	FC_DIAGNOSTIC               = 8
	FC_WRITE_MULTIPLE_REGISTERS = 16
)

/*
   Modbus error codes
*/
var ERRORS = map[uint32]string{
	1:  "Illegal Function",
	2:  "Illegal Data Address",
	3:  "Illegal Data Value",
	4:  "Slave Device Failure",
	5:  "Acknowledge",
	6:  "Slave Device Busy",
	7:  "Negative Acknowledge",
	8:  "Memory Parity Error",
	10: "Gateway Path Unavailable",
	11: "Gateway Target Device Failed to Respond",
}

/*
   Send a command frame to slave and returns a response

   @param      modbus - modbus connection to send to
               command - command frame to send

   @return     response - response from slave
               err - error from response
*/
func Query(mb Modbus, command []byte) (response []byte, err error) {
	//fmt.Println ("=====")
	err = mb.Send(command)
	if err != nil {
		return
	}

	response, err = mb.Read()

	if len(response) > 0 {
		functionCode := uint32(response[0])
		//fmt.Printf ("fc=%d\n", functionCode)

		switch {
		case functionCode >= uint32(0x80):
			errorCode := uint32(response[1])
			return nil, errors.New(ERRORS[errorCode])

		case functionCode == FC_READ_HOLDING_REGISTERS:
			fallthrough

		case functionCode == FC_READ_INPUT_REGISTERS:
			return response[2:], nil

		case functionCode == FC_WRITE_MULTIPLE_REGISTERS: //what to return?
			fallthrough

		case functionCode == FC_WRITE_SINGLE_REGISTER: //what to return?
			fallthrough

		case functionCode == FC_DIAGNOSTIC: //what to return?
			return response[1:], nil
		}
	}

	return
}

type modbus struct {
	addressMode  string
	queryTimeout uint64 // in nanoseconds
	queryRetries uint32
	slaveAddr    byte
}

func MakeModbus(addressMode string, queryTimeout uint64, queryRetries uint32) *modbus {

	// 20-Apr-11 AS - Shouldn't this be &&?  Otherwise it will *ALWAYS* force
	// FOUR_BYTE_MODE.

	if addressMode != FOUR_BYTE_MODE || addressMode != TWO_BYTE_MODE {
		addressMode = FOUR_BYTE_MODE
	}
	return &modbus{addressMode, queryTimeout, queryRetries, 0}
}

/*
   Read Holding Registers

   @param      addr        - start address of registers
               numRegister - number of registers to read

   @return     ModBus command frame for reading multiple registers
*/
func (m *modbus) ReadHoldingRegisters(addr uint32, numRegister uint32) []byte {
	pdu := make([]byte, 5) // []byte{}
	//pdu = bytes.AddByte( pdu, byte(FC_READ_HOLDING_REGISTERS) )
	//pdu = bytes.Add( pdu, ToWord( int32(addr) )
	pdu[0] = byte(FC_READ_HOLDING_REGISTERS)
	pdu[1], pdu[2] = ToWord(int32(addr))
	pdu[3], pdu[4] = ToWord(int32(numRegister))

	return pdu
}

/*
   Write Multiple Regisers

   @param      addr        - start address of registers
               values      - an array of interger values to write to registers

   @return     ModBus command frame for writing multiple registers
*/
func (m *modbus) WriteMultipleRegisters(addr uint32, values []uint16) []byte {

	len := len(values)
	pdu := make([]byte, len*2+6) // total buffer size

	pdu[0] = byte(FC_WRITE_MULTIPLE_REGISTERS) // Function code	
	pdu[1], pdu[2] = ToWord(int32(addr))       // Starting Address
	pdu[3], pdu[4] = ToWord(int32(len))        // Quantity of registers
	pdu[5] = byte(len * 2)                     // byte count

	for i := 0; i < len; i++ { // data bytes
		pdu[i*2+6] = byte(values[i] >> 8)
		pdu[i*2+7] = byte(values[i])
	}

	return pdu
}

/*
   Write Single Register

   @param      addr        - start address of register
               value       - integer value to write to register

   @return     ModBus command frame for writing single register
*/
func (m *modbus) WriteSingleRegister(addr uint32, value uint16) []byte {
	pdu := make([]byte, 5)
	pdu[0] = byte(FC_WRITE_SINGLE_REGISTER)
	pdu[1], pdu[2] = ToWord(int32(addr))
	pdu[3], pdu[4] = ToWord(int32(value))

	return pdu
}

/*
   Diagnostic

   @param      value       - test value for test

   @return     ModBus command frame for echo back test
*/
func (m *modbus) Diagnostic(value int32) []byte {
	pdu := make([]byte, 5)
	pdu[0] = byte(FC_DIAGNOSTIC)
	pdu[1], pdu[2] = ToWord(0)
	pdu[3], pdu[4] = ToWord(value)

	return pdu
}
