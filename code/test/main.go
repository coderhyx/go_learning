package main

func createSlice() []int {
	slice := make([]int, 0, 10)
	// ...
	return slice
}

func createMap() map[int]string {
	map1 := make(map[int]string, 1)
	return map1
}

func createInterface() interface{} {
	var i interface{}
	return i
}

func main() {
	//createSlice()
	createMap()
	//createInterface()
}
