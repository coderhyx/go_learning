package main

import (
	"fmt"
	"time"
)

func main() {
	worker()
}

func worker() {
	ticker := time.Tick(3 * time.Second)
	for {
		select {
		case <-ticker:
			// 执行定时任务
			fmt.Println("执行 1s 定时任务")
		}
	}
}
