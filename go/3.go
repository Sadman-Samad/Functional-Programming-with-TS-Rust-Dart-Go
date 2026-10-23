package main

import "fmt"

type Op func(uint32, uint32) uint32

func main() {
	v := Op(func(a, b uint32) uint32 { return a + b })

	can := func(a, b uint32, c Op) uint32 { return c(a, b) }

	fmt.Println(can(2, 3, v))
	fmt.Println(can(4, 5, func(a, b uint32) uint32 { return a * b }))
}
