package fingertable

import (
	"fmt"
	K "go_dht/constants"
	SSA "go_dht/shasumarith"
	"testing"
)

func succFn(key [K.ShaSize]byte) (string, string) {
	return "localhost", "8080"
}

func TestFT(t *testing.T) {
	ft := New(SSA.FromInt(0), succFn)
	for i := uint32(0); i < K.ShaNumBits; i++ {
		fmt.Printf("i = %d, host = %s\n", i, ft.table[i].hostname)
	}
	for i := uint32(0); i < uint32(K.ShaNumBits); i++ {
		host, port, err := ft.Find(SSA.Pow2(i))
		if err != nil {
			t.Errorf("Got wrong error on i = %d, key = %v , host = %s, port = %s\n", i, SSA.Pow2(i), host, port)
		} else {
			fmt.Printf("i = %d, key = %v , host = %s, port = %s\n", i, SSA.Pow2(i), host, port)
		}

	}
	host, port, err := ft.Find(SSA.FromInt(0))
	if err != nil {
		t.Errorf("Got wrong error on key = %v , host = %s, port = %s\n", SSA.FromInt(0), host, port)
	} else {
		fmt.Printf("key = %v , host = %s, port = %s\n", SSA.FromInt(0), host, port)
	}
}
