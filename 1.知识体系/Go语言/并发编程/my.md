# 2023-8-27 刘常青
## go编程入门

## 并发编程
# 并发同步概述
什么是并发同步？
并发同步是指如何控制若干并发计算（在Go中，即协程），从而
避免在它们之间产生数据竞争的现象；
避免在它们无所事事的时候消耗CPU资源。
并发同步有时候也称为数据同步。

# 通道用例大全
本文余下的内容将展示很多通道用例。 希望这篇文章能够说服你接收下面的观点：
使用通道进行异步和并发编程是简单和惬意的；
通道同步技术比被很多其它语言采用的其它同步方案（比如角色模型和async/await模式）有着更多的应用场景和更多的使用变种。
请注意，本文的目的是展示尽量多的通道用例。但是，我们应该知道通道并不是Go支持的唯一同步技术，并且通道并不是在任何情况下都是最佳的同步技术。 请阅读原子操作和其它并发同步技术来了解更多的Go支持的同步技术。

将通道用做future/promise
很多其它流行语言支持future/promise来实现异步（并发）编程。 Future/promise常常用在请求/回应场合。

返回单向接收通道做为函数返回结果
在下面这个例子中，sumSquares函数调用的两个实参请求并发进行。 每个通道读取操作将阻塞到请求返回结果为止。 两个实参总共需要大约3秒钟（而不是6秒钟）准备完毕（以较慢的一个为准）。

package main

import (
"time"
"math/rand"
"fmt"
)

func longTimeRequest() <-chan int32 {
r := make(chan int32)

	go func() {
		time.Sleep(time.Second * 3) // 模拟一个工作负载
		r <- rand.Int31n(100)
	}()

	return r
}

func sumSquares(a, b int32) int32 {
return a*a + b*b
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	a, b := longTimeRequest(), longTimeRequest()
	fmt.Println(sumSquares(<-a, <-b))
}
将单向发送通道类型用做函数实参
和上例一样，在下面这个例子中，sumSquares函数调用的两个实参的请求也是并发进行的。 和上例不同的是longTimeRequest函数接收一个单向发送通道类型参数而不是返回一个单向接收通道结果。

package main

import (
"time"
"math/rand"
"fmt"
)

func longTimeRequest(r chan<- int32)  {
time.Sleep(time.Second * 3) // 模拟一个工作负载
r <- rand.Int31n(100)
}

func sumSquares(a, b int32) int32 {
return a*a + b*b
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	ra, rb := make(chan int32), make(chan int32)
	go longTimeRequest(ra)
	go longTimeRequest(rb)

	fmt.Println(sumSquares(<-ra, <-rb))
}
对于上面这个特定的例子，我们可以只使用一个通道来接收回应结果，因为两个参数的作用是对等的。
...

	results := make(chan int32, 2) // 缓冲与否不重要
	go longTimeRequest(results)
	go longTimeRequest(results)

	fmt.Println(sumSquares(<-results, <-results))
}
这可以看作是后面将要提到的数据聚合的一个应用。

采用最快回应
本用例可以看作是上例中只使用一个通道变种的增强。

有时候，一份数据可能同时从多个数据源获取。这些数据源将返回相同的数据。 因为各种因素，这些数据源的回应速度参差不一，甚至某个特定数据源的多次回应速度之间也可能相差很大。 同时从多个数据源获取一份相同的数据可以有效保障低延迟。我们只需采用最快的回应并舍弃其它较慢回应。

注意：如果有N个数据源，为了防止被舍弃的回应对应的协程永久阻塞，则传输数据用的通道必须为一个容量至少为N-1的缓冲通道。

package main

import (
"fmt"
"time"
"math/rand"
)

func source(c chan<- int32) {
ra, rb := rand.Int31(), rand.Intn(3) + 1
// 睡眠1秒/2秒/3秒
time.Sleep(time.Duration(rb) * time.Second)
c <- ra
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	startTime := time.Now()
	c := make(chan int32, 5) // 必须用一个缓冲通道
	for i := 0; i < cap(c); i++ {
		go source(c)
	}
	rnd := <- c // 只有第一个回应被使用了
	fmt.Println(time.Since(startTime))
	fmt.Println(rnd)
}
“采用最快回应”用例还有一些其它实现方式，本文后面将会谈及。

更多“请求/回应”用例变种
做为函数参数和返回结果使用的通道可以是缓冲的，从而使得请求协程不需阻塞到它所发送的数据被接收为止。

有时，一个请求可能并不保证返回一份有效的数据。对于这种情形，我们可以使用一个形如struct{v T; err error}的结构体类型或者一个空接口类型做为通道的元素类型以用来区分回应的值是否有效。

有时，一个请求可能需要比预期更长的用时才能回应，甚至永远都得不到回应。 我们可以使用本文后面将要介绍的超时机制来应对这样的情况。

有时，回应方可能会不断地返回一系列值，这也同时属于后面将要介绍的数据流的一个用例。

使用通道实现通知
通知可以被看作是特殊的请求/回应用例。在一个通知用例中，我们并不关心回应的值，我们只关心回应是否已发生。 所以我们常常使用空结构体类型struct{}来做为通道的元素类型，因为空结构体类型的尺寸为零，能够节省一些内存（虽然常常很少量）。

向一个通道发送一个值来实现单对单通知
我们已知道，如果一个通道中无值可接收，则此通道上的下一个接收操作将阻塞到另一个协程发送一个值到此通道为止。 所以一个协程可以向此通道发送一个值来通知另一个等待着从此通道接收数据的协程。

在下面这个例子中，通道done被用来做为一个信号通道来实现单对单通知。
package main

import (
"crypto/rand"
"fmt"
"os"
"sort"
)

func main() {
values := make([]byte, 32 * 1024 * 1024)
if _, err := rand.Read(values); err != nil {
fmt.Println(err)
os.Exit(1)
}

	done := make(chan struct{}) // 也可以是缓冲的

	// 排序协程
	go func() {
		sort.Slice(values, func(i, j int) bool {
			return values[i] < values[j]
		})
		done <- struct{}{} // 通知排序已完成
	}()

	// 并发地做一些其它事情...

	<- done // 等待通知
	fmt.Println(values[0], values[len(values)-1])
}
从一个通道接收一个值来实现单对单通知
如果一个通道的数据缓冲队列已满（非缓冲的通道的数据缓冲队列总是满的）但它的发送协程队列为空，则向此通道发送一个值将阻塞，直到另外一个协程从此通道接收一个值为止。 所以我们可以通过从一个通道接收数据来实现单对单通知。一般我们使用非缓冲通道来实现这样的通知。

这种通知方式不如上例中介绍的方式使用得广泛。

package main

import (
"fmt"
"time"
)

func main() {
done := make(chan struct{})
// 此信号通道也可以缓冲为1。如果这样，则在下面
// 这个协程创建之前，我们必须向其中写入一个值。

	go func() {
		fmt.Print("Hello")
		// 模拟一个工作负载。
		time.Sleep(time.Second * 2)

		// 使用一个接收操作来通知主协程。
		<- done
	}()

	done <- struct{}{} // 阻塞在此，等待通知
	fmt.Println(" world!")
}
另一个事实是，上面的两种单对单通知方式其实并没有本质的区别。 它们都可以被概括为较快者等待较慢者发出通知。

多对单和单对多通知
略微扩展一下上面两个用例，我们可以很轻松地实现多对单和单对多通知。
package main

import "log"
import "time"

type T = struct{}

func worker(id int, ready <-chan T, done chan<- T) {
<-ready // 阻塞在此，等待通知
log.Print("Worker#", id, "开始工作")
// 模拟一个工作负载。
time.Sleep(time.Second * time.Duration(id+1))
log.Print("Worker#", id, "工作完成")
done <- T{} // 通知主协程（N-to-1）
}

func main() {
log.SetFlags(0)

	ready, done := make(chan T), make(chan T)
	go worker(0, ready, done)
	go worker(1, ready, done)
	go worker(2, ready, done)

	// 模拟一个初始化过程
	time.Sleep(time.Second * 3 / 2)
	// 单对多通知
	ready <- T{}; ready <- T{}; ready <- T{}
	// 等待被多对单通知
	<-done; <-done; <-done
}
事实上，上例中展示的多对单和单对多通知实现方式在实践中用的并不多。 在实践中，我们多使用sync.WaitGroup来实现多对单通知，使用关闭一个通道的方式来实现单对多通知（详见下一个用例）。

通过关闭一个通道来实现群发通知
上一个用例中的单对多通知实现在实践中很少用，因为通过关闭一个通道的方式在来实现单对多通知的方式更简单。 我们已经知道，从一个已关闭的通道可以接收到无穷个值，我们可以利用这一特性来实现群发通知。

我们可以把上一个例子中的三个数据发送操作ready <- struct{}{}替换为一个通道关闭操作close(ready)来达到同样的单对多通知效果。
...
close(ready) // 群发通知
...
当然，我们也可以通过关闭一个通道来实现单对单通知。事实上，关闭通道是实践中用得最多通知实现方式。

从一个已关闭的通道可以接收到无穷个值这一特性也将被用在很多其它在后面将要介绍的用例中。 实际上，这一特性被广泛地使用于标准库包中。比如，context标准库包使用了此特性来传达操作取消消息。

定时通知（timer）
用通道实现一个一次性的定时通知器是很简单的。 下面是一个自定义实现：
package main

import (
"fmt"
"time"
)

