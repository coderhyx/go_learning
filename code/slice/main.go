package main

func main() {
	//arr := [5]int{1, 2, 3, 4, 5}
	//s1 := arr[1:3:3]
	//fmt.Println(cap(s1), len(s1))
	//s2 := s1[1:2:2]
	//fmt.Println(cap(s2), len(s2))
	//fmt.Println(&s2[0])
	//address(s2)
	var s3 []int

	s3[0] = 1
}

func address(s []int) {
	println(&s[0])
}
