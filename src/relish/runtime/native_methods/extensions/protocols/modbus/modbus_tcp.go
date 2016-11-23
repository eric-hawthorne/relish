package modbus

/*
   This is an implementation of modbus protocol over TCP/IP.
*/

import (
	"errors"
	"fmt"
	"net"
        "io"
        "syscall"
	"strconv"
	"sync"
)

//const (
//    NO_CONNECTION       = "No connection established."
//    ERR_TRANSACTION_ID  = "Transaction ID mismatched."
//    ERR_SLAVE_ADDR      = "Incorrect slave address."
//    ERR_READ_TIMEOUT    = "Timed out during read attempt."
//)

// Transaction ID for Modbus TCP.
var transactionId int32 = 0

type ModbusTCP struct {
	*modbus
	tcpConn net.Conn
	tID     int32
	ipAddrAndPort string
}


// Map from ip-address:port to whether or not multiple modbus requests to that ip-address:port
// should share a single always-open TCP connection. If an ip-address:port is in this map,
// it means yes use the kept alive connection. If ip-address:port is not in this map, then 
// a single TCP connection should be created for each modbus request to the address:port.
//
var useKeptAliveConnection map[string]bool = make(map[string]bool)

// A map from ip-address:port to an open tcp connection to that ip address.
// See comment above.
//
var openTcpConnections map[string]net.Conn = make(map[string]net.Conn)

var openConnectionMutex sync.Mutex

func MaintainOpenConnection(ipAddrAndPort string) {
   openConnectionMutex.Lock()
   useKeptAliveConnection[ipAddrAndPort] = true
   openConnectionMutex.Unlock()
}

func DiscardOpenConnection(ipAddrAndPort string) {
   openConnectionMutex.Lock()	
   useKeptAliveConnection[ipAddrAndPort] = false
   openTcpConnections[ipAddrAndPort].Close()   
   delete(openTcpConnections,ipAddrAndPort)
   openConnectionMutex.Unlock()   

}


/*
   This creates a Modbus over TCP client.
*/
func MakeModbusTCP(addressMode string, queryTimeout uint64, queryRetries uint32) *ModbusTCP {
	mTCP := &ModbusTCP{MakeModbus(addressMode, queryTimeout, queryRetries), nil, 0, ""}

	return mTCP
}

/*
   Creates a TCP connection to slave on specified IP address and port

   @param  ippAddr     - IP address
           port
           slaveAddr   - Slave address

   @return err         - connection error
*/
func (mTCP *ModbusTCP) Connect(ipAddr string, port uint32, slaveAddr uint32) (err error) {
   
    mTCP.ipAddrAndPort = ipAddr+":"+strconv.FormatUint(uint64(port), 10)

    openConnectionMutex.Lock()	
    defer openConnectionMutex.Unlock()
    connection, found := openTcpConnections[mTCP.ipAddrAndPort]
    if ! found {

		connection, err = net.Dial("tcp", mTCP.ipAddrAndPort)
		if err != nil {
			return
		}
        
        if useKeptAliveConnection[mTCP.ipAddrAndPort] {
        	openTcpConnections[mTCP.ipAddrAndPort] = connection
        }
    }


	mTCP.tcpConn = connection

	mTCP.slaveAddr = byte(slaveAddr) //TODO: test for address > 255 ?
	return
}


/*
Applicable only to kept-alive connections.
Assumes the connection is dead in some way. 
Attempts a close on the connection and an opening of a new tcp connection.
*/
func (mTCP *ModbusTCP) RepairConnection() (err error) {
    openConnectionMutex.Lock()	
    defer openConnectionMutex.Unlock()
    mTCP.tcpConn.Close() // can fail
    var connection net.Conn
    connection, err = net.Dial("tcp", mTCP.ipAddrAndPort)
    if err != nil {
        return
    }

    if useKeptAliveConnection[mTCP.ipAddrAndPort] {
       delete(openTcpConnections,mTCP.ipAddrAndPort)
       openTcpConnections[mTCP.ipAddrAndPort] = connection
    }
    mTCP.tcpConn = connection
    return
}


