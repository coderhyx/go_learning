package main

func main() {
	arr := []int{1, 2, 3, 4, 5}
	for i, i2 := range arr {
		//arr[i] = i2 + 1
		i2 = i2 + 1
	}
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