func AfterDuration(d time.Duration) <- chan struct{} {
c := make(chan struct{}, 1)
go func() {
time.Sleep(d)
c <- struct{}{}
}()
return c
}

func main() {
fmt.Println("Hi!")
<- AfterDuration(time.Second)
fmt.Println("Hello!")
<- AfterDuration(time.Second)
fmt.Println("Bye!")
}
事实上，time标准库包中的After函数提供了和上例中AfterDuration同样的功能。 在实践中，我们应该尽量使用time.After函数以使代码看上去更干净。

注意，操作<-time.After(aDuration)将使当前协程进入阻塞状态，而一个time.Sleep(aDuration)函数调用不会如此。

<-time.After(aDuration)经常被使用在后面将要介绍的超时机制实现中。

将通道用做互斥锁（mutex）
上面的某个例子提到了容量为1的缓冲通道可以用做一次性二元信号量。 事实上，容量为1的缓冲通道也可以用做多次性二元信号量（即互斥锁）尽管这样的互斥锁效率不如sync标准库包中提供的互斥锁高效。

有两种方式将一个容量为1的缓冲通道用做互斥锁：
通过发送操作来加锁，通过接收操作来解锁；
通过接收操作来加锁，通过发送操作来解锁。
下面是一个通过发送操作来加锁的例子。
package main

import "fmt"

func main() {
mutex := make(chan struct{}, 1) // 容量必须为1

	counter := 0
	increase := func() {
		mutex <- struct{}{} // 加锁
		counter++
		<-mutex // 解锁
	}

	increase1000 := func(done chan<- struct{}) {
		for i := 0; i < 1000; i++ {
			increase()
		}
		done <- struct{}{}
	}

	done := make(chan struct{})
	go increase1000(done)
	go increase1000(done)
	<-done; <-done
	fmt.Println(counter) // 2000
}
下面是一个通过接收操作来加锁的例子，其中只显示了相对于上例而修改了的部分。
...
func main() {
mutex := make(chan struct{}, 1)
mutex <- struct{}{} // 此行是必需的

	counter := 0
	increase := func() {
		<-mutex // 加锁
		counter++
		mutex <- struct{}{} // 解锁
	}
...
将通道用做计数信号量（counting semaphore）
缓冲通道可以被用做计数信号量。 计数信号量可以被视为多主锁。如果一个缓冲通道的容量为N，那么它可以被看作是一个在任何时刻最多可有N个主人的锁。 上面提到的二元信号量是特殊的计数信号量，每个二元信号量在任一时刻最多只能有一个主人。

计数信号量经常被使用于限制最大并发数。

和将通道用做互斥锁一样，也有两种方式用来获取一个用做计数信号量的通道的一份所有权。
通过发送操作来获取所有权，通过接收操作来释放所有权；
通过接收操作来获取所有权，通过发送操作来释放所有权。
下面是一个通过接收操作来获取所有权的例子：
package main

import (
"log"
"time"
"math/rand"
)

type Seat int
type Bar chan Seat

func (bar Bar) ServeCustomer(c int) {
log.Print("顾客#", c, "进入酒吧")
seat := <- bar // 需要一个位子来喝酒
log.Print("++ customer#", c, " drinks at seat#", seat)
log.Print("++ 顾客#", c, "在第", seat, "个座位开始饮酒")
time.Sleep(time.Second * time.Duration(2 + rand.Intn(6)))
log.Print("-- 顾客#", c, "离开了第", seat, "个座位")
bar <- seat // 释放座位，离开酒吧
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	bar24x7 := make(Bar, 10) // 此酒吧有10个座位
	// 摆放10个座位。
	for seatId := 0; seatId < cap(bar24x7); seatId++ {
		bar24x7 <- Seat(seatId) // 均不会阻塞
	}

	for customerId := 0; ; customerId++ {
		time.Sleep(time.Second)
		go bar24x7.ServeCustomer(customerId)
	}
	for {time.Sleep(time.Second)} // 睡眠不属于阻塞状态
}
在上例中，只有获得一个座位的顾客才能开始饮酒。 所以在任一时刻同时在喝酒的顾客数不会超过座位数10。

上例main函数中的最后一行for循环是为了防止程序退出。 后面将介绍一种更好的实现此目的的方法。

在上例中，尽管在任一时刻同时在喝酒的顾客数不会超过座位数10，但是在某一时刻可能有多于10个顾客进入了酒吧，因为某些顾客在排队等位子。 在上例中，每个顾客对应着一个协程。虽然协程的开销比系统线程小得多，但是如果协程的数量很多，则它们的总体开销还是不能忽略不计的。 所以，最好当有空位的时候才创建顾客协程。
... // 省略了和上例相同的代码

func (bar Bar) ServeCustomerAtSeat(c int, seat Seat) {
log.Print("++ 顾客#", c, "在第", seat, "个座位开始饮酒")
time.Sleep(time.Second * time.Duration(2 + rand.Intn(6)))
log.Print("-- 顾客#", c, "离开了第", seat, "个座位")
bar <- seat // 释放座位，离开酒吧
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	bar24x7 := make(Bar, 10)
	for seatId := 0; seatId < cap(bar24x7); seatId++ {
		bar24x7 <- Seat(seatId)
	}

	// 这个for循环和上例不一样。
	for customerId := 0; ; customerId++ {
		time.Sleep(time.Second)
		seat := <- bar24x7 // 需要一个空位招待顾客
		go bar24x7.ServeCustomerAtSeat(customerId, seat)
	}
	for {time.Sleep(time.Second)}
}
在上面这个修改后的例子中，在任一时刻最多只有10个顾客协程在运行（但是在程序的生命期内，仍旧会有大量的顾客协程不断被创建和销毁）。

在下面这个更加高效的实现中，在程序的生命期内最多只会有10个顾客协程被创建出来。
... // 省略了和上例相同的代码

func (bar Bar) ServeCustomerAtSeat(consumers chan int) {
for c := range consumers {
seatId := <- bar
log.Print("++ 顾客#", c, "在第", seatId, "个座位开始饮酒")
time.Sleep(time.Second * time.Duration(2 + rand.Intn(6)))
log.Print("-- 顾客#", c, "离开了第", seatId, "个座位")
bar <- seatId // 释放座位，离开酒吧
}
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	bar24x7 := make(Bar, 10)
	for seatId := 0; seatId < cap(bar24x7); seatId++ {
		bar24x7 <- Seat(seatId)
	}

	consumers := make(chan int)
	for i := 0; i < cap(bar24x7); i++ {
		go bar24x7.ServeCustomerAtSeat(consumers)
	}
	
	for customerId := 0; ; customerId++ {
		time.Sleep(time.Second)
		consumers <- customerId
	}
}
题外话：当然，如果我们并不关心座位号（这种情况在编程实践中很常见），则实际上bar24x7计数信号量是完全不需要的：
... // 省略了和上例相同的代码

func ServeCustomer(consumers chan int) {
for c := range consumers {
log.Print("++ 顾客#", c, "开始在酒吧饮酒")
time.Sleep(time.Second * time.Duration(2 + rand.Intn(6)))
log.Print("-- 顾客#", c, "离开了酒吧")
}
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	const BarSeatCount = 10
	consumers := make(chan int)
	for i := 0; i < BarSeatCount; i++ {
		go ServeCustomer(consumers)
	}
	
	for customerId := 0; ; customerId++ {
		time.Sleep(time.Second)
		consumers <- customerId
	}
}
通过发送操作来获取所有权的实现相对简单一些，省去了摆放座位的步骤。
package main

import (
"log"
"time"
"math/rand"
)

type Customer struct{id int}
type Bar chan Customer

func (bar Bar) ServeCustomer(c Customer) {
log.Print("++ 顾客#", c.id, "开始饮酒")
time.Sleep(time.Second * time.Duration(3 + rand.Intn(16)))
log.Print("-- 顾客#", c.id, "离开酒吧")
<- bar // 离开酒吧，腾出位子
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	bar24x7 := make(Bar, 10) // 最对同时服务10位顾客
	for customerId := 0; ; customerId++ {
		time.Sleep(time.Second * 2)
		customer := Customer{customerId}
		bar24x7 <- customer // 等待进入酒吧
		go bar24x7.ServeCustomer(customer)
	}
	for {time.Sleep(time.Second)}
}
对话（或称乒乓）
两个协程可以通过一个通道进行对话，整个过程宛如打乒乓球一样。 下面是一个这样的例子，它将打印出一系列斐波那契（Fibonacci）数。
package main

import "fmt"
import "time"
import "os"

type Ball uint64

func Play(playerName string, table chan Ball) {
var lastValue Ball = 1
for {
ball := <- table // 接球
fmt.Println(playerName, ball)
ball += lastValue
if ball < lastValue { // 溢出结束
os.Exit(0)
}
lastValue = ball
table <- ball // 回球
time.Sleep(time.Second)
}
}

func main() {
table := make(chan Ball)
go func() {
table <- 1 // （裁判）发球
}()
go Play("A:", table)
Play("B:", table)
}
使用通道传送传输通道
一个通道类型的元素类型可以是另一个通道类型。 在下面这个例子中， 单向发送通道类型chan<- int是另一个通道类型chan chan<- int的元素类型。
package main

import "fmt"

var counter = func (n int) chan<- chan<- int {
requests := make(chan chan<- int)
go func() {
for request := range requests {
if request == nil {
n++ // 递增计数
} else {
request <- n // 返回当前计数
}
}
}()
return requests // 隐式转换到类型chan<- (chan<- int)
}(0)

func main() {
increase1000 := func(done chan<- struct{}) {
for i := 0; i < 1000; i++ {
counter <- nil
}
done <- struct{}{}
}

	done := make(chan struct{})
	go increase1000(done)
	go increase1000(done)
	<-done; <-done

	request := make(chan int, 1)
	counter <- request
	fmt.Println(<-request) // 2000
}
尽管对于上面这个用例来说，使用通道传送传输通道这种方式并非是最有效的实现方式，但是这种方式肯定有最适合它的用武之地。

检查通道的长度和容量
我们可以使用内置函数cap和len来查看一个通道的容量和当前长度。 但是在实践中我们很少这样做。我们很少使用内置函数cap的原因是一个通道的容量常常是已知的或者不重要的。 我们很少使用内置函数len的原因是一个len调用的结果并不能总能准确地反映出的一个通道的当前长度。

但有时确实有一些场景需要调用这两个函数。比如，有时一个协程欲将一个未关闭的并且不会再向其中发送数据的缓冲通道中的所有数据接收出来，在确保只有此一个协程从此通道接收数据的情况下，我们可以用下面的代码来实现之：
for len(c) > 0 {
value := <-c
// 使用value ...
}
我们也可以用本文后面将要介绍的尝试接收机制来实现这一需求。两者的运行效率差不多，但尝试接收机制的优点是多个协程可以并发地进行读取操作。

有时一个协程欲将一个缓冲通道写满而又不阻塞，在确保只有此一个协程向此通道发送数据的情况下，我们可以用下面的代码实现这一目的：
for len(c) < cap(c) {
c <- aValue
}
当然，我们也可以使用后面将要介绍的尝试发送机制来实现这一需求。

使当前协程永久阻塞
Go中的选择机制（select）是一个非常独特的特性。它给并发编程带来了很多新的模式和技巧。

我们可以用一个无分支的select流程控制代码块使当前协程永久处于阻塞状态。 这是select流程控制的最简单的应用。 事实上，上面很多例子中的for {time.Sleep(time.Second)}都可以换为select{}。

一般，select{}用在主协程中以防止程序退出。

一个例子：
package main

import "runtime"

func DoSomething() {
for {
// 做点什么...

		runtime.Gosched() // 防止本协程霸占CPU不放
	}
}

func main() {
go DoSomething()
go DoSomething()
select{}
}
顺便说一句，另外还有一些使当前协程永久阻塞的方法，但是select{}是最简单的方法。

尝试发送和尝试接收
含有一个default分支和一个case分支的select代码块可以被用做一个尝试发送或者尝试接收操作，取决于case关键字后跟随的是一个发送操作还是一个接收操作。
如果case关键字后跟随的是一个发送操作，则此select代码块为一个尝试发送操作。 如果case分支的发送操作是阻塞的，则default分支将被执行，发送失败；否则发送成功，case分支得到执行。
如果case关键字后跟随的是一个接收操作，则此select代码块为一个尝试接收操作。 如果case分支的接收操作是阻塞的，则default分支将被执行，接收失败；否则接收成功，case分支得到执行。
尝试发送和尝试接收代码块永不阻塞。

标准编译器对尝试发送和尝试接收代码块做了特别的优化，使得它们的执行效率比多case分支的普通select代码块执行效率高得多。

下例演示了尝试发送和尝试接收代码块的工作原理。
package main

import "fmt"

func main() {
type Book struct{id int}
bookshelf := make(chan Book, 3)

	for i := 0; i < cap(bookshelf) * 2; i++ {
		select {
		case bookshelf <- Book{id: i}:
			fmt.Println("成功将书放在书架上", i)
		default:
			fmt.Println("书架已经被占满了")
		}
	}

	for i := 0; i < cap(bookshelf) * 2; i++ {
		select {
		case book := <-bookshelf:
			fmt.Println("成功从书架上取下一本书", book.id)
		default:
			fmt.Println("书架上已经没有书了")
		}
	}
}
输出结果：
成功将书放在书架上 0
成功将书放在书架上 1
成功将书放在书架上 2
书架已经被占满了
书架已经被占满了
书架已经被占满了
成功从书架上取下一本书 0
成功从书架上取下一本书 1
成功从书架上取下一本书 2
书架上已经没有书了
书架上已经没有书了
书架上已经没有书了
后面的很多用例还要用到尝试发送和尝试接收代码块。

无阻塞地检查一个通道是否已经关闭
假设我们可以保证没有任何协程会向一个通道发送数据，则我们可以使用下面的代码来（并发安全地）检查此通道是否已经关闭，此检查不会阻塞当前协程。
func IsClosed(c chan T) bool {
select {
case <-c:
return true
default:
}
return false
}
此方法常用来查看某个期待中的通知是否已经来临。此通知将由另一个协程通过关闭一个通道来发送。

峰值限制（peak/burst limiting）
将通道用做计数信号量用例和通道尝试（发送或者接收）操作结合起来可用实现峰值限制。 峰值限制的目的是防止过大的并发请求数。

下面是对将通道用做计数信号量一节中的最后一个例子的简单修改，从而使得顾客不再等待而是离去或者寻找其它酒吧。
...
bar24x7 := make(Bar, 10) // 此酒吧只能同时招待10个顾客
for customerId := 0; ; customerId++ {
time.Sleep(time.Second)
consumer := Consumer{customerId}
select {
case bar24x7 <- consumer: // 试图进入此酒吧
go bar24x7.ServeConsumer(consumer)
default:
log.Print("顾客#", customerId, "不愿等待而离去")
}
}
...
另一种“采用最快回应”的实现方式
在上面的“采用最快回应”用例一节已经提到，我们也可以使用选择机制来实现“采用最快回应”用例。 每个数据源协程只需使用一个缓冲为1的通道并向其尝试发送回应数据即可。示例代码如下：
package main

import (
"fmt"
"math/rand"
"time"
)

func source(c chan<- int32) {
ra, rb := rand.Int31(), rand.Intn(3)+1
// 休眠1秒/2秒/3秒
time.Sleep(time.Duration(rb) * time.Second)
select {
case c <- ra:
default:
}
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	c := make(chan int32, 1) // 此通道容量必须至少为1
	for i := 0; i < 5; i++ {
		go source(c)
	}
	rnd := <-c // 只采用第一个成功发送的回应数据
	fmt.Println(rnd)
}
注意，使用选择机制来实现“采用最快回应”的代码中使用的通道的容量必须至少为1，以保证最快回应总能够发送成功。 否则，如果数据请求者因为种种原因未及时准备好接收，则所有回应者的尝试发送都将失败，从而所有回应的数据都将被错过。

第三种“采用最快回应”的实现方式
如果一个“采用最快回应”用例中的数据源的数量很少，比如两个或三个，我们可以让每个数据源使用一个单独的缓冲通道来回应数据，然后使用一个select代码块来同时接收这三个通道。 示例代码如下：
package main

import (
"fmt"
"math/rand"
"time"
)

func source() <-chan int32 {
c := make(chan int32, 1) // 必须为一个缓冲通道
go func() {
ra, rb := rand.Int31(), rand.Intn(3)+1
time.Sleep(time.Duration(rb) * time.Second)
c <- ra
}()
return c
}

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	var rnd int32
	// 阻塞在此直到某个数据源率先回应。
	select{
	case rnd = <-source():
	case rnd = <-source():
	case rnd = <-source():
	}
	fmt.Println(rnd)
}
注意：如果上例中使用的通道是非缓冲的，未被选中的case分支对应的两个source函数调用中开辟的协程将处于永久阻塞状态，从而造成内存泄露。

