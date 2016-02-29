package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println(string([]byte(hello())))
	fmt.Println(http.Handler(nil))
}

func hello() string {
	return "hello"
}
