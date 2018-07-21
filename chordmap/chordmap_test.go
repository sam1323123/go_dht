package chordmap

import (
	"fmt"
	K "go_dht/constants"
	SSA "go_dht/shasumarith"
	"testing"
)

/*
func TestShaSum(t *testing.T) {
	a := [ShaSize]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	exp := ""
	res := ""
	for range a {
		exp += "000"
	}
	res = ShaSumToStr(a)
	if res != exp {
		t.Errorf("ShaSum(%v) = %s", a, res)
	}
	a = [ShaSize]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100}
	exp = exp[0:57] + "100"
	res = ShaSumToStr(a)
	if res != exp {
		t.Errorf("ShaSum(%v) = %s", a, res)
	}
}
*/

func TestRange(t *testing.T) {
	start := SSA.MaxVal()
	end := SSA.FromInt(0)
	if !InRangeHelp(start, start, end) {
		t.Error("Range test 1 failed")
	}
	if InRangeHelp(end, start, end) {
		t.Error("Range test 2 failed")
	}
	start = SSA.Sub(SSA.MaxVal(), SSA.FromInt(1))
	end = SSA.FromInt(2)
	if !InRangeHelp(SSA.FromInt(1), start, end) {
		t.Error("Range test 3 failed")
	}
	if !InRangeHelp(SSA.MaxVal(), start, end) {
		t.Error("Range test 4 failed")
	}
	start = SSA.FromInt(0)
	end = SSA.Sub(SSA.MaxVal(), SSA.FromInt(1))
	if !InRangeHelp(SSA.Pow2(159), start, end) {
		t.Error("Range test 5 failed")
	}
	if InRangeHelp(SSA.MaxVal(), start, end) {
		t.Error("Range test 6 failed")
	}

}

func TestHashTable(t *testing.T) {
	cms := New([K.ShaSize]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[K.ShaSize]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	err := cms.Put("key", "value")
	if err != nil {
		t.Errorf("Should not Get %s for key = %s, value = %s\n", err.Error(), "key", "value")
	}
	value, err := cms.Get("key")
	if value != "value" {
		t.Errorf("Value should be %s not %s\n", "value", value)
	}
	cms.Delete("key")
	value, err = cms.Get("key")
	switch err.(type) {
	case *CMRangeError:
		t.Errorf("Should not get range error")
	case *CMKeyError:
		fmt.Println("PASSED. Correct Error type.")
	default:
		break

	}
}
