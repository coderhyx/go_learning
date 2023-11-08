package main

import (
	"fmt"
	"unsafe"
)

func main() {

	str := "Hello, World!"
	ptr := *(*uintptr)(unsafe.Pointer(&str))

	// 验证指针是否指向数组
	isArray := isPointerToArray(ptr)
	fmt.Println("Is pointer pointing to an array:", isArray)
}

func isPointerToArray(ptr uintptr) bool {
	// 假设一个字符串的指针指向的数组长度不会超过100
	const maxArrayLen = 100

	// 创建一个指向字节数组的指针
	arrayPtr := (*[maxArrayLen]byte)(unsafe.Pointer(ptr))

	// 验证指针是否指向数组
	return arrayPtr != nil
}