/*
   Closes the TCP connection.
   WARNING. Do not call this if your ModbusTCP is using a shared open (kept-alive) TCP connection.
*/
func (mTCP *ModbusTCP) Close() {
	if mTCP.tcpConn != nil {
        openConnectionMutex.Lock()	
        defer openConnectionMutex.Unlock()		
        if ! useKeptAliveConnection[mTCP.ipAddrAndPort] {
		   mTCP.tcpConn.Close()
	    }
		mTCP.tcpConn = nil
	}
}

/*
   Sends a command to slave

   @param  pdu (protocol data unit) - data to send

   @return error
*/
func (mTCP *ModbusTCP) Send(pdu []byte) (err error) {
	if mTCP.tcpConn == nil {
		return errors.New(NO_CONNECTION)
	}

	pduLength := len(pdu)

	message := make([]byte, 7+pduLength)

	transactionId = transactionId + 1
	if transactionId > 65535 {
		transactionId = 0
	}

	mTCP.tID = transactionId

	// Modbus TCP Header
	message[0], message[1] = ToWord(mTCP.tID)
	message[2], message[3] = ToWord(0) // protocol id for Modbus TCP
	message[4], message[5] = ToWord(int32(pduLength + 1))
	message[6] = mTCP.slaveAddr

	copy(message[7:], pdu)

	//fmt.Printf("Sending: %x\n",message)

	var n int
	n, err = mTCP.tcpConn.Write(message)

	fmt.Printf("Sent %x, %v bytes sent.", message, n)
	if err != nil {
		fmt.Println( "Error:", err )
		if err == io.EOF || err == syscall.EINVAL {
                   err = errors.New(CONNECTION_DEAD)
                }
	} else  {
		fmt.Println( "" )
	}

	return
}

/*
   Reads a response from slave

   @return response
*/
func (mTCP *ModbusTCP) Read() (response []byte, err error) {
	if mTCP.tcpConn == nil {
		return []byte{}, errors.New(NO_CONNECTION)
	}

	header := make([]byte, 7)
	for numRetries := mTCP.queryRetries + 1; numRetries > 0; numRetries-- {
		if numRetries <= mTCP.queryRetries {
			fmt.Printf("Error: %v. Retrying header read.\n", err)
			header = make([]byte, 7)
		}

		// Read TCP/IP header
		_, err = mTCP.tcpConn.Read(header)
		if err == nil {
			break
		}
	}

	if err != nil {
		if timeoutErr, ok := err.(net.Error); ok && timeoutErr.Timeout() {
			err = errors.New(ERR_READ_TIMEOUT)
		} else if err == io.EOF || err == syscall.EINVAL {
                        err = errors.New(CONNECTION_DEAD)
                }
		return
	} else {

		id := ToInt(header[0:2])
		if id == mTCP.tID {
			fmt.Println("Transaction ID correct.")
		} else {
			//fmt.Printf( "Transaction ID mismatch: %v.\n", mTCP.tID )
			return []byte{}, errors.New(ERR_TRANSACTION_ID)
		}

		protocol := ToInt(header[2:4])
		if protocol == 0 {
			//fmt.Println( "Protocol correct.")
		}

		length := ToInt(header[4:6])
		fmt.Printf("Response is %v bytes.\n", length)

		slaveAddr := header[6]
		if slaveAddr == mTCP.slaveAddr {
			//fmt.Println( "Slave address correct." )
		} else {
			return []byte{}, errors.New(ERR_SLAVE_ADDR)
		}

		response = make([]byte, length-1) // -1 because slave address had been read already

		for numRetries := mTCP.queryRetries + 1; numRetries > 0; numRetries-- {
			// Read modbus response
			_, err = mTCP.tcpConn.Read(response)
			if err == nil {
				fmt.Printf("Received %x\n", response)
				break
			}
		}

		if err != nil {
			if timeoutErr, ok := err.(net.Error); ok && timeoutErr.Timeout() {
				err = errors.New(ERR_READ_TIMEOUT)
		        } else if err == io.EOF || err == syscall.EINVAL {
                           err = errors.New(CONNECTION_DEAD)
                        }
		}
	}

	return
}