本小节和上一小节中展示的两种方法也可以用来实现多对单通知。

超时机制（timeout）
在一些请求/回应用例中，一个请求可能因为种种原因导致需要超出预期的时长才能得到回应，有时甚至永远得不到回应。 对于这样的情形，我们可以使用一个超时方案给请求者返回一个错误信息。 使用选择机制可以很轻松地实现这样的一个超时方案。

下面这个例子展示了如何实现一个支持超时设置的请求：
func requestWithTimeout(timeout time.Duration) (int, error) {
c := make(chan int)
go doRequest(c) // 可能需要超出预期的时长回应

	select {
	case data := <-c:
		return data, nil
	case <-time.After(timeout):
		return 0, errors.New("超时了！")
	}
}
脉搏器（ticker）
我们可以使用尝试发送操作来实现一个每隔一定时间发送一个信号的脉搏器。
package main

import "fmt"
import "time"

func Tick(d time.Duration) <-chan struct{} {
c := make(chan struct{}, 1) // 容量最好为1
go func() {
for {
time.Sleep(d)
select {
case c <- struct{}{}:
default:
}
}
}()
return c
}

func main() {
t := time.Now()
for range Tick(time.Second) {
fmt.Println(time.Since(t))
}
}
事实上，time标准库包中的Tick函数提供了同样的功能，但效率更高。 我们应该尽量使用标准库包中的实现。

速率限制（rate limiting）
上面已经展示了如何使用尝试发送实现峰值限制。 同样地，我们也可以使用使用尝试机制来实现速率限制，但需要前面刚提到的定时器实现的配合。 速率限制常用来限制吞吐和确保在一段时间内的资源使用不会超标。

下面的例子借鉴了官方Go维基中的例子。 在此例中，任何一分钟时段内处理的请求数不会超过200。
package main

import "fmt"
import "time"

