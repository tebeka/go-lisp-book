package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Printf(">>> ")
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		code, err := parse(s.Text())
		if err != nil {
			fmt.Printf("syntax error: %s\n", err)
		}

		val, err := code.Eval()
		if err != nil {
			fmt.Printf("eval error: %s\n", err)
		}
		fmt.Println(val)
		fmt.Printf(">>> ")
	}

	switch s.Err() {
	case io.EOF, nil:
		// OK
	default:
		fmt.Printf("input error: %s\n", s.Err())
	}
}
