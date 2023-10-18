package channel

import (
	"fmt"
	"time"
)

// 读已关闭
func ReadCloseCh1() {
	ch := make(chan int, 1000)
	close(ch)
	ch <- 1
}

func ReadCloseCh2() {
	ch := make(chan int, 1000)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
		}
	}()
	go func() {
		for {
			a, ok := <-ch
			if !ok {
				fmt.Println("close")
				return
			}
			fmt.Println("a: ", a)
		}
	}()
	close(ch)
	fmt.Println("ok")
	time.Sleep(time.Second * 100)
}