type Request interface{}
func handle(r Request) {fmt.Println(r.(int))}

const RateLimitPeriod = time.Minute
const RateLimit = 200 // 任何一分钟内最多处理200个请求

func handleRequests(requests <-chan Request) {
quotas := make(chan time.Time, RateLimit)

	go func() {
		tick := time.NewTicker(RateLimitPeriod / RateLimit)
		defer tick.Stop()
		for t := range tick.C {
			select {
			case quotas <- t:
			default:
			}
		}
	}()

	for r := range requests {
		<-quotas
		go handle(r)
	}
}

func main() {
requests := make(chan Request)
go handleRequests(requests)
// time.Sleep(time.Minute)
for i := 0; ; i++ {requests <- i}
}
上例的代码虽然可以保证任何一分钟时段内处理的请求数不会超过200，但是如果在开始的一分钟内没有任何请求，则接下来的某个瞬时时间点可能会同时处理最多200个请求（试着将time.Sleep行的注释去掉看看）。 这可能会造成卡顿情况。我们可以将速率限制和峰值限制一并使用来避免出现这样的情况。

开关
通道一文提到了向一个nil通道发送数据或者从中接收数据都属于阻塞操作。 利用这一事实，我们可以将一个select流程控制中的case操作中涉及的通道设置为不同的值，以使此select流程控制选择执行不同的分支。

下面是另一个乒乓模拟游戏的实现。此实现使用了选择机制。在此例子中，两个case操作中的通道有且只有一个为nil，所以只能是不为nil的通道对应的分支被选中。 每个循环步将对调这两个case操作中的通道，从而改变两个分支的可被选中状态。
package main

import "fmt"
import "time"
import "os"

type Ball uint8
func Play(playerName string, table chan Ball, serve bool) {
var receive, send chan Ball
if serve {
receive, send = nil, table
} else {
receive, send = table, nil
}
var lastValue Ball = 1
for {
select {
case send <- lastValue:
case value := <- receive:
fmt.Println(playerName, value)
value += lastValue
if value < lastValue { // 溢出了
os.Exit(0)
}
lastValue = value
}
receive, send = send, receive // 开关切换
time.Sleep(time.Second)
}
}

func main() {
table := make(chan Ball)
go Play("A:", table, false)
Play("B:", table, true)
}
下面是另一个也展示了开关效果的但简单得多的（非并发的）小例子。 此程序将不断打印出1212...。 它在实践中没有太多实用价值，这里只是为了学习的目的才展示之。
package main

import "fmt"
import "time"

func main() {
for c := make(chan struct{}, 1); true; {
select {
case c <- struct{}{}:
fmt.Print("1")
case <-c:
fmt.Print("2")
}
time.Sleep(time.Second)
}
}
控制代码被执行的几率
我们可以通过在一个select流程控制中使用重复的case操作来增加对应分支中的代码的执行几率。

一个例子：
package main

import "fmt"

func main() {
foo, bar := make(chan struct{}), make(chan struct{})
close(foo); close(bar) // 仅为演示目的
x, y := 0.0, 0.0
f := func(){x++}
g := func(){y++}
for i := 0; i < 100000; i++ {
select {
case <-foo: f()
case <-foo: f()
case <-bar: g()
}
}
fmt.Println(x/y) // 大致为2
}
在上面这个例子中，函数f的调用执行几率大致为函数g的两倍。

从动态数量的分支中选择
每个select控制流程中的分支数量在运行中是固定的，但是我们可以使用reflect标准库包中提供的功能在运行时刻来构建动态分支数量的select控制流程。 但是请注意：一个select控制流程中的分支越多，此select控制流程的执行效率就越低（这是我们常常只使用不多于三个分支的select控制流程的原因）。

reflect标准库包中也提供了模拟尝试发送和尝试接收代码块的TrySend和TryRecv函数。

数据流操纵
本节将介绍一些使用通道进行数据流处理的用例。

一般来说，一个数据流处理程序由多个模块组成。不同的模块执行分配给它们的不同的任务。 每个模块由一个或者数个并行工作的协程组成。实践中常见的工作任务包括：
数据生成/搜集/加载；
数据服务/存盘；
数据计算/处理；
数据验证/过滤；
数据聚合/分流；
数据组合/拆分；
数据复制/增殖；
等等。
一个模块中的工作协程从一些其它模块接收数据做为输入，并向另一些模块发送输出数据。 换句话数，一个模块可能同时兼任数据消费者和数据产生者的角色。

多个模块一起组成了一个数据流处理系统。

下面将展示一些模块工作协程的实现。这些实现仅仅是为了解释目的，所以它们都很简单，并且它们可能并不高效。

数据生成/搜集/加载
一个数据产生者可能通过以下途径生成数据：
加载一个文件、或者读取一个数据库、或者用爬虫抓取网页数据；
从一个软件或者硬件系统搜集各种数据；
产生一系列随机数；
等等。
这里，我们使用一个随机数产生器做为一个数据产生者的例子。 此数据产生者函数没有输入，只有输出。
import (
"crypto/rand"
"encoding/binary"
)

func RandomGenerator() <-chan uint64 {
c := make(chan uint64)
go func() {
rnds := make([]byte, 8)
for {
_, err := rand.Read(rnds)
if err != nil {
close(c)
break
}
c <- binary.BigEndian.Uint64(rnds)
}
}()
return c
}
事实上，此随机数产生器是一个多返回值的future/promise。

一个数据产生者可以在任何时刻关闭返回的通道以结束数据生成。

数据聚合
一个数据聚合模块的工作协程将多个数据流合为一个数据流。 假设数据类型为int64，下面这个函数将任意数量的数据流合为一个。
func Aggregator(inputs ...<-chan uint64) <-chan uint64 {
out := make(chan uint64)
for _, in := range inputs {
go func(in <-chan uint64) {
for {
out <- <-in // <=> out <- (<-in)
}
}(in)
}
return out
}
一个更完美的实现需要考虑一个输入数据流是否已经关闭。（下面要介绍的其它工作协程同理。）
import "sync"

func Aggregator(inputs ...<-chan uint64) <-chan uint64 {
output := make(chan uint64)
var wg sync.WaitGroup
for _, in := range inputs {
wg.Add(1)
go func(int <-chan uint64) {
defer wg.Done()
// 如果通道in被关闭，此循环将最终结束。
for x := range in {
output <- x
}
}(in)
}
go func() {
wg.Wait()
close(output)
}()
return output
}
如果被聚合的数据流的数量很小，我们也可以使用一个select控制流程代码块来聚合这些数据流。
// 假设数据流的数量为2。
...
output := make(chan uint64)
go func() {
inA, inB := inputs[0], inputs[1]
for {
select {
case v := <- inA: output <- v
case v := <- inB: output <- v
}
}
}
...
数据分流
数据分流是数据聚合的逆过程。数据分流的实现很简单，但在实践中用的并不多。
func Divisor(input <-chan uint64, outputs ...chan<- uint64) {
for _, out := range outputs {
go func(o chan<- uint64) {
for {
o <- <-input // <=> o <- (<-input)
}
}(out)
}
}
数据合成
数据合成将多个数据流中读取的数据合成一个。

下面是一个数据合成工作函数的实现中，从两个不同数据流读取的两个uint64值组成了一个新的uint64值。 当然，在实践中，数据的组合比这复杂得多。
func Composor(inA, inB <-chan uint64) <-chan uint64 {
output := make(chan uint64)
go func() {
for {
a1, b, a2 := <-inA, <-inB, <-inA
output <- a1 ^ b & a2
}
}()
return output
}
数据分解
数据分解是数据合成的逆过程。一个数据分解者从一个通道读取一份数据，并将此数据分解为多份数据。 这里就不举例了。
数据复制/增殖
数据复制（增殖）可以看作是特殊的数据分解。一份输入数据将被复制多份并输出给多个数据流。

一个例子：
func Duplicator(in <-chan uint64) (<-chan uint64, <-chan uint64) {
outA, outB := make(chan uint64), make(chan uint64)
go func() {
for x := range in {
outA <- x
outB <- x
}
}()
return outA, outB
}
数据计算/分析
数据计算和数据分析模块的功能因具体程序不同而有很大的差异。 一般来说，数据分析者接收一份数据并对之加工处理后转换为另一份数据。

下面的简单示例中，每个输入的uint64值将被进行位反转后输出。
func Calculator(in <-chan uint64, out chan uint64) (<-chan uint64) {
if out == nil {
out = make(chan uint64)
}
go func() {
for x := range in {
out <- ^x
}
}()
return out
}
数据验证/过滤
一个数据验证或过滤者的任务是检查输入数据的合理性并抛弃不合理的数据。 比如，下面的工作者协程将抛弃所有的非素数。
import "math/big"

func Filter0(input <-chan uint64, output chan uint64) <-chan uint64 {
if output == nil {
output = make(chan uint64)
}
go func() {
bigInt := big.NewInt(0)
for x := range input {
bigInt.SetUint64(x)
if bigInt.ProbablyPrime(1) {
output <- x
}
}
}()
return output
}

func Filter(input <-chan uint64) <-chan uint64 {
return Filter0(input, nil)
}
请注意这两个函数版本分别被本文下面最后展示的两个例子所使用。

数据服务/存盘
一般，一个数据服务或者存盘模块为一个数据流系统中的最后一个模块。 这里的实现值是简单地将数据输出到终端。
import "fmt"

