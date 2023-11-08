package main

import (
	"fmt"
	"unsafe"
)

var o *int

func main() {

	n := 1
	o = &n
	var s1 byte = 'a'
	//s2 := []byte(s1)
	sizeType(s1)

	var s2 rune = 'a'

	//s3 := []rune(s1)
	sizeRune(s2)
}

func sizeType(s byte) {
	fmt.Println(unsafe.Sizeof(s))
}

func sizeTypeSlice(s []byte) {
	fmt.Println(unsafe.Sizeof(s))
}

func sizeRune(s rune) {
	fmt.Println(unsafe.Sizeof(s))
}
func sizeRuneSlice(s []rune) {
	fmt.Println(unsafe.Sizeof(s))
}
