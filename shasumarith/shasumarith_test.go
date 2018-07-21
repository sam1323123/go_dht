package shasumarith

import (
	//"fmt"
	"go_dht/constants"
	"testing"
)

func TestFromInt(t *testing.T) {
	var n uint32 = 255 << 8
	ret := FromInt(n)
	if ret[constants.ShaSize-2] != 255 || ret[constants.ShaSize-1] != 0 {
		t.Errorf("From Int not correct\n")
	}
}

func TestArith(t *testing.T) {
	a := Pow2(uint32(constants.ShaNumBits - 1)) // MSB set
	// fmt.Printf("%v", a)
	b := Div2(a)
	c := Sub(a, b)
	if Cmp(b, c) != Equal {
		t.Errorf("Test 1 failed in TestMain")
	}
}