func Printer(input <-chan uint64) {
for x := range input {
fmt.Println(x)
}
}
组装数据流系统
现在，让我们使用上面的模块工作者函数实现来组装一些数据流系统。 组装数据流仅仅是创建一些工作者协程函数调用，并为这些调用指定输入数据流和输出数据流。

数据流系统例子1（一个流线型系统）：
package main

... // 上面的模块工作者函数实现

func main() {
Printer(
Filter(
Calculator(
RandomGenerator(), nil,
),
),
)
}
上面这个流线型系统描绘在下图中：
线性数据流
数据流系统例子2（一个单向无环图系统）：
package main

... // 上面的模块工作者函数实现

func main() {
filterA := Filter(RandomGenerator())
filterB := Filter(RandomGenerator())
filterC := Filter(RandomGenerator())
filter := Aggregator(filterA, filterB, filterC)
calculatorA := Calculator(filter, nil)
calculatorB := Calculator(filter, nil)
calculator := Aggregator(calculatorA, calculatorB)
Printer(calculator)
}
上面这个单向无环图系统描绘在下图中：
有向无环数据流
更复杂的数据流系统可以表示为任何拓扑结构的图。比如一个复杂的数据流系统可能有多个输出模块。 但是有环拓扑结构的数据流系统在实践中很少用。

从上面两个例子可以看出，使用通道来构建数据流系统是很简单和直观的。

从上例可以看出，通过使用数据聚合模块，我们可以很轻松地实现各个模块的工作协程数量的扇入（fan-in）和扇出（fan-out）。

事实上，我们也可以使用一个简单的通道来代替数据聚合模块的角色。比如，下面的代码使用两个通道代替了上例中的两个数据聚合器。
package main

... // 上面的模块工作者函数实现

func main() {
c1 := make(chan uint64, 100)
Filter0(RandomGenerator(), c1) // filterA
Filter0(RandomGenerator(), c1) // filterB
Filter0(RandomGenerator(), c1) // filterC
c2 := make(chan uint64, 100)
Calculator(c1, c2) // calculatorA
Calculator(c1, c2) // calculatorB
Printer(c2)
}

## 如何优雅地关闭通道
在本文发表数日前，我曾写了一篇文章来解释通道的规则。 那篇文章在reddit和HN上获得了很多点赞，但也有很多人对Go通道的细节设计提出了一些批评意见。

这些批评主要针对于通道设计中的下列细节：
没有一个简单和通用的方法用来在不改变一个通道的状态的情况下检查这个通道是否已经关闭。
关闭一个已经关闭的通道将产生一个恐慌，所以在不知道一个通道是否已经关闭的时候关闭此通道是很危险的。
向一个已关闭的通道发送数据将产生一个恐慌，所以在不知道一个通道是否已经关闭的时候向此通道发送数据是很危险的。
这些批评看上去有几分道理（实际上属于对通道的不正确使用导致的偏见）。 是的，Go语言中并没有提供一个内置函数来检查一个通道是否已经关闭。

在Go中，如果我们能够保证从不会向一个通道发送数据，那么有一个简单的方法来判断此通道是否已经关闭。 此方法已经在上一篇文章通道用例大全中展示过了。 这里为了本文的连贯性，在下面的例子中重新列出了此方法。
package main

import "fmt"

type T int

func IsClosed(ch <-chan T) bool {
select {
case <-ch:
return true
default:
}

	return false
}

func main() {
c := make(chan T)
fmt.Println(IsClosed(c)) // false
close(c)
fmt.Println(IsClosed(c)) // true
}
如前所述，此方法并不是一个通用的检查通道是否已经关闭的方法。

事实上，即使有一个内置closed函数用来检查一个通道是否已经关闭，它的有用性也是十分有限的。 原因是当此函数的一个调用的结果返回时，被查询的通道的状态可能已经又改变了，导致此调用结果并不能反映出被查询的通道的最新状态。 虽然我们可以根据一个调用closed(ch)的返回结果为true而得出我们不应该再向通道ch发送数据的结论， 但是我们不能根据一个调用closed(ch)的返回结果为false而得出我们可以继续向通道ch发送数据的结论。

通道关闭原则
一个常用的使用Go通道的原则是不要在数据接收方或者在有多个发送者的情况下关闭通道。 换句话说，我们只应该让一个通道唯一的发送者关闭此通道。

下面我们将称此原则为通道关闭原则。

当然，这并不是一个通用的关闭通道的原则。通用的原则是不要关闭已关闭的通道。 如果我们能够保证从某个时刻之后，再没有协程将向一个未关闭的非nil通道发送数据，则一个协程可以安全地关闭此通道。 然而，做出这样的保证常常需要很大的努力，从而导致代码过度复杂。 另一方面，遵循通道关闭原则是一件相对简单的事儿。

粗鲁地关闭通道的方法
如果由于某种原因，你一定非要从数据接收方或者让众多发送者中的一个关闭一个通道，你可以使用恢复机制来防止可能产生的恐慌而导致程序崩溃。 下面就是这样的一个实现（假设通道的元素类型为T）。
func SafeClose(ch chan T) (justClosed bool) {
defer func() {
if recover() != nil {
// 一个函数的返回结果可以在defer调用中修改。
justClosed = false
}
}()

	// 假设ch != nil。
	close(ch)   // 如果ch已关闭，则产生一个恐慌。
	return true // <=> justClosed = true; return
}
此方法违反了通道关闭原则。

同样的方法可以用来粗鲁地向一个关闭状态未知的通道发送数据。
func SafeSend(ch chan T, value T) (closed bool) {
defer func() {
if recover() != nil {
closed = true
}
}()

	ch <- value  // 如果ch已关闭，则产生一个恐慌。
	return false // <=> closed = false; return
}
这样的粗鲁方法不仅违反了通道关闭原则，而且Go白皮书和标准编译器不保证它的实现中不存在数据竞争。

礼貌地关闭通道的方法
很多Go程序员喜欢使用sync.Once来关闭通道。
type MyChannel struct {
C    chan T
once sync.Once
}

func NewMyChannel() *MyChannel {
return &MyChannel{C: make(chan T)}
}

func (mc *MyChannel) SafeClose() {
mc.once.Do(func() {
close(mc.C)
})
}
当然，我们也可以使用sync.Mutex来防止多次关闭一个通道。
type MyChannel struct {
C      chan T
closed bool
mutex  sync.Mutex
}

func NewMyChannel() *MyChannel {
return &MyChannel{C: make(chan T)}
}

func (mc *MyChannel) SafeClose() {
mc.mutex.Lock()
defer mc.mutex.Unlock()
if !mc.closed {
close(mc.C)
mc.closed = true
}
}

func (mc *MyChannel) IsClosed() bool {
mc.mutex.Lock()
defer mc.mutex.Unlock()
return mc.closed
}
这些实现确实比上一节中的方法礼貌一些，但是它们不能完全有效地避免数据竞争。 目前的Go白皮书并不保证发生在一个通道上的并发关闭操作和发送操作不会产生数据竞争。 如果一个SafeClose函数和同一个通道上的发送操作同时运行，则数据竞争可能发生（虽然这样的数据竞争一般并不会带来什么危害）。

优雅地关闭通道的方法
上一节中介绍的SafeSend函数有一个弊端，它的调用不能做为case操作而被使用在select代码块中。 另外，很多Go程序员（包括我）认为上面两节展示的关闭通道的方法不是很优雅。 本节下面将介绍一些在各种情形下使用纯通道操作来关闭通道的方法。

（为了演示程序的完整性，下面这些例子中使用到了sync.WaitGroup。在实践中，sync.WaitGroup并不是必需的。）

情形一：M个接收者和一个发送者。发送者通过关闭用来传输数据的通道来传递发送结束信号
这是最简单的一种情形。当发送者欲结束发送，让它关闭用来传输数据的通道即可。
package main

import (
"time"
"math/rand"
"sync"
"log"
)

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
log.SetFlags(0)

	// ...
	const Max = 100000
	const NumReceivers = 100

	wgReceivers := sync.WaitGroup{}
	wgReceivers.Add(NumReceivers)

	// ...
	dataCh := make(chan int)

	// 发送者
	go func() {
		for {
			if value := rand.Intn(Max); value == 0 {
				// 此唯一的发送者可以安全地关闭此数据通道。
				close(dataCh)
				return
			} else {
				dataCh <- value
			}
		}
	}()

	// 接收者
	for i := 0; i < NumReceivers; i++ {
		go func() {
			defer wgReceivers.Done()

			// 接收数据直到通道dataCh已关闭
			// 并且dataCh的缓冲队列已空。
			for value := range dataCh {
				log.Println(value)
			}
		}()
	}

	wgReceivers.Wait()
}
情形二：一个接收者和N个发送者，此唯一接收者通过关闭一个额外的信号通道来通知发送者不要再发送数据了
此情形比上一种情形复杂一些。我们不能让接收者关闭用来传输数据的通道来停止数据传输，因为这样做违反了通道关闭原则。 但是我们可以让接收者关闭一个额外的信号通道来通知发送者不要再发送数据了。
package main

