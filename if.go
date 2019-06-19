package main

import (
	"fmt"
)

func div(a, b int) int {
	return a / b
}


func main() {
	if true {
		fmt.Printf("OK")
	} else {
		fmt.Println(div(1, 0))
	}
}
