package shasumarith

import (
	"go_dht/constants"
)

type Ord int

const (
	Less    Ord = -1
	Equal   Ord = 0
	Greater Ord = 1
)

/*
Convert an unsigned int n into a Sha Byte Array of length constants.ShaSize representing the int.
Returns the byte array
*/
func FromInt(n uint32) [constants.ShaSize]byte {
	var ret [constants.ShaSize]byte
	const size_int = 4 // num bytes in an uint32
	for i := uint32(0); i < constants.ShaSize; i++ {
		// fill with zeros first
		ret[i] = byte(0)
	}
	for i := uint32(1); i <= size_int; i++ {
		ret[constants.ShaSize-i] = byte(n & 255) // extract the last 8 bits in n
		n = n >> 8                               // 8 bits in a byte
	}
	return ret
}

/*
Returns the highest value possible
*/
func MaxVal() [constants.ShaSize]byte {
	var max_byte byte = 0xff
	var ret [constants.ShaSize]byte
	for i := range ret {
		ret[i] = max_byte
	}
	return ret
}

/*
Returns a/2. Simply does a right shift on the array.
*/
func Div2(a [constants.ShaSize]byte) [constants.ShaSize]byte {
	var shift_out bool = false // true if a bit was shifted out
	var ret [constants.ShaSize]byte
	for i, x := range a {
		if shift_out {
			ret[i] = byte(0x80 | (uint8(x) >> 1))
		} else {
			ret[i] = byte((uint8(x) >> 1))
		}
		shift_out = ((x & 1) > 0) // true if lsb of x is set
	}
	return ret
}

/*
Calculates 2^i and returns the result as byte array. If i >= than ShaNumBits, returns 0
*/
func Pow2(i uint32) [constants.ShaSize]byte {
	if i >= uint32(constants.ShaNumBits) {
		return FromInt(0)
	}
	// calculate which byte to place the set bit
	ret := FromInt(0)
	size_byte := uint32(8)
	byte_i := (160 - 1 - i) / size_byte // index of the byte to set
	bit_i := i % size_byte              // bit within the byte to set
	ret[byte_i] = byte(1) << bit_i
	return ret
}

/*
Returns a + b. If overflow occurs, returns (a+b) % (2^constants.ShaSize) i.e ignores the carry out bit.
Each arg treated as an unsigned number
*/
func Add(a [constants.ShaSize]byte, b [constants.ShaSize]byte) [constants.ShaSize]byte {
	size_byte := uint32(8)
	var c_out uint32 = 0
	var temp uint32 // holds temp sum
	ret := FromInt(0)
	for i := int(constants.ShaSize - 1); i >= 0; i-- {
		temp = uint32(a[i]) + uint32(b[i]) + c_out
		c_out = temp >> size_byte
		ret[i] = byte(temp & 0xff) // get the last 8 bits
	}
	return ret
}

func Not(a [constants.ShaSize]byte) [constants.ShaSize]byte {
	var ret [constants.ShaSize]byte
	for i, x := range a {
		ret[i] = 0xff &^ x // AND(0xff, NOT(ret[i]))
	}
	return ret
}

/*
Returns a - b. Only works if b <= a as a and b are treated as unsigned values
*/
func Sub(a [constants.ShaSize]byte, b [constants.ShaSize]byte) [constants.ShaSize]byte {
	neg_b := Add(Not(b), FromInt(1)) // negative b, i.e 2's complement without the sign bit
	return Add(a, neg_b)
}

/* compares 2 sha arrays.
* Comparison done by comparing every single char from left to right until one is greater than the other
* Returns Less is a < b, Equal if a==b ...
 */
func Cmp(a [constants.ShaSize]byte, b [constants.ShaSize]byte) Ord {
	for i := range a {
		if a[i] < b[i] {
			return Less
		} else if a[i] > b[i] {
			return Greater
		}
	}
	return Equal
}

/*
checks if shasum >= lo and <= hi using ShaCmp. hi must be >= lo i.e wrap arounds not handled
*/
func InRange(shasum [constants.ShaSize]byte, lo [constants.ShaSize]byte, hi [constants.ShaSize]byte) bool {
	return Cmp(shasum, lo) != Less && Cmp(shasum, hi) != Greater
}