import (
"time"
"math/rand"
"sync"
"log"
)

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
log.SetFlags(0)

	// ...
	const Max = 100000
	const NumSenders = 1000

	wgReceivers := sync.WaitGroup{}
	wgReceivers.Add(1)

	// ...
	dataCh := make(chan int)
	stopCh := make(chan struct{})
		// stopCh是一个额外的信号通道。它的
		// 发送者为dataCh数据通道的接收者。
		// 它的接收者为dataCh数据通道的发送者。

	// 发送者
	for i := 0; i < NumSenders; i++ {
		go func() {
			for {
				// 这里的第一个尝试接收用来让此发送者
				// 协程尽早地退出。对于这个特定的例子，
				// 此select代码块并非必需。
				select {
				case <- stopCh:
					return
				default:
				}

				// 即使stopCh已经关闭，此第二个select
				// 代码块中的第一个分支仍很有可能在若干个
				// 循环步内依然不会被选中。如果这是不可接受
				// 的，则上面的第一个select代码块是必需的。
				select {
				case <- stopCh:
					return
				case dataCh <- rand.Intn(Max):
				}
			}
		}()
	}

	// 接收者
	go func() {
		defer wgReceivers.Done()

		for value := range dataCh {
			if value == Max-1 {
				// 此唯一的接收者同时也是stopCh通道的
				// 唯一发送者。尽管它不能安全地关闭dataCh数
				// 据通道，但它可以安全地关闭stopCh通道。
				close(stopCh)
				return
			}

			log.Println(value)
		}
	}()

	// ...
	wgReceivers.Wait()
}
如此例中的注释所述，对于此额外的信号通道stopCh，它只有一个发送者，即dataCh数据通道的唯一接收者。 dataCh数据通道的接收者关闭了信号通道stopCh，这是不违反通道关闭原则的。

在此例中，数据通道dataCh并没有被关闭。是的，我们不必关闭它。 当一个通道不再被任何协程所使用后，它将逐渐被垃圾回收掉，无论它是否已经被关闭。 所以这里的优雅性体现在通过不关闭一个通道来停止使用此通道。

情形三：M个接收者和N个发送者。它们中的任何协程都可以让一个中间调解协程帮忙发出停止数据传送的信号
这是最复杂的一种情形。我们不能让接收者和发送者中的任何一个关闭用来传输数据的通道，我们也不能让多个接收者之一关闭一个额外的信号通道。 这两种做法都违反了通道关闭原则。 然而，我们可以引入一个中间调解者角色并让其关闭额外的信号通道来通知所有的接收者和发送者结束工作。 具体实现见下例。注意其中使用了一个尝试发送操作来向中间调解者发送信号。
package main

import (
"time"
"math/rand"
"sync"
"log"
"strconv"
)

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
log.SetFlags(0)

	// ...
	const Max = 100000
	const NumReceivers = 10
	const NumSenders = 1000

	wgReceivers := sync.WaitGroup{}
	wgReceivers.Add(NumReceivers)

	// ...
	dataCh := make(chan int)
	stopCh := make(chan struct{})
		// stopCh是一个额外的信号通道。它的发送
		// 者为中间调解者。它的接收者为dataCh
		// 数据通道的所有的发送者和接收者。
	toStop := make(chan string, 1)
		// toStop是一个用来通知中间调解者让其
		// 关闭信号通道stopCh的第二个信号通道。
		// 此第二个信号通道的发送者为dataCh数据
		// 通道的所有的发送者和接收者，它的接收者
		// 为中间调解者。它必须为一个缓冲通道。

	var stoppedBy string

	// 中间调解者
	go func() {
		stoppedBy = <-toStop
		close(stopCh)
	}()

	// 发送者
	for i := 0; i < NumSenders; i++ {
		go func(id string) {
			for {
				value := rand.Intn(Max)
				if value == 0 {
					// 为了防止阻塞，这里使用了一个尝试
					// 发送操作来向中间调解者发送信号。
					select {
					case toStop <- "发送者#" + id:
					default:
					}
					return
				}

				// 此处的尝试接收操作是为了让此发送协程尽早
				// 退出。标准编译器对尝试接收和尝试发送做了
				// 特殊的优化，因而它们的速度很快。
				select {
				case <- stopCh:
					return
				default:
				}

				// 即使stopCh已关闭，如果这个select代码块
				// 中第二个分支的发送操作是非阻塞的，则第一个
				// 分支仍很有可能在若干个循环步内依然不会被选
				// 中。如果这是不可接受的，则上面的第一个尝试
				// 接收操作代码块是必需的。
				select {
				case <- stopCh:
					return
				case dataCh <- value:
				}
			}
		}(strconv.Itoa(i))
	}

	// 接收者
	for i := 0; i < NumReceivers; i++ {
		go func(id string) {
			defer wgReceivers.Done()

			for {
				// 和发送者协程一样，此处的尝试接收操作是为了
				// 让此接收协程尽早退出。
				select {
				case <- stopCh:
					return
				default:
				}

				// 即使stopCh已关闭，如果这个select代码块
				// 中第二个分支的接收操作是非阻塞的，则第一个
				// 分支仍很有可能在若干个循环步内依然不会被选
				// 中。如果这是不可接受的，则上面尝试接收操作
				// 代码块是必需的。
				select {
				case <- stopCh:
					return
				case value := <-dataCh:
					if value == Max-1 {
						// 为了防止阻塞，这里使用了一个尝试
						// 发送操作来向中间调解者发送信号。
						select {
						case toStop <- "接收者#" + id:
						default:
						}
						return
					}

					log.Println(value)
				}
			}
		}(strconv.Itoa(i))
	}

	// ...
	wgReceivers.Wait()
	log.Println("被" + stoppedBy + "终止了")
}
在此例中，通道关闭原则依旧得到了遵守。

请注意，信号通道toStop的容量必须至少为1。 如果它的容量为0，则在中间调解者还未准备好的情况下就已经有某个协程向toStop发送信号时，此信号将被抛弃。

我们也可以不使用尝试发送操作向中间调解者发送信号，但信号通道toStop的容量必须至少为数据发送者和数据接收者的数量之和，以防止向其发送数据时（有一个极其微小的可能）导致某些发送者和接收者协程永久阻塞。
...
toStop := make(chan string, NumReceivers + NumSenders)
...
value := rand.Intn(Max)
if value == 0 {
toStop <- "sender#" + id
return
}
...
if value == Max-1 {
toStop <- "receiver#" + id
return
}
...
情形四：“M个接收者和一个发送者”情形的一个变种：用来传输数据的通道的关闭请求由第三方发出
有时，数据通道（dataCh）的关闭请求需要由某个第三方协程发出。对于这种情形，我们可以使用一个额外的信号通道来通知唯一的发送者关闭数据通道（dataCh）。
package main

import (
"time"
"math/rand"
"sync"
"log"
)

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
log.SetFlags(0)

	// ...
	const Max = 100000
	const NumReceivers = 100
	const NumThirdParties = 15

	wgReceivers := sync.WaitGroup{}
	wgReceivers.Add(NumReceivers)

	// ...
	dataCh := make(chan int)
	closing := make(chan struct{}) // 信号通道
	closed := make(chan struct{})
	
	// 此stop函数可以被安全地多次调用。
	stop := func() {
		select {
		case closing<-struct{}{}:
			<-closed
		case <-closed:
		}
	}
	
	// 一些第三方协程
	for i := 0; i < NumThirdParties; i++ {
		go func() {
			r := 1 + rand.Intn(3)
			time.Sleep(time.Duration(r) * time.Second)
			stop()
		}()
	}

	// 发送者
	go func() {
		defer func() {
			close(closed)
			close(dataCh)
		}()

		for {
			select{
			case <-closing: return
			default:
			}

			select{
			case <-closing: return
			case dataCh <- rand.Intn(Max):
			}
		}
	}()

	// 接收者
	for i := 0; i < NumReceivers; i++ {
		go func() {
			defer wgReceivers.Done()

			for value := range dataCh {
				log.Println(value)
			}
		}()
	}

	wgReceivers.Wait()
}
上述代码中的stop函数中使用的技巧偷自Roger Peppe在此贴中的一个留言。

情形五：“N个发送者”的一个变种：用来传输数据的通道必须被关闭以通知各个接收者数据发送已经结束了
在上面的提到的“N个发送者”情形中，为了遵守通道关闭原则，我们避免了关闭数据通道（dataCh）。 但是有时候，数据通道（dataCh）必须被关闭以通知各个接收者数据发送已经结束。 对于这种“N个发送者”情形，我们可以使用一个中间通道将它们转化为“一个发送者”情形，然后继续使用上一节介绍的技巧来关闭此中间通道，从而避免了关闭原始的dataCh数据通道。
package main

import (
"time"
"math/rand"
"sync"
"log"
"strconv"
)

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
log.SetFlags(0)

	// ...
	const Max = 1000000
	const NumReceivers = 10
	const NumSenders = 1000
	const NumThirdParties = 15

	wgReceivers := sync.WaitGroup{}
	wgReceivers.Add(NumReceivers)

	// ...
	dataCh := make(chan int)   // 将被关闭
	middleCh := make(chan int) // 不会被关闭
	closing := make(chan string)
	closed := make(chan struct{})

	var stoppedBy string

	stop := func(by string) {
		select {
		case closing <- by:
			<-closed
		case <-closed:
		}
	}
	
	// 中间层
	go func() {
		exit := func(v int, needSend bool) {
			close(closed)
			if needSend {
				dataCh <- v
			}
			close(dataCh)
		}

		for {
			select {
			case stoppedBy = <-closing:
				exit(0, false)
				return
			case v := <- middleCh:
				select {
				case stoppedBy = <-closing:
					exit(v, true)
					return
				case dataCh <- v:
				}
			}
		}
	}()
	
	// 一些第三方协程
	for i := 0; i < NumThirdParties; i++ {
		go func(id string) {
			r := 1 + rand.Intn(3)
			time.Sleep(time.Duration(r) * time.Second)
			stop("3rd-party#" + id)
		}(strconv.Itoa(i))
	}

	// 发送者
	for i := 0; i < NumSenders; i++ {
		go func(id string) {
			for {
				value := rand.Intn(Max)
				if value == 0 {
					stop("sender#" + id)
					return
				}

				select {
				case <- closed:
					return
				default:
				}

				select {
				case <- closed:
					return
				case middleCh <- value:
				}
			}
		}(strconv.Itoa(i))
	}

	// 接收者
	for range [NumReceivers]struct{}{} {
		go func() {
			defer wgReceivers.Done()

			for value := range dataCh {
				log.Println(value)
			}
		}()
	}

	// ...
	wgReceivers.Wait()
	log.Println("stopped by", stoppedBy)
}
更多情形？
在日常编程中可能会遇到更多的变种情形，但是上面介绍的情形是最常见和最基本的。 通过聪明地使用通道（和其它并发同步技术），我相信，对于各种变种，我们总会找到相应的遵守通道关闭原则的解决方法。

