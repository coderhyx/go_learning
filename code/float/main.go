package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func main() {
	//var a float64 = 0.2
	//var b float64 = 0.7
	//fmt.Println("======>", a+b)
	// 将浮点数转换为高精度表示
	// 将浮点数转换为decimal类型
	a := decimal.NewFromFloat(0.2)
	b := decimal.NewFromFloat(0.7)

	// 计算decimal类型的和
	sum := a.Add(b)
	// 输出结果
	fmt.Println("0.2 + 0.7 =", sum)
}
