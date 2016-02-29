package main

import (
	"fmt"
)

func main() {
	fmt.Println(string([]byte(hello())))
}

func hello() string {
	return "hello"
}
