package modbus

import
    (
//    "fmt"
    "math"
//    "strconv"
//    "unsafe"
    )

/*
    Converts an integer to word

    @param  num - integer
    @return upper and lower byte representing integer
*/
func ToWord( num int32 ) (byte, byte) {
    return byte(num >> 8), byte(num)
}

/*
    Converts a byte array to integer

    @param  byte array
    @return integer
*/
func ToInt( arr []byte ) int32 {
    var num int32 = 0
    length := len(arr)

    for i := 0; i < length; i++ {
        num = num | ( int32(arr[length-1-i]) << uint32(i*8) )
//fmt.Printf ("num=%d\n", num)
    }

    return num
}

//******************************************************************************
// ToInt_2301
/**
** Convert a 4-byte slice into an integer by rearranging the bytes in a specific
** order.  Some MODBUS implementations do not use a consistent byte ordering.
**
** @param arr - Byte array to convert.
**
** @return An integer value.
*/

func
ToInt_2301 (arr []byte) int32 {
    sz := len (arr)
    if (sz != 4) {
        return (math.MinInt32) // Closest thing to an error
    } // if

    var num int32 =
        (int32(arr[2]) << uint32(24)) +
        (int32(arr[3]) << uint32(16)) +
        (int32(arr[0]) << uint32( 8)) +
        (int32(arr[1]) << uint32( 0))

    return (num)
} // ToInt_2301

//******************************************************************************
// ToInt_3210
/**
** Convert a 4-byte slice into an integer by rearranging the bytes in a specific
** order.  Some MODBUS implementations do not use a consistent byte ordering.
** 3210 represents a Bigendian int32.
**
** @param arr - Byte array to convert.
**
** @return An integer value.
*/

func
ToInt_3210 (arr []byte) int32 {
    sz := len (arr)
    if (sz != 4) {
        return (math.MinInt32) // Closest thing to an error
    } // if

    var num int32 =
        (int32(arr[3]) << uint32(24)) +
        (int32(arr[2]) << uint32(16)) +
        (int32(arr[1]) << uint32( 8)) +
        (int32(arr[0]) << uint32( 0))

    return (num)
} // ToInt_3210

//******************************************************************************
// ToInt_10
/**
** Convert a 2-byte slice into an integer where the bytes are lo/hi.  This mixed
** Endian crap is a major PITA.
**
** @param arr - Byte array to convert.
**
** @return An integer value.
*/

func
ToInt_10 (arr []byte) uint16 {
    sz := len (arr)
    if (sz != 2) {
        return (math.MaxInt16) // Closest thing to an error
    } // if

    var num uint16 =
        (uint16(arr[1]) << uint16( 8)) +
        (uint16(arr[0]) << uint16( 0))

    return (num)
} // ToInt_10

//******************************************************************************
// ToWord_10
/**
** Return the lower two bytes of an int32 in lo/hi order instead of hi/lo.
**
** @param num - The value to convert.
**
** @return Two bytes.
*/

func
ToWord_10 (num int32) (byte, byte) {
    return byte(num), byte(num >> 8)
} // ToWord_10

//******************************************************************************
// ToFloat_2301
/**
** Convert a byte array of 4 bytes into an IEEE 32 Float value.
**
** @param arr - Byte array to convert.
**
** @return A float value.
*/

func
ToFloat_2301 (arr []byte) float32 {
//    pba := unsafe.Pointer (&(arr[0]))
//    f := *(*float32)(pba)

    i := uint32(ToInt_2301 (arr)) // May be Windows/Linux specific in endianness
    f := math.Float32frombits (i)

//fmt.Printf ("f=%f\n", f)
    return (f)
} // ToFloat_2301

//******************************************************************************
// ToASCII
/**
** The Aurora VBINE Inverter has some values that come back as ASCII values.
** Basically, it's a string.
*/

func
ToASCII (arr []byte) string {
//    return (string(arr))
    var s string = ""
    for i := 0; i < len(arr); i++ {
        if (arr[i] > 0) {
//            s = s + fmt.Sprintf ("%c", arr[i])
//            s = s + strconv.Itoa (int(arr[i]))
            s = s + string(arr[i])
        } // if
    } // for
    return (s)
} // ToASCII
