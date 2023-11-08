package main

import "fmt"

type Student struct {
	Name string
	Age  int
}

type myInt int

var i myInt = 1

func main() {
	//var i interface{} = new(Student)

	var i interface{} = i
	//s := i.(Student)
	//
	fmt.Println(i)
}