结论
并没有什么情况非得逼得我们违反通道关闭原则。 如果你遇到了此情形，请考虑修改你的代码流程和结构设计。

## sync标准库包中提供的并发同步技术
通道用例大全一文中介绍了很多通过使用通道来实现并发同步的用例。 事实上，通道并不是Go支持的唯一的一种并发同步技术。而且对于一些特定的情形，通道并不是最有效和可读性最高的同步技术。 本文下面将介绍sync标准库包中提供的各种并发同步技术。相对于通道，这些技术对于某些情形更加适用。

sync标准库包提供了一些用于实现并发同步的类型。这些类型适用于各种不同的内存顺序需求。 对于这些特定的需求，这些类型使用起来比通道效率更高，代码实现更简洁。

（请注意：为了避免各种异常行为，最好不要复制sync标准库包中提供的类型的值。）

sync.WaitGroup（等待组）类型
每个sync.WaitGroup值在内部维护着一个计数，此计数的初始默认值为零。

*sync.WaitGroup类型有三个方法：Add(delta int)、Done()和Wait()。

对于一个可寻址的sync.WaitGroup值wg，
我们可以使用方法调用wg.Add(delta)来改变值wg维护的计数。
方法调用wg.Done()和wg.Add(-1)是完全等价的。
如果一个wg.Add(delta)或者wg.Done()调用将wg维护的计数更改成一个负数，一个恐慌将产生。
当一个协程调用了wg.Wait()时，
如果此时wg维护的计数为零，则此wg.Wait()此操作为一个空操作（no-op）；
否则（计数为一个正整数），此协程将进入阻塞状态。 当以后其它某个协程将此计数更改至0时（一般通过调用wg.Done()），此协程将重新进入运行状态（即wg.Wait()将返回）。
请注意wg.Add(delta)、wg.Done()和wg.Wait()分别是(&wg).Add(delta)、(&wg).Done()和(&wg).Wait()的简写形式。

一般，一个sync.WaitGroup值用来让某个协程等待其它若干协程都先完成它们各自的任务。 一个例子：
package main

import (
"fmt"
"math/rand"
"sync"
"time"
)

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	const N = 5
	var values [N]int32

	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		i := i
		go func() {
			values[i] = 50 + rand.Int31n(50)
			fmt.Println("Done:", i)
			wg.Done() // <=> wg.Add(-1)
		}()
	}

	wg.Wait()
	// 所有的元素都保证被初始化了。
	fmt.Println("values:", values)
}
在此例中，主协程等待着直到其它5个协程已经将各自负责的元素初始化完毕此会打印出各个元素值。 这里是一个可能的程序执行输出结果：
Done: 4
Done: 1
Done: 3
Done: 0
Done: 2
values: [71 89 50 62 60]
我们可以将上例中的Add方法调用拆分成多次调用：
...
var wg sync.WaitGroup
for i := 0; i < N; i++ {
wg.Add(1) // 将被执行5次
i := i
go func() {
values[i] = 50 + rand.Int31n(50)
wg.Done()
}()
}
...
一个*sync.WaitGroup值的Wait方法可以在多个协程中调用。 当对应的sync.WaitGroup值维护的计数降为0，这些协程都将得到一个（广播）通知而结束阻塞状态。
func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	const N = 5
	var values [N]int32

	var wgA, wgB sync.WaitGroup
	wgA.Add(N)
	wgB.Add(1)

	for i := 0; i < N; i++ {
		i := i
		go func() {
			wgB.Wait() // 等待广播通知
			log.Printf("values[%v]=%v \n", i, values[i])
			wgA.Done()
		}()
	}

	// 下面这个循环保证将在上面的任何一个
	// wg.Wait调用结束之前执行。
	for i := 0; i < N; i++ {
		values[i] = 50 + rand.Int31n(50)
	}
	wgB.Done() // 发出一个广播通知
	wgA.Wait()
}
一个WaitGroup可以在它的一个Wait方法返回之后被重用。 但是请注意，当一个WaitGroup值维护的基数为零时，它的带有正整数实参的Add方法调用不能和它的Wait方法调用并发运行，否则将可能出现数据竞争。

sync.Once类型
每个*sync.Once值有一个Do(f func())方法。 此方法只有一个类型为func()的参数。

对一个可寻址的sync.Once值o，o.Do()（即(&o).Do()的简写形式）方法调用可以在多个协程中被多次并发地执行， 这些方法调用的实参应该（但并不强制）为同一个函数值。 在这些方法调用中，有且只有一个调用的实参函数（值）将得到调用。 此被调用的实参函数保证在任何o.Do()方法调用返回之前退出。 换句话说，被调用的实参函数内的代码将在任何o.Do()方法返回调用之前被执行。

一般来说，一个sync.Once值被用来确保一段代码在一个并发程序中被执行且仅被执行一次。

一个例子：
package main

import (
"log"
"sync"
)

func main() {
log.SetFlags(0)

	x := 0
	doSomething := func() {
		x++
		log.Println("Hello")
	}

	var wg sync.WaitGroup
	var once sync.Once
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(doSomething)
			log.Println("world!")
		}()
	}

	wg.Wait()
	log.Println("x =", x) // x = 1
}
在此例中，Hello将仅被输出一次，而world!将被输出5次，并且Hello肯定在所有的5个world!之前输出。

sync.Mutex（互斥锁）和sync.RWMutex（读写锁）类型
*sync.Mutex和*sync.RWMutex类型都实现了sync.Locker接口类型。 所以这两个类型都有两个方法：Lock()和Unlock()，用来保护一份数据不会被多个使用者同时读取和修改。

除了Lock()和Unlock()这两个方法，*sync.RWMutex类型还有两个另外的方法：RLock()和RUnlock()，用来支持多个读取者并发读取一份数据但防止此份数据被某个数据写入者和其它数据访问者（包括读取者和写入者）同时使用。

（注意：这里的数据读取者和数据写入者不应该从字面上理解。有时候某些数据读取者可能修改数据，而有些数据写入者可能只读取数据。）

一个Mutex值常称为一个互斥锁。 一个Mutex零值为一个尚未加锁的互斥锁。 一个（可寻址的）Mutex值m只有在未加锁状态时才能通过m.Lock()方法调用被成功加锁。 换句话说，一旦m值被加了锁（亦即某个m.Lock()方法调用成功返回）， 一个新的加锁试图将导致当前协程进入阻塞状态，直到此Mutex值被解锁为止（通过m.Unlock()方法调用）。

注意：m.Lock()和m.Unlock()分别是(&m).Lock()和(&m).Unlock()的简写形式。

一个使用sync.Mutex的例子：
package main

import (
"fmt"
"runtime"
"sync"
)

type Counter struct {
m sync.Mutex
n uint64
}

func (c *Counter) Value() uint64 {
c.m.Lock()
defer c.m.Unlock()
return c.n
}

func (c *Counter) Increase(delta uint64) {
c.m.Lock()
c.n += delta
c.m.Unlock()
}

func main() {
var c Counter
for i := 0; i < 100; i++ {
go func() {
for k := 0; k < 100; k++ {
c.Increase(1)
}
}()
}

	// 此循环仅为演示目的。
	for c.Value() < 10000 {
		runtime.Gosched()
	}
	fmt.Println(c.Value()) // 10000
}
在上面这个例子中，一个Counter值使用了一个Mutex字段来确保它的字段n永远不会被多个协程同时使用。

一个RWMutex值常称为一个读写互斥锁，它的内部包含两个锁：一个写锁和一个读锁。 对于一个可寻址的RWMutex值rwm，数据写入者可以通过方法调用rwm.Lock()对rwm加写锁，或者通过rwm.RLock()方法调用对rwm加读锁。 方法调用rwm.Unlock()和rwm.RUnlock()用来解开rwm的写锁和读锁。 rwm的读锁维护着一个计数。当rwm.RLock()调用成功时，此计数增1；当rwm.Unlock()调用成功时，此计数减1； 一个零计数表示rwm的读锁处于未加锁状态；反之，一个非零计数（肯定大于零）表示rwm的读锁处于加锁状态。

注意rwm.Lock()、rwm.Unlock()、rwm.RLock()和rwm.RUnlock()分别是(&rwm).Lock()、(&rwm).Unlock()、(&rwm).RLock()和(&rwm).RUnlock()的简写形式。

