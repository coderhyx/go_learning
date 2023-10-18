package main

import "fmt"

func main() {
	arr := [5]int{1, 2, 3, 4, 5}
	s1 := arr[1:3:3]
	fmt.Println(cap(s1), len(s1))
	s2 := s1[1:2:2]
	fmt.Println(cap(s2), len(s2))

}
