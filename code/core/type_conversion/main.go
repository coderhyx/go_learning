package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

func main() {
	b1 := string2bytes("aaaa")
	t := reflect.TypeOf(b1)
	fmt.Println(t)

	m := map[string]int{"a": 1, "b": 2, "c": 3}
	length := len(m)
	fmt.Println("map的长度为：", length)
}

func string2bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
