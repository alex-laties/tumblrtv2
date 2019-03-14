package main

import "C"

//export HelloWorld
func HelloWorld() int64 {
	return 1
}

//export DoSomething
func DoSomething() int {
	return 1 + 1
}

func main() {}
