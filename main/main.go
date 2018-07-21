package main

import (
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
)

func shaSumToString(data [sha1.Size]byte) string {
	/*hexMap := map[byte]string{0: "0", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5",
	6: "6", 7: "7", 8: "8", 9: "9", 10: "a", 11: "b", 12: "c", 13: "d", 14: "e", 15: "f"}
	*/
	var ret []string
	for i := range data {
		fmt.Printf("i = %d, data = %d, str=%s\n", i, int(data[i]), strconv.Itoa(int(data[i])))
		ret = append(ret, strconv.Itoa(int(data[i])))
	}
	return strings.Join(ret, "")

}

func main() {
	b_arr := []byte("abc")
	fmt.Printf("len=%d cap=%d word=%s\n", len(b_arr), cap(b_arr), b_arr)
	fmt.Printf("Slice = %s\n", b_arr[0:2])
	fmt.Printf("Hash of \"%s\": %s\n", b_arr, shaSumToString(sha1.Sum(b_arr)))
}
