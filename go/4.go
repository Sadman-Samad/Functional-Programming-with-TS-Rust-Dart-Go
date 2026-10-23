package main

import "fmt"

func main() {
	plus := func(value uint8, fn func(uint8) uint8) uint8 {
		return fn(value)
	}

	double := func(n uint8) uint8 { return n * 2 }

	fmt.Println(plus(4, double))
}
