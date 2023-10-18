package channel

// 读已关闭
func CloseClosedCh1() {
	ch := make(chan int, 1000)
	close(ch)
	close(ch)
}
