package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var count int32
	fmt.Println("main start...")
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Println("goroutine:", i, "start...")
			//atomic.AddInt32(&count, 1)
			count = count + 1
			//fmt.Println("goroutine:", i, "count:", count, "end...")
		}(i)
	}
	wg.Wait()
	// 读取最终计数器的值
	finalCount := atomic.LoadInt32(&count)
	fmt.Printf("最终计数器的值: %d\n", finalCount)
	fmt.Println("main end...")
}