对于一个可寻址的RWMutex值rwm，下列规则存在：
rwm的写锁只有在它的写锁和读锁都处于未加锁状态时才能被成功加锁。 换句话说，rwm的写锁在任何时刻最多只能被一个数据写入者成功加锁，并且rwm的写锁和读锁不能同时处于加锁状态。
当rwm的写锁正处于加锁状态的时候，任何新的对之加写锁或者加读锁的操作试图都将导致当前协程进入阻塞状态，直到此写锁被解锁，这样的操作试图才有机会成功。
当rwm的读锁正处于加锁状态的时候，新的加写锁的操作试图将导致当前协程进入阻塞状态。 但是，一个新的加读锁的操作试图将成功，只要此操作试图发生在任何被阻塞的加写锁的操作试图之前（见下一条规则）。 换句话说，一个读写互斥锁的读锁可以同时被多个数据读取者同时加锁而持有。 当rwm的读锁维护的计数清零时，读锁将返回未加锁状态。
假设rwm的读锁正处于加锁状态的时候，为了防止后续数据写入者没有机会成功加写锁，后续发生在某个被阻塞的加写锁操作试图之后的所有加读锁的试图都将被阻塞。
假设rwm的写锁正处于加锁状态的时候，（至少对于标准编译器来说，）为了防止后续数据读取者没有机会成功加读锁，发生在此写锁下一次被解锁之前的所有加读锁的试图都将在此写锁下一次被解锁之后肯定取得成功，即使所有这些加读锁的试图发生在一些仍被阻塞的加写锁的试图之后。
后两条规则是为了确保数据读取者和写入者都有机会执行它们的操作。

请注意：一个锁并不会绑定到一个协程上，即一个锁并不记录哪个协程成功地加锁了它。 换句话说，一个锁的加锁者和此锁的解锁者可以不是同一个协程，尽管在实践中这种情况并不多见。

在上一个例子中，如果Value方法被十分频繁调用而Increase方法并不频繁被调用，则Counter类型的m字段的类型可以更改为sync.RWMutex，从而使得执行效率更高，如下面的代码所示。
...
type Counter struct {
//m sync.Mutex
m sync.RWMutex
n uint64
}

func (c *Counter) Value() uint64 {
//c.m.Lock()
//defer c.m.Unlock()
c.m.RLock()
defer c.m.RUnlock()
return c.n
}
...
sync.RWMutex值的另一个应用场景是将一个写任务分隔成若干小的写任务。下一节中展示了一个这样的例子。

根据上面列出的后两条规则，下面这个程序最有可能输出abdc。
package main

import (
"fmt"
"time"
"sync"
)

func main() {
var m sync.RWMutex
go func() {
m.RLock()
fmt.Print("a")
time.Sleep(time.Second)
m.RUnlock()
}()
go func() {
time.Sleep(time.Second * 1 / 4)
m.Lock()
fmt.Print("b")
time.Sleep(time.Second)
m.Unlock()
}()
go func() {
time.Sleep(time.Second * 2 / 4)
m.Lock()
fmt.Print("c")
m.Unlock()
}()
go func () {
time.Sleep(time.Second * 3 / 4)
m.RLock()
fmt.Print("d")
m.RUnlock()
}()
time.Sleep(time.Second * 3)
fmt.Println()
}
请注意，上例这个程序仅仅是为了解释和验证上面列出的读写锁的后两条加锁规则。 此程序使用了time.Sleep调用来做协程间的同步。这种所谓的同步方法不应该被使用在生产代码中。

sync.Mutex和sync.RWMutex值也可以用来实现通知，尽管这不是Go中最优雅的方法来实现通知。 下面是一个使用了Mutex值来实现通知的例子。
package main

import (
"fmt"
"sync"
"time"
)

func main() {
var m sync.Mutex
m.Lock()
go func() {
time.Sleep(time.Second)
fmt.Println("Hi")
m.Unlock() // 发出一个通知
}()
m.Lock() // 等待通知
fmt.Println("Bye")
}
在此例中，Hi将确保在Bye之前打印出来。 关于sync.Mutex和sync.RWMutex值相关的内存顺序保证，请阅读Go中的内存顺序保证一文。

sync.Cond类型
sync.Cond类型提供了一种有效的方式来实现多个协程间的通知。

每个sync.Cond值拥有一个sync.Locker类型的名为L的字段。 此字段的具体值常常为一个*sync.Mutex值或者*sync.RWMutex值。

*sync.Cond类型有三个方法：Wait()、Signal()和Broadcast()。

每个Cond值维护着一个先进先出等待协程队列。 对于一个可寻址的Cond值c，
c.Wait()必须在c.L字段值的锁处于加锁状态的时候调用；否则，c.Wait()调用将造成一个恐慌。 一个c.Wait()调用将
首先将当前协程推入到c所维护的等待协程队列；
然后调用c.L.Unlock()对c.L的锁解锁；
然后使当前协程进入阻塞状态；

（当前协程将被另一个协程通过c.Signal()或c.Broadcast()调用唤醒而重新进入运行状态。）

一旦当前协程重新进入运行状态，c.L.Lock()将被调用以试图重新对c.L字段值的锁加锁。 此c.Wait()调用将在此试图成功之后退出。
一个c.Signal()调用将唤醒并移除c所维护的等待协程队列中的第一个协程（如果此队列不为空的话）。
一个c.Broadcast()调用将唤醒并移除c所维护的等待协程队列中的所有协程（如果此队列不为空的话）。
请注意：c.Wait()、c.Signal()和c.Broadcast()分别为(&c).Wait()、(&c).Signal()和(&c).Broadcast()的简写形式。

c.Signal()和c.Broadcast()调用常用来通知某个条件的状态发生了变化。 一般说来，c.Wait()应该在一个检查某个条件是否已经得到满足的循环中调用。

下面是一个典型的sync.Cond用例。
package main

import (
"fmt"
"math/rand"
"sync"
"time"
)

func main() {
rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	const N = 10
	var values [N]string

	cond := sync.NewCond(&sync.Mutex{})

	for i := 0; i < N; i++ {
		d := time.Second * time.Duration(rand.Intn(10)) / 10
		go func(i int) {
			time.Sleep(d) // 模拟一个工作负载
			cond.L.Lock()
			// 下面的修改必须在cond.L被锁定的时候执行
			values[i] = string('a' + i)
			cond.Broadcast() // 可以在cond.L被解锁后发出通知
			cond.L.Unlock()
			// 上面的通知也可以在cond.L未锁定的时候发出。
			//cond.Broadcast() // 上面的调用也可以放在这里
		}(i)
	}

	// 此函数必须在cond.L被锁定的时候调用。
	checkCondition := func() bool {
		fmt.Println(values)
		for i := 0; i < N; i++ {
			if values[i] == "" {
				return false
			}
		}
		return true
	}

	cond.L.Lock()
	defer cond.L.Unlock()
	for !checkCondition() {
		cond.Wait() // 必须在cond.L被锁定的时候调用
	}
}
一个可能的输出：
[         ]
[     f    ]
[  c   f    ]
[  c   f  h  ]
[ b c   f  h  ]
[a b c   f  h  j]
[a b c   f g h i j]
[a b c  e f g h i j]
[a b c d e f g h i j]
因为上例中只有一个协程（主协程）在等待通知，所以其中的cond.Broadcast()调用也可以换为cond.Signal()。 如上例中的注释所示，cond.Broadcast()和cond.Signal()不必在cond.L的锁处于加锁状态时调用。

为了防止数据竞争，对自定义条件的修改必须在cond.L的锁处于加锁状态时才能执行。 另外，checkCondition函数和cond.Wait方法也必须在cond.L的锁处于加锁状态时才可被调用。

事实上，对于上面这个特定的例子，cond.L字段的也可以为一个*sync.RWMutex值。 对自定义条件的十个部分的修改可以在RWMutex值的读锁处于加锁状态时执行。这十个修改可以并发进行，因为它们是互不干扰的。 如下面的代码所示：

...
cond := sync.NewCond(&sync.RWMutex{})
cond.L.Lock()

	for i := 0; i < N; i++ {
		d := time.Second * time.Duration(rand.Intn(10)) / 10
		go func(i int) {
			time.Sleep(d)
			cond.L.(*sync.RWMutex).RLock()
			values[i] = string('a' + i)
			cond.L.(*sync.RWMutex).RUnlock()
			cond.Signal()
		}(i)
	}
...
在上面的代码中，此sync.RWMutex值的用法有些不符常规。 它的读锁被一些修改数组元素的协程所加锁并持有，而它的写锁被主协程加锁持有用来读取并检查各个数组元素的值。

Cond值所表示的自定义条件可以是一个虚无。对于这种情况，此Cond值纯粹被用来实现通知。 比如，下面这个程序将打印出abc或者bac。
package main

import (
"fmt"
"sync"
)

func main() {
wg := sync.WaitGroup{}
wg.Add(1)
cond := sync.NewCond(&sync.Mutex{})
cond.L.Lock()
go func() {
cond.L.Lock()
go func() {
cond.L.Lock()
cond.Broadcast()
cond.L.Unlock()
}()
cond.Wait()
fmt.Print("a")
cond.L.Unlock()
wg.Done()
}()
cond.Wait()
fmt.Print("b")
cond.L.Unlock()
wg.Wait()
fmt.Println("c")
}
如果需要，多个sync.Cond值可以共享一个sync.Locker值。但是这种情形在实践中并不多见。

## 原子操作
