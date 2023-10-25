package main

import (
	"fmt"
)

type data struct {
	name string
}
type printer interface {
	print()
}

// 这里不能使用指针
func (p *data) print() {
	fmt.Println("name: ", p.name)
}

/*
//解决方法一: 正确的语法, 不能使用指针
func (p data) print() {
fmt.Println("name: ", p.name)
}
*/
func main() {
	d1 := data{"one"}
	d1.print() // d1 变量可寻址，可直接调用指针 receiver 的方法
	/*
	   var in printer = data{"two"}
	   in.print() // Error: 类型不匹配
	*/

	//	m := map[string]data{
	//		"x": data{"three"},
	//	}
	//	// 对于 map 存储struct的数据类型, 不能使用指针
	//	m["x"].print() //Error: m["x"] 是不可寻址的 // 变动频繁
	//
	//	// 解决方法二:
	//	r := m["x"]
	//	r.print()
}
