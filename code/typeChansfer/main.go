package main

import (
	"fmt"
	"strconv"
)

func main() {
	var old string = "123"
	age, err := strconv.Atoi(old) //字符串型转换为整型
	if err != nil {
		fmt.Println("Atoi 返回值需要用两个变量接收")
	}
	fmt.Println(age)
}
