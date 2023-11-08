package main

import (
	"fmt"
	"time"
)

func sendData(ch chan<- int) {
	// 向只写通道发送数据
	for i := 0; i < 5; i++ {
		ch <- i
		time.Sleep(time.Second)
	}
	// 关闭只写通道，表示发送结束
	close(ch)
}

func main() {
	ch := make(chan int)

	// 启动一个 goroutine 发送数据到只写通道
	go sendData(ch)

	// 主 goroutine 等待数据并打印
	for value := range ch {
		fmt.Println("Received:", value)
	}
}
