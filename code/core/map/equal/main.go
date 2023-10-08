package main

import (
	"fmt"
	"reflect"
)

func main() {
	var m map[string]int
	var n map[string]int

	fmt.Println(m == nil)
	fmt.Println(n == nil)

	m1 := make(map[string]int, 0)
	m1["a"] = 1
	//m2 := m1
	//fmt.Println(m2 == m1)

	// 不能通过编译
	//fmt.Println(m == n)

	map1 := map[string]int{"a": 1, "b": 2}
	map2 := map[string]int{"b": 2, "a": 1}

	if reflect.DeepEqual(map1, map2) {
		fmt.Println("两个map相等")
	} else {
		fmt.Println("两个map不相等")
	}
}
